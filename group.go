package flux

import "net/http"

type RouteGroup struct {
	basePath         string
	routes           []*Route
	groupMiddlewares []HandlerFunc
}

// Attach a middleware chain to group (will be applied to each route that belongs to a group)
func (rg *RouteGroup) Use(middlewares ...HandlerFunc) *RouteGroup {
	rg.groupMiddlewares = append(rg.groupMiddlewares, middlewares...)
	return rg
}

// POST request registration
func (rg *RouteGroup) POST(path string, handler func(*Context)) *Route {
	return rg.register(path, http.MethodPost, handler)
}

// GET request registration
func (rg *RouteGroup) GET(path string, handler func(*Context)) *Route {
	return rg.register(path, http.MethodGet, handler)
}

// DELETE request registration
func (rg *RouteGroup) DELETE(path string, handler func(*Context)) *Route {
	return rg.register(path, http.MethodDelete, handler)
}

// PUT request registration
func (rg *RouteGroup) PUT(path string, handler func(*Context)) *Route {
	return rg.register(path, http.MethodPut, handler)
}

// PATCH request registration
func (rg *RouteGroup) PATCH(path string, handler func(*Context)) *Route {
	return rg.register(path, http.MethodPatch, handler)
}

func (rg *RouteGroup) register(path, method string, handler func(*Context)) *Route {
	route := &Route{
		path:             path,
		method:           method,
		handler:          handler,
		middlewaresChain: rg.groupMiddlewares,
		requireAuth:      false,
		allowedRoles:     nil,
	}

	rg.routes = append(rg.routes, route)

	return route
}
