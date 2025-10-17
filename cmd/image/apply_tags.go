package image

import (
	"github.com/spf13/cobra"

	"github.com/mmorhun/konflux-task-cli/pkg/commands"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

// ApplyTagsCmd represents the apply-tags command
var ApplyTagsCmd = &cobra.Command{
	Use:   "apply-tags",
	Short: "Creates more tags for provided image",
	Long: `Creates additional tags for the given image.
It might be useful when, for example, the build produces hash based tag, but 'latest' or some other tags needed.

Tags can be defined in two ways:
 - via tags parameter
 - via image label 'konflux.additional-tags' value.
`,
	Run: func(cmd *cobra.Command, args []string) {
		l.Logger.Info("Starting apply-tags")
		applyTags, err := commands.NewApplyTags(cmd)
		if err != nil {
			l.Logger.Fatal(err)
		}
		if err := applyTags.Run(); err != nil {
			l.Logger.Fatal(err)
		}
		l.Logger.Info("Finishing apply-tags")
	},
}

func init() {
	common.RegisterParameters(ApplyTagsCmd, commands.ApplyTagsParamsConfig)
}
