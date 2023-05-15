/*

go-tutl is a Trivial Unit Testing Library (for Go).  Example usage:

	package duration
	import (
		"testing"

		u "github.com/Unity-Technologies/go-tutl-internal"
		_ "github.com/Unity-Technologies/go-tutl-internal/hang" // ^C gives stack dumps.
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

See also "go doc github.com/Unity-Technologies/go-tutl-internal/hang".

*/
package tutl

import (
	"fmt"
	"io"
	"os"
)

// TestingT is an interface covering the methods of '*testing.T' that TUTL
// uses.  This makes it easier to test this test library.
//
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

func (out FakeTester) Helper() {}

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

// TUTL is a type used to allow an alternate calling style, especially for
// Is() and Like().
//
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
//          u "github.com/Unity-Technologies/go-tutl-internal"
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
//          "github.com/Unity-Technologies/go-tutl-internal"
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
// and other options.
//
// New() also copies the current settings from the global 'tutl.Default' into
// the returned object.
//
func New(t TestingT) TUTL { return TUTL{t, Default} }

// Same as the non-method tutl.Is() except the '*testing.T' argument is held
// in the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) Is(want, got interface{}, desc string) bool {
	u.Helper()
	return u.o.Is(want, got, desc, u)
}

// Same as the non-method tutl.IsNot() except the '*testing.T' argument is
// held in the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) IsNot(hate, got interface{}, desc string) bool {
	u.Helper()
	return u.o.IsNot(hate, got, desc, u)
}

// Same as the non-method tutl.HasType() except the '*testing.T' argument is
// held in the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) HasType(want string, got interface{}, desc string) bool {
	u.Helper()
	return u.o.HasType(want, got, desc, u)
}

// Same as the non-method tutl.Circa() except the '*testing.T' argument is
// held in the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) Circa(digits int, want, got float64, desc string) bool {
	u.Helper()
	return u.o.Circa(digits, want, got, desc, u)
}

// Same as the non-method tutl.Like() except the '*testing.T' argument is
// held in the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) Like(got interface{}, desc string, match ...string) int {
	u.Helper()
	return u.o.Like(got, desc, u, match...)
}

// Same as the non-method tutl.S() except that it honors the option settings
// of the invoking TUTL object, not of the 'tutl.Default' global.
//
func (u TUTL) S(vs ...interface{}) string { return u.o.S(vs...) }

// Same as the non-method tutl.V() except that it honors the option settings
// of the invoking TUTL object, not of the tutl.Default global.
//
func (u TUTL) V(v interface{}) string {
	return u.o.V(v)
}

// Same as the ReplaceNewlines() method on the 'tutl.Default' global,
// except it honors the settings from the invoking TUTL object.
//
func (u *TUTL) ReplaceNewlines(s string) string {
	return u.o.ReplaceNewlines(s)
}

// Same as the EscapeNewline() method on the 'tutl.Default' global,
// except it only changes the setting for the invoking TUTL object.
//
func (u *TUTL) EscapeNewline(b bool) { u.o.EscapeNewline(b) }

// SetLineWidth() is the same as setting the global 'tutl.Default.LineWidth'
// except it only changes the setting for the invoking TUTL object.
//
func (u *TUTL) SetLineWidth(w int) {
	u.o.LineWidth = w
}

// SetPathLength() is the same as setting the global 'tutl.Default.PathLength'
// except it only changes the setting for the invoking TUTL object.
//
func (u *TUTL) SetPathLength(l int) {
	u.o.PathLength = l
}

// SetDigits32() is the same as setting the global 'tutl.Default.Digits32'
// value, except it only changes the setting for the invoking TUTL object.
//
func (u TUTL) SetDigits32(d int) {
	u.o.Digits32 = d
}

// SetDigits64() is the same as setting the global 'tutl.Default.Digits64'
// value, except it only changes the setting for the invoking TUTL object.
//
func (u TUTL) SetDigits64(d int) {
	u.o.Digits64 = d
}

// Identical to the non-method tutl.DoubleQuote().
func (u TUTL) DoubleQuote(s string) string {
	return DoubleQuote(s)
}

// Identical to the non-method tutl.Escape().
func (u TUTL) Escape(r rune) string {
	return Escape(r)
}

// Identical to the non-method tutl.Rune().
func (u TUTL) Rune(r rune) string {
	return Rune(r)
}

// Identical to the non-method tutl.Char().
func (u TUTL) Char(c byte) string {
	return Char(c)
}

// GetPanic() calls the passed-in function and returns 'nil' or the argument
// that gets passed to panic() from within it.  This can be used in other
// test functions, for example:
//
//      u.Is(nil, u.GetPanic(func(){ obj.Method(nil) }), "Method panic")
//
func (_ TUTL) GetPanic(run func()) interface{} {
	return GetPanic(run)
}
