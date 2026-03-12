package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"picoclaw/agent/pkg/agent"
	"picoclaw/agent/pkg/config"
	"github.com/spf13/cobra"
)

// NewWorkspaceCommand creates the workspace management command
func NewWorkspaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage per-agent workspaces",
		Long:  `List, inspect, and manage per-agent workspace directories.`,
	}

	cmd.AddCommand(newWorkspaceListCommand())
	cmd.AddCommand(newWorkspaceInfoCommand())
	cmd.AddCommand(newWorkspaceCleanCommand())
	cmd.AddCommand(newWorkspaceMigrateCommand())

	return cmd
}

func newWorkspaceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all agent workspaces",
		Long:  `List all agent workspaces with their sizes and last modified times.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			workspaceRoot := cfg.WorkspacePath()
			agentsDir := filepath.Join(workspaceRoot, "agents")

			entries, err := os.ReadDir(agentsDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No agent workspaces found.")
					return nil
				}
				return fmt.Errorf("read agents directory: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "AGENT ID\tSIZE\tSESSIONS\tMEMORY\tLAST MODIFIED")

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				agentID := entry.Name()
				aw := agent.NewAgentWorkspace(workspaceRoot, agentID)

				// Get directory size
				size := getDirSize(aw.BasePath)
				sizeStr := formatSize(size)

				// Count sessions
				sessions := countFiles(aw.SessionDir, ".json")

				// Check memory
				hasMemory := fileExists(aw.GetMemoryPath())
				memoryStr := "no"
				if hasMemory {
					memoryStr = "yes"
				}

				// Get last modified
				info, _ := os.Stat(aw.BasePath)
				modTime := "unknown"
				if info != nil {
					modTime = info.ModTime().Format("2006-01-02")
				}

				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", agentID, sizeStr, sessions, memoryStr, modTime)
			}

			w.Flush()
			return nil
		},
	}
}

func newWorkspaceInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info [agent-id]",
		Short: "Show detailed info about an agent workspace",
		Long:  `Display detailed information about a specific agent's workspace.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]

			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			workspaceRoot := cfg.WorkspacePath()
			aw := agent.NewAgentWorkspace(workspaceRoot, agentID)

			fmt.Printf("Agent Workspace: %s\n", agentID)
			fmt.Printf("==================\n\n")
			fmt.Printf("Base Path: %s\n", aw.BasePath)
			fmt.Printf("Memory Dir: %s\n", aw.MemoryDir)
			fmt.Printf("Session Dir: %s\n", aw.SessionDir)
			fmt.Printf("\n")

			// Memory info
			memoryPath := aw.GetMemoryPath()
			if fileExists(memoryPath) {
				info, _ := os.Stat(memoryPath)
				fmt.Printf("MEMORY.md: %s\n", formatSize(info.Size()))

				// Show first few lines
				if data, err := os.ReadFile(memoryPath); err == nil && len(data) > 0 {
					lines := strings.Split(string(data), "\n")
					if len(lines) > 0 {
						fmt.Printf("  Title: %s\n", strings.TrimPrefix(lines[0], "# "))
					}
				}
			} else {
				fmt.Printf("MEMORY.md: not found\n")
			}

			// Sessions info
			sessions := countFiles(aw.SessionDir, ".json")
			fmt.Printf("\nSessions: %d\n", sessions)

			// Entries
			entriesPath := filepath.Join(aw.BasePath, "entries.json")
			if fileExists(entriesPath) {
				info, _ := os.Stat(entriesPath)
				fmt.Printf("Structured Memory: %s\n", formatSize(info.Size()))
			}

			// Total size
			totalSize := getDirSize(aw.BasePath)
			fmt.Printf("\nTotal Size: %s\n", formatSize(totalSize))

			return nil
		},
	}
}

func newWorkspaceCleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean [agent-id]",
		Short: "Clean up an agent workspace (removes all data)",
		Long: `Remove all data for a specific agent. This will delete:
- All session history
- Memory files
- Task archives
- Structured memory entries

Use with caution! This action cannot be undone.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]

			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			workspaceRoot := cfg.WorkspacePath()
			aw := agent.NewAgentWorkspace(workspaceRoot, agentID)

			// Confirm
			fmt.Printf("This will DELETE all data for agent '%s' at:\n", agentID)
			fmt.Printf("  %s\n\n", aw.BasePath)
			fmt.Print("Are you sure? Type the agent ID to confirm: ")

			var confirm string
			fmt.Scanln(&confirm)

			if confirm != agentID {
				fmt.Println("Confirmation failed. Aborting.")
				return nil
			}

			// Delete workspace
			if err := agent.DeleteAgentWorkspace(workspaceRoot, agentID); err != nil {
				return fmt.Errorf("delete workspace: %w", err)
			}

			fmt.Printf("Agent workspace '%s' has been deleted.\n", agentID)
			return nil
		},
	}
}

func newWorkspaceMigrateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate legacy shared data to per-agent workspaces",
		Long: `Migrate sessions and memory from the shared workspace to per-agent workspaces.
This will:
1. Detect which agents have been used
2. Move their sessions to per-agent directories
3. Copy shared MEMORY.md to each agent as a starting point`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			workspaceRoot := cfg.WorkspacePath()

			// Check for legacy data
			migrator := agent.NewWorkspaceMigration(workspaceRoot)
			legacyData := migrator.DetectLegacyData()

			fmt.Println("Migration Analysis")
			fmt.Println("==================")
			fmt.Printf("Legacy sessions found: %v\n", legacyData["has_legacy_sessions"])
			fmt.Printf("Legacy memory found: %v\n", legacyData["has_legacy_memory"])
			fmt.Printf("Session count: %d\n", legacyData["legacy_session_count"])

			agents := legacyData["legacy_agents_found"].([]string)
			fmt.Printf("Agents detected: %d\n", len(agents))
			for _, a := range agents {
				fmt.Printf("  - %s\n", a)
			}

			if len(agents) == 0 {
				fmt.Println("\nNo agents detected. Nothing to migrate.")
				return nil
			}

			fmt.Print("\nProceed with migration? [y/N]: ")
			var confirm string
			fmt.Scanln(&confirm)

			if strings.ToLower(confirm) != "y" {
				fmt.Println("Migration cancelled.")
				return nil
			}

			// Run migration
			results, err := migrator.MigrateAll()
			if err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			// Save report
			if err := migrator.SaveReport(); err != nil {
				fmt.Printf("Warning: could not save report: %v\n", err)
			}

			// Print results
			fmt.Println("\nMigration Results")
			fmt.Println("=================")

			totalSessions := 0
			totalMemory := 0

			for _, r := range results {
				fmt.Printf("\nAgent: %s\n", r.AgentID)
				fmt.Printf("  Sessions moved: %d\n", r.SessionsMoved)
				fmt.Printf("  Memory copied: %v\n", r.MemoryCopied)
				if len(r.Errors) > 0 {
					fmt.Printf("  Errors: %d\n", len(r.Errors))
					for _, e := range r.Errors {
						fmt.Printf("    - %s\n", e)
					}
				}
				totalSessions += r.SessionsMoved
				if r.MemoryCopied {
					totalMemory++
				}
			}

			fmt.Printf("\nTotal: %d agents, %d sessions moved, %d memories copied\n",
				len(results), totalSessions, totalMemory)
			fmt.Printf("\nMigration report saved to: %s\n",
				filepath.Join(workspaceRoot, "MIGRATION_REPORT.md"))

			return nil
		},
	}
}

// Helper functions

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func countFiles(dir, suffix string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			count++
		}
	}
	return count
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
