package tutl

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Options contains user preference options.  The 'tutl.Default' global
// is the Options used unless you make a copy of it and use the copy.
//
// Calling tutl.New(t) associates such a copy with the returned object so
// changes to preferences via the returned object don't modify 'Default'.
//
//      func TestFoo(t *testing.T) {
//          u := tutl.New(t)
//          u.SetLineWidth(120)
//          u.Is(want, got(), "") // Uses 120-character line width.
//          tutl.Is(want, got(), "", t) // Uses tutl.Default's line width.
//      }
//
// You can modify some options directly via 'tutl.Default', such as:
//
//      tutl.Default.LineWidth = 120
//
type Options struct {
	// Gets set to '\n' to NOT escape newlines (' ' to escape newlines).
	doNotEscape rune

	// LineWidth influences when "Got {got} not {want} for {title}" output
	// gets split onto multiple lines instead.  See Is() for details.
	//
	LineWidth   int
}

// The 'tutl.Default' global contains the user preferences to be used unless
// you make a copy and use it, such as via New() (see Options for more).
//
var Default = Options{
	doNotEscape: '\n', LineWidth: 72}

// V() just converts a value to a string.  It is similar to 'fmt.Sprint(v)'.
// But it treats '[]byte' values as 'string's.
//
func V(v interface{}) string {
	return Default.V(v)
}

// See tutl.V() for documentation.
func (o Options) V(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}
	return fmt.Sprint(v)
}

// DoubleQuote() returns the string enclosed in double quotes and with
// contained \ and " characters escaped.
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

// See tutl.EscapeNewline() for documentation.
func (o *Options) EscapeNewline(b bool) {
	if b {
		o.doNotEscape = ' '
	} else {
		o.doNotEscape = '\n'
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
// identical to 'fmt.Sprint()'.  S() never inserts spaces between your values
// (if you want spaces, it is easy for you to add them).  S() puts single
// quotes around 'byte' (and 'uint8') values.  S() treats '[]byte' values
// like 'string's.  S() puts double quotes around '[]byte' and 'error' values
// (escaping enclosed " and \ characters).
//
// S() escapes control characters except for newlines [but see
// EscapeNewline()].  S() also escapes non-UTF-8 byte sequences.
//
// If S() is passed a single argument that is a 'string', then it will put
// double quotes around it and escape any contained " and \ characters.
//
// Note that S() does not put single quotes around 'rune' values as 'rune'
// is just an alias for 'int32' so S('x') == S(int32('x')) == "120" while
// S("x"[0]) == S(byte('x')) == S(uint8('x')) = "'x'".
//
func S(vs ...interface{}) string {
	return Default.S(vs...)
}

// See tutl.S() for documentation.
func (o Options) S(vs ...interface{}) string {
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
			} else if r < 32 && r != o.doNotEscape || 0x7f <= r {
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
// Note that you pass 'want' before 'got' when calling Is() because the
// 'want' value is often a simple constant while 'got' can be a complex
// call and code is easier to read if you put shorter things first.  But
// the output shows 'got' before 'want' as "Got X not Y" is the shortest
// way to express that concept in English.
//
// Is() returns whether the test passed, which is useful for skipping tests
// that would make no sense to run given a prior failure or to display extra
// debug information only when a test fails.
//
func Is(want, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	return Default.Is(want, got, desc, t)
}

// See tutl.Is() for documentation.
func (o Options) Is(want, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	vwant := o.V(want)
	vgot := o.V(got)
	if vwant == vgot {
	//  t.Log("want:", vwant, " got:", vgot, " for:", desc)
		return true
	}
	line := "Got " + o.S(got) + " not " + o.S(want) + " for " + desc + "."
	wid := utf8.RuneCount([]byte(line))
	if strings.Contains(line, "\n") {
		wid = 1 + o.LineWidth // Force multi-line output
	}
	if wid <= o.LineWidth-20 {
		t.Error(line)
	} else if wid <= o.LineWidth {
		t.Error("\n" + line)
	} else {
		t.Errorf("\nGot %s\nnot %s\nfor %s.", o.S(got), o.S(want), desc)
	}
	return false
}

// IsNot() tests that the first two arguments are converted to different
// strings by V().  If they are not, then a diagnostic is displayed which
// also causes the unit test to fail.  The diagnostic is similar to
// "Got unwanted {got} for {desc}.\n" except that S() is used for 'got' so
// control characters will be escaped and their values may be in quotes.
//
// IsNot() returns whether the test passed, which is useful for skipping
// tests that would make no sense to run given a prior failure.
//
func IsNot(hate, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	return Default.IsNot(hate, got, desc, t)
}

// See tutl.IsNot() for documentation.
func (o Options) IsNot(hate, got interface{}, desc string, t TestingT) bool {
	t.Helper()
	vhate := o.V(hate)
	vgot := o.V(got)
	if vhate != vgot {
	//  t.Log("hate:", vhate, " got:", vgot, " for:", desc)
		return true
	}
	line := "Got unwanted " + o.S(got) + " for " + desc + "."
	t.Error(line)
	return false
}

// Like() is most often used to test error messages (or other complex
// strings).  It lets you perform multiple tests against a single value.
// Each test checks that the value converts into a string that either
// contains a specific sub-string (ignoring letter case) or that it matches
// a regular expression.  You must pass at least one string to be matched.
//
// Strings that start with "*" have the "*" stripped before a substring match
// is performed (ignoring letter case).  If a string does not start with a
// "*", then it must be a valid regular expression that will be matched
// against the value's string representation.
//
// Except that strings that start with "!" have that stripped before checking
// for a subsequent "*".  The "!" negates the match so that the test will
// only pass if the string does not match.  To specify a regular expression
// that starts with a "!" character, simply escape it as `\!` or "[!]".
//
// Like() returns the number of matches that failed.
//
// If 'got' is 'nil', the empty string, or becomes the empty string, then
// no comparisons are done and a single failure is reported (but the number
// returned is the number of match strings as it is assumed that none of
// them would have matched the empty string).
//
func Like(got interface{}, desc string, t TestingT, match ...string) int {
	t.Helper()
	return Default.Like(got, desc, t, match...)
}

// See tutl.Like() for documentation.
func (o Options) Like(
	got interface{}, desc string, t TestingT, match ...string,
) int {
	t.Helper()
	if 0 == len(match) {
		t.Errorf("Called Like() with too few arguments in test code.")
		return 1
	}

	sgot := o.V(got)
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
