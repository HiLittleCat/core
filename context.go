package core

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/json-iterator/go"
)

// Context contains all the data needed during the serving flow, including the standard http.ResponseWriter and *http.Request.
//
// The Data field can be used to pass all kind of data through the handlers stack.
type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	index          int                    // Keeps the actual handler index.
	handlersStack  *HandlersStack         // Keeps the reference to the actual handlers stack.
	written        bool                   // A flag to know if the response has been written.
	Params         Params                 // Path Value
	Data           map[string]interface{} // Custom Data
}

// ResFormat response data
type ResFormat struct {
	Ok      bool
	Data    interface{}
	Message string
}

type resOk struct {
	Ok   bool
	Data interface{}
}

type resFail struct {
	Ok      bool
	Message string
}

// Ok Response json
func (ctx *Context) Ok(data interface{}) {
	if ctx.written == true {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln("Context.Success: request has been writed")
		return
	}
	ctx.written = true
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(&resOk{Ok: true, Data: data})
	ctx.ResponseWriter.WriteHeader(http.StatusOK)
	_, err := ctx.ResponseWriter.Write(b)
	if err != nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}
}

// Fail Response fail
func (ctx *Context) Fail(err error) {
	if err == nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln("Context.Fail: err is nil")
		ctx.ResponseWriter.WriteHeader(err.(*ServerError).HTTPCode)
		_, err = ctx.ResponseWriter.Write(nil)
		return
	}

	message := err.Error()
	if ctx.written == true {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln("Context.Fail: request has been writed")
		return
	}
	ctx.written = true
	if _, ok := err.(*ServerError); ok == true {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(message)
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(&resFail{Ok: false, Message: message})

	coreErr, ok := err.(ICoreError)
	if ok == true {
		ctx.ResponseWriter.WriteHeader(coreErr.GetHTTPCode())
	} else {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
	}

	_, err = ctx.ResponseWriter.Write(b)
	if err != nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}
}

// ResStatus Response status code, use http.StatusText to write the response.
func (ctx *Context) ResStatus(code int) (int, error) {
	if ctx.written == true {
		return 0, errors.New("Context.ResStatus: request has been writed")
	}
	ctx.written = true
	ctx.ResponseWriter.WriteHeader(code)
	return fmt.Fprint(ctx.ResponseWriter, http.StatusText(code))
}

// Written tells if the response has been written.
func (ctx *Context) Written() bool {
	return ctx.written
}

// Next calls the next handler in the stack, but only if the response isn't already written.
func (ctx *Context) Next() {
	// Call the next handler only if there is one and the response hasn't been written.
	if !ctx.Written() && ctx.index < len(ctx.handlersStack.Handlers)-1 {
		ctx.index++
		ctx.handlersStack.Handlers[ctx.index](ctx)
	}
}

// Param returns the value of the URL param.
// It is a shortcut for c.Params.ByName(key)
//     router.GET("/user/:id", func(c *gin.Context) {
//         // a GET request to /user/john
//         id := c.Param("id") // id == "john"
//     })
func (ctx *Context) Param(key string) string {
	return ctx.Params.ByName(key)
}

// GetSession get session
func (ctx *Context) GetSession() IStore {
	store := ctx.Data["session"]
	if store == nil {
		return nil
	}
	st, ok := store.(IStore)
	if ok == false {
		return nil
	}
	return st
}

// SetSession set session
func (ctx *Context) SetSession(key string, values map[string]string) error {
	// 生成csrf token
	//values["token"] = ctx.genToken(key)
	sid := ctx.genSid(key)
	store, err := provider.Set(sid, values)
	if err != nil {
		return err
	}
	cookie := httpCookie
	cookie.Value = sid
	ctx.Data["session"] = store

	respCookie := ctx.ResponseWriter.Header().Get("Set-Cookie")
	if strings.HasPrefix(respCookie, cookie.Name) {
		ctx.ResponseWriter.Header().Del("Set-Cookie")
	}
	http.SetCookie(ctx.ResponseWriter, &cookie)
	return nil
}

// FreshSession set session
func (ctx *Context) FreshSession(key string) error {
	err := provider.UpExpire(key)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSession delete session
func (ctx *Context) DeleteSession(sid string) error {
	ctx.Data["session"] = nil
	provider.Destroy(sid)
	cookie := httpCookie
	cookie.MaxAge = -1
	http.SetCookie(ctx.ResponseWriter, &cookie)
	return nil
}

func (ctx *Context) genToken(key string) string {
	return key
}

func (ctx *Context) genSid(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Recover recovers form panics.
// It logs the stack and uses the PanicHandler (or a classic Internal Server Error) to write the response.
//
// Usage:
//
//	defer c.Recover()
func (ctx *Context) Recover() {
	if err := recover(); err != nil {
		if e, ok := err.(*ValidationError); ok == true {
			ctx.Fail(e)
			return
		}

		stack := make([]byte, 64<<10)
		stack = stack[:runtime.Stack(stack, false)]
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Errorln(string(stack))
		if !ctx.Written() {
			ctx.ResponseWriter.Header().Del("Content-Type")

			if ctx.handlersStack.PanicHandler != nil {
				ctx.Data["panic"] = err
				ctx.handlersStack.PanicHandler(ctx)
			} else {
				ctx.Fail((&ServerError{}).New(http.StatusText(http.StatusInternalServerError)))
			}
		}
	}
}

// ctxPool
var ctxPool = sync.Pool{
	New: func() interface{} {
		return &Context{
			Data:          make(map[string]interface{}),
			index:         -1, // Begin with -1 because Next will increment the index before calling the first handler.
			handlersStack: defaultHandlersStack,
		}
	},
}

func getContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := ctxPool.Get().(*Context)
	ctx.Request = r
	ctx.ResponseWriter = contextWriter{w, ctx}
	ctx.Data = make(map[string]interface{})
	return ctx
}

func putContext(ctx *Context) {
	if ctx.Request.Body != nil {
		ctx.Request.Body.Close()
	}
	ctx.Data = nil
	ctx.Params = nil
	ctx.ResponseWriter = nil
	ctx.Request = nil
	ctx.index = -1
	ctx.written = false
	ctxPool.Put(ctx)
}

// contextWriter represents a binder that catches a downstream response writing and sets the context's written flag on the first write.
type contextWriter struct {
	http.ResponseWriter
	context *Context
}

// Write sets the context's written flag before writing the response.
func (w contextWriter) Write(p []byte) (int, error) {
	w.context.written = true
	return w.ResponseWriter.Write(p)
}

// WriteHeader sets the context's written flag before writing the response header.
func (w contextWriter) WriteHeader(code int) {
	w.context.written = true
	w.ResponseWriter.WriteHeader(code)
}
