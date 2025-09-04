package commands

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	cliWrappers "github.com/mmorhun/konflux-task-cli/pkg/cliwrappers"
	"github.com/mmorhun/konflux-task-cli/pkg/common"
	l "github.com/mmorhun/konflux-task-cli/pkg/logger"
	"github.com/spf13/cobra"
)

// type GitCloneParamName string
// const (
// 	P_url     GitCloneParamName = "url"
// 	P_branch  GitCloneParamName = "branch"
// 	P_depth   GitCloneParamName = "depth"
// 	P_verbose GitCloneParamName = "verbose"
// )
// var GitCloneParamsInfo = map[GitCloneParamName]common.Parameter{

var GitCloneParamsConfig = map[string]common.Parameter{
	"url": {
		Name:       "url",
		EnvVarName: "GIT_REPO_URL",
		TypeKind:   reflect.String,
		Usage:      "Git URL to clone from",
		Required:   true,
	},
	"branch": {
		Name:         "branch",
		ShortName:    "b",
		EnvVarName:   "GIT_BRANCH",
		TypeKind:     reflect.String,
		DefaultValue: "main",
		Usage:        "Branch to clone from",
		Required:     false,
	},
	"depth": {
		Name:         "depth",
		ShortName:    "d",
		TypeKind:     reflect.Int,
		DefaultValue: "",
		Usage:        "Clone depth",
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

type GitCloneParams struct {
	RepoUrl string `paramName:"url"`
	Branch  string `paramName:"branch"`
	Depth   int    `paramName:"depth"`
	Verbose bool   `paramName:"verbose"`
}

type GitCloneResultFilesPath struct {
	Url         string `env:"RESULT_URL"`
	SourceDir   string `env:"RESULT_SOURCE_DIR"`
	Commit      string `env:"RESULT_COMMIT"`
	ShortCommit string `env:"RESULT_SHORT_COMMIT"`
}

type GitCloneCliWrappers struct {
	GitCli cliWrappers.GitCliInterface
}

type GitClone struct {
	Params        *GitCloneParams
	Results       *GitCloneResultFilesPath
	ResultsWriter common.ResultsWriterInterface
	CliWrappers   GitCloneCliWrappers
}

func NewGitClone(cmd *cobra.Command) (*GitClone, error) {
	gitClone := &GitClone{}

	params := &GitCloneParams{}
	if err := common.ParseParameters(cmd, GitCloneParamsConfig, params); err != nil {
		return nil, err
	}
	gitClone.Params = params

	results := &GitCloneResultFilesPath{}
	if err := common.ReadResultFilesPath(results); err != nil {
		return nil, err
	}
	gitClone.Results = results
	gitClone.ResultsWriter = common.NewResultsWriter(gitClone.Params.Verbose)

	if err := gitClone.initCliWrappers(); err != nil {
		return nil, err
	}

	return gitClone, nil
}

func (c *GitClone) initCliWrappers() error {
	executor := cliWrappers.NewCliExecutor(c.Params.Verbose)

	gitCli, err := cliWrappers.NewGitCli(executor, c.Params.Verbose)
	if err != nil {
		return err
	}
	c.CliWrappers.GitCli = gitCli
	return nil
}

func (c *GitClone) Run() error {
	if c.Params.Verbose {
		l.Logger.Infof("[param] repository: %s", c.Params.RepoUrl)
		if c.Params.Branch != "" {
			l.Logger.Infof("[param] branch: %s", c.Params.Branch)
		}
		if c.Params.Depth > 0 {
			l.Logger.Infof("[param] depth: %d", c.Params.Depth)
		}
	}

	if err := c.validateParams(); err != nil {
		return err
	}

	sourceDir, err := c.CliWrappers.GitCli.Clone(c.Params.RepoUrl, c.Params.Branch, c.Params.Depth)
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	commitSha, err := c.CliWrappers.GitCli.GetRepoHeadFullSha(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to get HEAD SHA: %w", err)
	}
	commitShortSha := commitSha[:7]

	if err := c.ResultsWriter.WriteResultString(c.Params.RepoUrl, c.Results.Url); err != nil {
		return err
	}
	if err := c.ResultsWriter.WriteResultString(sourceDir, c.Results.SourceDir); err != nil {
		return err
	}
	if err := c.ResultsWriter.WriteResultString(commitSha, c.Results.Commit); err != nil {
		return err
	}
	if err := c.ResultsWriter.WriteResultString(commitShortSha, c.Results.ShortCommit); err != nil {
		return err
	}

	if c.Params.Verbose {
		l.Logger.Infof("[result] url: %s", c.Params.RepoUrl)
		l.Logger.Infof("[result] source dir: %s", sourceDir)
		l.Logger.Infof("[result] commit: %s", commitSha)
		l.Logger.Infof("[result] short commit: %s", commitShortSha)
	}

	return nil
}

func (c *GitClone) validateParams() error {
	if c.Params.RepoUrl == "" {
		return errors.New("git repository url must be set")
	}
	if !strings.HasPrefix(c.Params.RepoUrl, "https://") {
		return errors.New("only https protocol is supported")
	}

	return nil
}
