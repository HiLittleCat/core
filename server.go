package core

import (
	"flag"
	"time"

	"github.com/astaxie/beego/grace"
)

var (
	// Production allows handlers know whether the server is running in a production environment.
	Production bool

	// Address is the TCP network address on which the server is listening and serving. Default is ":8080".
	Address = ":8080"

	// beforeRun stores a set of functions that are triggered just before running the server.
	beforeRun []func()
)

func init() {
	flag.BoolVar(&Production, "production", Production, "run the server in production environment")
	flag.StringVar(&Address, "address", Address, "the address to listen and serving on")
}

// BeforeRun adds a function that will be triggered just before running the server.
func BeforeRun(f func()) {
	beforeRun = append(beforeRun, f)
}

// Run starts the server for listening and serving.
func Run() {
	for _, f := range beforeRun {
		f()
	}

	// set server
	grace.DefaultReadTimeOut = 10 * time.Second
	// DefaultWriteTimeOut is the HTTP Write timeout
	grace.DefaultWriteTimeOut = 10 * time.Second
	// DefaultMaxHeaderBytes is the Max HTTP Herder size, default is 0, no limit
	grace.DefaultMaxHeaderBytes = 0
	// DefaultTimeout is the shutdown server's timeout. default is 60s

	panic(grace.ListenAndServe(Address, defaultHandlersStack))

}
