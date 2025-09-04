package cliwrappers

import (
	"errors"
	"fmt"
	"strconv"

	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

type SkopeoCliInterface interface {
	Copy(args *SkopeoCopyArgs) error
	Inspect(args *SkopeoInspectArgs) (string, error)
}

var _ SkopeoCliInterface = &SkopeoCli{}

type SkopeoCli struct {
	Executor CliExecutorInterface
	Verbose  bool
}

func NewSkopeoCli(executor CliExecutorInterface, verbose bool) (*SkopeoCli, error) {
	skopeoCliAvailable, err := CheckCliToolAvailable("skopeo")
	if err != nil {
		return nil, err
	}
	if !skopeoCliAvailable {
		return nil, errors.New("skopeo CLI is not available")
	}

	return &SkopeoCli{
		Executor: executor,
		Verbose:  verbose,
	}, nil
}

type SkopeoCopyArgMultiArch string

const (
	SkopeoCopyArgMultiArchSystem    SkopeoCopyArgMultiArch = "system"
	SkopeoCopyArgMultiArchAll       SkopeoCopyArgMultiArch = "all"
	SkopeoCopyArgMultiArchIndexOnly SkopeoCopyArgMultiArch = "index-only"
)

type SkopeoCopyArgs struct {
	BaseImage   string
	TargetImage string
	MultiArch   SkopeoCopyArgMultiArch
	RetryTimes  int
	ExtraArgs   []string
}

func (s *SkopeoCli) Copy(args *SkopeoCopyArgs) error {
	if args.BaseImage == "" {
		return errors.New("image to copy from must be set")
	}
	if args.TargetImage == "" {
		return errors.New("image to copy to must be set")
	}

	scopeoArgs := []string{"copy"}

	if args.MultiArch != "" {
		scopeoArgs = append(scopeoArgs, "--multi-arch", string(args.MultiArch))
	}
	if args.RetryTimes != 0 {
		scopeoArgs = append(scopeoArgs, "--retry-times", strconv.Itoa(args.RetryTimes))
	}

	if len(args.ExtraArgs) != 0 {
		scopeoArgs = append(scopeoArgs, args.ExtraArgs...)
	}

	dockerPrefix := "docker://"
	scopeoArgs = append(scopeoArgs, dockerPrefix+args.BaseImage, dockerPrefix+args.TargetImage)

	stdout, stderr, err := s.Executor.Execute("skopeo", scopeoArgs...)
	if err != nil {
		l.Logger.Errorf("[stdout]:\n%s", stdout.String())
		l.Logger.Errorf("[stderr]:\n%s", stderr.String())
		return fmt.Errorf("skopeo copy failed: %v", err)
	}

	if s.Verbose {
		l.Logger.Info("[stdout]:\n" + stdout.String())
	}

	return nil
}

type SkopeoInspectArgs struct {
	ImageRef   string
	RetryTimes int
	Raw        bool
	NoTags     bool
	Format     string
	ExtraArgs  []string
}

func (s *SkopeoCli) Inspect(args *SkopeoInspectArgs) (string, error) {
	if args.ImageRef == "" {
		return "", errors.New("no image to inspect")
	}

	scopeoArgs := []string{"inspect"}

	if args.RetryTimes != 0 {
		scopeoArgs = append(scopeoArgs, "--retry-times", strconv.Itoa(args.RetryTimes))
	}
	if args.Raw {
		scopeoArgs = append(scopeoArgs, "--raw")
	}
	if args.NoTags {
		scopeoArgs = append(scopeoArgs, "--no-tags")
	}
	if args.Format != "" {
		scopeoArgs = append(scopeoArgs, "--format", args.Format)
	}

	if len(args.ExtraArgs) != 0 {
		scopeoArgs = append(scopeoArgs, args.ExtraArgs...)
	}

	dockerPrefix := "docker://"
	scopeoArgs = append(scopeoArgs, dockerPrefix+args.ImageRef)

	stdout, stderr, err := s.Executor.Execute("skopeo", scopeoArgs...)
	if err != nil {
		l.Logger.Errorf("[stdout]:\n%s", stdout.String())
		l.Logger.Errorf("[stderr]:\n%s", stderr.String())
		return "", fmt.Errorf("skopeo inspect failed: %v", err)
	}

	if s.Verbose {
		l.Logger.Info("[stdout]:\n" + stdout.String())
	}

	return stdout.String(), nil
}
