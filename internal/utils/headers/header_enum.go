package headers

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// ContextHeaders is the key used to store and retrieve Header from context
	ContextHeaders ContextKey = "headers"
)
