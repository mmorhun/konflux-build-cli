## Konflux Tasks CLI PoC

This is proof of concept for Konflux task CLI written in Golang.

Do not expect polished solution, it's a prototype to evaluate.
Some things should be improved.

### How to build

```bash
go build -o konflux-task-cli main.go
```
or (in debug mode):
```bash
go build -gcflags "all=-N -l" -o konflux-task-cli main.go
```

### How to run / debug a command on host

Build the cli and setup the command environment.
You can do it manually or via a script, for example:

```bash
. hack/taskenv/git-clone.sh
```
and run a command:
```bash
./konflux-task-cli gitclone
```

Parameters can be passed via cli arguments or envinonment variables, cli arguments take precedence.

```bash
./konflux-task-cli image apply-tags --image-url quay.io/namespace/image:tag --digest sha256:abcde1234 --tags tag1 tag2
```

Note, that running some commands on host might cause some issues, since the command might work with home directory, etc.
See integration tests section for running / debugging in a container.

### How to run unit tests

Note, unit tests are implemented only for `git clone`.
That shows how mocks and everything works.

Unit tests are implemented twice:
 - Using golang standard testing lib and Gomega (`*_test.go`)
 - Using Ginkgo and Gomega (`*_ginkgo_test.go` and `*_ginkgo_suite_test.go`)

Only one approach should be chosen for the final solution.

To run all unit tests:
```bash
go test ./pkg/...
```

To run / debug a specific test or run all tests in a single file, it's most convenient to use UI of your IDE.
To run specific test from console execute:
```bash
go test -run ^TestGitClone_Success$ ./pkg/...
```
or for all tests in single package:
```bash
go test ./pkg/commands
```

### How to run integration tests

All integration tests are located in `integration_tests` directory.

Integration tests are run in a container which is started (and cleaned) by each test.

Running specific integration tests are identical to the unit tests commands, however, some preparation migth be required.
For example, to test functionality that works with registry one needs to prepare image repository and provide credentials in environment variables:
```bash
export QUAY_ROBOT_NAME=account+robot_name
export QUAY_ROBOT_TOKEN=*****
```
then just run the test.

In order to dubug the cli in test:
 - Compile the cli for debug
 - Change `Debug` package variable to `true`. Note, all cli invocations will wait until the debugger is connected.
 - Run the test itself in debug mode
 - Connect with debugger to remote `dlv` on port `2345`

### e2e tests

In `integration_tests` directory, there is `pipeline_test.go` that runs whole pipeline, each task one by one, passing data between containers.

## Developing commands

The cli is built on top of `cobra` go library.
Commands can be grouped into subcommands, e.g. `cli image build --args`.
All Cobra commands are located in `cmd` package, however,
actual logic is implemented in `pkg/commands` which relies on `pkg/cliwrappers` to execute other clis.

### Scaffolding a new subcommand

It can be done via `cobra-cli`:
```bash
cobra-cli add <command_name>
```
or manually, just copying any command file from `cmd` package and renaming the command to yours.
Example:
```golang
package cmd

var mycommandCmd = &cobra.Command{
	Use:   "my-command",
	Short: "Short description here",
	Long: `Long, multi line description could be here.`,
	Run: func(cmd *cobra.Command, args []string) {
		myCommand, err := commands.NewMyCommand(cmd)
		if err != nil {
			l.Logger.Error(err)
			return
		}
		if err := myCommand.Run(); err != nil {
			l.Logger.Error(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(mycommandCmd)

	common.RegisterParameters(mycommandCmd, commands.MyCommandParamsConfig)
}
```

All `cobra` commands in `cmd` package are headers which is tipical for all commands.
Actual implementation and cli arguments are defined in a convenient parameters array in the command itself `commands` package.
In fact, it's not nessesary to dig into cobra things much, because the cli has own wrappers over cobra parameters.
One just need to define parameters and results and cli will do the rest.
Definition of parameters, results and command constructor are typical for all commands too as shown below:
```golang
package commands

MyCommandParamsConfig = map[string]common.Parameter{
	"url": {
		Name:       "url",
		ShortName:  "u",
		EnvVarName: "URL",
		TypeKind:   reflect.String,
		Usage:      "URL to process",
		Required:   true,
	},
	"verbose": {
		Name:         "verbose",
		ShortName:    "v",
		EnvVarName:   "VERBOSE",
		TypeKind:     reflect.Bool,
		Usage:        "Activates verbose mode",
		DefaultValue: "false",
	},
}

type MyCommandParams struct {
	Url string   `paramName:"url"`
	Verbose bool `paramName:"verbose"`
}

type MyCommandResultFilesPath struct {
	Location string `env:"RESULT_LOCATION"`
	Hash string     `env:"RESULT_HASH"`
}

type MyCommandCliWrappers struct {
	SomeCli cliWrappers.SomeCliInterface
}

type MyCommand struct {
	Params        *MyCommandParams
	Results       *MyCommandResultFilesPath
	ResultsWriter common.ResultsWriterInterface
	CliWrappers   MyCommandCliWrappers
}

func NewMyCommand(cmd *cobra.Command) (*MyCommand, error) {
	myCommand := &MyCommand{}

	params := &MyCommandParams{}
	if err := common.ParseParameters(cmd, MyCommandParamsConfig, params); err != nil {
		return nil, err
	}
	myCommand.Params = params

	results := &MyCommandResultFilesPath{}
	if err := common.ReadResultFilesPath(results); err != nil {
		return nil, err
	}
	myCommand.Results = results
	myCommand.ResultsWriter = common.NewResultsWriter(myCommand.Params.Verbose)

	if err := myCommand.initCliWrappers(); err != nil {
		return nil, err
	}

	return myCommand, nil
}

func (c *MyCommand) Run() error {
	if err := c.validateParams(); err != nil {
		return err
	}

	// Logic here

	location, err := c.CliWrappers.SomeCli.DoSomething(c.Params.Url)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	if err := c.ResultsWriter.WriteResultString(location, c.Results.Location); err != nil {
		return err
	}

	return nil
}
```

## Tekton

The cli does not depend on Tekton, but made with it in mind, so can be easily used in Tekton tasks.
A simplified example:
```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: apply-tags
spec:
  description: Applies additional tags to the built image.
  params:
  - name: IMAGE_URL
    description: Image repository and tag reference of the the built image.
    type: string
  - name: IMAGE_DIGEST
    description: Image digest of the built image.
    type: string
  - name: ADDITIONAL_TAGS
    description: Additional tags that will be applied to the image in the registry.
    type: array
    default: []
  steps:
    - name: apply-additional-tags
      image: quai.io/org/tekton-catalog/apply-tags:latest
      command: ["konflux-build-cli", "image", "apply-tags"]
      args:
        - --image-url
        - $(params.IMAGE_URL)
        - --digest
        - $(params.IMAGE_DIGEST)
        - --tags
        - $(params.ADDITIONAL_TAGS[*])
```
