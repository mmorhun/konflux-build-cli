package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mmorhun/konflux-task-cli/pkg/commands"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

var gitcloneCmd = &cobra.Command{
	Use:   "gitclone",
	Short: "Clones a git repository",
	Long: `A Konflux helper command to clone git repository.
Mandatory parameters are "url" and "branch".
Note, parameters could be passed as flag or via environmebt variables.
Flags take precedence over environment variable.

The command requires git cli installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		l.Logger.Info("Starting git clone")
		gitClone, err := commands.NewGitClone(cmd)
		if err != nil {
			l.Logger.Fatal(err)
		}
		if err := gitClone.Run(); err != nil {
			l.Logger.Fatal(err)
		}
		l.Logger.Info("Finishing git clone")
	},
}

func init() {
	rootCmd.AddCommand(gitcloneCmd)

	common.RegisterParameters(gitcloneCmd, commands.GitCloneParamsConfig)
	// The above could be done manually:
	// gitcloneCmd.Flags().String("url", "", "Git URL to clone from")
	// gitcloneCmd.Flags().String("branch", "main", "Branch to clone from")
	// gitcloneCmd.Flags().IntP("depth", "d", 0, "Clone depth")
	// gitcloneCmd.Flags().Bool("verbose", false, "Activates verbose mode")
}
