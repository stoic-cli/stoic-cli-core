package tool

// Checkout represents a materialization of a specific version of a tool in the
// filesystem.
type Checkout interface {
	// Path returns the filesystem path for the checkout.
	Path() string

	// Version returns the version of the tool in the checkout.
	Version() Version
}
