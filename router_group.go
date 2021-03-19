package base

import (
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
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

		defer func() {
			if e := recover(); e != nil {
				err = ErrUnknown
				log.Println(e)
				if v, ok := e.(StatefulError); ok {
					err = v
				}
				// TODO: logs request and stack
				c.JSON(err.Status(), err)
				c.Abort()
				return
			}

			if err != nil {
				c.JSON(err.Status(), err)
				c.Abort()
				return
			}
		}()

		inputs := make([]reflect.Value, inNum)
		for i := 1; i < inNum; i++ {
			input := reflect.New(inputTypes[i])
			iface := input.Interface()

			if _, ok := iface.(JSONRequester); ok && c.ShouldBindJSON(iface) != nil {
				panic(ErrParseRequest.SetDetails("json"))
			}

			log.Printf("req%d: %+v\n", i, iface)
			if _, ok := iface.(QueryOrFormRequester); ok && c.ShouldBindQuery(iface) != nil {
				panic(ErrParseRequest.SetDetails("query or form"))
			}
			if _, ok := iface.(PathRequester); ok && c.ShouldBindUri(iface) != nil {
				panic(ErrParseRequest.SetDetails("path"))
			}
			if _, ok := iface.(HeaderRequester); ok && c.ShouldBindHeader(iface) != nil {
				panic(ErrParseRequest.SetDetails("header"))
			}

			inputs[i] = input.Elem()
			log.Printf("req%d: %+v\n", i, iface)
		}
		inputs[0] = reflect.ValueOf(&Context{Context: c})

		outs := handle.Call(inputs)
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

	method(url, func(c *gin.Context) {
		var result interface{}
		var err StatefulError

		defer func() {
			if e := recover(); e != nil {
				err = ErrUnknown
				log.Println(e)
				if v, ok := e.(StatefulError); ok {
					err = v
				}
				// TODO: logs request and stack
				c.JSON(err.Status(), err)
				return
			}

			if err != nil {
				c.JSON(err.Status(), err)
				return
			}

			c.JSON(http.StatusOK, result)
		}()

		inputs := make([]reflect.Value, inNum)
		for i := 1; i < inNum; i++ {
			input := reflect.New(inputTypes[i])
			iface := input.Interface()
			if v, ok := iface.(Defaulter); ok {
				v.Default()
			}

			if _, ok := iface.(JSONRequester); ok && c.ShouldBindJSON(iface) != nil {
				panic(ErrParseRequest.SetDetails("json"))
			}

			log.Printf("req%d: %+v\n", i, iface)
			if _, ok := iface.(QueryOrFormRequester); ok && c.ShouldBindQuery(iface) != nil {
				panic(ErrParseRequest.SetDetails("query or form"))
			}
			if _, ok := iface.(PathRequester); ok && c.ShouldBindUri(iface) != nil {
				panic(ErrParseRequest.SetDetails("path"))
			}
			if _, ok := iface.(HeaderRequester); ok && c.ShouldBindHeader(iface) != nil {
				panic(ErrParseRequest.SetDetails("header"))
			}
			if v, ok := iface.(PostValidator); ok {
				if err := v.PostValidator(); err != nil {
					panic(err.SetDetails("post validator error"))
				}
			}

			inputs[i] = input.Elem()
			log.Printf("req%d: %+v\n", i, iface)
		}
		inputs[0] = reflect.ValueOf(&Context{Context: c})

		outs := handle.Call(inputs)
		result = outs[0].Interface()
		if ei := outs[1].Interface(); ei != nil {
			err = ei.(StatefulError)
		}
	})

	return nil
}
