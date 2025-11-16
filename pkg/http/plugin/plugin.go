package plugin

// ServerInterface defines the interface that plugins can use to interact with the server
// This avoids import cycles
type ServerInterface interface {
	App() any // Returns the underlying Fiber app
	Use(middleware ...any)
	GET(path string, handler any)
	POST(path string, handler any)
	PUT(path string, handler any)
	DELETE(path string, handler any)
	PATCH(path string, handler any)
	Group(prefix string, fn any)
}

// Plugin defines the interface for server plugins
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Install installs the plugin on the server
	Install(server ServerInterface) error
}
