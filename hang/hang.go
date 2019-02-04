/*

Include:

    import (
        _ "github.com/TyeMcQueen/go-tutl/hang" // ^C gives stack dumps.
    )

in just one of your *_test.go files so that you can interrupt (such as
via typing Ctrl-C) an infinite loop or otherwise hanging test run and be
shown, in response, the stack traces of everything that is running.

*/
package hang

import (
    "os"
    "testing"

    "github.com/TyeMcQueen/go-tutl"
)

// If your tests hang, interrupt them (Ctrl-C) to get stack dumps of what is
// running.
func TestMain(m *testing.M) {
    go tutl.ShowStackOnInterrupt()
    os.Exit(m.Run())
}
