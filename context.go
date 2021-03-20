package base

import (
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ContextFields = "fields"
	ContextUser   = "user-id"
)

type Context struct {
	*gin.Context
	fields []zap.Field
	mu     sync.RWMutex
}

func NewContext(ctx *gin.Context) *Context {
	fields := []zap.Field{}

	if v, ok := ctx.Get(ContextFields); ok {
		fields = v.([]zap.Field)
	}

	return &Context{
		Context: ctx,
		fields:  fields,
		mu:      sync.RWMutex{},
	}
}

func (r *Context) SetField(key string, val interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.fields = append(r.fields, zap.Any(key, val))
	r.Set(ContextFields, r.fields)
	r.Set(key, val)
}

func (r *Context) Sweeten() []zap.Field {
	return r.fields
}

func (r *Context) User() string {
	ret := ""
	if v, ok := r.Get(ContextUser); ok {
		ret = v.(string)
	}

	return ret
}
