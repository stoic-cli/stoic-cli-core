package shellrunner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/shlex"
	"github.com/mitchellh/mapstructure"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/format"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

func NewRunner(s stoic.Stoic, config format.ToolConfig) (tool.Runner, error) {
	var options shellRunnerOptions
	err := mapstructure.Decode(config.Runner.Options, &options)
	if err != nil {
		return nil, err
	}
	return shellRunner{options}, nil
}

type shellRunnerOptions struct {
	Setup            string
	Command          string
	SetupEnvironment map[string]string `mapstructure:"setup-environment"`
	Environment      map[string]string
}

type shellRunner struct {
	options shellRunnerOptions
}

func (sr shellRunner) shellCommand(
	checkout tool.Checkout, shellCommand string, environment map[string]string) (*exec.Cmd, error) {
	if shellCommand == "" {
		return nil, nil
	}

	splitCommand, err := shlex.Split(shellCommand)
	if err != nil {
		return nil, err
	}

	command := splitCommand[0]
	if !filepath.IsAbs(command) {
		command = filepath.Join(checkout.Path(), command)
	}

	cmd := exec.Command(command)
	cmd.Args = splitCommand
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if len(environment) != 0 {
		cmd.Env = os.Environ()
		for k, v := range environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return cmd, nil
}

func (sr shellRunner) Setup(checkout tool.Checkout) error {
	cmd, err := sr.shellCommand(checkout, sr.options.Setup, sr.options.SetupEnvironment)
	if err != nil {
		return err
	}
	return cmd.Run()
}

func (sr shellRunner) Run(checkout tool.Checkout, name string, args []string) error {
	cmd, err := sr.shellCommand(checkout, sr.options.Command, sr.options.Environment)
	if err != nil {
		return err
	}

	cmd.Args = append(cmd.Args, args...)
	return cmd.Run()
}
