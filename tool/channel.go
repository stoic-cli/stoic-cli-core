package tool

const DefaultChannel = Channel("")

// Channel is an opaque identifier to represent a specific tool distribution
// channel. This defines the granularity at which upstream versions are tracked
// and updated.
//
// Channel can be used, for instance, to distinguish between stable and
// pre-release channels, or between branches in a git repository.
type Channel string
