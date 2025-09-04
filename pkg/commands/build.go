package commands

import (
	"reflect"
	"strings"

	cliWrappers "github.com/mmorhun/konflux-task-cli/pkg/cliwrappers"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
	"github.com/spf13/cobra"

	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

var ImageBuildParamsConfig = map[string]common.Parameter{
	"image": {
		Name:       "image",
		ShortName:  "i",
		EnvVarName: "IMAGE",
		TypeKind:   reflect.String,
		Usage:      "Image to produce",
		Required:   true,
	},
	"source-dir": {
		Name:       "source-dir",
		ShortName:  "s",
		EnvVarName: "SOURCE_DIR",
		TypeKind:   reflect.String,
		Usage:      "Path to source directory",
		Required:   true,
	},
	"dockerfile": {
		Name:       "dockerfile",
		ShortName:  "d",
		EnvVarName: "DOCKERFILE",
		TypeKind:   reflect.String,
		Usage:      "Path to Dockerfile",
	},
	"labels": {
		Name:         "labels",
		ShortName:    "l",
		EnvVarName:   "LABELS",
		TypeKind:     reflect.Array,
		DefaultValue: "",
		Usage:        "Labels to add to the image",
	},
	"annotations": {
		Name:         "annotations",
		ShortName:    "a",
		EnvVarName:   "ANNOTATIONS",
		TypeKind:     reflect.Array,
		DefaultValue: "",
		Usage:        "Annotations to add to the image",
	},
	"verbose": {
		Name:         "verbose",
		ShortName:    "v",
		EnvVarName:   "VERBOSE",
		TypeKind:     reflect.Bool,
		DefaultValue: "false",
		Usage:        "Activates verbose mode",
	},
}

type ImageBuildParams struct {
	Image          string   `paramName:"image"`
	DockerfilePath string   `paramName:"dockerfile"`
	SourceDir      string   `paramName:"source-dir"`
	Labels         []string `paramName:"labels"`
	Annotations    []string `paramName:"annotations"`
	Verbose        bool     `paramName:"verbose"`
}

type ImageBuildResultFilesPath struct {
	ImageUrl string `env:"RESULT_IMAGE_URL"`
	Digest   string `env:"RESULT_IMAGE_DIGEST"`
}

type ImageBuildCliWrappers struct {
	BuildahCli cliWrappers.BuildahCliInterface
}

type ImageBuild struct {
	Params        *ImageBuildParams
	Results       *ImageBuildResultFilesPath
	ResultsWriter common.ResultsWriterInterface
	CliWrappers   ImageBuildCliWrappers
}

func NewImageBuild(cmd *cobra.Command) (*ImageBuild, error) {
	imageBuild := &ImageBuild{}

	params := &ImageBuildParams{}
	if err := common.ParseParameters(cmd, ImageBuildParamsConfig, params); err != nil {
		return nil, err
	}
	imageBuild.Params = params

	results := &ImageBuildResultFilesPath{}
	if err := common.ReadResultFilesPath(results); err != nil {
		return nil, err
	}
	imageBuild.Results = results
	imageBuild.ResultsWriter = common.NewResultsWriter(imageBuild.Params.Verbose)

	if err := imageBuild.initCliWrappers(); err != nil {
		return nil, err
	}

	return imageBuild, nil
}

func (c *ImageBuild) initCliWrappers() error {
	executor := cliWrappers.NewCliExecutor(c.Params.Verbose)

	buildahCli, err := cliWrappers.NewBuildahCli(executor, c.Params.Verbose)
	if err != nil {
		return err
	}
	c.CliWrappers.BuildahCli = buildahCli
	return nil
}

func (c *ImageBuild) Run() error {
	if c.Params.Verbose {
		l.Logger.Infof("[param] Image: %s", c.Params.Image)
		l.Logger.Infof("[param] Source directory: %s", c.Params.SourceDir)
		if c.Params.DockerfilePath != "" {
			l.Logger.Infof("[param] Dockerfile: %s", c.Params.DockerfilePath)
		}
		if len(c.Params.Labels) > 0 {
			l.Logger.Infof("[param] Labels: %s", strings.Join(c.Params.Labels, ", "))
		}
		if len(c.Params.Annotations) > 0 {
			l.Logger.Infof("[param] Annotations: %s", strings.Join(c.Params.Annotations, ", "))
		}
	}

	if err := c.validateParams(); err != nil {
		return err
	}

	buildArgs := &cliWrappers.BuildahBuildArgs{
		Image:          c.Params.Image,
		DockerfilePath: c.Params.DockerfilePath,
		SourceDir:      c.Params.SourceDir,
		Labels:         c.Params.Labels,
		Annotations:    c.Params.Annotations,
	}
	image, _, err := c.CliWrappers.BuildahCli.Build(buildArgs)
	if err != nil {
		return err
	}

	digest, err := c.CliWrappers.BuildahCli.Push(c.Params.Image)
	if err != nil {
		return err
	}

	if err := c.ResultsWriter.WriteResultString(image, c.Results.ImageUrl); err != nil {
		return err
	}
	if err := c.ResultsWriter.WriteResultString(digest, c.Results.Digest); err != nil {
		return err
	}

	if c.Params.Verbose {
		l.Logger.Infof("[result] Image URL: %s", image)
		l.Logger.Infof("[result] Image digest: %s", digest)
	}

	return nil
}

func (c *ImageBuild) validateParams() error {
	return nil
}
