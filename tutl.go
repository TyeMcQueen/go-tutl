/*

go-tutl is a Trivial Unit Testing Library (for Go).  Example usage:

	package duration
	import (
		"testing"

		u "github.com/TyeMcQueen/go-tutl"
		_ "github.com/TyeMcQueen/go-tutl/hang" // ^C gives stack dumps.
	)

	func TestDur(t *testing.T) {
		u.Is("1m 1s", DurationAsText(61), "61", t)

		got, err := TextAsDuration("1h 5s")
		u.Is(nil, err, "Error from '1h 5s'", t)
		u.Is(60*60+5, got, "'1h 5s'", t)

		got, err = TextAsDuration("3 fortnight")
		u.Like(err, "Error from '3 fortnight'", t,
			"(Unknown|Invalid) unit", "*fortnight")
	}

See also "go doc github.com/TyeMcQueen/go-tutl/hang".

*/
package tutl

import (
	"fmt"
	"io"
	"os"
)

// An interface covering the methods of *testing.T that TUTL uses.  This
// makes it easier to test this test library.
type TestingT interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Failed() bool
}

// A FakeTester is a replacement for a '*testing.T' so that you can use
// TUTL's functionality outside of a real 'go test' run.
//
type FakeTester struct {
	Output    io.Writer
	HasFailed bool
}

// The 'tutl.StdoutTester' is a replacement for a '*testing.T' that just
// writes output to 'os.Stdout'.
//
var StdoutTester = FakeTester{os.Stdout, false}

func (out FakeTester) Helper() { }

func (out FakeTester) Log(args ...interface{}) {
	fmt.Fprintln(out.Output, args...)
}

func (out FakeTester) Logf(format string, args ...interface{}) {
	if "" == format || '\n' != format[len(format)-1] {
		format += "\n"
	}
	fmt.Fprintf(out.Output, format, args...)
}

func (out FakeTester) Error(args ...interface{}) {
	out.Log(args...)
	out.HasFailed = true
}

func (out FakeTester) Errorf(format string, args ...interface{}) {
	out.Logf(format, args...)
	out.HasFailed = true
}

func (out FakeTester) Failed() bool {
	return out.HasFailed
}

// A type to allow an alternate calling style, especially for Is() and Like().
type TUTL struct {
	TestingT
	o Options
}

// A unit test can have a huge number of calls to Is().  Having to remember
// to pass in the *testing.T argument can be inconvenient.  TUTL offers an
// alternate calling method that replaces the huge number of such extra
// arguments with a single line of code.  This example code:
//
//      import (
//          "testing"
//          u "github.com/TyeMcQueen/go-tutl"
//          ^^  Import alias
//      )
//
//      func TestDur(t *testing.T) {
//          u.Is("1m 1s", DurationAsText(61), "61", t)
//          //                                    ^^^  Extra argument
//          u.Like(Valid("3f", "Error from '3f'", t, "(Unknown|Invalid) unit")
//          //                                    ^^^  Extra argument
//      }
//
// would become:
//
//      import (
//          "testing"
//          "github.com/TyeMcQueen/go-tutl"
//      )
//
//      func TestDur(t *testing.T) {
//          var u = tutl.New(t)
//          // ^^^^^^^^^^^^^^^^ Added line
//          u.Is("1m 1s", DurationAsText(61), "61")
//          u.Like(Valid("3f", "Error from '3f'", "(Unknown|Invalid) unit")
//      }
//
// Whether to use an import alias or New() (or neither) is mostly a personal
// preference.  Though, using New() also limits the scope of EscapeNewline()
// MaxLineLine().
//
func New(t TestingT) TUTL { return TUTL{t, Default} }

// Same as the non-method Is() except the *testing.T argument is held in
// the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) Is(want, got interface{}, desc string) bool {
	u.Helper()
	return u.o.Is(want, got, desc, u)
}

// Same as the non-method IsNot() except the *testing.T argument is held in
// the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) IsNot(hate, got interface{}, desc string) bool {
	u.Helper()
	return u.o.IsNot(hate, got, desc, u)
}

// Same as the non-method Like() except the *testing.T argument is held in
// the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) Like(got interface{}, desc string, match ...string) int {
	u.Helper()
	return u.o.Like(got, desc, u, match...)
}

// New() copies the EscapeNewline() state from the Default Conext but future
// calls to EscapeNewline() only impact one Options, not any other copies.
//
func (u *TUTL) EscapeNewline(b bool) { u.o.EscapeNewline(b) }

// Same as the non-method S() except that it honors the state from the
// method version of EscapeNewline() [called on the same object].  New()
// copies the EscapeNewline() state from the Default Conext but future
// calls to EscapeNewline() only impact one Options, not any other copies.
//
func (u TUTL) S(vs ...interface{}) string { return u.o.S(vs...) }

// Sets the maximum line width for single-line test failure output, measured
// in UTF-8 runes.  If either of your values being compared would be displayed
// with unescaped newlines, then single-line output will not be used.
//
func (u *TUTL) SetLineWidth(w int) {
	u.o.LineWidth = w
}

func (u TUTL) V(v interface{}) string       { return V(v) }
func (u TUTL) DoubleQuote(s string) string  { return DoubleQuote(s) }
func (u TUTL) Escape(r rune) string         { return Escape(r) }
func (u TUTL) Rune(r rune) string           { return Rune(r) }
func (u TUTL) Char(c byte) string           { return Char(c) }
