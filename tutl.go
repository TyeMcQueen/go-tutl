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
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"
)

type Context struct {
	doNotEscape rune
	LineWidth   int
}

// You can just directly set the maximum line width for single-line test
// failure output [see also TUTL.SetLineWidth()] like:
//
//      tutl.Default = 120
//
var Default = Context{doNotEscape: '\n', LineWidth: 72}


// An interface covering the methods of *testing.T that TUTL uses.  This
// makes it easier to test this test library.
type TestingT interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}


// V() just converts a value to a string.  It is similar to
// fmt.Sprintf("%v", v).  But it treats []byte values as strings.
//
func V(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}
	return fmt.Sprintf("%v", v)
}


// Returns the string enclosed in double quotes and with contained \ and "
// characters escaped.
//
func DoubleQuote(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", s)
}


// After calling EscapeNewline(true), S() will escape '\n' characters.  You
// can call EscapeNewline(false) to restore the default behavior.
//
func EscapeNewline(b bool) { Default.EscapeNewline(b) }


func (c *Context) EscapeNewline(b bool) {
	if b {
		c.doNotEscape = ' '
	} else {
		c.doNotEscape = '\n'
	}
}


// Escape() returns a string containing the passed-in rune, unless it is a
// control character.  Runes '\n', '\r', and '\t' each return a 2-character
// string (\n, \r, or \t).  Other 7-bit control characters are turned into
// strings like \x1B.  The 8-bit control characters are turned into strings
// like \u009B.  EscapeNewline(false) does not affect Escape().
//
func Escape(r rune) string {
	switch r {
	case '\n':  return `\n`
	case '\r':  return `\r`
	case '\t':  return `\t`
	}
	if r < 32 || 0x7F == r {
		return fmt.Sprintf("\\x%02X", r)
	} else if 0x80 <= r && r < 0xa0 {
		return fmt.Sprintf("\\u00%02X", r)
	}
	return fmt.Sprintf("%c", r)
}


// Rune() returns a string consisting of the rune enclosed in single quotes,
// except that control characters are escaped [see Escape()].
//
// Note that neither ' nor \ characters are escaped so Char('\'') returns
// "'''" (3 apostrophes) and Char('\\') returns `'\'` (partly because `'\''`
// and `'\\'` are rather ugly).
//
func Rune(r rune) string {
	return fmt.Sprintf("'%s'", Escape(r))
}


// Char(c) is similar to Rune(rune(c)), except it escapes all byte values
// of 0x80 and above into 6-character strings like '\x9B' (rather then
// converting them UTF-8).
//
func Char(c byte) string {
	if 0xA0 <= c {
		return fmt.Sprintf("'\\x%02X'", c)
	}
	return Rune(rune(c))
}


// S() returns a single string composed by converting each argument into
// a string and concatenating all of those strings.  It is similar to but not
// identical to fmt.Sprint().  S() never inserts spaces between your values
// (if you want spaces, it is easy for you to add them).  S() puts single
// quotes around byte values.  S() treats []byte values like strings.  S()
// puts double quotes around []byte and error values (escaping enclosed "
// and \ characters).
//
// S() escapes control characters except for newlines [but see
// EscapeNewline()].  S() also escapes non-UTF-8 byte sequences.
//
// If S() is passed a single argument that is a string, then it will put
// double quotes around it and escape any contained " and \ characters.
//
// Note that S() does not put single quotes around rune values as "rune"
// is just an alias for "int32" so S('x') == 'x' == 120 while
// S("x"[0]) == "'x'".
//
func S(vs ...interface{}) string {
	return Default.S(vs...)
}


func (c Context) S(vs ...interface{}) string {
	ss := make([]string, len(vs))
	for j, i := range vs {
		s := ""
		switch v := i.(type) {
		case byte:
			s = Char(v)
		case error:
			s = DoubleQuote(v.Error())
		case []byte:
			s = DoubleQuote(string(v))
		case string:
			if 1 == len(vs) {
				s = DoubleQuote(v)
			} else {
				s = v
			}
		default:
			s = fmt.Sprintf("%v", i)
		}
		buf := make([]byte, 0, len(s))
		for i, r := range s {
			if 0xFFFD == r {
				buf = append(buf, []byte(fmt.Sprintf("\\x%02X", s[i]))...)
			} else if r < 32 && r != c.doNotEscape || 0x7f <= r {
				buf = append(buf, []byte(Escape(r))...)
			} else {
				buf = append(buf, byte(r))
			}
		}
		ss[j] = string(buf)
	}
	return strings.Join(ss, "")
}


// Is() tests that the first two arguments are converted to the same string
// by V().  If they are not, then a diagnostic is displayed which also causes
// the unit test to fail.
//
// The diagnostic is similar to "Got {got} not {want} for {desc}.\n" except
// that; 1) S() is used for 'got' and 'want' so control characters will be
// escaped and their values may be in quotes and 2) it will be split onto
// multiple lines if the values involved are long enough.
//
// Is() returns whether the test passed, which is useful for skipping tests
// that would make no sense to run given a prior failure or to display extra
// debug information only when a test fails.
//
func Is(want, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	return Default.Is(want, got, desc, t)
}


func (c Context) Is(want, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	vwant := V(want)
	vgot := V(got)
	if vwant == vgot {
	//  t.Log("want:", vwant, " got:", vgot, " for:", desc)
		return true
	}
	line := "Got " + c.S(got) + " not " + c.S(want) + " for " + desc + "."
	wid := utf8.RuneCount([]byte(line))
	if strings.Contains(line, "\n") {
		wid = 1 + c.LineWidth
	}
	if wid <= c.LineWidth-20 {
		t.Error(line)
	} else if wid <= c.LineWidth {
		t.Error("\n" + line)
	} else {
		t.Errorf("\nGot %s\nnot %s\nfor %s.", c.S(got), c.S(want), desc)
	}
	return false
}


// IsNot() tests that the first two arguments are converted to different
// strings by V().  If they are not, then a diagnostic is displayed which also
// causes the unit test to fail.  The diagnostic is similar to
// "Got unwanted {got} for {desc}.\n" except that S() is used for 'got' so
// control characters will be escaped and their values may be in quotes.
//
// IsNot() returns whether the test passed, which is useful for skipping tests
// that would make no sense to run given a prior failure.
//
func IsNot(hate, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	return Default.IsNot(hate, got, desc, t)
}


func (c Context) IsNot(hate, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	vhate := V(hate)
	vgot := V(got)
	if vhate != vgot {
	//  t.Log("hate:", vhate, " got:", vgot, " for:", desc)
		return true
	}
	line := "Got unwanted " + c.S(got) + " for " + desc + "."
	t.Error(line)
	return false
}


// Like() is most often used to test error messages.  It lets you perform
// multiple tests against a single value.  Each test checks that the value
// converts into a string that either contains a specific sub-string (ignoring
// letter case) or that it matches a regular expression.  You must pass
// at least one string to be matched.
//
// Strings that start with "*" have the "*" stripped before a substring match
// is performed (ignoring letter case).  If a string does not start with a
// "*", then it must be a valid regular expression that will be matched
// against the value's string representation.
//
// Except that strings that start with "!" have that stripped before checking
// for a subsequent "*".  The "!" negates the match so that the test will only
// pass if the string does not match.  To specify a regular expression that
// starts with a "!" character, simply escape it as `\!` or "[!]".
//
// Like() returns the number of matches that failed.
//
// If 'got' is nil, the empty string, or becomes the empty string, then no
// comparisons are done and a single failure is reported (but the number
// returned is the number of match strings as it is assumed that none of them
// would have matched the empty string).
//
func Like(got interface{}, desc string, t TestingT, match ...string) int {
	t.Helper()
	return Default.Like(got, desc, t, match...)
}


func (c Context) Like(
	got interface{}, desc string, t TestingT, match ...string,
) int {
	t.Helper()
	if 0 == len(match) {
		t.Errorf("Called Like() with too few arguments in test code.")
		return 1
	}

	sgot := V(got)
	empty := ""
	if nil == got {
		empty = "nil"
	} else if s, ok := got.(string); ok && "" == s {
		empty = "empty string"
	} else if "" == sgot {
		empty = "blank"
	}
	if "" != empty {
		t.Errorf("No string to check what it is Like(); got %s.", empty)
		return len(match)
	}

	failed := 0
	invalid := 0
	lgot := strings.ToLower(sgot)
	for _, m := range match {
		if "" == m || "!" == m {
			t.Error(`Match strings passed to Like() must not be empty nor "!"`)
			return len(match)
		}
		negate := false
		if '!' == m[0] {
			m = m[1:]
			negate = true
		}
		if '*' == m[0] {
			lwant := strings.ToLower(m[1:])
			if negate == strings.Contains(lgot, lwant) {
				failed++
				if negate {
					t.Errorf("Found unwanted <%s>...", m[1:])
				} else {
					t.Errorf("No <%s>...", m[1:])
				}
			}
		} else if re, err := regexp.Compile(m); nil != err {
			invalid++
			t.Errorf("Invalid regexp (%s) in test code: %v", m, err)
		} else if negate == ( "" != re.FindString(sgot) ) {
			failed++
			if negate {
				t.Errorf("Like unwanted /%s/...", m)
			} else {
				t.Errorf("Not like /%s/...", m)
			}
		}
	}
	if 0 < failed {
		t.Errorf("...in <%s> for %s.", sgot, desc)
	}
	return failed+invalid
}


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
// See also "go doc github.com/TyeMcQueen/go-tutl/hang".
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


// A type to allow an alternate calling style, especially for Is() and Like().
type TUTL struct {
	TestingT
	c Context
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
	return u.c.Is(want, got, desc, u)
}


// Same as the non-method IsNot() except the *testing.T argument is held in
// the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) IsNot(hate, got interface{}, desc string) bool {
	u.Helper()
	return u.c.IsNot(hate, got, desc, u)
}


// Same as the non-method Like() except the *testing.T argument is held in
// the TUTL object and so does not need to be passed as an argument.
//
func (u TUTL) Like(got interface{}, desc string, match ...string) int {
	u.Helper()
	return u.c.Like(got, desc, u, match...)
}


// New() copies the EscapeNewline() state from the Default Conext but future
// calls to EscapeNewline() only impact one Context, not any other copies.
//
func (u *TUTL) EscapeNewline(b bool) { u.c.EscapeNewline(b) }


// Same as the non-method S() except that it honors the state from the
// method version of EscapeNewline() [called on the same object].  New()
// copies the EscapeNewline() state from the Default Conext but future
// calls to EscapeNewline() only impact one Context, not any other copies.
//
func (u TUTL) S(vs ...interface{}) string { return u.c.S(vs...) }


// Sets the maximum line width for single-line test failure output, measured
// in UTF-8 runes.  If either of your values being compared would be displayed
// with unescaped newlines, then single-line output will not be used.
//
func (u *TUTL) SetLineWidth(w int) {
	u.c.LineWidth = w
}


func (u TUTL) V(v interface{}) string       { return V(v) }
func (u TUTL) DoubleQuote(s string) string  { return DoubleQuote(s) }
func (u TUTL) Escape(r rune) string         { return Escape(r) }
func (u TUTL) Rune(r rune) string           { return Rune(r) }
func (u TUTL) Char(c byte) string           { return Char(c) }
