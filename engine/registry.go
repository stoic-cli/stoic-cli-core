package engine

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

const (
	toolGetterRegistry = registryName("getter")
	toolRunnerRegistry = registryName("runner")
)

var (
	registries []registryRecord
)

type registryName string

type registryRecord struct {
	Name    registryName
	Records map[string]interface{}
}

type GetterCtor func(stoic.Stoic, stoic.Tool) (tool.Getter, error)
type RunnerCtor func(stoic.Stoic, stoic.Tool) (tool.Runner, error)

type toolGetterRecord struct{ Ctor GetterCtor }
type toolRunnerRecord struct{ Ctor RunnerCtor }

func addToRegistry(registryName registryName, name string, record interface{}) {
	for i := range registries {
		if registries[i].Name == registryName {
			if _, ok := registries[i].Records[name]; ok {
				panic(fmt.Sprintf("element already registered in %s registry", name))
			}
			registries[i].Records[name] = record
			return
		}
	}

	registries = append(registries, registryRecord{
		Name:    registryName,
		Records: map[string]interface{}{name: record},
	})
}

func findInRegistry(registryName registryName, name string) interface{} {
	for i := range registries {
		if registries[i].Name == registryName {
			return registries[i].Records[name]
		}
	}
	return nil
}

func RegisterGetter(name string, ctor GetterCtor) {
	addToRegistry(toolGetterRegistry, name, toolGetterRecord{ctor})
}
func RegisterRunner(name string, ctor RunnerCtor) {
	addToRegistry(toolRunnerRegistry, name, toolRunnerRecord{ctor})
}

func (e *engine) getterFor(tool stoic.Tool) (tool.Getter, error) {
	typ := tool.Config().Getter.Type
	getter := findInRegistry(toolGetterRegistry, typ)
	if getter != nil {
		return getter.(toolGetterRecord).Ctor(e, tool)
	}
	return nil, errors.Errorf("unknown getter type, '%v'", typ)
}

func (e *engine) runnerFor(tool stoic.Tool) (tool.Runner, error) {
	typ := tool.Config().Runner.Type
	runner := findInRegistry(toolRunnerRegistry, typ)
	if runner != nil {
		return runner.(toolRunnerRecord).Ctor(e, tool)
	}
	return nil, errors.Errorf("unknown runner type, '%v'", typ)
}
