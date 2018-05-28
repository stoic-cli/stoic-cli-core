package engine

import (
	git "github.com/stoic-cli/stoic-cli-core/get-git"
	github "github.com/stoic-cli/stoic-cli-core/get-github-release"
	python "github.com/stoic-cli/stoic-cli-core/run-python"
	shell "github.com/stoic-cli/stoic-cli-core/run-shell"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

const (
	GitGetterType           = "git"
	GithubReleaseGetterType = "github-release"

	ShellRunnerType   = "shell"
	PythonRunnerType  = "python"
	Python2RunnerType = "python2"
	Python3RunnerType = "python3"
)

const (
	DefaultToolUpdateFrequency = tool.UpdateWeekly

	DefaultToolGetterType = GitGetterType
	DefaultToolRunnerType = ShellRunnerType
)

func init() {
	RegisterGetter(GitGetterType, git.NewGetter)
	RegisterGetter(GithubReleaseGetterType, github.NewGetter)

	RegisterRunner(ShellRunnerType, shell.NewRunner)
	RegisterRunner(PythonRunnerType, python.NewPythonRunner)
	RegisterRunner(Python2RunnerType, python.NewPython2Runner)
	RegisterRunner(Python3RunnerType, python.NewPython3Runner)
}
