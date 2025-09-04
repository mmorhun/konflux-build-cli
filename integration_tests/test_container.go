package integration_tests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	cliWrappers "github.com/mmorhun/konflux-task-cli/pkg/cliwrappers"
)

var containerTool string = "docker"

type TestContainer struct {
	name       string
	image      string
	workdir    string
	privileged bool
	env        map[string]string
	volumes    map[string]string
	ports      map[string]string
	results    map[string]string

	verbose  bool
	executor cliWrappers.CliExecutorInterface

	isStarted bool // TODO add checks in methods
}

func NewTestContainer(name, image string, verbose bool) *TestContainer {
	return &TestContainer{
		executor: cliWrappers.NewCliExecutor(verbose),
		name:     name,
		image:    image,
		env:      make(map[string]string),
		volumes:  make(map[string]string),
		ports:    make(map[string]string),
		results:  make(map[string]string),
	}
}

func (c *TestContainer) SetWorkdir(workdir string) {
	c.workdir = workdir
}

func (c *TestContainer) AddEnv(key, value string) {
	c.env[key] = value
}

func (c *TestContainer) AddVolume(hostPath, containerPath string) {
	c.volumes[hostPath] = containerPath
}

func (c *TestContainer) AddPort(hostPort, containerPort string) {
	c.ports[hostPort] = containerPort
}

func (c *TestContainer) containerExists(isRunning bool) (bool, error) {
	args := []string{"ps", "-q"}
	if !isRunning {
		args = append(args, "-a")
	}
	args = append(args, "-f", "name="+c.name)

	stdout, stderr, err := c.executor.Execute(containerTool, args...)
	if c.verbose || err != nil {
		fmt.Printf("[stdout]:\n%s\n", stdout.String())
		fmt.Printf("[stderr]:\n%s\n", stderr.String())
		return false, err
	}
	return len(stdout.String()) > 0, nil
}

func (c *TestContainer) checkContainer() error {
	existRunning, err := c.containerExists(true)
	if err != nil {
		return err
	}
	if existRunning {
		return fmt.Errorf("container with name '%s' exists and running", c.name)
	}

	existStopped, err := c.containerExists(false)
	if err != nil {
		return err
	}
	if existStopped {
		return c.Delete()
	}
	return nil
}

func (c *TestContainer) Start() error {
	if err := c.checkContainer(); err != nil {
		return err
	}

	args := []string{"run", "--detach", "--name", c.name}
	for name, value := range c.env {
		args = append(args, "-e", name+"="+value)
	}
	// Results set via env vars
	for name, value := range c.results {
		args = append(args, "-e", name+"="+value)
	}
	for hostPath, containerPath := range c.volumes {
		args = append(args, "-v", hostPath+":"+containerPath)
	}
	for hostPort, containerPort := range c.ports {
		args = append(args, "-p", hostPort+":"+containerPort)
	}
	if c.workdir != "" {
		args = append(args, "--workdir", c.workdir)
	}
	if c.privileged {
		args = append(args, "--privileged")
	}

	args = append(args, "--entrypoint", "sleep", c.image, "infinity")

	stdout, stderr, err := c.executor.Execute(containerTool, args...)
	if c.verbose || err != nil {
		fmt.Printf("[stdout]:\n%s\n", stdout.String())
		fmt.Printf("[stderr]:\n%s\n", stderr.String())
	}
	c.isStarted = true
	return err
}

func (c *TestContainer) Delete() error {
	stdout, stderr, err := c.executor.Execute(containerTool, "rm", "-f", c.name)
	if c.verbose || err != nil {
		fmt.Printf("[stdout]:\n%s\n", stdout.String())
		fmt.Printf("[stderr]:\n%s\n", stderr.String())
	}
	return err
}

func (c *TestContainer) CopyFileIntoContainer(hostPath, containerPath string) error {
	stdout, stderr, err := c.executor.Execute(containerTool, "cp", hostPath, c.name+":"+containerPath)
	if c.verbose || err != nil {
		fmt.Printf("[stdout]:\n%s\n", stdout.String())
		fmt.Printf("[stderr]:\n%s\n", stderr.String())
	}
	return err
}

func (c *TestContainer) GetFileContent(path string) (string, error) {
	stdout, stderr, err := c.executor.Execute(containerTool, "exec", c.name, "cat", path)
	if c.verbose || err != nil {
		fmt.Printf("[stdout]:\n%s\n", stdout.String())
		fmt.Printf("[stderr]:\n%s\n", stderr.String())
		if strings.Contains(stderr.String(), "No such file or directory") {
			return "", fmt.Errorf("no such file or directory: '%s'", path)
		}
		return "", err
	}
	return stdout.String(), nil
}

func (c *TestContainer) ExecuteAndWait(command string, args ...string) error {
	execArgs := []string{"exec", "-t", c.name}
	execArgs = append(execArgs, command)
	execArgs = append(execArgs, args...)

	stdout, stderr, err := c.executor.Execute(containerTool, execArgs...)
	if c.verbose || err != nil {
		fmt.Printf("[stdout]:\n%s\n", stdout.String())
		fmt.Printf("[stderr]:\n%s\n", stderr.String())
	}
	return err
}

func (c *TestContainer) DebugCli(cliArgs ...string) error {
	dlvPath, err := getDlvPath()
	if err != nil {
		return err
	}
	err = c.CopyFileIntoContainer(dlvPath, "/usr/bin/")
	if err != nil {
		return err
	}

	execArgs := []string{"exec", "-t", c.name}
	execArgs = append(execArgs, "dlv", "--listen=0.0.0.0:2345", "--headless=true", "--log=true", "--api-version=2", "exec", "/usr/bin/"+KonfluxCli)
	if len(cliArgs) > 0 {
		execArgs = append(execArgs, "--")
		execArgs = append(execArgs, cliArgs...)
	}

	stdout, stderr, err := c.executor.Execute(containerTool, execArgs...)
	if c.verbose || err != nil {
		fmt.Printf("[stdout]:\n%s\n", stdout.String())
		fmt.Printf("[stderr]:\n%s\n", stderr.String())
	}
	return err
}

func (c *TestContainer) AddTaskResult(result string) {
	c.results[result] = path.Join(ResultsPathInContainer, result)
}

func (c *TestContainer) GetTaskResultValue(result string) (string, error) {
	resultFile, resultRegistered := c.results[result]
	if !resultRegistered {
		return "", fmt.Errorf("result '%s' is not registered", result)
	}
	resultValue, err := c.GetFileContent(resultFile)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such file or directory") {
			return "", fmt.Errorf("result '%s' is not created", result)
		}
		return "", err
	}
	return resultValue, nil
}

func (c *TestContainer) InjectDockerAuth(registry, login, password string) error {
	type dockerConfigAuth struct {
		Auth string `json:"auth"`
	}
	type dockerConfigJson struct {
		Auths map[string]dockerConfigAuth `json:"auths"`
	}

	auth := dockerConfigAuth{Auth: base64.StdEncoding.EncodeToString([]byte(login + ":" + password))}
	auths := dockerConfigJson{Auths: map[string]dockerConfigAuth{registry: auth}}
	authContent, err := json.Marshal(auths)
	if err != nil {
		return err
	}

	filePath, err := SaveToTempFile(authContent)
	if err != nil {
		return err
	}
	defer func() { os.Remove(filePath) }()

	dockerDir := "/root/.docker"
	if err := c.ExecuteAndWait("mkdir", "-p", dockerDir); err != nil {
		return err
	}
	if err := c.CopyFileIntoContainer(filePath, path.Join(dockerDir, "config.json")); err != nil {
		return err
	}

	return nil
}
