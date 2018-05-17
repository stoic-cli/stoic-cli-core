package engine

import (
	"github.com/stoic-cli/stoic-cli-core/get-git"
	"github.com/stoic-cli/stoic-cli-core/run-shell"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

const (
	DefaultToolUpdateFrequency = tool.UpdateWeekly

	DefaultToolGetterType = "git"
	DefaultToolRunnerType = "shell"
)

func init() {
	RegisterGetter("git", gitgetter.NewGetter)
	RegisterRunner("shell", shellrunner.NewRunner)
}
