package cliwrappers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
)

type GitCliInterface interface {
	Clone(url, branch string, depth int) (string, error)
	GetRepoHeadFullSha(gitRepoDir string) (string, error)
}

var _ GitCliInterface = &GitCli{}

type GitCli struct {
	Executor CliExecutorInterface
	Verbose  bool
}

func NewGitCli(executor CliExecutorInterface, verbose bool) (*GitCli, error) {
	gitCliAvailable, err := CheckCliToolAvailable("git")
	if err != nil {
		return nil, err
	}
	if !gitCliAvailable {
		return nil, errors.New("git CLI is not available")
	}

	return &GitCli{
		Executor: executor,
		Verbose:  verbose,
	}, nil
}

// Clone clones given git repository and returns path to the repository root folder.
// Returns name of the clonned source directory.
func (g *GitCli) Clone(url, branch string, depth int) (string, error) {
	if url == "" {
		return "", errors.New("url must be set to clone")
	}
	gitArgs := []string{"clone", url}

	if branch == "" {
		branch = "main"
	}
	gitArgs = append(gitArgs, "--branch", branch)

	if depth != 0 {
		gitArgs = append(gitArgs, "--depth", strconv.Itoa(depth))
	}

	stdout, stderr, err := g.Executor.Execute("git", gitArgs...)
	if err != nil {
		l.Logger.Errorf("[stdout]:\n%s", stdout.String())
		l.Logger.Errorf("[stderr]:\n%s", stderr.String())
		return "", fmt.Errorf("git clone failed: %v", err)
	}

	if g.Verbose {
		l.Logger.Info("[stdout]:\n" + stderr.String())
	}

	// Parse output for "Cloning into 'repository-name'..."
	repoDir, err := parseRepoDir(stderr.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to obtain cloned repository directory: %w", err)
	}

	return repoDir, nil
}

// parseRepoDir parses git clone output and returns directory name into which the git repository was cloned.
func parseRepoDir(output []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	re := regexp.MustCompile(`Cloning into '(.+)'`)

	// Check all lines in case of additional git config that prints headers.
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) == 2 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not parse 'Cloning into' line")
}

func (g *GitCli) GetRepoHeadFullSha(gitRepoDir string) (string, error) {
	stdout, stderr, err := g.Executor.ExecuteInDir(gitRepoDir, "git", "rev-parse", "HEAD")
	if err != nil {
		l.Logger.Errorf("[stdout]:\n%s", stdout.String())
		l.Logger.Errorf("[stderr]:\n%s", stderr.String())
		return "", fmt.Errorf("git rev-parse failed: %v", err)
	}

	if g.Verbose {
		l.Logger.Info("[stdout]:\n" + stdout.String())
	}

	fullSha := strings.TrimSpace(string(stdout.String()))
	return fullSha, nil
}
