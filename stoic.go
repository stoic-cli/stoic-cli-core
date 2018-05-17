package stoic

type Stoic interface {
	Root() string
	ConfigFile() string

	RunTool(name string, args []string) error
}
