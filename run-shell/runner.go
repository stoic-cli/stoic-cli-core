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

	if options.SetupEnvironment == nil {
		options.SetupEnvironment = map[string]string{}
	}
	if options.SetupParameters == nil {
		options.SetupParameters = map[string]interface{}{}
	}
	if options.Environment == nil {
		options.Environment = map[string]string{}
	}
	if options.Parameters == nil {
		options.Parameters = map[string]interface{}{}
	}

	return Runner{s, options}, nil
}

type Options struct {
	Setup            string
	SetupEnvironment map[string]string      `mapstructure:"setup-environment"`
	SetupParameters  map[string]interface{} `mapstructure:"setup-parameters"`

	Command     string
	Environment map[string]string
	Parameters  map[string]interface{}
}

type Runner struct {
	Stoic   stoic.Stoic
	Options Options
}

func expandString(tmplStr string, parameters map[string]interface{}) (string, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return tmplStr, err
	}

	var builder strings.Builder
	err = tmpl.Execute(&builder, parameters)
	if err != nil {
		return tmplStr, err
	}

	return builder.String(), nil
}

func (sr Runner) shellCommand(
	checkout tool.Checkout,
	shellCommand string,
	environment map[string]string,
	userParameters map[string]interface{},
) (*exec.Cmd, error) {
	if shellCommand == "" {
		return nil, nil
	}

	parameters := sr.Stoic.Parameters()

	parameters["Checkout"] = checkout.Path()
	parameters["Version"] = string(checkout.Version())

	for k, v := range userParameters {
		parameters[k] = v
	}

	cmdAndArgs, err := shlex.Split(shellCommand)
	if err != nil {
		return nil, err
	}

	for i := range cmdAndArgs {
		cmdAndArgs[i], err = expandString(cmdAndArgs[i], parameters)
		if err != nil {
			return nil, err
		}
	}

	command := cmdAndArgs[0]
	if !filepath.IsAbs(command) {
		command = filepath.Join(checkout.Path(), command)
	}

	cmd := exec.Command(command)
	cmd.Args = cmdAndArgs

	if len(environment) != 0 {
		cmd.Env = os.Environ()
		for k, v := range environment {
			v, err = expandString(v, parameters)
			if err != nil {
				return nil, err
			}
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return cmd, nil
}

func (sr Runner) Setup(checkout tool.Checkout) error {
	cmd, err := sr.shellCommand(
		checkout, sr.Options.Setup, sr.Options.SetupEnvironment,
		sr.Options.SetupParameters)

	if err != nil {
		return err
	}
	if cmd == nil {
		return nil
	}

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	cmd.Dir = checkout.Path()
	return cmd.Run()
}

func (sr Runner) Run(checkout tool.Checkout, name string, args []string) error {
	cmd, err := sr.shellCommand(
		checkout, sr.Options.Command, sr.Options.Environment,
		sr.Options.Parameters)

	if err != nil {
		return err
	}
	if cmd == nil {
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Args = append(cmd.Args, args...)
	return cmd.Run()
}
