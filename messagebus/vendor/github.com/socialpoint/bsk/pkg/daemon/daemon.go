package daemon

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/socialpoint/bsk/pkg/httpx"
	"github.com/socialpoint/bsk/pkg/server"
)

const serverTimeout = time.Second * 30

// Application is something runnable
type Application server.Runner

// HTTPApplication is an Apllication that exposes HTTP port
type HTTPApplication interface {
	RegisterHTTP(*httpx.Router)
}

// Daemon coordinates several concurrently running apps
type Daemon struct {
	port int
	apps []Application
}

// New returns a Daemon
func New(port int, apps ...Application) *Daemon {
	return &Daemon{
		port: port,
		apps: apps,
	}
}

// Run creates a new router, and then for every app of the deamon if iis an
// HTTPApplication it registers it with the router (usually configuring
// endpoints) and runs it.
// Then registers the root path endpoint with an StatusOKHandler for the ELB health check.
// Finally starts the http Server with the router http Handler.
func (a *Daemon) Run(ctx context.Context) {
	r := httpx.NewRouter()

	for _, app := range a.apps {
		if httpApp, ok := app.(HTTPApplication); ok {
			httpApp.RegisterHTTP(r)
		}

		go app.Run(ctx)
	}

	// ELBs need a 200OK in the port they are balancing to
	// see http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-healthchecks.html#health-check-configuration
	r.Route("/", httpx.StatusOKHandler)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(a.port)))
	if err != nil {
		panic("error running HTTP server")
	}

	srv := &http.Server{
		Handler:      r,
		ReadTimeout:  serverTimeout,
		WriteTimeout: serverTimeout,
	}
	go func() {
		err := srv.Serve(listener)
		if err != nil {
			panic("error running HTTP server")
		}
	}()
}

// NewSetup returns a Setup with configured Options
func NewSetup(opts ...SetupOption) *Setup {
	setup := &Setup{}

	for _, o := range opts {
		o(setup)
	}

	return setup
}

// Setup is a set of applications
type Setup []Application

// SetupOption configures a Setup
type SetupOption func(*Setup)
