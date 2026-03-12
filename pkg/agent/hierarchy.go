package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"picoclaw/agent/pkg/logger"
	"picoclaw/agent/pkg/providers"
)

// AgentRole defines the role of an agent in the hierarchy
type AgentRole string

const (
	// RoleManager manages and coordinates other agents
	RoleManager AgentRole = "manager"
	// RolePlanner plans and breaks down complex tasks
	RolePlanner AgentRole = "planner"
	// RoleExecutor executes tasks assigned by planner
	RoleExecutor AgentRole = "executor"
	// RoleSpecialist has specialized capabilities
	RoleSpecialist AgentRole = "specialist"
	// RoleWorker performs general tasks
	RoleWorker AgentRole = "worker"
)

// HierarchyNode represents a node in the agent hierarchy
type HierarchyNode struct {
	AgentID    string
	ParentID   string
	Children   []string
	Role       AgentRole
	Capabilities []string
	mu         sync.RWMutex
}

// HierarchyManager manages the agent hierarchy
type HierarchyManager struct {
	mu       sync.RWMutex
	nodes    map[string]*HierarchyNode
	registry *AgentRegistry
	loop    *AgentLoop
}

// NewHierarchyManager creates a new HierarchyManager
func NewHierarchyManager(registry *AgentRegistry) *HierarchyManager {
	return &HierarchyManager{
		nodes:    make(map[string]*HierarchyNode),
		registry: registry,
	}
}

// SetAgentLoop sets the agent loop reference for executing tasks
func (hm *HierarchyManager) SetAgentLoop(loop *AgentLoop) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.loop = loop
}

// RegisterNode registers an agent in the hierarchy
func (hm *HierarchyManager) RegisterNode(agentID, parentID string, role AgentRole, capabilities []string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	node := &HierarchyNode{
		AgentID:      agentID,
		ParentID:     parentID,
		Role:         role,
		Capabilities: capabilities,
		Children:     []string{},
	}

	// Add to parent's children
	if parentID != "" {
		if parent, ok := hm.nodes[parentID]; ok {
			parent.Children = append(parent.Children, agentID)
		}
	}

	hm.nodes[agentID] = node
	logger.InfoCF("agent", "Registered hierarchy node",
		map[string]any{
			"agent_id":    agentID,
			"parent_id":  parentID,
			"role":       role,
			"capabilities": capabilities,
		})
}

// GetNode returns the hierarchy node for an agent
func (hm *HierarchyManager) GetNode(agentID string) (*HierarchyNode, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	node, ok := hm.nodes[agentID]
	return node, ok
}

// GetParent returns the parent agent ID
func (hm *HierarchyManager) GetParent(agentID string) (string, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	if node, ok := hm.nodes[agentID]; ok {
		return node.ParentID, node.ParentID != ""
	}
	return "", false
}

// GetChildren returns the child agent IDs
func (hm *HierarchyManager) GetChildren(agentID string) []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	if node, ok := hm.nodes[agentID]; ok {
		children := make([]string, len(node.Children))
		copy(children, node.Children)
		return children
	}
	return nil
}

// GetRole returns the role of an agent
func (hm *HierarchyManager) GetRole(agentID string) (AgentRole, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	if node, ok := hm.nodes[agentID]; ok {
		return node.Role, true
	}
	return "", false
}

// GetDescendants returns all agents below this agent in the hierarchy
func (hm *HierarchyManager) GetDescendants(agentID string) []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var result []string
	hm.collectDescendants(agentID, &result)
	return result
}

func (hm *HierarchyManager) collectDescendants(agentID string, result *[]string) {
	if node, ok := hm.nodes[agentID]; ok {
		for _, childID := range node.Children {
			*result = append(*result, childID)
			hm.collectDescendants(childID, result)
		}
	}
}

// GetAncestors returns all agents above this agent in the hierarchy
func (hm *HierarchyManager) GetAncestors(agentID string) []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var result []string
	currentID := agentID
	for {
		if node, ok := hm.nodes[currentID]; ok && node.ParentID != "" {
			result = append(result, node.ParentID)
			currentID = node.ParentID
		} else {
			break
		}
	}
	return result
}

// GetManagers returns all manager agents in the hierarchy
func (hm *HierarchyManager) GetManagers() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var managers []string
	for id, node := range hm.nodes {
		if node.Role == RoleManager {
			managers = append(managers, id)
		}
	}
	return managers
}

// ExecuteHierarchicalTask executes a task using the agent hierarchy
func (hm *HierarchyManager) ExecuteHierarchicalTask(
	ctx context.Context,
	rootAgentID string,
	task string,
) (string, error) {
	node, ok := hm.nodes[rootAgentID]
	if !ok {
		return "", fmt.Errorf("agent %s not found in hierarchy", rootAgentID)
	}

	switch node.Role {
	case RoleManager:
		return hm.executeManagerTask(ctx, rootAgentID, task)
	case RolePlanner:
		return hm.executePlannerTask(ctx, rootAgentID, task)
	case RoleExecutor, RoleWorker:
		return hm.executeWorkerTask(ctx, rootAgentID, task)
	default:
		return hm.executeWorkerTask(ctx, rootAgentID, task)
	}
}

func (hm *HierarchyManager) executeManagerTask(ctx context.Context, agentID string, task string) (string, error) {
	logger.InfoCF("agent", "Manager executing task decomposition",
		map[string]any{
			"agent_id": agentID,
			"task":     task[:min(50, len(task))],
		})

	// Manager breaks down task and delegates to children
	children := hm.GetChildren(agentID)
	if len(children) == 0 {
		// No children, execute directly
		return hm.executeWorkerTask(ctx, agentID, task)
	}

	// Analyze task to determine which children to involve
	taskLower := strings.ToLower(task)
	var subtasks []string

	// Simple task decomposition heuristics
	if strings.Contains(taskLower, "and") || strings.Contains(taskLower, ", and") {
		parts := strings.Split(task, " and ")
		subtasks = parts
	} else if strings.Contains(taskLower, " then ") {
		parts := strings.Split(task, " then ")
		subtasks = parts
	} else {
		subtasks = []string{task}
	}

	// Execute subtasks in parallel using children
	var wg sync.WaitGroup
	results := make([]string, len(subtasks))
	errors := make([]error, len(subtasks))

	for i, subtask := range subtasks {
		if i >= len(children) {
			break
		}
		wg.Add(1)
		go func(idx int, st string, childID string) {
			defer wg.Done()
			result, err := hm.ExecuteHierarchicalTask(ctx, childID, st)
			results[idx] = result
			errors[idx] = err
		}(i, subtask, children[i])
	}

	wg.Wait()

	// Collect results
	var successfulResults []string
	for i, result := range results {
		if errors[i] != nil {
			logger.WarnCF("agent", "Subtask failed",
				map[string]any{"error": errors[i].Error()})
			continue
		}
		if result != "" {
			successfulResults = append(successfulResults, result)
		}
	}

	if len(successfulResults) == 0 && len(errors) > 0 {
		return "", fmt.Errorf("all subtasks failed: %v", errors[0])
	}

	// Combine results
	return strings.Join(successfulResults, "\n\n---\n\n"), nil
}

func (hm *HierarchyManager) executePlannerTask(ctx context.Context, agentID string, task string) (string, error) {
	logger.InfoCF("agent", "Planner analyzing and breaking down task",
		map[string]any{"agent_id": agentID})

	// Execute actual work
	children := hm.GetChildren(agentID)
	if len(children) > 0 {
		// Delegate to executor children
		return hm.ExecuteHierarchicalTask(ctx, children[0], task)
	}

	// No children, execute directly
	return hm.executeWorkerTask(ctx, agentID, task)
}

func (hm *HierarchyManager) executeWorkerTask(ctx context.Context, agentID string, task string) (string, error) {
	logger.DebugCF("agent", "Worker executing task",
		map[string]any{"agent_id": agentID, "task_len": len(task)})

	// Use agent loop to execute task if available
	hm.mu.RLock()
	loop := hm.loop
	hm.mu.RUnlock()

	if loop == nil {
		// Fallback: use registry to get agent
		agent, ok := hm.registry.GetAgent(agentID)
		if !ok {
			return "", fmt.Errorf("agent %s not found", agentID)
		}

		// Execute via provider directly
		messages := []providers.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: task},
		}

		resp, err := agent.Provider.Chat(ctx, messages, nil, agent.Model, map[string]any{
			"max_tokens": agent.MaxTokens,
		})
		if err != nil {
			return "", err
		}
		return resp.Content, nil
	}

	// Execute via agent loop
	result, err := loop.ProcessDirect(ctx, task, "hierarchy-"+agentID, nil)
	return result, err
}

// BuildHierarchy builds the hierarchy from the initialized agent registry
func (hm *HierarchyManager) BuildHierarchy() {
	if hm.registry == nil {
		return
	}
	for _, id := range hm.registry.ListAgentIDs() {
		if agent, ok := hm.registry.GetAgent(id); ok {
			ac := agent.Config
			role := RoleWorker
			if ac.Role != "" {
				role = AgentRole(ac.Role)
			}
			
			var capabilities []string
			if ac.Capabilities != nil {
				capabilities = ac.Capabilities
			}

			hm.RegisterNode(ac.ID, ac.ParentID, role, capabilities)
		}
	}
}

// FindBestAgent finds the best agent for a task based on capabilities
func (hm *HierarchyManager) FindBestAgent(task string) (string, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	taskLower := strings.ToLower(task)
	bestScore := -1
	bestAgentID := ""

	// Score each agent based on capability match
	for agentID, node := range hm.nodes {
		score := 0

		// Check capability matches
		for _, cap := range node.Capabilities {
			if strings.Contains(taskLower, cap) {
				score += 10
			}
		}

		// Prefer specialized roles for complex tasks
		if strings.Contains(taskLower, "code") || strings.Contains(taskLower, "implement") {
			if node.Role == RoleSpecialist || node.Role == RoleExecutor {
				score += 5
			}
		}

		if strings.Contains(taskLower, "plan") || strings.Contains(taskLower, "analyze") {
			if node.Role == RolePlanner {
				score += 5
			}
		}

		if score > bestScore {
			bestScore = score
			bestAgentID = agentID
		}
	}

	return bestAgentID, bestAgentID != ""
}

// GetHierarchyTree returns a tree representation of the hierarchy
func (hm *HierarchyManager) GetHierarchyTree() string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var sb strings.Builder

	// Find root nodes (nodes without parents)
	var roots []string
	for id := range hm.nodes {
		isRoot := true
		for _, node := range hm.nodes {
			for _, child := range node.Children {
				if child == id {
					isRoot = false
					break
				}
			}
		}
		if isRoot {
			roots = append(roots, id)
		}
	}

	// Build tree
	for _, root := range roots {
		hm.buildTree(&sb, root, 0)
	}

	return sb.String()
}

func (hm *HierarchyManager) buildTree(sb *strings.Builder, agentID string, level int) {
	node, ok := hm.nodes[agentID]
	if !ok {
		return
	}

	// Print with indentation
	indent := strings.Repeat("  ", level)
	sb.WriteString(fmt.Sprintf("%s- %s (%s)\n", indent, agentID, node.Role))

	// Print children
	for _, childID := range node.Children {
		hm.buildTree(sb, childID, level+1)
	}
}

// HierarchicalExecutionResult contains the result of hierarchical execution
type HierarchicalExecutionResult struct {
	AgentID    string
	Role       AgentRole
	Result     string
	Error      error
	Children   []HierarchicalExecutionResult
	Duration   int64
}
