package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"picoclaw/agent/pkg/config"
)

type ExecTool struct {
	workingDir          string
	timeout             time.Duration
	denyPatterns        []*regexp.Regexp
	allowPatterns       []*regexp.Regexp
	customAllowPatterns []*regexp.Regexp
	restrictToWorkspace bool
}

var (
	// criticalDenyPatterns — ALWAYS blocked regardless of safety level.
	// These are irreversible, destructive operations that can cause data loss
	// or render the system unusable.
	criticalDenyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\brm\s+-[rf]{1,2}\b`),          // Recursive/forced file deletion
		regexp.MustCompile(`\bdel\s+/[fq]\b`),              // Windows forced deletion
		regexp.MustCompile(`\brmdir\s+/s\b`),               // Windows recursive dir removal
		regexp.MustCompile(`\b(format|mkfs|diskpart)\b\s`), // Disk formatting
		regexp.MustCompile(`\bdd\s+if=`),                   // Raw disk writes
		regexp.MustCompile(`>\s*/dev/(sd[a-z]|hd[a-z]|vd[a-z]|xvd[a-z]|nvme\d|mmcblk\d|loop\d|dm-\d|md\d|sr\d|nbd\d)`), // Block device writes
		regexp.MustCompile(`\b(shutdown|reboot|poweroff)\b`),                                                           // System power control
		regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),                                                                      // Fork bomb
		regexp.MustCompile(`\bcurl\b.*\|\s*(sh|bash)`),                                                                 // Remote code execution
		regexp.MustCompile(`\bwget\b.*\|\s*(sh|bash)`),                                                                 // Remote code execution
		regexp.MustCompile(`;\s*rm\s+-[rf]`),                                                                           // Chained destructive rm
		regexp.MustCompile(`&&\s*rm\s+-[rf]`),                                                                          // Chained destructive rm
		regexp.MustCompile(`\|\|\s*rm\s+-[rf]`),                                                                        // Chained destructive rm
	}

	// cautiousDenyPatterns — blocked in "balanced" and "strict" modes.
	// These prevent privilege escalation and system-level changes.
	cautiousDenyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\bsudo\s+(su|bash|sh|zsh|csh|fish)\b`), // Shell escalation
		regexp.MustCompile(`\bsudo\s+(-i|--login|-s)\b`),           // Login shell escalation
		regexp.MustCompile(`\bchmod\s+[0-7]{3,4}\s+/`),             // Permissions on root paths
		regexp.MustCompile(`\bchown\s+-R\b`),                       // Recursive ownership change
		regexp.MustCompile(`\bapt\s+(remove|purge)\b`),             // Package removal
		regexp.MustCompile(`\byum\s+remove\b`),                     // Package removal
		regexp.MustCompile(`\bdnf\s+remove\b`),                     // Package removal
	}

	// strictDenyPatterns — blocked ONLY in "strict" mode (legacy behavior).
	// These block common dev operations that are safe for personal use.
	strictDenyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\$\([^)]+\)`),                // Command substitution $(...)
		regexp.MustCompile(`\$\{[^}]+\}`),                // Variable expansion ${...}
		regexp.MustCompile("`[^`]+`"),                    // Backtick substitution
		regexp.MustCompile(`\|\s*sh\b`),                  // Pipe to sh
		regexp.MustCompile(`\|\s*bash\b`),                // Pipe to bash
		regexp.MustCompile(`<<\s*EOF`),                   // Heredoc
		regexp.MustCompile(`\$\(\s*cat\s+`),              // Command substitution with cat
		regexp.MustCompile(`\$\(\s*curl\s+`),             // Command substitution with curl
		regexp.MustCompile(`\$\(\s*wget\s+`),             // Command substitution with wget
		regexp.MustCompile(`\$\(\s*which\s+`),            // Command substitution with which
		regexp.MustCompile(`\bpkill\b`),                  // Process kill by name
		regexp.MustCompile(`\bkillall\b`),                // Kill all by name
		regexp.MustCompile(`\bkill\s+-[9]\b`),            // Force kill
		regexp.MustCompile(`\bnpm\s+install\s+-g\b`),     // Global npm install
		regexp.MustCompile(`\bpip\s+install\s+--user\b`), // User pip install
		regexp.MustCompile(`\bdocker\s+run\b`),           // Docker run
		regexp.MustCompile(`\bdocker\s+exec\b`),          // Docker exec
		regexp.MustCompile(`\bgit\s+push\b`),             // Git push
		regexp.MustCompile(`\bgit\s+force\b`),            // Git force
		regexp.MustCompile(`\bssh\b.*@`),                 // SSH connections
		regexp.MustCompile(`\beval\b`),                   // Eval
		regexp.MustCompile(`\bsource\s+.*\.sh\b`),        // Source shell scripts
	}

	// defaultDenyPatterns is the legacy combined list (used when no config provided).
	// Equivalent to "strict" safety level.
	defaultDenyPatterns = func() []*regexp.Regexp {
		all := make([]*regexp.Regexp, 0, len(criticalDenyPatterns)+len(cautiousDenyPatterns)+len(strictDenyPatterns))
		all = append(all, criticalDenyPatterns...)
		all = append(all, cautiousDenyPatterns...)
		all = append(all, strictDenyPatterns...)
		return all
	}()

	// absolutePathPattern matches absolute file paths in commands (Unix and Windows).
	absolutePathPattern = regexp.MustCompile(`[A-Za-z]:\\[^\\\"']+|/[^\s\"']+`)

	// urlPattern detects URLs so their path components aren't treated as filesystem paths.
	urlPattern = regexp.MustCompile(`(?:https?|ftp|ssh|git)://[^\s"']+`)

	// safePaths are kernel pseudo-devices that are always safe to reference in
	// commands, regardless of workspace restriction. They contain no user data
	// and cannot cause destructive writes.
	safePaths = map[string]bool{
		"/dev/null":    true,
		"/dev/zero":    true,
		"/dev/random":  true,
		"/dev/urandom": true,
		"/dev/stdin":   true,
		"/dev/stdout":  true,
		"/dev/stderr":  true,
	}
)

func NewExecTool(workingDir string, restrict bool) (*ExecTool, error) {
	return NewExecToolWithConfig(workingDir, restrict, nil)
}

func NewExecToolWithConfig(workingDir string, restrict bool, cfg *config.Config) (*ExecTool, error) {
	denyPatterns := make([]*regexp.Regexp, 0)
	customAllowPatterns := make([]*regexp.Regexp, 0)

	if cfg != nil {
		execConfig := cfg.Tools.Exec
		if execConfig.EnableDenyPatterns {
			denyPatterns = buildDenyPatterns(execConfig.SafetyLevel)
			if len(execConfig.CustomDenyPatterns) > 0 {
				fmt.Printf("Using custom deny patterns: %v\n", execConfig.CustomDenyPatterns)
				for _, pattern := range execConfig.CustomDenyPatterns {
					re, err := regexp.Compile(pattern)
					if err != nil {
						return nil, fmt.Errorf("invalid custom deny pattern %q: %w", pattern, err)
					}
					denyPatterns = append(denyPatterns, re)
				}
			}
		} else {
			fmt.Println("Warning: deny patterns are disabled. All commands will be allowed.")
		}
		for _, pattern := range execConfig.CustomAllowPatterns {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid custom allow pattern %q: %w", pattern, err)
			}
			customAllowPatterns = append(customAllowPatterns, re)
		}
	} else {
		denyPatterns = append(denyPatterns, defaultDenyPatterns...)
	}

	return &ExecTool{
		workingDir:          workingDir,
		timeout:             60 * time.Second,
		denyPatterns:        denyPatterns,
		allowPatterns:       nil,
		customAllowPatterns: customAllowPatterns,
		restrictToWorkspace: restrict,
	}, nil
}

// buildDenyPatterns returns the deny pattern list for a given safety level.
//
//	"permissive" — critical only (rm -rf, disk wipe, fork bomb, curl|bash)
//	"balanced"   — critical + cautious (+ privilege escalation, package removal)
//	"strict"     — all patterns (legacy, very restrictive)
//
// Empty or unknown defaults to "balanced".
func buildDenyPatterns(level string) []*regexp.Regexp {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "permissive":
		out := make([]*regexp.Regexp, len(criticalDenyPatterns))
		copy(out, criticalDenyPatterns)
		return out
	case "strict":
		out := make([]*regexp.Regexp, 0, len(criticalDenyPatterns)+len(cautiousDenyPatterns)+len(strictDenyPatterns))
		out = append(out, criticalDenyPatterns...)
		out = append(out, cautiousDenyPatterns...)
		out = append(out, strictDenyPatterns...)
		return out
	default: // "balanced" or empty
		out := make([]*regexp.Regexp, 0, len(criticalDenyPatterns)+len(cautiousDenyPatterns))
		out = append(out, criticalDenyPatterns...)
		out = append(out, cautiousDenyPatterns...)
		return out
	}
}

func (t *ExecTool) Name() string {
	return "exec"
}

func (t *ExecTool) Description() string {
	return "Execute a shell command and return its output. Use with caution."
}

func (t *ExecTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"working_dir": map[string]any{
				"type":        "string",
				"description": "Optional working directory for the command",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ExecTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	command, ok := args["command"].(string)
	if !ok {
		return ErrorResult("command is required")
	}

	cwd := t.workingDir
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		if t.restrictToWorkspace && t.workingDir != "" {
			resolvedWD, err := validatePath(wd, t.workingDir, true)
			if err != nil {
				return ErrorResult("Command blocked by safety guard (" + err.Error() + ")")
			}
			cwd = resolvedWD
		} else {
			cwd = wd
		}
	}

	if cwd == "" {
		wd, err := os.Getwd()
		if err == nil {
			cwd = wd
		}
	} else {
		// Auto-create working directory if it doesn't exist, 
		// because fork/exec will fail before the command can run (even if the command has 'mkdir -p ...')
		if _, err := os.Stat(cwd); os.IsNotExist(err) {
			if err := os.MkdirAll(cwd, 0755); err != nil {
				return ErrorResult(fmt.Sprintf("failed to create working directory: %v", err))
			}
		}
	}

	if guardError := t.guardCommand(command, cwd); guardError != "" {
		return ErrorResult(guardError)
	}

	// timeout == 0 means no timeout
	var cmdCtx context.Context
	var cancel context.CancelFunc
	if t.timeout > 0 {
		cmdCtx, cancel = context.WithTimeout(ctx, t.timeout)
	} else {
		cmdCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", command)
	}
	if cwd != "" {
		cmd.Dir = cwd
	}

	prepareCommandForTermination(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return ErrorResult(fmt.Sprintf("failed to start command: %v", err))
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var err error
	select {
	case err = <-done:
	case <-cmdCtx.Done():
		_ = terminateProcessTree(cmd)
		select {
		case err = <-done:
		case <-time.After(2 * time.Second):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			err = <-done
		}
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}

	if err != nil {
		if errors.Is(cmdCtx.Err(), context.DeadlineExceeded) {
			msg := fmt.Sprintf("Command timed out after %v", t.timeout)
			return &ToolResult{
				ForLLM:  msg,
				ForUser: msg,
				IsError: true,
			}
		}
		output += fmt.Sprintf("\nExit code: %v", err)
	}

	if output == "" {
		output = "(no output)"
	}

	maxLen := 10000
	if len(output) > maxLen {
		output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
	}

	if err != nil {
		return &ToolResult{
			ForLLM:  output,
			ForUser: output,
			IsError: true,
		}
	}

	return &ToolResult{
		ForLLM:  output,
		ForUser: output,
		IsError: false,
	}
}

func (t *ExecTool) guardCommand(command, cwd string) string {
	cmd := strings.TrimSpace(command)
	lower := strings.ToLower(cmd)

	// Custom allow patterns exempt a command from deny checks.
	explicitlyAllowed := false
	for _, pattern := range t.customAllowPatterns {
		if pattern.MatchString(lower) {
			explicitlyAllowed = true
			break
		}
	}

	if !explicitlyAllowed {
		for _, pattern := range t.denyPatterns {
			if pattern.MatchString(lower) {
				return "Command blocked by safety guard (dangerous pattern detected)"
			}
		}
	}

	if len(t.allowPatterns) > 0 {
		allowed := false
		for _, pattern := range t.allowPatterns {
			if pattern.MatchString(lower) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "Command blocked by safety guard (not in allowlist)"
		}
	}

	if t.restrictToWorkspace {
		if strings.Contains(cmd, "..\\") || strings.Contains(cmd, "../") {
			return "Command blocked by safety guard (path traversal detected)"
		}

		cwdPath, err := filepath.Abs(cwd)
		if err != nil {
			return ""
		}

		// Find URL ranges so their path components aren't treated as filesystem paths
		urlRanges := urlPattern.FindAllStringIndex(cmd, -1)

		matches := absolutePathPattern.FindAllStringIndex(cmd, -1)

		for _, loc := range matches {
			raw := cmd[loc[0]:loc[1]]

			// Skip paths that fall within a URL
			inURL := false
			for _, ur := range urlRanges {
				if loc[0] >= ur[0] && loc[0] < ur[1] {
					inURL = true
					break
				}
			}
			if inURL {
				continue
			}

			p, err := filepath.Abs(raw)
			if err != nil {
				continue
			}

			if safePaths[p] {
				continue
			}

			rel, err := filepath.Rel(cwdPath, p)
			if err != nil {
				continue
			}

			if strings.HasPrefix(rel, "..") {
				return "Command blocked by safety guard (path outside working dir)"
			}
		}
	}

	return ""
}

func (t *ExecTool) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}

func (t *ExecTool) SetRestrictToWorkspace(restrict bool) {
	t.restrictToWorkspace = restrict
}

func (t *ExecTool) SetAllowPatterns(patterns []string) error {
	t.allowPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid allow pattern %q: %w", p, err)
		}
		t.allowPatterns = append(t.allowPatterns, re)
	}
	return nil
}
