package hooks

// HookType represents the type of lifecycle hook
type HookType string

const (
	// BeforeStart is called before the server starts
	BeforeStart HookType = "before_start"

	// AfterStart is called after the server successfully starts
	AfterStart HookType = "after_start"

	// BeforeShutdown is called before the server begins shutdown
	BeforeShutdown HookType = "before_shutdown"

	// AfterShutdown is called after the server fully shuts down
	AfterShutdown HookType = "after_shutdown"
)

// HookFunc is a function that can be registered as a lifecycle hook
type HookFunc func() error

