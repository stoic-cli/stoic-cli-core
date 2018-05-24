package engine

import (
	git "github.com/stoic-cli/stoic-cli-core/get-git"
	github "github.com/stoic-cli/stoic-cli-core/get-github-release"
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
}
