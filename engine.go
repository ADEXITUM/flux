package flux

import (
	"io"
	"log"
	"net/http"
	"time"
)

type Engine struct {
	routes            []*Route
	routeGroups       []*RouteGroup
	authFunc          *AuthFunc
	globalMiddlewares []HandlerFunc
}

// New creates a new router instance
func New() *Engine {
	return &Engine{}
}

// UseAuth attaches an auth function that will be invoked after the middleware chain
//
// It updates the Client struct of Context
func (e *Engine) UseAuth(authFunc AuthFunc) {
	e.authFunc = &authFunc
}

// AllowAllCORS attaches a CORS middleware allowing all origins
func (e *Engine) AllowAllCORS() {
	e.globalMiddlewares = append([]HandlerFunc{allowAllCORS}, e.globalMiddlewares...)
}

// Use attaches a chain of global middlewares that will execute before each route
func (e *Engine) Use(middlewares ...HandlerFunc) {
	e.globalMiddlewares = append(e.globalMiddlewares, middlewares...)
}

// register is a common function to register a route
func (e *Engine) register(path, method string, handler HandlerFunc) *Route {
	route := &Route{
		path:             path,
		method:           method,
		handler:          handler,
		middlewaresChain: nil,
		requireAuth:      false,
		allowedRoles:     nil,
	}
	e.routes = append(e.routes, route)
	return route
}

// Group creates a group of routes with the same prefix
//
// (supposed to use path without /)
func (e *Engine) Group(path string) *RouteGroup {
	group := &RouteGroup{
		basePath:         "/" + path,
		groupMiddlewares: e.globalMiddlewares,
	}
	e.routeGroups = append(e.routeGroups, group)
	return group
}

// POST registers a POST route
func (e *Engine) POST(path string, handler HandlerFunc) *Route {
	return e.register(path, http.MethodPost, handler)
}

// GET registers a GET route
func (e *Engine) GET(path string, handler HandlerFunc) *Route {
	return e.register(path, http.MethodGet, handler)
}

// DELETE registers a DELETE route
func (e *Engine) DELETE(path string, handler HandlerFunc) *Route {
	return e.register(path, http.MethodDelete, handler)
}

// PUT registers a PUT route
func (e *Engine) PUT(path string, handler HandlerFunc) *Route {
	return e.register(path, http.MethodPut, handler)
}

// PATCH registers a PATCH route
func (e *Engine) PATCH(path string, handler HandlerFunc) *Route {
	return e.register(path, http.MethodPatch, handler)
}

func (e *Engine) handle(route *Route) {
	http.HandleFunc(route.path, func(wr http.ResponseWriter, req *http.Request) {
		if !e.isPathExist(req.URL.Path) {
			http.NotFound(wr, req)
			return
		}

		if req.Method != route.method {
			wr.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		defer e.recoverPanic(req)

		c := e.newContext(wr, req, route)
		if route.requireAuth && e.authFunc != nil {
			(*e.authFunc)(c)
		}

		if err := c.parseBody(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, H{"error": "couldn't parse request body"})
		}

		c.Next()
	})
}

// newContext initializes a new Context for the given request and response
func (e *Engine) newContext(wr http.ResponseWriter, req *http.Request, route *Route) *Context {
	return &Context{
		Request:       req,
		Writer:        wr,
		FullPath:      route.path,
		Client:        &Client{SessionToken: e.extractToken(req)},
		AllowedRoles:  route.allowedRoles,
		HandlersChain: route.middlewaresChain,
		StatusCode:    http.StatusOK,
		Index:         0,
		CreatedAt:     time.Now(),
	}
}

// recoverPanic handles recovery from panics during request processing
func (e *Engine) recoverPanic(req *http.Request) {
	if err := recover(); err != nil {
		body, _ := io.ReadAll(req.Body)
		log.Printf("[PANIC RECOVERED]:\n%s\nEndpoint: %s\nBody: %s\n\n", err, req.URL.Path, string(body))
	}
}

// Apply registers all defined routes with their handlers
func (e *Engine) Apply() {
	for _, route := range e.routes {
		route.middlewaresChain = append(route.middlewaresChain, route.handler)
		e.handle(route)
	}

	for _, group := range e.routeGroups {
		for _, route := range group.routes {
			route.middlewaresChain = append(route.middlewaresChain, route.handler)
			route.path = group.basePath + route.path
			e.handle(route)
		}
	}
}
