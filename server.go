package core

import (
	"flag"
	"fmt"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"

	"os"
	"time"

	"gopkg.in/tylerb/graceful.v1"
)

var (
	// OpenCommandLine Open command line params.
	OpenCommandLine bool

	// Production allows handlers know whether the server is running in a production environment.
	Production bool

	// Address is the TCP network address on which the server is listening and serving. Default is ":8080".
	Address = ":8080"

	// beforeRun stores a set of functions that are triggered just before running the server.
	beforeRun []func()

	// Timeout is the duration to allow outstanding requests to survive
	// before forcefully terminating them.
	Timeout = 30 * time.Second

	// ListenLimit Limit the number of outstanding requests
	ListenLimit = 5000

	// ReadTimeout Maximum duration for reading the full request (including body); ns|µs|ms|s|m|h
	ReadTimeout = 5 * time.Second

	// WriteTimeout Maximum duration for writing the full response (including body); ns|µs|ms|s|m|h
	WriteTimeout = 10 * time.Second

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration

	// MultipartMaxmemoryMb Maximum size of memory that can be used when receiving uploaded files
	MultipartMaxmemoryMb int

	// MaxHeaderBytes Max HTTP Herder size, default is 0, no limit
	MaxHeaderBytes = 1 << 20
)

func init() {
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

	// parse command line params.
	if OpenCommandLine {
		flag.StringVar(&Address, "address", ":8080", "-address=:8080")
		flag.BoolVar(&Production, "production", false, "-production=false")
		flag.Parse()
	}

	log.Warnln(fmt.Sprintf("Serving %s with pid %d. Production is %t.", Address, os.Getpid(), Production))

	// set default router.
	Use(Routers.RouteHandler)

	// set graceful server.
	srv := &graceful.Server{
		ListenLimit: ListenLimit,
		ConnState: func(conn net.Conn, state http.ConnState) {
			// conn has a new state
		},
		Server: &http.Server{
			Addr:           Address,
			Handler:        defaultHandlersStack,
			ReadTimeout:    ReadTimeout,
			WriteTimeout:   WriteTimeout,
			IdleTimeout:    IdleTimeout,
			MaxHeaderBytes: MaxHeaderBytes,
		},
	}
	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalln(err)
	}
	log.Warnln("Server stoped.")

}
