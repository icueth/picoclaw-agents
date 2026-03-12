package agent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"picoclaw/agent/pkg/config"
	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/providers"
	"picoclaw/agent/pkg/skills"
)

type ContextBuilder struct {
	workspace    string // Shared workspace directory for all agents
	skillsLoader *skills.SkillsLoader
	memory       *MemoryStore
	taskArchive  *TaskArchive
	modelName    string // Current model name for identity purposes

	// New: MemoryManager for SQLite + RAG + Embedding (Layer B)
	// This provides advanced memory capabilities alongside legacy systems
	memoryManager *MemoryManager

	agentID string // Current agent ID for specialized prompt loading
	department string // Department for loading department-specific persona files
	embeddedPrompt string // Embedded persona prompt from builtin configuration
	
	// Per-agent workspace for isolated sessions and memory
	agentWorkspace *AgentWorkspace
	
	// Department shared memory for cross-agent knowledge sharing
	deptMemory *DepartmentMemory

	// Phase 5: Lightweight A2A mode for reduced token usage
	// When enabled, builds a minimal system prompt optimized for A2A collaboration
	a2aMode          bool   // Enable A2A lightweight mode
	a2aModeCached    bool   // Whether A2A mode is cached
	a2aCachedPrompt  string // Cached A2A system prompt

	// Cache for system prompt to avoid rebuilding on every call.
	// This fixes issue #607: repeated reprocessing of the entire context.
	// The cache auto-invalidates when workspace source files change (mtime check).
	systemPromptMutex  sync.RWMutex
	cachedSystemPrompt string
	cachedAt           time.Time // max observed mtime across tracked paths at cache build time

	// existedAtCache tracks which source file paths existed the last time the
	// cache was built. This lets sourceFilesChanged detect files that are newly
	// created (didn't exist at cache time, now exist) or deleted (existed at
	// cache time, now gone) — both of which should trigger a cache rebuild.
	existedAtCache map[string]bool

	// skillFilesAtCache snapshots the skill tree file set and mtimes at cache
	// build time. This catches nested file creations/deletions/mtime changes
	// that may not update the top-level skill root directory mtime.
	skillFilesAtCache map[string]time.Time
}

func getGlobalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".picoclaw")
}

func NewContextBuilder(workspace string) *ContextBuilder {
	// builtin skills: skills directory in current project
	// Use the skills/ directory under the current working directory
	builtinSkillsDir := strings.TrimSpace(os.Getenv("PICOCLAW_BUILTIN_SKILLS"))
	if builtinSkillsDir == "" {
		wd, _ := os.Getwd()
		builtinSkillsDir = filepath.Join(wd, "skills")
	}
	globalSkillsDir := filepath.Join(getGlobalConfigDir(), "skills")

	return &ContextBuilder{
		workspace:     workspace,
		skillsLoader:  skills.NewSkillsLoader(workspace, globalSkillsDir, builtinSkillsDir),
		memory:        nil, // Will be set when agentWorkspace is set
		taskArchive:   nil, // Will be set when agentWorkspace is set
		modelName:     "", // Will be set by SetModelName when agent is initialized
		memoryManager: nil, // Will be initialized by SetMemoryManager
		agentWorkspace: nil, // Will be set by SetAgentWorkspace
	}
}

// SetAgentWorkspace initializes per-agent storage paths.
// This must be called before using memory or task archive.
func (cb *ContextBuilder) SetAgentWorkspace(aw *AgentWorkspace) {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()
	
	cb.agentWorkspace = aw
	// Initialize per-agent memory stores
	cb.memory = NewMemoryStoreWithOptions(aw.MemoryDir, false) // MEMORY.md directly in memory dir
	cb.taskArchive = NewTaskArchive(aw.BasePath)
	
	logger.InfoCF("context", "Agent workspace set",
		map[string]any{
			"agent_id": aw.AgentID,
			"memory":   aw.MemoryDir,
			"sessions": aw.SessionDir,
		})
}

// SetAgentDir is DEPRECATED.
// All agents now use shared workspace for IDENTITY.md, SOUL.md, MEMORY.md.
// This function is kept for backward compatibility but does nothing.
func (cb *ContextBuilder) SetAgentDir(agentDir string) {
	// No-op: agent-specific directories are no longer used
}

// SetMemoryManager sets the MemoryManager for advanced memory capabilities.
// This should be called after the database and config are initialized.
func (cb *ContextBuilder) SetMemoryManager(mm *MemoryManager) {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()

	cb.memoryManager = mm
	logger.InfoCF("context", "MemoryManager set in ContextBuilder",
		map[string]any{
			"rag_enabled": mm != nil && mm.IsRAGEnabled(),
		})
}

// SetAgentID sets the agent ID for specialized prompt loading.
// Also auto-initializes agent workspace if not already set.
func (cb *ContextBuilder) SetAgentID(id string) {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()
	cb.agentID = id
	
	// Auto-initialize agent workspace if not set
	if cb.agentWorkspace == nil && id != "" {
		aw := NewAgentWorkspace(cb.workspace, id)
		cb.agentWorkspace = aw
		// Initialize per-agent memory stores
		cb.memory = NewMemoryStoreWithOptions(aw.MemoryDir, false)
		cb.taskArchive = NewTaskArchive(aw.BasePath)
	}
}

// SetDepartment sets the department for loading department-specific persona files.
// Also initializes department shared memory for cross-agent knowledge sharing.
// Files are loaded from workspace/agents/{department}/IDENTITY.md, SOUL.md, etc.
func (cb *ContextBuilder) SetDepartment(dept string) {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()
	cb.department = dept
	
	// Initialize department shared memory
	if dept != "" {
		cb.deptMemory = NewDepartmentMemory(cb.workspace, dept)
		logger.InfoCF("context", "Department memory initialized",
			map[string]any{
				"department": dept,
			})
	}
}

// GetAgentWorkspace returns the per-agent workspace (may be nil if not initialized)
func (cb *ContextBuilder) GetAgentWorkspace() *AgentWorkspace {
	cb.systemPromptMutex.RLock()
	defer cb.systemPromptMutex.RUnlock()
	return cb.agentWorkspace
}

// GetDepartmentMemory returns the department shared memory (may be nil if not initialized)
func (cb *ContextBuilder) GetDepartmentMemory() *DepartmentMemory {
	cb.systemPromptMutex.RLock()
	defer cb.systemPromptMutex.RUnlock()
	return cb.deptMemory
}

// SetA2AMode enables/disables lightweight A2A mode for reduced token usage.
// When enabled, BuildSystemPrompt() returns a minimal prompt optimized for A2A collaboration.
func (cb *ContextBuilder) SetA2AMode(enabled bool) {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()

	if cb.a2aMode != enabled {
		cb.a2aMode = enabled
		// Invalidate A2A cache when mode changes
		cb.a2aModeCached = false
		cb.a2aCachedPrompt = ""
		// Also invalidate main cache since mode affects output
		cb.cachedSystemPrompt = ""
		cb.cachedAt = time.Time{}

		logger.InfoCF("context", "A2A mode changed",
			map[string]any{
				"enabled":  enabled,
				"agent_id": cb.agentID,
			})
	}
}

// IsA2AMode returns whether A2A lightweight mode is enabled
func (cb *ContextBuilder) IsA2AMode() bool {
	cb.systemPromptMutex.RLock()
	defer cb.systemPromptMutex.RUnlock()
	return cb.a2aMode
}

// SetEmbeddedPrompt sets the embedded persona prompt markdown.
func (cb *ContextBuilder) SetEmbeddedPrompt(prompt string) {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()
	cb.embeddedPrompt = prompt
}

func (cb *ContextBuilder) getIdentity() string {
	workspacePath, _ := filepath.Abs(filepath.Join(cb.workspace))

	// Include model name and capabilities in identity
	modelInfo := ""
	capabilityInfo := ""
	if cb.modelName != "" {
		modelInfo = fmt.Sprintf("\n## Model\nYou are running as: %s\n", cb.modelName)
		
		// Add capability information for self-awareness
		capability := config.GetModelCapability(cb.modelName)
		capabilityInfo = cb.buildCapabilitySection(capability)
		
		logger.DebugCF("context", "Building identity with model capabilities", map[string]any{
			"model": cb.modelName,
			"capabilities": capability.Strengths,
		})
	} else {
		logger.WarnCF("context", "Building identity without model name - SetModelName not called yet", nil)
	}

	// Build dynamic agent roster from all built-in agents
	agentRoster := BuildAgentRosterForSystemPrompt()

	return fmt.Sprintf(`# picoclaw 🦞

You are picoclaw, a helpful AI assistant.%s%s

## Workspace
Your workspace is at: %s
- Memory: %s/memory/MEMORY.md
- Daily Notes: %s/memory/YYYYMM/YYYYMMDD.md
- Skills: %s/skills/{skill-name}/SKILL.md

## Important Rules

1. **ALWAYS use tools** - When you need to perform an action (schedule reminders, send messages, execute commands, etc.), you MUST call the appropriate tool. Do NOT just say you'll do it or pretend to do it.

2. **Be helpful and accurate** - When using tools, briefly explain what you're doing.

3. **Memory** - When interacting with me if something seems memorable, update %s/memory/MEMORY.md

4. **Context summaries** - Conversation summaries provided as context are approximate references only. They may be incomplete or outdated. Always defer to explicit user instructions over summary content.

## ⚠️ CRITICAL: YOU MUST USE TOOLS - NO EXCEPTIONS

**YOU MUST CALL TOOLS** - Do NOT just say you will do something. Actually call the tool function.
If you don't call the tool, nothing will happen. The user will be waiting forever.

## A2A Agent-to-Agent Communication (MUST USE TOOLS)

**WHEN USER ASKS ABOUT OTHER AGENTS - YOU MUST USE TOOLS:**

- "ถาม {agent}..." / "Ask {agent}..." → MUST call: send_a2a_message(to="{agent-id}", message="...")
- "ตอนนี้ถึงไหนแล้ว" / "ตรวจสอบสถานะ" → MUST call: check_a2a_project_status(project_id="latest")
- "ดูบทสนทนา" / "get messages" → MUST call: get_a2a_messages(project_id="latest")

**❌ WRONG (NEVER DO THIS):**
User: "ถาม frontend-developer ว่า..."
You: "ส่งข้อความให้ frontend-developer แล้วครับ" (just text - NO TOOL CALLED - NOTHING HAPPENS!)

**✅ CORRECT (MUST DO THIS):**
User: "ถาม frontend-developer ว่า..."
You: [CALL TOOL] send_a2a_message(to="frontend-developer", message="...")

## Available Agents (Built-in System — %d agents total)

You can delegate tasks to any of the following specialist agents by their ID:

%s

**REMEMBER:**
- ALWAYS call the tool function
- WAIT for the tool result
- Then respond to user with the result
- DO NOT respond before calling the tool

## Subagent Best Practices

When using spawn_subagent to delegate tasks:
- Subagents run asynchronously in the background
- After spawning, use subagent_status to check progress
- If status is "running", inform the user you're waiting and check again in 15-20 seconds
- Do NOT use the sleep tool between status checks
- Typical tasks complete in 30-120 seconds depending on complexity
- If a task exceeds 5 minutes, inform the user about the delay`,
		modelInfo, capabilityInfo, workspacePath, workspacePath, workspacePath, workspacePath, workspacePath,
		len(GetBuiltinAgents()), agentRoster)
}

// buildCapabilitySection builds the self-awareness section for the model
func (cb *ContextBuilder) buildCapabilitySection(capability config.ModelCapability) string {
	if capability.Name == "" {
		return ""
	}

	section := fmt.Sprintf(`
## Your Capabilities & Limitations

**Model:** %s
**Context Window:** %d tokens
**Max Task Duration:** %s

### Your Strengths
`, capability.Name, capability.ContextWindow, capability.MaxTaskDuration.String())

	for _, strength := range capability.Strengths {
		section += fmt.Sprintf("- %s\n", strength)
	}

	section += "\n### Your Weaknesses\n"
	for _, weakness := range capability.Weaknesses {
		section += fmt.Sprintf("- %s\n", weakness)
	}

	section += "\n### When You MUST Delegate to a Subagent\n"
	section += "You are REQUIRED to spawn a subagent when:\n\n"
	
	if !capability.IsGoodAtCoding {
		section += "1. **Coding Tasks** - Any programming, debugging, code review, or software development\n"
	}
	if !capability.IsGoodAtReasoning {
		section += "2. **Complex Reasoning** - Architecture design, system planning, or complex problem-solving\n"
	}
	if !capability.IsGoodAtAnalysis {
		section += "3. **Deep Analysis** - Research, data analysis, or investigation tasks\n"
	}
	
	section += `4. **Tasks Exceeding Your Duration** - If estimated time > your max task duration
5. **High Complexity Tasks** - Multi-step tasks requiring specialized knowledge
6. **When Uncertain** - If you are not confident about the best approach

### Available Subagent Roles
- **coder**: Specialized in code generation, debugging, and software development
- **planner**: Specialized in architecture, design, and strategic planning  
- **researcher**: Specialized in research, analysis, and investigation
- **reviewer**: Specialized in code review, quality assurance, and auditing
- **architect**: Specialized in system design and high-level architecture

### Delegation Decision Rule
**WHEN IN DOUBT, ALWAYS DELEGATE.** It is better to use a specialist than to produce suboptimal results.

To delegate: Use the spawn_subagent tool with the appropriate role.
`

	return section
}

// SetModelName sets the model name for identity purposes.
// This should be called when the agent's model changes.
func (cb *ContextBuilder) SetModelName(modelName string) {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()
	
	if cb.modelName != modelName {
		cb.modelName = modelName
		// Invalidate cache since identity has changed
		cb.cachedSystemPrompt = ""
		cb.cachedAt = time.Time{}
		logger.InfoCF("context", "Model name updated, system prompt cache invalidated", map[string]any{
			"model": modelName,
		})
	}
}

func (cb *ContextBuilder) BuildSystemPrompt() string {
	// Phase 5: Use lightweight A2A mode if enabled
	if cb.a2aMode {
		return cb.buildA2ASystemPrompt()
	}

	parts := []string{}

	// Core identity section
	parts = append(parts, cb.getIdentity())

	// Bootstrap files
	bootstrapContent := cb.LoadBootstrapFiles()
	if bootstrapContent != "" {
		parts = append(parts, bootstrapContent)
	}

	// Skills - show summary, AI can read full content with read_file tool
	skillsSummary := cb.skillsLoader.BuildSkillsSummary()
	if skillsSummary != "" {
		parts = append(parts, fmt.Sprintf(`# Skills

The following skills extend your capabilities. To use a skill, read its SKILL.md file using the read_file tool.

%s`, skillsSummary))
	}

	// Join with "---" separator
	// Note: memory is loaded dynamically in BuildMessages (capability-aware), not cached here.
	return strings.Join(parts, "\n\n---\n\n")
}

// buildA2ASystemPrompt creates a lightweight system prompt optimized for A2A collaboration
// This reduces token usage by ~30-50% compared to the full system prompt
func (cb *ContextBuilder) buildA2ASystemPrompt() string {
	// Check cache first
	if cb.a2aModeCached && cb.a2aCachedPrompt != "" {
		return cb.a2aCachedPrompt
	}

	parts := []string{}

	// Minimal identity - just name and core role
	parts = append(parts, cb.getA2AIdentity())

	// Embedded prompt only (skip file-based persona files in A2A mode)
	if cb.embeddedPrompt != "" {
		parts = append(parts, fmt.Sprintf("## Your Role\n\n%s", cb.embeddedPrompt))
	}

	// Minimal skills info - just names, no descriptions
	if cb.skillsLoader != nil {
		skills := cb.skillsLoader.ListSkills()
		if len(skills) > 0 {
			var skillNames []string
			for _, s := range skills {
				skillNames = append(skillNames, s.Name)
			}
			parts = append(parts, fmt.Sprintf("## Available Skills\n\n%s", strings.Join(skillNames, ", ")))
		}
	}

	// A2A-specific instructions
	parts = append(parts, `## A2A Collaboration

You are participating in Agent-to-Agent (A2A) collaboration.
- Focus on your assigned task
- Communicate clearly and concisely
- Use tools proactively to complete tasks`)

	prompt := strings.Join(parts, "\n\n---\n\n")

	// Cache the A2A prompt
	cb.a2aCachedPrompt = prompt
	cb.a2aModeCached = true

	return prompt
}

// getA2AIdentity returns a minimal identity for A2A mode
func (cb *ContextBuilder) getA2AIdentity() string {
	workspacePath, _ := filepath.Abs(filepath.Join(cb.workspace))

	return fmt.Sprintf(`# %s 🦞

You are %s, an AI assistant collaborating with other agents.
Workspace: %s`,
		cb.agentID,
		cb.agentID,
		workspacePath,
	)
}

// BuildSystemPromptWithCache returns the cached system prompt if available
// and source files haven't changed, otherwise builds and caches it.
// Source file changes are detected via mtime checks (cheap stat calls).
func (cb *ContextBuilder) BuildSystemPromptWithCache() string {
	// Try read lock first — fast path when cache is valid
	cb.systemPromptMutex.RLock()
	if cb.cachedSystemPrompt != "" && !cb.sourceFilesChangedLocked() {
		result := cb.cachedSystemPrompt
		cb.systemPromptMutex.RUnlock()
		return result
	}
	cb.systemPromptMutex.RUnlock()

	// Acquire write lock for building
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()

	// Double-check: another goroutine may have rebuilt while we waited
	if cb.cachedSystemPrompt != "" && !cb.sourceFilesChangedLocked() {
		return cb.cachedSystemPrompt
	}

	// Snapshot the baseline (existence + max mtime) BEFORE building the prompt.
	// This way cachedAt reflects the pre-build state: if a file is modified
	// during BuildSystemPrompt, its new mtime will be > baseline.maxMtime,
	// so the next sourceFilesChangedLocked check will correctly trigger a
	// rebuild. The alternative (baseline after build) risks caching stale
	// content with a too-new baseline, making the staleness invisible.
	baseline := cb.buildCacheBaseline()
	prompt := cb.BuildSystemPrompt()
	cb.cachedSystemPrompt = prompt
	cb.cachedAt = baseline.maxMtime
	cb.existedAtCache = baseline.existed
	cb.skillFilesAtCache = baseline.skillFiles

	logger.DebugCF("agent", "System prompt cached",
		map[string]any{
			"length": len(prompt),
		})

	return prompt
}

// InvalidateCache clears the cached system prompt.
// Normally not needed because the cache auto-invalidates via mtime checks,
// but this is useful for tests or explicit reload commands.
func (cb *ContextBuilder) InvalidateCache() {
	cb.systemPromptMutex.Lock()
	defer cb.systemPromptMutex.Unlock()

	cb.cachedSystemPrompt = ""
	cb.cachedAt = time.Time{}
	cb.existedAtCache = nil
	cb.skillFilesAtCache = nil

	logger.DebugCF("agent", "System prompt cache invalidated", nil)
}

// sourcePaths returns non-skill workspace source files tracked for cache
// invalidation (bootstrap files). Memory is loaded dynamically per request
// and is NOT part of the cached system prompt, so MEMORY.md is excluded.
func (cb *ContextBuilder) sourcePaths() []string {
	// Priority: department-specific > workspace-level shared files
	// Agent-specific prompts come from builtin agents (embeddedPrompt)
	
	if cb.department != "" {
		deptDir := filepath.Join(cb.workspace, "agents", cb.department)
		// Check if department directory exists
		if info, err := os.Stat(deptDir); err == nil && info.IsDir() {
			return []string{
				filepath.Join(deptDir, "AGENTS.md"),
				filepath.Join(deptDir, "SOUL.md"),
				filepath.Join(deptDir, "USER.md"),
				filepath.Join(deptDir, "IDENTITY.md"),
				// Also include workspace-level as fallback
				filepath.Join(cb.workspace, "AGENTS.md"),
				filepath.Join(cb.workspace, "SOUL.md"),
				filepath.Join(cb.workspace, "USER.md"),
				filepath.Join(cb.workspace, "IDENTITY.md"),
			}
		}
	}
	
	// Fallback to workspace-level shared files
	return []string{
		filepath.Join(cb.workspace, "AGENTS.md"),
		filepath.Join(cb.workspace, "SOUL.md"),
		filepath.Join(cb.workspace, "USER.md"),
		filepath.Join(cb.workspace, "IDENTITY.md"),
	}
}

// skillRoots returns all skill root directories that can affect
// BuildSkillsSummary output (workspace/global/builtin).
func (cb *ContextBuilder) skillRoots() []string {
	if cb.skillsLoader == nil {
		return []string{filepath.Join(cb.workspace, "skills")}
	}

	roots := cb.skillsLoader.SkillRoots()
	if len(roots) == 0 {
		return []string{filepath.Join(cb.workspace, "skills")}
	}
	return roots
}

// cacheBaseline holds the file existence snapshot and the latest observed
// mtime across all tracked paths. Used as the cache reference point.
type cacheBaseline struct {
	existed    map[string]bool
	skillFiles map[string]time.Time
	maxMtime   time.Time
}

// buildCacheBaseline records which tracked paths currently exist and computes
// the latest mtime across all tracked files + skills directory contents.
// Called under write lock when the cache is built.
func (cb *ContextBuilder) buildCacheBaseline() cacheBaseline {
	skillRoots := cb.skillRoots()

	// All paths whose existence we track: source files + all skill roots.
	allPaths := append(cb.sourcePaths(), skillRoots...)

	existed := make(map[string]bool, len(allPaths))
	skillFiles := make(map[string]time.Time)
	var maxMtime time.Time

	for _, p := range allPaths {
		info, err := os.Stat(p)
		existed[p] = err == nil
		if err == nil && info.ModTime().After(maxMtime) {
			maxMtime = info.ModTime()
		}
	}

	// Walk all skill roots recursively to snapshot skill files and mtimes.
	// Use os.Stat (not d.Info) for consistency with sourceFilesChanged checks.
	for _, root := range skillRoots {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr == nil && !d.IsDir() {
				if info, err := os.Stat(path); err == nil {
					skillFiles[path] = info.ModTime()
					if info.ModTime().After(maxMtime) {
						maxMtime = info.ModTime()
					}
				}
			}
			return nil
		})
	}

	// If no tracked files exist yet (empty workspace), maxMtime is zero.
	// Use a very old non-zero time so that:
	// 1. cachedAt.IsZero() won't trigger perpetual rebuilds.
	// 2. Any real file created afterwards has mtime > cachedAt, so it
	//    will be detected by fileChangedSince (unlike time.Now() which
	//    could race with a file whose mtime <= Now).
	if maxMtime.IsZero() {
		maxMtime = time.Unix(1, 0)
	}

	return cacheBaseline{existed: existed, skillFiles: skillFiles, maxMtime: maxMtime}
}

// sourceFilesChangedLocked checks whether any workspace source file has been
// modified, created, or deleted since the cache was last built.
//
// IMPORTANT: The caller MUST hold at least a read lock on systemPromptMutex.
// Go's sync.RWMutex is not reentrant, so this function must NOT acquire the
// lock itself (it would deadlock when called from BuildSystemPromptWithCache
// which already holds RLock or Lock).
func (cb *ContextBuilder) sourceFilesChangedLocked() bool {
	if cb.cachedAt.IsZero() {
		return true
	}

	// Check tracked source files (bootstrap + memory).
	if slices.ContainsFunc(cb.sourcePaths(), cb.fileChangedSince) {
		return true
	}

	// --- Skill roots (workspace/global/builtin) ---
	//
	// For each root:
	// 1. Creation/deletion and root directory mtime changes are tracked by fileChangedSince.
	// 2. Nested file create/delete/mtime changes are tracked by the skill file snapshot.
	for _, root := range cb.skillRoots() {
		if cb.fileChangedSince(root) {
			return true
		}
	}
	if skillFilesChangedSince(cb.skillRoots(), cb.skillFilesAtCache) {
		return true
	}

	return false
}

// fileChangedSince returns true if a tracked source file has been modified,
// newly created, or deleted since the cache was built.
//
// Four cases:
//   - existed at cache time, exists now -> check mtime
//   - existed at cache time, gone now   -> changed (deleted)
//   - absent at cache time,  exists now -> changed (created)
//   - absent at cache time,  gone now   -> no change
func (cb *ContextBuilder) fileChangedSince(path string) bool {
	// Defensive: if existedAtCache was never initialized, treat as changed
	// so the cache rebuilds rather than silently serving stale data.
	if cb.existedAtCache == nil {
		return true
	}

	existedBefore := cb.existedAtCache[path]
	info, err := os.Stat(path)
	existsNow := err == nil

	if existedBefore != existsNow {
		return true // file was created or deleted
	}
	if !existsNow {
		return false // didn't exist before, doesn't exist now
	}
	return info.ModTime().After(cb.cachedAt)
}

// errWalkStop is a sentinel error used to stop filepath.WalkDir early.
// Using a dedicated error (instead of fs.SkipAll) makes the early-exit
// intent explicit and avoids the nilerr linter warning that would fire
// if the callback returned nil when its err parameter is non-nil.
var errWalkStop = errors.New("walk stop")

// skillFilesChangedSince compares the current recursive skill file tree
// against the cache-time snapshot. Any create/delete/mtime drift invalidates
// the cache.
func skillFilesChangedSince(skillRoots []string, filesAtCache map[string]time.Time) bool {
	// Defensive: if the snapshot was never initialized, force rebuild.
	if filesAtCache == nil {
		return true
	}

	// Check cached files still exist and keep the same mtime.
	for path, cachedMtime := range filesAtCache {
		info, err := os.Stat(path)
		if err != nil {
			// A previously tracked file disappeared (or became inaccessible):
			// either way, cached skill summary may now be stale.
			return true
		}
		if !info.ModTime().Equal(cachedMtime) {
			return true
		}
	}

	// Check no new files appeared under any skill root.
	changed := false
	for _, root := range skillRoots {
		if strings.TrimSpace(root) == "" {
			continue
		}

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				// Treat unexpected walk errors as changed to avoid stale cache.
				if !os.IsNotExist(walkErr) {
					changed = true
					return errWalkStop
				}
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if _, ok := filesAtCache[path]; !ok {
				changed = true
				return errWalkStop
			}
			return nil
		})

		if changed {
			return true
		}
		if err != nil && !errors.Is(err, errWalkStop) && !os.IsNotExist(err) {
			logger.DebugCF("agent", "skills walk error", map[string]any{"error": err.Error()})
			return true
		}
	}

	return false
}

func (cb *ContextBuilder) LoadBootstrapFiles() string {
	// Use sourcePaths to get the correct paths from shared workspace
	paths := cb.sourcePaths()
	
	var sb strings.Builder
	
	// Add embedded prompt first if available
	embedded := cb.embeddedPrompt

	
	if embedded != "" {
		fmt.Fprintf(&sb, "## PERSONA\n\n%s\n\n", embedded)
	}

	for _, filePath := range paths {
		filename := filepath.Base(filePath)
		if data, err := os.ReadFile(filePath); err == nil {
			fmt.Fprintf(&sb, "## %s\n\n%s\n\n", filename, data)
		}
	}

	return sb.String()
}

// buildDynamicContext returns a short dynamic context string with per-request info.
// This changes every request (time, session) so it is NOT part of the cached prompt.
// LLM-side KV cache reuse is achieved by each provider adapter's native mechanism:
//   - Anthropic: per-block cache_control (ephemeral) on the static SystemParts block
//   - OpenAI / Codex: prompt_cache_key for prefix-based caching
//
// See: https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching
// See: https://platform.openai.com/docs/guides/prompt-caching
func (cb *ContextBuilder) buildDynamicContext(channel, chatID string) string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	rt := fmt.Sprintf("%s %s, Go %s", runtime.GOOS, runtime.GOARCH, runtime.Version())

	var sb strings.Builder
	fmt.Fprintf(&sb, "## Current Time\n%s\n\n## Runtime\n%s", now, rt)

	// Include model name in dynamic context so it's always up-to-date
	// This is more reliable than static identity which may be cached before SetModelName is called
	if cb.modelName != "" {
		fmt.Fprintf(&sb, "\n\n## Model\nYou are running as: %s", cb.modelName)
	}

	if channel != "" && chatID != "" {
		fmt.Fprintf(&sb, "\n\n## Current Session\nChannel: %s\nChat ID: %s", channel, chatID)
	}

	return sb.String()
}

func (cb *ContextBuilder) BuildMessages(
	history []providers.Message,
	summary string,
	currentMessage string,
	media []string,
	channel, chatID string,
	dynamicPrompt string,
	capability string,
) []providers.Message {
	messages := []providers.Message{}

	// The static part (identity, bootstrap, skills, memory) is cached locally to
	// avoid repeated file I/O and string building on every call (fixes issue #607).
	// Dynamic parts (time, session, summary) are appended per request.
	// Everything is sent as a single system message for provider compatibility:
	// - Anthropic adapter extracts messages[0] (Role=="system") and maps its content
	//   to the top-level "system" parameter in the Messages API request. A single
	//   contiguous system block makes this extraction straightforward.
	// - Codex maps only the first system message to its instructions field.
	// - OpenAI-compat passes messages through as-is.
	staticPrompt := cb.BuildSystemPromptWithCache()

	// Build short dynamic context (time, runtime, session) — changes per request
	dynamicCtx := cb.buildDynamicContext(channel, chatID)

	// Compose a single system message: static (cached) + dynamic + optional summary.
	// Keeping all system content in one message ensures every provider adapter can
	// extract it correctly (Anthropic adapter -> top-level system param,
	// Codex -> instructions field).
	//
	// SystemParts carries the same content as structured blocks so that
	// cache-aware adapters (Anthropic) can set per-block cache_control.
	// The static block is marked "ephemeral" — its prefix hash is stable
	// across requests, enabling LLM-side KV cache reuse.
	stringParts := []string{staticPrompt, dynamicCtx}

	// Load memory dynamically based on current capability (not cached).
	// Layer 1: Try MemoryManager (SQLite + RAG + Embedding) first
	// Layer 2: Try StructuredMemory (v2)
	// Layer 3: Fall back to legacy MemoryStore
	
	// Use per-agent memory path if available
	memoryBasePath := cb.workspace
	if cb.agentWorkspace != nil {
		memoryBasePath = cb.agentWorkspace.BasePath
	}
	sm := NewStructuredMemory(memoryBasePath)
	var memoryContext string
	var ragContext string

	// Layer 1: MemoryManager with RAG
	if cb.memoryManager != nil && cb.memoryManager.IsRAGEnabled() {
		logger.DebugCF("context", "Querying RAG memory", map[string]any{"query": currentMessage})
		ragResult, err := cb.memoryManager.QueryMemory(currentMessage, 5)
		if err == nil && len(ragResult.Documents) > 0 {
			var ragParts []string
			ragParts = append(ragParts, "### RAG Memory (Semantic Search)")
			for _, doc := range ragResult.Documents {
				ragParts = append(ragParts, fmt.Sprintf("- [%s] %s", doc.Metadata.Source, doc.Content))
			}
			ragContext = strings.Join(ragParts, "\n")
			logger.InfoCF("context", "RAG memory retrieved and injected into context",
				map[string]any{"documents": len(ragResult.Documents), "query": currentMessage})
		} else if err != nil {
			logger.WarnCF("context", "RAG query failed", map[string]any{"error": err.Error()})
		} else {
			logger.DebugCF("context", "RAG query returned no results", nil)
		}
	} else {
		if cb.memoryManager == nil {
			logger.DebugCF("context", "MemoryManager not available, skipping RAG", nil)
		} else {
			logger.DebugCF("context", "RAG not enabled, using legacy memory only", nil)
		}
	}

	// Layer 2 & 3: Legacy memory systems
	if sm.Count() > 0 {
		relevant := sm.RetrieveRelevant(capability, currentMessage, 15)
		memoryContext = FormatForPrompt(relevant)
	} else if cb.memory != nil {
		memoryContext = cb.memory.GetMemoryForCapability(capability)
	}

	var taskContext string
	if cb.taskArchive != nil {
		taskContext = cb.taskArchive.GetRelevantContext(capability, currentMessage, 3)
	}
	sessionContext := sm.GetRecentSessionsForPrompt(3)

	// Layer 4: Department shared memory (cross-agent knowledge)
	var deptContext string
	if cb.deptMemory != nil && currentMessage != "" {
		deptContext = cb.deptMemory.BuildContext(currentMessage)
		if deptContext != "" {
			logger.DebugCF("context", "Department memory injected",
				map[string]any{
					"department": cb.department,
					"chars":      len(deptContext),
				})
		}
	}

	// Combine RAG context with legacy memory
	if ragContext != "" {
		if memoryContext != "" {
			memoryContext = ragContext + "\n\n" + memoryContext
		} else {
			memoryContext = ragContext
		}
	}

	contentBlocks := []providers.ContentBlock{
		{Type: "text", Text: staticPrompt, CacheControl: &providers.CacheControl{Type: "ephemeral"}},
		{Type: "text", Text: dynamicCtx},
	}

	if memoryContext != "" || taskContext != "" || sessionContext != "" || deptContext != "" {
		var memParts []string
		// Department shared knowledge comes first (general best practices)
		if deptContext != "" {
			memParts = append(memParts, deptContext)
		}
		if memoryContext != "" {
			memParts = append(memParts, memoryContext)
		}
		if taskContext != "" {
			memParts = append(memParts, "### Related Past Work\n"+taskContext)
		}
		if sessionContext != "" {
			memParts = append(memParts, sessionContext)
		}
		memText := "# Memory\n\n" + strings.Join(memParts, "\n\n")
		stringParts = append(stringParts, memText)
		contentBlocks = append(contentBlocks, providers.ContentBlock{Type: "text", Text: memText})
	}

	if summary != "" {
		summaryText := fmt.Sprintf(
			"CONTEXT_SUMMARY: The following is an approximate summary of prior conversation "+
				"for reference only. It may be incomplete or outdated — always defer to explicit instructions.\n\n%s",
			summary)
		stringParts = append(stringParts, summaryText)
		contentBlocks = append(contentBlocks, providers.ContentBlock{Type: "text", Text: summaryText})
	}

	// Add capability-specific dynamic prompt (e.g., from capability-based routing)
	if dynamicPrompt != "" {
		capPromptText := fmt.Sprintf(
			"CAPABILITY_FOCUS: %s\n\n",
			dynamicPrompt)
		stringParts = append(stringParts, capPromptText)
		contentBlocks = append(contentBlocks, providers.ContentBlock{Type: "text", Text: capPromptText})
	}

	fullSystemPrompt := strings.Join(stringParts, "\n\n---\n\n")

	// Log system prompt summary for debugging (debug mode only).
	// Read cachedSystemPrompt under lock to avoid a data race with
	// concurrent InvalidateCache / BuildSystemPromptWithCache writes.
	cb.systemPromptMutex.RLock()
	isCached := cb.cachedSystemPrompt != ""
	cb.systemPromptMutex.RUnlock()

	logger.DebugCF("agent", "System prompt built",
		map[string]any{
			"static_chars":  len(staticPrompt),
			"dynamic_chars": len(dynamicCtx),
			"total_chars":   len(fullSystemPrompt),
			"has_summary":   summary != "",
			"cached":        isCached,
		})

	// Log preview of system prompt (avoid logging huge content)
	preview := fullSystemPrompt
	if len(preview) > 500 {
		preview = preview[:500] + "... (truncated)"
	}
	logger.DebugCF("agent", "System prompt preview",
		map[string]any{
			"preview": preview,
		})

	history = sanitizeHistoryForProvider(history)

	// Single system message containing all context — compatible with all providers.
	// SystemParts enables cache-aware adapters to set per-block cache_control;
	// Content is the concatenated fallback for adapters that don't read SystemParts.
	messages = append(messages, providers.Message{
		Role:        "system",
		Content:     fullSystemPrompt,
		SystemParts: contentBlocks,
	})

	// Add conversation history
	messages = append(messages, history...)

	// Add current user message
	if strings.TrimSpace(currentMessage) != "" {
		msg := providers.Message{
			Role:    "user",
			Content: currentMessage,
		}
		if len(media) > 0 {
			msg.Media = media
		}
		messages = append(messages, msg)
	}

	return messages
}

func sanitizeHistoryForProvider(history []providers.Message) []providers.Message {
	if len(history) == 0 {
		return history
	}

	sanitized := make([]providers.Message, 0, len(history))
	for _, msg := range history {
		switch msg.Role {
		case "system":
			// Drop system messages from history. BuildMessages always
			// constructs its own single system message (static + dynamic +
			// summary); extra system messages would break providers that
			// only accept one (Anthropic, Codex).
			logger.DebugCF("agent", "Dropping system message from history", map[string]any{})
			continue

		case "tool":
			if len(sanitized) == 0 {
				logger.DebugCF("agent", "Dropping orphaned leading tool message", map[string]any{})
				continue
			}
			// Walk backwards to find the nearest assistant message,
			// skipping over any preceding tool messages (multi-tool-call case).
			foundAssistant := false
			for i := len(sanitized) - 1; i >= 0; i-- {
				if sanitized[i].Role == "tool" {
					continue
				}
				if sanitized[i].Role == "assistant" && len(sanitized[i].ToolCalls) > 0 {
					foundAssistant = true
				}
				break
			}
			if !foundAssistant {
				logger.DebugCF("agent", "Dropping orphaned tool message", map[string]any{})
				continue
			}
			sanitized = append(sanitized, msg)

		case "assistant":
			if len(msg.ToolCalls) > 0 {
				if len(sanitized) == 0 {
					logger.DebugCF("agent", "Dropping assistant tool-call turn at history start", map[string]any{})
					continue
				}
				prev := sanitized[len(sanitized)-1]
				if prev.Role != "user" && prev.Role != "tool" {
					logger.DebugCF(
						"agent",
						"Dropping assistant tool-call turn with invalid predecessor",
						map[string]any{"prev_role": prev.Role},
					)
					continue
				}
			}
			sanitized = append(sanitized, msg)

		default:
			sanitized = append(sanitized, msg)
		}
	}

	return sanitized
}

func (cb *ContextBuilder) AddToolResult(
	messages []providers.Message,
	toolCallID, toolName, result string,
) []providers.Message {
	messages = append(messages, providers.Message{
		Role:       "tool",
		Content:    result,
		ToolCallID: toolCallID,
	})
	return messages
}

func (cb *ContextBuilder) AddAssistantMessage(
	messages []providers.Message,
	content string,
	toolCalls []map[string]any,
) []providers.Message {
	msg := providers.Message{
		Role:    "assistant",
		Content: content,
	}
	// Always add assistant message, whether or not it has tool calls
	messages = append(messages, msg)
	return messages
}

// GetSkillsInfo returns information about loaded skills.
func (cb *ContextBuilder) GetSkillsInfo() map[string]any {
	allSkills := cb.skillsLoader.ListSkills()
	skillNames := make([]string, 0, len(allSkills))
	for _, s := range allSkills {
		skillNames = append(skillNames, s.Name)
	}
	return map[string]any{
		"total":     len(allSkills),
		"available": len(allSkills),
		"names":     skillNames,
	}
}
