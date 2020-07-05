/*

Include:

	import (
		_ "github.com/TyeMcQueen/go-tutl/hang" // ^C gives stack dumps.
	)

in just one of your *_test.go files (per module) so that you can interrupt
(such as via typing Ctrl-C) an infinite loop or otherwise hanging test run
and be shown, in response, the stack traces of everything that is running.

If you have your own TestMain() function, then just call:

	go tutl.ShowStackOnInterrupt()

from it rather than loading this module (which would fail).

Note that loading this module does nothing useful in non-test code.

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
