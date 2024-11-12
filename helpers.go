package flux

import (
	"net/http"
	"reflect"
	"regexp"
)

// Get plain Bearer token
func (r *Engine) extractToken(req *http.Request) string {
	token := req.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		return token[7:]
	}

	return token
}

// Check is given path is defined
func (r *Engine) isPathExist(path string) bool {
	for _, route := range r.routes {
		if route.path == path {
			return true
		}
	}

	for _, group := range r.routeGroups {
		for _, route := range group.routes {
			if route.path == path {
				return true
			}
		}
	}

	return false
}

// Check if email has corrent format
func isValidEmail(email string) bool {
	var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// Check if value is empty
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan, reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}
