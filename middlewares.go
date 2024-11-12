package flux

import (
	"net/http"
)

// BUG: doesn't front-end doesn't see the change of CORS policy
func allowAllCORS(c *Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	if c.Request.Method == http.MethodOptions {
		c.Writer.WriteHeader(http.StatusOK)
		return
	}

	c.Next()
}
