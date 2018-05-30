package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/stoic-cli/stoic-cli-core"
	shell "github.com/stoic-cli/stoic-cli-core/run-shell"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

func NewRunner(stoic stoic.Stoic, tool stoic.Tool) (tool.Runner, error) {
	importPath := tool.Config().Endpoint
	buildEnviron := append(os.Environ(),
		"GOARCH="+runtime.GOARCH,
		"GOOS="+runtime.GOOS,
		// TODO: set GOARM
	)
	binary := filepath.Join("bin", path.Base(importPath))

	if _, ok := tool.Config().Runner.Options["command"]; !ok {
		tool.Config().Runner.Options["command"] = binary
	}

	shellRunner, err := shell.NewRunner(stoic, tool)
	if err != nil {
		return nil, err
	}
	sr, ok := shellRunner.(shell.Runner)
	if !ok {
		return nil, fmt.Errorf(
			"unable to cast shell runner of type %T to shell.Runner",
			shellRunner)
	}
	return &runner{sr, importPath, buildEnviron, binary}, nil
}

type runner struct {
	ShellRunner  shell.Runner
	ImportPath   string
	BuildEnviron []string
	Binary       string
}

func (r runner) Setup(checkout tool.Checkout) error {
	gopath := checkout.Path()

	build := exec.Command("go", "build", "-o", r.Binary, r.ImportPath)
	build.Dir = gopath
	build.Env = append(r.BuildEnviron, "GOPATH="+gopath)
	build.Stderr = os.Stderr
	build.Stdout = os.Stderr

	if err := build.Run(); err != nil {
		return err
	}

	r.ShellRunner.Options.SetupEnvironment["GOPATH"] = gopath
	return r.ShellRunner.Setup(checkout)
}

func (r runner) Run(checkout tool.Checkout, name string, args []string) error {
	return r.ShellRunner.Run(checkout, name, args)
}
