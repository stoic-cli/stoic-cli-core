package engine

import (
	git "github.com/stoic-cli/stoic-cli-core/get-git"
	github "github.com/stoic-cli/stoic-cli-core/get-github-release"
	python "github.com/stoic-cli/stoic-cli-core/run-python"
	shell "github.com/stoic-cli/stoic-cli-core/run-shell"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

const (
	DefaultToolUpdateFrequency = tool.UpdateWeekly

	DefaultToolGetterType = "git"
	DefaultToolRunnerType = "shell"
)

func init() {
	RegisterGetter("git", git.NewGetter)
	RegisterGetter("github-release", github.NewGetter)

	RegisterRunner("shell", shell.NewRunner)
	RegisterRunner("python", python.NewPythonRunner)
	RegisterRunner("python2", python.NewPython2Runner)
	RegisterRunner("python3", python.NewPython3Runner)
}
