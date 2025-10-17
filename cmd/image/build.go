package image

import (
	"github.com/spf13/cobra"

	"github.com/mmorhun/konflux-task-cli/pkg/commands"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

// BuildCmd represents the build command
var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a container image",
	Run: func(cmd *cobra.Command, args []string) {
		l.Logger.Info("Starting image build")
		imageBuild, err := commands.NewImageBuild(cmd)
		if err != nil {
			l.Logger.Fatal(err)
		}
		if err := imageBuild.Run(); err != nil {
			l.Logger.Fatal(err)
		}
		l.Logger.Info("Finishing image build")
	},
}

func init() {
	common.RegisterParameters(BuildCmd, commands.ImageBuildParamsConfig)
}
