package agent

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSubagentCommand creates commands for managing subagents.
// This is a reference implementation showing the CLI interface for spawn/agents management.
func NewSubagentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subagent",
		Short: "Manage subagents and shared context",
		Long:  "Commands for managing subagents, shared context, and inter-agent communication",
	}

	cmd.AddCommand(
		newListAgentsCommand(),
		newSpawnCommand(),
		newContextCommand(),
		newMessageLogCommand(),
	)

	return cmd
}

func newListAgentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all active subagents",
		Example: `  picoclaw agent subagent list
  picoclaw agent subagent list --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")

			// TODO: Integrate with actual SubagentManager
			fmt.Println("Active Subagents:")
			fmt.Println("  - subagent-1 (coder)    status: running   created: 2026-03-04T10:00:00Z")
			fmt.Println("  - subagent-2 (reviewer) status: completed  created: 2026-03-04T09:30:00Z")

			if verbose {
				fmt.Println("\nVerbose output:")
				fmt.Println("  Subagent ID: subagent-1")
				fmt.Println("  Label: coder")
				fmt.Println("  Type: subagent")
				fmt.Println("  Status: running")
				fmt.Println("  Parent: main")
				fmt.Println("  Capabilities: [shell, edit, read, write]")
			}

			return nil
		},
	}
	cmd.Flags().BoolP("verbose", "v", false, "Show verbose output")
	return cmd
}

func newSpawnCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spawn <task>",
		Short: "Spawn a new subagent with a task",
		Args:  cobra.MinimumNArgs(1),
		Example: `  picoclaw agent subagent spawn "Review code in /path/to/file"
  picoclaw agent subagent spawn --label mytask "Do something"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			task := args[0]
			label, _ := cmd.Flags().GetString("label")
			agentID, _ := cmd.Flags().GetString("agent")

			// TODO: Integrate with actual SubagentManager.Spawn()
			fmt.Printf("Spawning subagent...\n")
			if label != "" {
				fmt.Printf("  Label: %s\n", label)
			}
			fmt.Printf("  Task: %s\n", task)
			if agentID != "" {
				fmt.Printf("  Target Agent: %s\n", agentID)
			}
			fmt.Println("\nSubagent spawned successfully!")
			fmt.Println("Use 'picoclaw agent subagent list' to view status")

			return nil
		},
	}
	cmd.Flags().StringP("label", "l", "", "Label for the subagent")
	cmd.Flags().StringP("agent", "a", "", "Target agent ID")
	return cmd
}

func newContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context [key]",
		Short: "View or manage shared context",
		Args:  cobra.MaximumNArgs(1),
		Example: `  picoclaw agent subagent context
  picoclaw agent subagent context mykey
  picoclaw agent subagent context --set mykey myvalue
  picoclaw agent subagent context --delete mykey`,
		RunE: func(cmd *cobra.Command, args []string) error {
			setKey, _ := cmd.Flags().GetString("set")
			deleteKey, _ := cmd.Flags().GetBool("delete")

			// TODO: Integrate with actual SharedContext
			if setKey != "" {
				fmt.Printf("Setting context: %s\n", setKey)
				return nil
			}

			if deleteKey {
				key, _ := cmd.Flags().GetString("key")
				fmt.Printf("Deleting context key: %s\n", key)
				return nil
			}

			if len(args) > 0 {
				key := args[0]
				fmt.Printf("Context value for '%s': <value>\n", key)
				return nil
			}

			// List all context
			fmt.Println("Shared Context:")
			fmt.Println("  task:subagent-1:result = 'Review completed with 3 suggestions'")
			fmt.Println("  shared_data:files = [file1.go, file2.go]")
			fmt.Println("  session:started = '2026-03-04T10:00:00Z'")

			return nil
		},
	}
	cmd.Flags().StringP("set", "s", "", "Set a context value")
	cmd.Flags().BoolP("delete", "d", false, "Delete a context key")
	cmd.Flags().StringP("key", "k", "", "Key for delete operation")
	return cmd
}

func newMessageLogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "View message log between agents",
		Example: `  picoclaw agent subagent log
  picoclaw agent subagent log --agent subagent-1
  picoclaw agent subagent log --since 2026-03-04T09:00:00Z`,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID, _ := cmd.Flags().GetString("agent")
			since, _ := cmd.Flags().GetString("since")

			// TODO: Integrate with actual SharedContext.GetMessageLog()
			fmt.Println("Message Log:")

			if agentID != "" {
				fmt.Printf("  (filtered by agent: %s)\n", agentID)
			}
			if since != "" {
				fmt.Printf("  (since: %s)\n", since)
			}

			fmt.Println("  [10:05:00] subagent-1 -> main: Task 'review' completed")
			fmt.Println("  [10:04:30] main -> subagent-1: Please review file.go")
			fmt.Println("  [10:04:00] subagent-1 -> broadcast: Starting review task")
			fmt.Println("  [10:03:00] main -> subagent-1: Spawned for review task")

			return nil
		},
	}
	cmd.Flags().StringP("agent", "a", "", "Filter by agent ID")
	cmd.Flags().StringP("since", "t", "", "Filter by timestamp")
	return cmd
}
