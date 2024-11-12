package flux

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Client struct {
	UserID       int64
	RoleID       int8
	SessionToken string
}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request

	FullPath string

	// Client data (supposed to be filled in Engine.AuthFunc or in middlewares)
	Client *Client

	// Body is already parsed req.Body data into []byte
	Body []byte

	// MultipartForm is already parsed multipart/formdata
	MultipartForm *multipart.Form

	// Mutex guards CustomData keys
	mu sync.RWMutex
	// CustomData is custom data map (you may sue it in your middlewares)
	CustomData map[string]any

	HandlersChain []HandlerFunc // Stack of middlewares to call
	Index         int           // Keeps track of which middleware to call next
	Aborted       bool
	StatusCode    int
	AllowedRoles  []int8

	CreatedAt time.Time // May be used to track current request duration
}

// Execute next handler in a chain
func (c *Context) Next() {
	if c.Index < len(c.HandlersChain) {
		if c.Aborted {
			return
		}
		handler := c.HandlersChain[c.Index]
		c.Index++
		handler(c)
	}
}

// Stop all pending requests from execution
func (c *Context) Abort() {
	c.Aborted = true
}

// Stop all pending requests from execution
func (c *Context) AbortWithStatus(status int) {
	c.Abort()
	c.Status(status)
}

// Stop all pending requests from execution
func (c *Context) AbortWithStatusJSON(status int, json any) {
	c.Abort()
	c.JSON(status, json)
}

// Set custom values into CustomData map
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.CustomData == nil {
		c.CustomData = make(map[string]any)
	}
	c.CustomData[key] = value
}

// Get custom value from CustomData map by key
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.CustomData[key]
	return
}

// Sets the HTTP response code
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// Serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json"
func (c *Context) JSON(code int, data any) error {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Status(code)
	return json.NewEncoder(c.Writer).Encode(data)
}

// Unmarshal context.Body into given struct
func (c *Context) BindJSON(obj any) error {
	return json.Unmarshal(c.Body, obj)
}

// BindJSON with check for custom struct bindings
func (c *Context) ShouldBindJSON(obj interface{}) error {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("object must be a pointer")
	}

	if err := c.BindJSON(obj); err != nil {
		return err
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference the pointer to access the struct
	}

	return validateStruct(val)
}

func validateStruct(val reflect.Value) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("binding")

		if tag == "" {
			continue
		}

		tags := strings.Split(tag, ",")
		for _, rule := range tags {
			switch rule {
			case "required":
				if isEmpty(field) {
					return errors.New(typ.Field(i).Name + " is required")
				}
			case "email":
				if !isValidEmail(field.String()) {
					return errors.New("invalid email format")
				}
			}
		}
	}
	return nil
}

// Parse req.Body into context.Body []byte
func (c *Context) parseBody() error {
	if c.Body != nil {
		return nil
	}

	if ct := c.Request.Header.Get("Content-Type"); ct != "" {
		mt, _, err := mime.ParseMediaType(c.Request.Header.Get("Content-Type"))
		if err != nil {
			return err
		}

		if mt == "multipart/form-data" {
			if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
				return err
			}

			c.MultipartForm = c.Request.MultipartForm
			return nil
		}
	}

	if err := c.Request.ParseForm(); err != nil {
		return err
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}

	c.Body = body
	return nil
}
