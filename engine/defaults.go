package engine

import (
	git "github.com/stoic-cli/stoic-cli-core/get-git"
	github "github.com/stoic-cli/stoic-cli-core/get-github-release"
	goget "github.com/stoic-cli/stoic-cli-core/get-go-get"
	gobuild "github.com/stoic-cli/stoic-cli-core/run-go-build"
	python "github.com/stoic-cli/stoic-cli-core/run-python"
	shell "github.com/stoic-cli/stoic-cli-core/run-shell"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

const (
	GitGetterType           = "git"
	GithubReleaseGetterType = "github-release"
	GoGetGetterType         = "go-get"

	GoBuildRunnerType = "go-build"
	Python2RunnerType = "python2"
	Python3RunnerType = "python3"
	PythonRunnerType  = "python"
	ShellRunnerType   = "shell"
)

const (
	DefaultToolUpdateFrequency = tool.UpdateWeekly

	DefaultToolGetterType = GitGetterType
	DefaultToolRunnerType = ShellRunnerType
)

func init() {
	RegisterGetter(GitGetterType, git.NewGetter)
	RegisterGetter(GithubReleaseGetterType, github.NewGetter)
	RegisterGetter(GoGetGetterType, goget.NewGetter)

	RegisterRunner(ShellRunnerType, shell.NewRunner)
	RegisterRunner(PythonRunnerType, python.NewPythonRunner)
	RegisterRunner(Python2RunnerType, python.NewPython2Runner)
	RegisterRunner(Python3RunnerType, python.NewPython3Runner)
	RegisterRunner(GoBuildRunnerType, gobuild.NewRunner)
}
