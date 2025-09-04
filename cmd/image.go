package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mmorhun/konflux-task-cli/cmd/image"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "A subcommand groop to work with container images",
}

func init() {
	rootCmd.AddCommand(imageCmd)

	imageCmd.AddCommand(image.BuildCmd)
	imageCmd.AddCommand(image.ApplyTagsCmd)
}
