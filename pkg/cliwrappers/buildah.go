package cliwrappers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

type BuildahCliInterface interface {
	Build(args *BuildahBuildArgs) (string, string, error)
	Push(image string) (string, error)
}

var _ BuildahCliInterface = &BuildahCli{}

type BuildahCli struct {
	Executor CliExecutorInterface
	Verbose  bool
}

func NewBuildahCli(executor CliExecutorInterface, verbose bool) (*BuildahCli, error) {
	buildahCliAvailable, err := CheckCliToolAvailable("buildah")
	if err != nil {
		return nil, err
	}
	if !buildahCliAvailable {
		return nil, errors.New("buildah CLI is not available")
	}

	unshareCliAvailable, err := CheckCliToolAvailable("unshare")
	if err != nil {
		return nil, err
	}
	if !unshareCliAvailable {
		return nil, errors.New("unshare CLI is not available")
	}

	return &BuildahCli{
		Executor: executor,
		Verbose:  verbose,
	}, nil
}

type BuildahBuildArgs struct {
	Image          string
	DockerfilePath string
	SourceDir      string
	Annotations    []string
	Labels         []string
}

// Build builds the image and returns the built image and its digest
func (b *BuildahCli) Build(args *BuildahBuildArgs) (string, string, error) {
	if args.Image == "" {
		return "", "", errors.New("image to build must be set")
	}

	buildahArgs := []string{"build", "--no-cache", "--ulimit", "nofile=4096:4096", "--http-proxy=false"}
	if args.DockerfilePath != "" {
		buildahArgs = append(buildahArgs, "-f", args.DockerfilePath)
	}
	for _, label := range args.Labels {
		buildahArgs = append(buildahArgs, "--label", label)
	}
	for _, annotation := range args.Annotations {
		buildahArgs = append(buildahArgs, "--annotation", annotation)
	}
	buildahArgs = append(buildahArgs, "-t", args.Image, ".")

	buildahCmd := "buildah " + strings.Join(buildahArgs, " ")

	unshareArgs := []string{"-Uf", "--keep-caps", "-r", "--map-users", "1,1,65536", "--map-groups", "1,1,65536", "--mount"}
	if args.SourceDir != "" {
		unshareArgs = append(unshareArgs, "-w", args.SourceDir)
	}
	unshareArgs = append(unshareArgs, "--", "sh", "-c", buildahCmd)

	stdout, stderr, _, err := b.Executor.Execute("unshare", unshareArgs...)
	if err != nil {
		l.Logger.Errorf("[stdout]:\n%s", stdout)
		l.Logger.Errorf("[stderr]:\n%s", stderr)
		return "", "", fmt.Errorf("unshare ... buildah build failed: %v", err)
	}

	image, localDigest, err := parseImageAndDigest(stdout)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse image digest: %w", err)
	}

	if b.Verbose {
		l.Logger.Info("[stdout]:\n" + stdout)
	}

	return image, "sha256:" + localDigest, nil
}

func parseImageAndDigest(output string) (string, string, error) {
	imageRegex := regexp.MustCompile(`Successfully tagged\s+([^\s]+)`)
	digestRegex := regexp.MustCompile(`([a-f0-9]{64})`)

	imageMatch := imageRegex.FindStringSubmatch(output)
	if len(imageMatch) < 2 {
		return "", "", errors.New("image reference not found")
	}
	imageName := imageMatch[1]

	// Find all possible digests and take the last one
	digestMatches := digestRegex.FindAllString(output, -1)
	if len(digestMatches) == 0 {
		return "", "", errors.New("digest not found")
	}
	digest := digestMatches[len(digestMatches)-1]

	return imageName, digest, nil
}

// Push image to remote registry and returns remote image digest
func (b *BuildahCli) Push(image string) (string, error) {
	const digestFile = "/tmp/digestfile"
	stdout, stderr, _, err := b.Executor.Execute("buildah", "push", "--digestfile", digestFile, image)
	if err != nil {
		l.Logger.Errorf("[stdout]:\n%s", stdout)
		l.Logger.Errorf("[stderr]:\n%s", stderr)
		return "", fmt.Errorf("buildah push failed: %v", err)
	}

	if b.Verbose {
		l.Logger.Info("[stdout]:\n" + stdout)
	}

	stdout, stderr, _, err = b.Executor.Execute("cat", digestFile)
	if err != nil {
		l.Logger.Errorf("[stdout]:\n%s", stdout)
		l.Logger.Errorf("[stderr]:\n%s", stderr)
		return "", fmt.Errorf("failed to read digest file: %v", err)
	}

	return stdout, nil
}
