package base

import (
	"context"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mad-app/base/blog"
	"go.uber.org/zap"
)

type JSONRequester interface {
	JSONRequest()
}

type QueryOrFormRequester interface {
	QueryOrFormRequest()
}

type PathRequester interface {
	PathRequest()
}

type HeaderRequester interface {
	HeaderRequest()
}

type GroupRouter interface {
	Route(rg *RouteGroup)
}

type Defaulter interface {
	Default()
}

type PostValidator interface {
	PostValidator() StatefulError
}

type APIContext struct {
	context.Context

	method  string
	url     string
	handler string
}

func (r *APIContext) Sweeten() []zap.Field {
	return []zap.Field{
		zap.Any("method", r.method),
		zap.Any("url", r.url),
		zap.Any("handler", r.handler),
	}
}

type RouteGroup struct {
	rg *gin.RouterGroup
	gr GroupRouter
}

func (r *RouteGroup) Group(path string, gr GroupRouter) *RouteGroup {
	ret := &RouteGroup{
		rg: r.rg.Group(path),
		gr: gr,
	}

	gr.Route(ret)

	return ret
}

func (r *RouteGroup) Use(handler interface{}) error {
	handleType := reflect.TypeOf(handler)

	if handleType.Kind() != reflect.Func {
		panic("invalid handler type; handler must be a function")
	}
	if handleType.NumOut() != 1 {
		panic("handler must returns two outputs; owned response and base error.")
	}

	if !handleType.Out(0).Implements(reflect.TypeOf((*StatefulError)(nil)).Elem()) {
		panic("the first return parameter must implement functions of  `base.StatefulError`.")
	}

	handle := reflect.ValueOf(handler)
	if handle.IsZero() || handle.IsNil() || !handle.IsValid() {
		panic("handler must be set before registration; could not be nil")
	}

	inputTypes := []reflect.Type{}
	inNum := handleType.NumIn()
	for i := 0; i < inNum; i++ {
		typ := handleType.In(i)
		if typ.Kind() == reflect.Ptr && i != 0 {
			panic("all request parameters except the first context have to be non-pointer variables")
		} else if typ.Kind() != reflect.Ptr && i == 0 {
			panic("the first parameter of handler must be `*base.Context`")
		}

		inputTypes = append(inputTypes, typ)
	}

	r.rg.Use(func(c *gin.Context) {
		var err StatefulError
		ctx := NewContext(c)

		defer r.success(ctx, err)

		outs := handle.Call(r.parseRequest(ctx, inNum, inputTypes))
		if ei := outs[0].Interface(); ei != nil {
			err = ei.(StatefulError)
		}
	})

	return nil

}

func (r *RouteGroup) POST(url string, handler interface{}, opts ...interface{}) error {
	return r.createHandler(r.rg.POST, url, handler)
}

func (r *RouteGroup) GET(url string, handler interface{}, opts ...interface{}) error {
	return r.createHandler(r.rg.GET, url, handler)
}

func (r *RouteGroup) DELETE(url string, handler interface{}, opts ...interface{}) error {
	return r.createHandler(r.rg.DELETE, url, handler)
}

func (r *RouteGroup) PUT(url string, handler interface{}, opts ...interface{}) error {
	return r.createHandler(r.rg.PUT, url, handler)
}

func (r *RouteGroup) PATCH(url string, handler interface{}, opts ...interface{}) error {
	return r.createHandler(r.rg.PATCH, url, handler)
}

func (r *RouteGroup) HEAD(url string, handler interface{}, opts ...interface{}) error {
	return r.createHandler(r.rg.HEAD, url, handler)
}

func (r *RouteGroup) createHandler(method func(string, ...gin.HandlerFunc) gin.IRoutes, url string, handler interface{}) error {
	handle := reflect.ValueOf(handler)
	mns := strings.Split(runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name(), ".")
	mn := strings.Trim(mns[len(mns)-1], "-fm")
	hn := strings.Trim(runtime.FuncForPC(handle.Pointer()).Name(), "-fm")
	ac := &APIContext{
		method:  mn,
		url:     r.rg.BasePath() + "/" + url,
		handler: hn,
	}
	blog.Infof(ac, "Registered: %s %s%s", mn, r.rg.BasePath(), url)
	handleType := reflect.TypeOf(handler)

	if handleType.Kind() != reflect.Func {
		panic("invalid handler type; handler must be a function")
	}
	if handleType.NumOut() != 2 {
		panic("handler must returns two outputs; owned response and base error.")
	}

	if !handleType.Out(1).Implements(reflect.TypeOf((*StatefulError)(nil)).Elem()) {
		panic("the second return parameter must implement functions of  `base.StatefulError`.")
	}

	if handle.IsZero() || handle.IsNil() || !handle.IsValid() {
		panic("handler must be set before registration; could not be nil")
	}

	inputTypes := []reflect.Type{}
	inNum := handleType.NumIn()
	for i := 0; i < inNum; i++ {
		typ := handleType.In(i)
		if typ.Kind() == reflect.Ptr && i != 0 {
			panic("all request parameters except the first context have to be non-pointer variables")
		} else if typ.Kind() != reflect.Ptr && i == 0 {
			panic("the first parameter of handler must be `*base.Context`")
		}

		inputTypes = append(inputTypes, typ)
	}

	method(url, func(c *gin.Context) {
		var result interface{}
		var err StatefulError
		ctx := NewContext(c)

		defer func() {
			if r.success(ctx, err) {
				c.JSON(http.StatusOK, result)
			}
		}()

		outs := handle.Call(r.parseRequest(ctx, inNum, inputTypes))
		result = outs[0].Interface()
		if ei := outs[1].Interface(); ei != nil {
			err = ei.(StatefulError)
		}
	})

	return nil
}

func (r *RouteGroup) parseRequest(ctx *Context, inNum int, inputTypes []reflect.Type) []reflect.Value {
	c := ctx.Context
	inputs := make([]reflect.Value, inNum)
	for i := 1; i < inNum; i++ {
		input := reflect.New(inputTypes[i])
		iface := input.Interface()

		if _, ok := iface.(JSONRequester); ok {
			if err := c.ShouldBindJSON(iface); err != nil {
				panic(ErrParseRequest.SetDetails(err.Error()))
			}
		}

		blog.Debugf(ctx, "req%d: %+v\n", i, iface)
		if _, ok := iface.(QueryOrFormRequester); ok {
			if err := c.ShouldBindQuery(iface); err != nil {
				panic(ErrParseRequest.SetDetails(err.Error()))
			}
		}
		if _, ok := iface.(PathRequester); ok {
			if err := c.ShouldBindUri(iface); err != nil {
				panic(ErrParseRequest.SetDetails(err.Error()))
			}
		}
		if _, ok := iface.(HeaderRequester); ok {
			if err := c.ShouldBindHeader(iface); err != nil {
				panic(ErrParseRequest.SetDetails(err.Error()))
			}
		}

		inputs[i] = input.Elem()
		blog.Debugf(ctx, "req%d: %+v\n", i, iface)
	}

	inputs[0] = reflect.ValueOf(ctx)

	return inputs
}
func (r *RouteGroup) success(ctx *Context, err StatefulError) bool {
	c := ctx.Context
	if e := recover(); e != nil {
		err = ErrUnknown
		if v, ok := e.(StatefulError); ok {
			err = v
		}
	}

	if err != nil {
		blog.Error(ctx, err.Error())
		c.JSON(err.Status(), err)
		c.Abort()
		return false
	}

	return true
}
