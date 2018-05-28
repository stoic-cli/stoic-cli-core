package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/stoic-cli/stoic-cli-core"
	shell "github.com/stoic-cli/stoic-cli-core/run-shell"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

func newRunner(s stoic.Stoic, t stoic.Tool) (tool.Runner, error) {
	var options PythonOptions
	err := mapstructure.Decode(t.Config().Runner.Options, &options)
	if err != nil {
		return nil, err
	}

	python, ok := t.Config().Runner.Options["python"].(string)
	if !ok {
		return nil, errors.New("python executable not specified")
	}
	absolutePython, err := lookupAbsolutePath(python)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to find python executable %s: %v", python, err)
	}

	if options.EntryPoint != "" {
		if options.Command != "" {
			return nil, fmt.Errorf(
				"entry-point and command cannot both be specified simultaneously\n"+
					"\tentry-point: %v\n"+
					"\tcommand: %v\n",
				options.EntryPoint, options.Command)
		}

		var equals, colon int
		equals = strings.IndexByte(options.EntryPoint, '=')

		if equals == -1 {
			colon = strings.IndexByte(options.EntryPoint, ':')
		} else {
			colon = strings.IndexByte(options.EntryPoint[equals+1:], ':')
		}

		if colon == -1 {
			return nil, fmt.Errorf(
				"expected entry-point of the form 'module[.submodule]:function', got: %v",
				options.EntryPoint)
		}

		if equals == -1 {
			options.EntryPoint = t.Name() + "=" + options.EntryPoint
		}
	}

	root := filepath.Join(s.Root(), "python")

	shellRunner, err := shell.NewRunner(s, t)
	if err != nil {
		return nil, err
	}
	sr, ok := shellRunner.(shell.Runner)
	if !ok {
		return nil, fmt.Errorf(
			"unable to cast shell runner of type %T to shell.Runner",
			shellRunner)
	}
	return runner{sr, root, absolutePython, options}, nil
}

func NewPythonRunner(s stoic.Stoic, t stoic.Tool) (tool.Runner, error) {
	options := t.Config().Runner.Options
	if _, ok := options["python"]; !ok {
		options["python"] = "python"
	}
	return newRunner(s, t)
}

func NewPython2Runner(s stoic.Stoic, t stoic.Tool) (tool.Runner, error) {
	t.Config().Runner.Options["python"] = "python2.7"
	return newRunner(s, t)
}

func NewPython3Runner(s stoic.Stoic, t stoic.Tool) (tool.Runner, error) {
	t.Config().Runner.Options["python"] = "python3"
	return newRunner(s, t)
}

type PythonOptions struct {
	Python           string `mapstructure:",omitempty"`
	RequirementsFile string `mapstructure:"requirements,omitempty"`
	ModulePath       string `mapstructure:"module-path,omitempty"`
	EntryPoint       string `mapstructure:"entry-point,omitempty"`
	Command          string `mapstructure:",omitempty"`
}

type runner struct {
	ShellRunner shell.Runner

	Root   string
	Python string

	PythonOptions
}

func (r runner) Setup(checkout tool.Checkout) error {
	pe, err := setupPythonEnv(r.Root, r.Python, r.ShellRunner.Stoic.Cache())
	if err != nil {
		return err
	}

	requirementsFile := filepath.Join(checkout.Path(), r.RequirementsFile)
	ve, err := setupVirtualEnv(pe, requirementsFile)
	if err != nil {
		return err
	}

	setupCommand := exec.Command(ve.Python(),
		"-m", "pip", "check", "--quiet")

	setupCommand.Stdout = os.Stderr
	setupCommand.Stderr = os.Stderr
	setupCommand.Env = ve.Environ()

	err = setupCommand.Run()
	if err != nil {
		return err
	}

	r.ShellRunner.Options.SetupEnvironment["PIP_CACHE_DIR"] = pe.PipCache()
	r.ShellRunner.Options.SetupEnvironment["PYTHONPATH"] = pe.SitePackages()
	return r.ShellRunner.Setup(checkout)
}

func (r runner) Run(checkout tool.Checkout, name string, args []string) error {
	pe, err := setupPythonEnv(r.Root, r.Python, r.ShellRunner.Stoic.Cache())
	if err != nil {
		return err
	}

	requirementsFile := filepath.Join(checkout.Path(), r.RequirementsFile)
	ve, err := setupVirtualEnv(pe, requirementsFile)
	if err != nil {
		return err
	}

	if r.EntryPoint != "" {
		r.ShellRunner.Options.Parameters["EntryPoint"] = r.EntryPoint
		r.ShellRunner.Options.Command = "{{.Python}} -c \"" +
			"import functools as f, importlib as i, sys as s;" +

			"ep = s.argv[1];" +

			"n, _, ep = ep.partition('=');" +
			"mn, _, fn = ep.partition(':');" +

			"e = f.reduce(getattr, fn.split('.'), i.import_module(mn));" +
			"s.argv[0:2] = [n];" +

			"s.exit(e());" +
			"\" {{.EntryPoint}}"
	}

	pythonPath := pe.SitePackages()
	if r.ModulePath != "" {
		modulePath := filepath.Join(checkout.Path(), r.ModulePath)
		pythonPath = modulePath + string(os.PathListSeparator) + pythonPath
	}

	r.ShellRunner.Options.Parameters["Python"] = ve.Python()
	r.ShellRunner.Options.Parameters["VirtualEnv"] = ve.Root()
	r.ShellRunner.Options.Parameters["VirtualEnvScripts"] = ve.Scripts()

	// TODO: Should filter out PYTHONHOME from environment, if set

	r.ShellRunner.Options.Environment["PATH"] = ve.EnvPath()
	r.ShellRunner.Options.Environment["PYTHONPATH"] = pythonPath
	r.ShellRunner.Options.Environment["VIRTUAL_ENV"] = ve.Root()
	return r.ShellRunner.Run(checkout, name, args)
}
