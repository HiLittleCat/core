package core

import (
	"net/http"
)

// HandlersStack contains a set of handlers.
type HandlersStack struct {
	Handlers     []func(*Context) // The handlers stack.
	PanicHandler func(*Context)   // The handler called in case of panic. Useful to send custom server error information. Context.Data["panic"] contains the panic error.
}

// defaultHandlersStack contains the default handlers stack used for serving.
var defaultHandlersStack = NewHandlersStack()

// NewHandlersStack returns a new NewHandlersStack.
func NewHandlersStack() *HandlersStack {
	return new(HandlersStack)
}

// Use adds a handler to the handlers stack.
func (hs *HandlersStack) Use(h func(*Context)) {
	hs.Handlers = append(hs.Handlers, h)
}

// Use adds a handler to the default handlers stack.
func Use(h func(*Context)) {
	defaultHandlersStack.Use(h)
}

// HandlePanic sets the panic handler of the handlers stack.
//
// Context.Data["panic"] contains the panic error.
func (hs *HandlersStack) HandlePanic(h func(*Context)) {
	hs.PanicHandler = h
}

// HandlePanic sets the panic handler of the default handlers stack.
//
// Context.Data["panic"] contains the panic error.
func HandlePanic(h func(*Context)) {
	defaultHandlersStack.HandlePanic(h)
}

// ServeHTTP makes a context for the request, sets some good practice default headers and enters the handlers stack.
func (hs *HandlersStack) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get a context for the request from ctxPool.
	c := getContext(w, r)

	// Set some "good practice" default headers.
	c.ResponseWriter.Header().Set("Cache-Control", "no-cache")
	c.ResponseWriter.Header().Set("Content-Type", "application/json")
	c.ResponseWriter.Header().Set("Connection", "keep-alive")
	c.ResponseWriter.Header().Set("Vary", "Accept-Encoding")

	// Always recover form panics.
	defer c.Recover()

	// Enter the handlers stack.
	c.Next()

	// Put the context to ctxPool
	putContext(c)
}
