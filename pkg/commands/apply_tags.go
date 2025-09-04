package commands

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	cliWrappers "github.com/mmorhun/konflux-task-cli/pkg/cliwrappers"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
	"github.com/spf13/cobra"

	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

var ApplyTagsParamsConfig = map[string]common.Parameter{
	"image-url": {
		Name:       "image-url",
		ShortName:  "i",
		EnvVarName: "IMAGE_URL",
		TypeKind:   reflect.String,
		Usage:      "Image URL to add tags to",
		Required:   true,
	},
	"digest": {
		Name:       "digest",
		ShortName:  "d",
		EnvVarName: "IMAGE_DIGEST",
		TypeKind:   reflect.String,
		Usage:      "Image digest to add tags to",
		Required:   true,
	},
	"tags": {
		Name:         "tags",
		ShortName:    "t",
		EnvVarName:   "TAGS",
		TypeKind:     reflect.Array,
		DefaultValue: "",
		Usage:        "Tags to add to the given image",
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

type ApplyTagsParams struct {
	ImageUrl string   `paramName:"image-url"`
	Digest   string   `paramName:"digest"`
	NewTags  []string `paramName:"tags"`
	Verbose  bool     `paramName:"verbose"`
}

type ApplyTagsCliWrappers struct {
	SkopeoCli cliWrappers.SkopeoCliInterface
}

type ApplyTags struct {
	Params      *ApplyTagsParams
	CliWrappers ApplyTagsCliWrappers

	imageWithoutTag string
	imageByDigest   string
}

func NewApplyTags(cmd *cobra.Command) (*ApplyTags, error) {
	applyTags := &ApplyTags{}

	params := &ApplyTagsParams{}
	if err := common.ParseParameters(cmd, ApplyTagsParamsConfig, params); err != nil {
		return nil, err
	}
	applyTags.Params = params

	if err := applyTags.initCliWrappers(); err != nil {
		return nil, err
	}

	return applyTags, nil
}

func (c *ApplyTags) initCliWrappers() error {
	executor := cliWrappers.NewCliExecutor(c.Params.Verbose)

	skopeoCli, err := cliWrappers.NewSkopeoCli(executor, c.Params.Verbose)
	if err != nil {
		return err
	}
	c.CliWrappers.SkopeoCli = skopeoCli
	return nil
}

func (c *ApplyTags) Run() error {
	if c.Params.Verbose {
		l.Logger.Infof("[param] Image repository: %s", c.Params.ImageUrl)
		l.Logger.Infof("[param] Image digest: %s", c.Params.Digest)
		if len(c.Params.NewTags) > 0 {
			l.Logger.Infof("[param] Tags: %s", strings.Join(c.Params.NewTags, ", "))
		}
	}

	if err := c.validateParams(); err != nil {
		return err
	}

	c.imageWithoutTag = c.stripTag(c.Params.ImageUrl)
	c.imageByDigest = c.imageWithoutTag + "@" + c.Params.Digest

	if err := c.applyTagsFromParam(); err != nil {
		return err
	}

	if err := c.applyTagsFromLabel(); err != nil {
		return err
	}

	return nil
}

func (c *ApplyTags) applyTagsFromParam() error {
	if len(c.Params.NewTags) > 0 {
		l.Logger.Infof("Applying following tags from parameter: %s", strings.Join(c.Params.NewTags, ", "))
		return c.applyTags(c.Params.NewTags)
	}
	l.Logger.Info("No additional tags provided by tags parameter")
	return nil
}

func (c *ApplyTags) applyTagsFromLabel() error {
	inspectArgs := &cliWrappers.SkopeoInspectArgs{
		ImageRef:   c.imageByDigest,
		Format:     `{{ index .Labels "konflux.additional-tags" }}`,
		RetryTimes: 3,
		NoTags:     true,
	}
	output, err := c.CliWrappers.SkopeoCli.Inspect(inspectArgs)
	if err != nil {
		return err
	}
	tagSeparatorRegex := regexp.MustCompile(`[\s,]+`)
	tags := tagSeparatorRegex.Split(output, -1)
	l.Logger.Infof("Applying following tags from label: %s", strings.Join(tags, ", "))

	return c.applyTags(tags)
}

func (c *ApplyTags) applyTags(tags []string) error {
	args := &cliWrappers.SkopeoCopyArgs{
		BaseImage:  c.imageByDigest,
		MultiArch:  cliWrappers.SkopeoCopyArgMultiArchIndexOnly,
		RetryTimes: 3,
	}
	for _, tag := range tags {
		if !c.isTagValid(tag) {
			return fmt.Errorf("tag '%s' is not valid", tag)
		}

		l.Logger.Infof("Applying tag '%s'", tag)
		args.TargetImage = c.imageWithoutTag + ":" + tag
		if err := c.CliWrappers.SkopeoCli.Copy(args); err != nil {
			return err
		}
		l.Logger.Infof("Tag '%s' pushed", tag)
	}
	return nil
}

func (c *ApplyTags) isTagValid(tag string) bool {
	tagPattern := "^[a-zA-Z0-9_][a-zA-Z0-9._-]{0,127}$"
	tagRegex := regexp.MustCompile(tagPattern)
	return tagRegex.MatchString(tag)
}

func (c *ApplyTags) validateParams() error {
	digestPattern := `^sha256:[a-f0-9]{64}$`
	digestRegex := regexp.MustCompile(digestPattern)

	if !digestRegex.MatchString(c.Params.Digest) {
		return fmt.Errorf("image digest '%s' is invalid", c.Params.Digest)
	}

	return nil
}

func (c *ApplyTags) stripTag(imageURL string) string {
	index := strings.LastIndex(imageURL, ":")
	if index == -1 {
		return imageURL
	}
	return imageURL[:index]
}
