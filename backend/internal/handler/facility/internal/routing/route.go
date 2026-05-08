package routing

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Definition struct {
	Method     string
	Path       string
	Permission string
	Handler    gin.HandlerFunc
}

func Get(path string, permission string, handler gin.HandlerFunc) Definition {
	return route(http.MethodGet, path, permission, handler)
}

func Post(path string, permission string, handler gin.HandlerFunc) Definition {
	return route(http.MethodPost, path, permission, handler)
}

func Put(path string, permission string, handler gin.HandlerFunc) Definition {
	return route(http.MethodPut, path, permission, handler)
}

func Patch(path string, permission string, handler gin.HandlerFunc) Definition {
	return route(http.MethodPatch, path, permission, handler)
}

func Delete(path string, permission string, handler gin.HandlerFunc) Definition {
	return route(http.MethodDelete, path, permission, handler)
}

func route(method string, path string, permission string, handler gin.HandlerFunc) Definition {
	return Definition{
		Method:     method,
		Path:       path,
		Permission: permission,
		Handler:    handler,
	}
}
