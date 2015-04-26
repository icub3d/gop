// Package signalhandler provides an all-in-one solution for simple
// signal handling. It basically implements the common idiomatic was of
// using os/signal.
package signalhandler

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// These are common signals. You can find more in the packages os
// and syscall.
const (
	SigHup  = syscall.SIGHUP  // Reload the config.
	SigUsr1 = syscall.SIGUSR1 // Reopen the logs.
	SigTerm = syscall.SIGTERM // gracefully die.
	SigKill = syscall.SIGKILL // bad day.
)

// SignalHandler is the interface used for handling incoming signals
// from the operating system.
type SignalHandler interface {
	// Watch registers the given function when the given signal is
	// called. Multiple functions can be registered to a signal and if
	// the same function is registered multiple times, it will be called
	// multiple times.
	Watch(os.Signal, func())

	// Stop stops watching for incoming signals.
	Stop()
}

// Handler is our implementation of the SignalHandler interface.
type handler struct {
	// Incoming is the incoming channel.
	incoming chan os.Signal

	// funcs is map of functions mapped to signals.
	funcs map[os.Signal][]func()

	// We'll use this to stop the goroutine that's waiting on signals.
	stop chan struct{}
}

func (h *handler) Watch(sig os.Signal, f func()) {
	if funcs, ok := h.funcs[sig]; !ok {
		h.funcs[sig] = []func(){f}
	} else {
		h.funcs[sig] = append(funcs, f)
	}
	signal.Notify(h.incoming, sig)
}

func (h *handler) Stop() {
	signal.Stop(h.incoming)
	close(h.stop)
}

// Listen is the main loop that listens for signals until stop is
// called.
func (h *handler) listen() {
	for {
		select {
		case sig := <-h.incoming:
			funcs, ok := h.funcs[sig]
			fmt.Println(sig, len(funcs))
			if ok {
				// Call all the registered functions.
				for _, f := range funcs {
					f()
				}
			}
		case <-h.stop:
			return
		}
	}
}

// New create a new signal handler which is listening for
// signal. Calls to Watch() will add functions when signals come down
// the pipe. Stop() should be called when you are done listening for
// signals.
func New() SignalHandler {
	h := &handler{
		incoming: make(chan os.Signal, 20),
		funcs:    make(map[os.Signal][]func()),
		stop:     make(chan struct{}),
	}
	go h.listen()
	return h
}
