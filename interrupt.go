package tutl

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
)

var atInterrupt = make([]func(), 0, 16)
var aiMu sync.Mutex
var running = 0
var skip = true

// If you have a TestMain() function, then you can add
//
//      go tutl.ShowStackOnInterrupt()
//
// to it to allow you to interrupt it (such as via typing Ctrl-C) in order
// to see stack traces of everything that is running.  This is particularly
// useful if your code has an infinite loop.
//
// See also "go doc github.com/Unity-Technologies/go-tutl-internal/hang".
//
// ShowStackOnInterrupt() can also be used from non-test programs.
//
// ShowStackOnInterrupt(false) has a special meaning; see AtInterrupt().
//
func ShowStackOnInterrupt(show ...bool) {
	aiMu.Lock()
	if 0 == len(show) || show[0] {
		skip = false
	}
	if 0 < running {
		aiMu.Unlock()
		return
	}
	running++
	aiMu.Unlock()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	_ = <-sig

	aiMu.Lock()
	// Make a reversed copy of the atInterrupt slice:
	cp := make([]func(), len(atInterrupt))
	for i, ai := range atInterrupt {
		cp[len(cp)-1-i] = ai
	}
	aiMu.Unlock()

	// Call functions registered via AtInterrupt(), in reverse order:
	for _, ai := range cp {
		ai()
	}

	if skip {
		fmt.Fprintln(os.Stderr, "Interrupted.")
		os.Exit(1)
	}
	debug.SetTraceback("all")
	panic("Interrupted")
}

// AtInterrupt registers a function to be called if the test run is
// interrupted (by the user typing Ctrl-C or whatever sends SIGINT).
// The function registered first is run last.  You must have called
// ShowStackOnInterrupt() or AtInterrupt() will do nothing useful.
//
// The function passed in is also returned.  This can be useful for
// clean-up code that should be run whether the test run is interrupted
// or not:
//
//      defer tutl.AtInterrupt(cleanup)()
//
// If you want to empower AtInterrupt() even if ShowStackOnInterrupt() has
// not been enabled, then your code can call:
//
//      go tutl.ShowStackOnInterrupt(false)
//
// This does nothing if ShowStackOnInterrupt() had already been called
// (in particular, it does not disable the showing of stack traces).
// Otherwise, it does the work that it would normally do except, if a
// SIGINT is received, it will only run any functions registered via
// AtInterrupt() but will not show stack traces.
//
func AtInterrupt(f func()) func() {
	aiMu.Lock()
	defer aiMu.Unlock()
	atInterrupt = append(atInterrupt, f)
	return f
}
