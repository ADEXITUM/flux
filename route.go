package flux

type Route struct {
	path             string
	method           string
	handler          func(*Context)
	middlewaresChain []HandlerFunc
	requireAuth      bool
	allowedRoles     []int8
}

// Guard route with router.AuthFunc
func (r *Route) Auth() *Route {
	r.requireAuth = true

	return r
}

// Guard route with roles
//
// Sets requireAuth = true internally
func (r *Route) Roles(roles ...int8) {
	r.requireAuth = true
	r.allowedRoles = append(r.allowedRoles, roles...)
}

// Attach a middleware chain to route
func (r *Route) Use(middlewares ...HandlerFunc) {
	r.middlewaresChain = append(r.middlewaresChain, middlewares...)
}
