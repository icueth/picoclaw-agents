package skills

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInstallBuiltinCommand(workspaceFn func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install-builtin",
		Short:   "Install all builtin skills to global directory (shared across all agents)",
		Example: `picoclaw skills install-builtin`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Install to global skills directory so all agents can share
			home, _ := os.UserHomeDir()
			globalSkillsDir := filepath.Join(home, ".picoclaw", "skills")
			SkillsInstallBuiltinCmd(globalSkillsDir)
			return nil
		},
	}

	return cmd
}
