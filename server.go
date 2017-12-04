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
	// Production allows handlers know whether the server is running in a production environment.
	Production bool

	// Address is the TCP network address on which the server is listening and serving. Default is ":8080".
	Address = ":8080"

	// beforeRun stores a set of functions that are triggered just before running the server.
	beforeRun []func()

	// Maximum duration for reading the full request (including body); ns|µs|ms|s|m|h
	ReadTimeout time.Duration

	// Maximum duration for writing the full response (including body); ns|µs|ms|s|m|h
	WriteTimeout time.Duration

	// Maximum size of memory that can be used when receiving uploaded files
	MultipartMaxmemoryMb int

	//Max HTTP Herder size, default is 0, no limit
	MaxHeaderBytes int
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

	log.Println(fmt.Sprintf("Serving %s with pid %d.", Address, os.Getpid()))

	srv := &graceful.Server{
		Timeout: 10 * time.Second,

		ConnState: func(conn net.Conn, state http.ConnState) {
			// conn has a new state
		},

		Server: &http.Server{
			Addr:    Address,
			Handler: defaultHandlersStack,
		},
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Server stoped.")

}
