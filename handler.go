package core

import (
	"net/http"
)

// HandlersStack contains a set of handlers.
type HandlersStack struct {
	Handlers     []RouterHandler // The handlers stack.
	PanicHandler RouterHandler   // The handler called in case of panic. Useful to send custom server error information. Context.Data["panic"] contains the panic error.
}

// defaultHandlersStack contains the default handlers stack used for serving.
var defaultHandlersStack = NewHandlersStack()

// NewHandlersStack returns a new NewHandlersStack.
func NewHandlersStack() *HandlersStack {
	return new(HandlersStack)
}

// Use adds a handler to the handlers stack.
func (hs *HandlersStack) Use(h RouterHandler) {
	hs.Handlers = append(hs.Handlers, h)
}

// Use adds a handler to the default handlers stack.
func Use(h RouterHandler) {
	defaultHandlersStack.Use(h)
}

// HandlePanic sets the panic handler of the handlers stack.
//
// Context.Data["panic"] contains the panic error.
func (hs *HandlersStack) HandlePanic(h RouterHandler) {
	hs.PanicHandler = h
}

// HandlePanic sets the panic handler of the default handlers stack.
//
// Context.Data["panic"] contains the panic error.
func HandlePanic(h RouterHandler) {
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
	//c.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	c.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "X-Requested-With")
	c.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")

	// Always recover form panics.
	defer c.Recover()

	// Enter the handlers stack.
	c.Next()

	// Respnose data
	// if c.written == false {
	// 	c.Fail(errors.New("not written"))
	// }
	// Put the context to ctxPool
	putContext(c)
}
