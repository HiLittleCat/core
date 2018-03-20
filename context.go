package core

import (
	"errors"
	"fmt"
	"sync"

	"net/http"
	"runtime"

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
	Data           map[string]interface{} // Custom Data
}

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
	for k := range ctx.Data {
		delete(ctx.Data, k)
	}
	ctx.index = -1
	ctx.written = false
	return ctx
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
func (ctx *Context) Ok(status int, data interface{}) (int, error) {
	if ctx.written == true {
		return 0, errors.New("Context.Success: request has been writed")
	}
	ctx.written = true
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(&resOk{Ok: true, Data: data})
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	ctx.ResponseWriter.WriteHeader(status)
	return ctx.ResponseWriter.Write(b)
}

// Fail Response fail
func (ctx *Context) Fail(status int, err error) (int, error) {
	message := err.Error()
	if ctx.written == true {
		return 0, errors.New("Context.JSON: request has been writed")
	}
	ctx.written = true
	if err != nil {
		if _, ok := err.(*ServerError); ok == true {
			log.WithFields(log.Fields{"Controller": ctx.Data["Controller"], "Method": ctx.Data["Method"], "err": err}).Warnln(message)
		}
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(&resFail{Ok: false, Message: ctx.Request.URL.Path + ": " + message})
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	ctx.ResponseWriter.WriteHeader(status)
	return ctx.ResponseWriter.Write(b)
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

// Recover recovers form panics.
// It logs the stack and uses the PanicHandler (or a classic Internal Server Error) to write the response.
//
// Usage:
//
//	defer c.Recover()
func (ctx *Context) Recover() {
	if err := recover(); err != nil {
		if e, ok := err.(ValidationError); ok == true {
			ctx.Fail(http.StatusBadRequest, &e)
			return
		}
		stack := make([]byte, 64<<10)
		stack = stack[:runtime.Stack(stack, false)]
		log.Errorf("%v \n %s", err, stack)
		if !ctx.Written() {
			ctx.ResponseWriter.Header().Del("Content-Type")

			if ctx.handlersStack.PanicHandler != nil {
				ctx.Data["panic"] = err
				ctx.handlersStack.PanicHandler(ctx)
			} else {
				ctx.Fail(http.StatusInternalServerError, &ServerError{Message: http.StatusText(http.StatusInternalServerError)})
				//http.Error(c.ResponseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	}
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
