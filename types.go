package flux

type HandlerFunc func(*Context)
type AuthFunc func(*Context)
type Middleware func(*Context)

// H is a shortcut for map[string]any
type H map[string]any
