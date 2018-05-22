package tool

type Getter interface {
	FetchLatest() (Version, error)
	FetchVersion(version Version) error

	CheckoutTo(version Version, path string) error
}
