// A couple of simple helper functions to make it easy to get CPU or blocking
// profile data from your tests.
package profile

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/Unity-Technologies/go-tutl-internal"
)

func die(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	if 0 == len(args) {
		fmt.Fprint(os.Stderr, format)
	} else {
		fmt.Fprintf(os.Stderr, format, args...)
	}
	os.Exit(1)
}

// To save CPU profile data from your program, add code like the following to
// your main() function:
//
//      import(
//          "os"
//          "github.com/Unity-Technologies/go-tutl-internal"
//          "github.com/Unity-Technologies/go-tutl-internal/profile"
//      )
//
//      func main() {
//          // ...
//          if path := os.Getenv("CPU_PROFILE"); "" != path {
//              go tutl.ShowStackTraceOnInterrupt(false)
//              defer profile.ProfieCPU(path)()
//          }
//          // ...
//      }
//
// The call to ShowStackOnInterrupt() ensures the CPU profile data will be
// properly flushed even if you interrupt (SIGINT, Ctrl-C) your test run.
//
func ProfileCPU(file string) func() {
	fh, err := os.Create(file)
	if err != nil {
		die("Can't create CPU profile, %s: %v", file, err)
	}
	if err = pprof.StartCPUProfile(fh); err != nil {
		die("Can't start CPU profile: %v", err)
	}
	return tutl.AtInterrupt(func() {
		pprof.StopCPUProfile()
		fh.Close()
	})
}

// To save block profile data (how much time is being spent waiting) from
// your program, add code like the following to your main() function:
//
//      import(
//          "os"
//          "github.com/Unity-Technologies/go-tutl-internal/profile"
//      )
//
//      func main() {
//          // ...
//          if path := os.Getenv("BLOCK_PROFILE"); "" != path {
//              go tutl.ShowStackTraceOnInterrupt(false)
//              defer profile.ProfieBlockings(path)()
//          }
//          // ...
//      }
//
// The call to ShowStackOnInterrupt() ensures the blocking profile data
// will be saved even if you interrupt (SIGINT, Ctrl-C) your test run.
//
func ProfileBlockings(file string) func() {
	fh, err := os.Create(file)
	if err != nil {
		die("Can't create block profile, %s: %v", file, err)
	}
	runtime.SetBlockProfileRate(1)
	return tutl.AtInterrupt(func() {
		runtime.SetBlockProfileRate(0)
		fmt.Fprintf(os.Stderr, "Saving blockings profiles to %s...\n", file)
		pprof.Lookup("block").WriteTo(fh, 0)
		fh.Close()
	})
}
