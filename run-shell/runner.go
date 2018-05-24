package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/shlex"
	"github.com/mitchellh/mapstructure"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

func NewRunner(s stoic.Stoic, tool stoic.Tool) (tool.Runner, error) {
	var options Options
	err := mapstructure.Decode(tool.Config().Runner.Options, &options)
	if err != nil {
		return nil, err
	}
	return Runner{s, options}, nil
}

type Options struct {
	Setup            string
	Command          string
	SetupEnvironment map[string]string `mapstructure:"setup-environment"`
	Environment      map[string]string
}

type Runner struct {
	Stoic   stoic.Stoic
	Options Options
}

func (sr Runner) shellCommand(
	checkout tool.Checkout, shellCommand string, environment map[string]string) (*exec.Cmd, error) {
	if shellCommand == "" {
		return nil, nil
	}

	parameters := sr.Stoic.Parameters()
	parameters["Checkout"] = checkout.Path()
	parameters["Version"] = string(checkout.Version())

	cmdAndArgs, err := shlex.Split(shellCommand)
	if err != nil {
		return nil, err
	}

	for i := range cmdAndArgs {
		tmpl, err := template.New("").Parse(cmdAndArgs[i])
		if err != nil {
			return nil, err
		}

		var builder strings.Builder
		err = tmpl.Execute(&builder, parameters)
		if err != nil {
			return nil, err
		}

		cmdAndArgs[i] = builder.String()
	}

	command := cmdAndArgs[0]
	if !filepath.IsAbs(command) {
		command = filepath.Join(checkout.Path(), command)
	}

	cmd := exec.Command(command)
	cmd.Args = cmdAndArgs
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

func (sr Runner) Setup(checkout tool.Checkout) error {
	cmd, err := sr.shellCommand(checkout, sr.Options.Setup, sr.Options.SetupEnvironment)
	if err != nil {
		return err
	}
	if cmd == nil {
		return nil
	}

	cmd.Dir = checkout.Path()
	return cmd.Run()
}

func (sr Runner) Run(checkout tool.Checkout, name string, args []string) error {
	cmd, err := sr.shellCommand(checkout, sr.Options.Command, sr.Options.Environment)
	if err != nil {
		return err
	}
	if cmd == nil {
		return nil
	}

	cmd.Args = append(cmd.Args, args...)
	return cmd.Run()
}
