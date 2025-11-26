package tutl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	"sigs.k8s.io/yaml"
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
	// gets split onto multiple lines instead.  If that string is longer
	// than LineWidth, then it gets split into "Got ...\nnot ...\n...".
	//
	// This also happens if you aren't escaping newlines and either value
	// contains a newline (and the newlines get indentation added so that
	// the output is easier to understand).
	//
	// If the diagnostic line is no longer than LineWidth but is longer than
	// LineWidth-PathLength, then a newline gets prepended to it as the
	// prepended source info would likely cause the diagnostic to wrap.
	//
	LineWidth int

	// PathLength is the maximum expected length of the path to the
	// *_test.go file being run plus the line number that 'go test'
	// prepends to each diagnostic.  It defaults to 20.
	//
	PathLength int

	// Digits32 specifies how many significant digits to use when comparing
	// 'float32' values.  In particular, if a 'float32' or '[]float32' value
	// is passed to V(), then no more than Digits32 significant digits are
	// used in the resulting string.  Other data structures that contain
	// 'float32' values are not impacted.
	//
	// If Digits32 is 0, then the default value of 5 is used.  If Digits32
	// is negative or more than 7, then 'fmt.Sprint()' is used which may
	// use even 8 digits for some values (such as 1/3) so that 2 'float32'
	// values that are even very slightly different will produce different
	// strings (a 'float32' is accurate to only slightly more than 7 digits).
	//
	Digits32 int

	// Digits64 specifies how many significant digits to use when comparing
	// 'float64' values.  In particular, if a 'float64' or '[]float64' value
	// is passed to V(), then no more than Digits64 significant digits are
	// used in the resulting string.  Other data structures that contain
	// 'float64' values are not impacted.
	//
	// If Digits64 is 0, then the default value of 12 is used.  If Digits64
	// is negative or more than 16, then 'fmt.Sprint()' is used which may
	// use up to 16 digits so that 2 'float64' values that are even very
	// slightly different will produce different strings (a 'float64' is
	// accurate to only slightly less than 16 digits).
	//
	Digits64 int
}

const MaxDigits32 = 7
const MaxDigits64 = 15

// The 'tutl.Default' global contains the user preferences to be used unless
// you make a copy and use it, such as via New() (see Options for more).
//
var Default = Options{
	doNotEscape: '\n', LineWidth: 72, PathLength: 20, Digits32: 5, Digits64: 12}

// V() just converts a value to a string.  It is similar to 'fmt.Sprint(v)'.
// But it treats '[]byte' values as 'string's.  It also (by default) uses
// fewer significant digits when converting 'float32', 'float64',
// '[]float32', and '[]float64' values (see Options for details).
//
func V(v any) string {
	return Default.V(v)
}

// See tutl.V() for documentation.
func (o Options) V(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case float32:
		d := o.Digits32
		if 0 == d {
			d = 5
		} else if d < 0 || MaxDigits32 < d {
			return fmt.Sprint(t)
		}
		return fmt.Sprintf("%.*g", d, t)
	case float64:
		d := o.Digits64
		if 0 == d {
			d = 12
		} else if d < 0 || MaxDigits64 < d {
			return fmt.Sprint(t)
		}
		return fmt.Sprintf("%.*g", d, t)
	case []float32:
		s := make([]string, len(t))
		for i, f := range t {
			s[i] = o.V(f)
		}
		return strings.Join(s, ",")
	case []float64:
		s := make([]string, len(t))
		for i, f := range t {
			s[i] = o.V(f)
		}
		return strings.Join(s, ",")
	}
	return fmt.Sprint(v)
}

// DoubleQuote() returns the string enclosed in ‟ and ”, or, if the string
// contains either of those characters, then it will enclose it in double
// quotes (") and with contained \ and " characters escaped (preceeded by \).
//
func DoubleQuote(s string) string {
	if ! strings.ContainsAny(s, "‟”") {
		return "‟" + s + "”"
	}
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", s)
}

// ReplaceNewlines() returns a string with each newline replaced with either
// an escaped newline (a \ then an 'n') or with the string "\n...." (so that
// subsequent lines of a multi-line value are indented to make them easier
// to distinguish from subsequent lines of a test diagnostic).
//
func ReplaceNewlines(s string) string { return Default.ReplaceNewlines(s) }

// See tutl.ReplaceNewlines() for documentation.
func (o *Options) ReplaceNewlines(s string) string {
	if '\n' == o.doNotEscape {
		return strings.Replace(s, "\n", "\n....", -1)
	}
	return strings.Replace(s, "\n", "\\n", -1)
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

func SetDigits32(d int) { Default.SetDigits32(d) }
func SetDigits64(d int) { Default.SetDigits64(d) }

func (o *Options) SetDigits32(d int) { o.Digits32 = d }
func (o *Options) SetDigits64(d int) { o.Digits64 = d }

// Escape() returns a string containing the passed-in rune, unless it is a
// control character.  Runes '\n', '\r', and '\t' each return a 2-character
// string (\n, \r, or \t).  Other 7-bit control characters are turned into
// strings like \x1B.  The 8-bit control characters are turned into strings
// like \u009B.  EscapeNewline(false) does not affect Escape().
//
func Escape(r rune) string {
	switch r {
	case '\n':
		return `\n`
	case '\r':
		return `\r`
	case '\t':
		return `\t`
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

// ListToJson() takes a list of values alternating between a string containing
// a fragment of YAML followed by an arbitrary value to be converted to a JSON
// string. The resulting strings are concatenated together, interpretted as
// YAML, and converted to JSON (which is returned). This is a very compact
// form of templating for producing JSON. For example:
//
//      tutl.ListToJson("{ID:", id, ",Name:", name, ",Rank:", rank, "}")
//
// might return
//
//      `{ "ID": "876-TTJ-23", "Name": "Ray \"Bob\" Keen", "Rank": 12 }`
//
// A space is always added after JSON items and is added before if the prior
// character was not from " \t\n" (since YAML treats `Key:"Val"` as 1 string).
//
// An argument that is a string cast to the 'tutl.LiteralYaml' type is always
// treated as a fragment of YAML to be appended unchanged and also resets
// argument counting so the next argument must be a string to be concatenated
// unchanged. If this 2nd string is not also cast to LiteralYaml, then the
// following argument will be converted to JSON, resuming the alternating
// pattern. This can be useful when you are building a long YAML/JSON string:
//
//      const na = tutl.LiteralYaml(""), nl = tutl.LiteralYaml("\n")
//
//      tutl.Has(tutl.ListToJson(
//          "{ID:", id, ",Name:", name, ",Ranks:", ranks, ",", na,
//          "Active:", active, ",Exclude:[]}"),
//          got, desc, t)
//
//      tutl.Has(tutl.ListToJson(
//          "Config:", nl,
//          "  Name:", configName, nl,
//          "  Tier: dev", nl,
//          ...
//          "  Domains: ", domains),
//          got, desc, t)
//
// If an attempt to convert an argument to JSON fails, then a test failure
// will be logged and 'nil' will be returned. Similarly, if the resulting
// YAML is invalid, then a test failure is logged along with the full
// not-YAML string (and 'nil' is returned).
//
func ListToJson(t TestingT, args ...any) []byte {
	t.Helper()
	return Default.ListToJson(t, args...)
}

func (o Options) ListToJson(t TestingT, args ...any) []byte {
	t.Helper()
	buf := bytes.Buffer{}
	literal := true
	for _, arg := range args {
		if v, ok := arg.(LiteralYaml); ok {
			buf.WriteString(string(v))
			literal = true
			continue
		} else if ! literal {
			js, err := json.Marshal(arg)
			if ! o.Is(
				nil, err, fmt.Sprintf(
					"ListToJson(): Can't convert type %T to JSON: %s", arg, err,
				), t,
			) {
				return nil
			}
			end := buf.Len() - 1
			if 0 <= end {
				last := buf.Bytes()[end]
				if ! strings.Contains(" \t\n", string(last)) {
					buf.WriteString(" ")
				}
			}
			buf.Write(js)
			buf.WriteString(" ")
			literal = true
			continue
		}
		literal = false
		switch v := arg.(type) {
		case string:
			buf.WriteString(v)
		case []byte:
			buf.Write(v)
		default:
			if ! o.Is(
				"string|[]byte", fmt.Sprintf("%T", arg),
				"ListToJson(): Invalid type for literal YAML fragment", t,
			) {
				return nil
			}
		}
	}
	js, err := yaml.YAMLToJSON(buf.Bytes())
	if ! o.Is(
		nil, err, fmt.Sprintf("ListToJson(): Invalid YAML: %v", err), t,
	) {
		t.Log(S("Invalid YAML=(\n", buf.Bytes(), "\n)."))
		return nil
	}
	return js
}

// GetPanic() calls the passed-in function and returns 'nil' or the argument
// that gets passed to panic() from within it.  This can be used in other
// test functions, for example:
//
//      u := tutl.New(t)
//      u.Is(nil, u.GetPanic(func(){ obj.Method(nil) }), "Method panic")
//
func GetPanic(run func()) (failure any) {
	defer func() {
		failure = recover()
	}()
	run()
	return
}

// S() returns a single string composed by converting each argument into
// a string and concatenating all of those strings.  It is similar to but not
// identical to 'fmt.Sprint()'.  S() never inserts spaces between your values
// (if you want spaces, it is easy for you to add them).  S() puts single
// quotes around 'byte' (and 'uint8') values.  S() treats '[]byte' values
// like 'string's.  S() puts double quotes around '[]byte' and 'error' values
// [see DoubleQuote()].
//
// S() escapes control characters except for newlines [but see
// EscapeNewline()].  S() also escapes non-UTF-8 byte sequences.
//
// If S() is passed a single argument that is a 'string', then it will put
// double quotes around it [see DoubleQuote()].
//
// See V() for how 'float32', 'float64', '[]float32', or '[]float64' values
// are converted.
//
// Note that S() does not put single quotes around 'rune' values as 'rune'
// is just an alias for 'int32' so S('x') == S(int32('x')) == "120" while
// S("x"[0]) == S(byte('x')) == S(uint8('x')) = "'x'".
//
func S(vs ...any) string {
	return Default.S(vs...)
}

// See tutl.S() for documentation.
func (o Options) S(vs ...any) string {
	ss := make([]string, len(vs))
	for j, ix := range vs {
		s := ""
		switch v := ix.(type) {
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
		case float32, float64, []float32, []float64:
			s = o.V(ix)
		default:
			s = fmt.Sprintf("%v", ix)
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
func Is(want, got any, desc string, t TestingT) bool {
	t.Helper()
	return Default.Is(want, got, desc, t)
}

// See tutl.Is() for documentation.
func (o Options) Is(want, got any, desc string, t TestingT) bool {
	t.Helper()
	vwant := o.V(want)
	vgot := o.V(got)
	if vwant == vgot {
		//  t.Log("want:", vwant, " got:", vgot, " for:", desc)
		return true
	}
	sGot := o.S(got)
	sWant := o.S(want)
	line := "Got " + sGot + " not " + sWant + " for " + desc + "."
	wid := utf8.RuneCountInString(line)
	if strings.Contains(line, "\n") {
		sGot = o.ReplaceNewlines(sGot)
		sWant = o.ReplaceNewlines(sWant)
		t.Errorf("\nGot %s\nnot %s\nfor %s.", sGot, sWant, desc)
		return false
	}
	if wid <= o.LineWidth-o.PathLength {
		t.Error(line)
	} else if wid <= o.LineWidth {
		t.Error("\n" + line)
	} else {
		t.Errorf("\nGot %s\nnot %s\nfor %s.", sGot, sWant, desc)
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
func IsNot(hate, got any, desc string, t TestingT) bool {
	t.Helper()
	return Default.IsNot(hate, got, desc, t)
}

// See tutl.IsNot() for documentation.
func (o Options) IsNot(hate, got any, desc string, t TestingT) bool {
	t.Helper()
	vhate := o.V(hate)
	vgot := o.V(got)
	if vhate != vgot {
		//  t.Log("hate:", vhate, " got:", vgot, " for:", desc)
		return true
	}
	t.Error(
		"Got unwanted " + o.ReplaceNewlines(o.S(got)) + " for " + desc + ".")
	return false
}

// HasType() tests that the type of the 2nd argument ('got') is equal to the
// first argument ('want', a string).  That is, it checks that
// 'want == fmt.Sprintf("%T", got)'.  If not, then a diagnostic is displayed
// which also causes the unit test to fail.
//
// The diagnostic is similar to "Got {got} not {want} for {desc}.\n" where
// '{got}' is the data type of 'got' and '{want}' is just the 'want' string.
//
// If 'got' is an 'interface' type, then the type string will be the type of
// the underlying object (or "nil").  If you actually wish to compare the
// 'interface' type, then place '&' before 'got' and prepend "*" to 'want':
//
//      got := GetReader() // Returns io.Reader interface to an *os.File
//      tutl.HasType("*os.File", got, "underlying type is *os.File", t)
//      tutl.HasType("*io.Reader", &got, "interface type is io.Reader", t)
//      //            ^            ^ insert these to test interface type
//
// HasType() returns whether the test passed, which is useful for skipping
// tests that would make no sense to run given a prior failure.
//
func HasType(want string, got any, desc string, t TestingT) bool {
	t.Helper()
	return Default.HasType(want, got, desc, t)
}

// See tutl.HasType() for documentation.
func (o Options) HasType(
	want string, got any, desc string, t TestingT,
) bool {
	t.Helper()
	tgot := "nil"
	if nil != got {
		tgot = fmt.Sprintf("%T", got)
	}
	return o.Is(want, tgot, desc, t)
}

// ToMap() takes a value to be converted to a tutl.Map type
// ('map[string]any'). The value can already be a tutl.Map type, can be
// JSON (a string or a []byte), or can be a data type that can be converted
// to JSON.
//
// If 'value' needs to be converted to JSON but marhsaling fails, then
// a test failure is logged and 'nil' is returned. This is also done if
// unmarshaling JSON into the map is required but fails.
//
func ToMap(value any, t TestingT) Map {
	t.Helper()
	return Default.ToMap(value, t)
}

// See tutl.ToMap() for documentation.
func (o Options) ToMap(value any, t TestingT) (retMap Map) {
	t.Helper()
	var js []byte
	switch v := value.(type) {
	case string:
		js = []byte(v)
	case []byte:
		js = v
	case *bytes.Buffer:
		js = v.Bytes()
	default:
		js2, err := json.Marshal(value)
		if ! Is(
			nil, err, fmt.Sprintf(
				"ToMap() JSON marshal of type %T", value),
			t,
		) {
			return nil
		}
		js = js2
	}
	err := json.Unmarshal(js, &retMap)
	if ! Is(nil, err, "ToMap() got invalid JSON", t) {
		t.Log("JSON was: (\n", js, "\n).")
		return nil
	}
	return
}

// Element() looks up an element in a struct/map[string]/JSON 'value' using a
// 'key' string. Element() is designed to be used by Has()/Lacks()/Covers(),
// but can be used directly.
//
// If 'key' starts with ".", then the remainder is treated as a list of
// subkeys separated by ".". tutl.Element(value, ".Config.source") can return:
//
//      value.Config.source
//      value["Config"]["source"]
//      value["Config"].source
//      value.Config["source"]
//
// If a key or subkey gets used against a struct that has no field by that
// name, then a test failure is logged and 'nil' is returned. A lookup on a
// map with no such key simply returns a 'nil' (logging nothing). If the
// final lookup returns a non-scalar (an array, chan, func, map, pointer,
// slice, struct, or interface holding a non-scalar) that holds the zero
// value for that type, then 'nil' is returned instead (and not a non-nil
// interface to a 'nil' of whatever type).
//
func Element(value any, key string, t TestingT) any {
	t.Helper()
	return Default.Element(value, key, t)
}

// See tutl.Element() for documentation.
func (o Options) Element(value any, key string, t TestingT) any {
	t.Helper()
	if ! strings.HasPrefix(key, ".") {
		return o.oneElement(value, key, "", t)
	}
	parts := strings.Split(strings.TrimPrefix(key, "."), ".")
	for _, part := range parts {
		value = o.oneElement(value, part, key, t)
		if value == nil {
			return nil
		}
	}
	return value
}

func (o Options) oneElement(value any, subkey, key string, t TestingT) any {
	t.Helper()
	switch v := value.(type) {
	case map[string]any:
		ret, _ := v[subkey]
		return ret
	}
	refVal := reflect.ValueOf(value)
	if reflect.Map == refVal.Kind() {
		ret := refVal.MapIndex(reflect.ValueOf(subkey))
		if ret.IsZero() {
			return nil
		}
		return ret.Interface()
	} else if reflect.Struct == refVal.Kind() {
		ret := refVal.FieldByName(subkey)
		desc := S("Element(): No '", subkey, "' in struct")
		if key != "" {
			desc = S(desc, " field (key=", key, ")")
		}
		if ! o.Is(true, ret.IsValid(), desc, t) {
			return nil
		}
		return ret.Interface()
	}
	o.Is("Map|Struct", refVal.Kind(), fmt.Sprintf(
		"Element(): Wrong type (%T) to look up '%s' element (key=%s)",
		value, subkey, key), t)
	return nil
}

// Covers() verifies that 'got' is a superset of 'want'. 'want' is a map/JSON
// (see tutl.ToMap) representing required parts of the 'got' data structure.
// The map and 'got' are traversed together [a part of 'got' being accessed
// using the map's key string passed to tutl.Element()]. This traversal
// recurses if the map's value is another map.
//
// When the next value in the map is not another map, the values are
// compared, similar to:
//
//      tutl.Is(want[key], tutl.Element(got, key), ...)
//
// 'got' can be JSON, any 'map' type (not just tutl.Map), or a 'struct' type.
//
// Covers() returns the count of failing tests.
//
func Covers(want any, got any, desc string, t TestingT) int {
	t.Helper()
	return Default.Covers(want, got, desc, t)
}

// See tutl.Covers() for documentation.
func (o Options) Covers(
	want any, got any, desc string, t TestingT,
) (fails int) {
	t.Helper()
	wantMap := o.ToMap(want, t)
	fails += o.oneCover(wantMap, got, desc, "", t)
	return fails
}

func (o Options) oneCover(
	wantMap Map, gotAny any, desc, prefix string, t TestingT,
) int {
	t.Helper()
	fails := 0
	for key, want := range wantMap {
		got := o.Element(gotAny, key, t)
		if w, ok := want.(Map); ok {
			fails += o.oneCover(w, got, desc, prefix + key + ".", t)
			continue
		}
		if ! o.Is(want, got, o.S(desc, ": ", prefix, key), t) {
			fails++
		}
	}
	return fails
}

// Has() takes a value ('got') and a list of key/value pairs. Each key is
// used to access some part of 'got' and compare it to the paired value:
//
//      tutl.Is(value, tutl.Element(got, key), ...)
//
// For example:
//
//      got := GetConfig() // Get something to test
//      tutl.Has("Default config", got, t,
//          "Active", true, "MaxCount", 10_000, ".Tier.Name", "dev")
//
// 'got' can be JSON (see tutl.ToMap) or a 'struct' data type.
// The above code can call the equivalent of (see also tutl.Element):
//
//      tutl.Is(true, got.Active, "Defautl Config: Active", t)
// or
//      tutl.Is(10_000, got["MaxCount"], "Defautl Config: MaxCount", t)
// or
//      tutl.Is("dev", got.Tier.Name, "Defautl Config: Tier.Name", t)
// or
//      tutl.Is("dev", got.Tier["Name"], "Defautl Config: Tier.Name", t)
//
// Has() returns the count of failing tests, or -1 if it was called
// incorrectly.
//
func Has(t TestingT, desc string, got any, pairs ...any) int {
	t.Helper()
	return Default.Has(t, desc, got, pairs...)
}

// See tutl.Has() for documentation.
func (o Options) Has(t TestingT, desc string, gotAny any, pairs ...any) int {
	t.Helper()
	fails := 0
	if 1 == len(pairs) & 1 {
		o.Is("Even number", len(pairs), "Number of 'pairs' scalars to Has()", t)
		return -1
	}
	key := ""
	for i, arg := range pairs {
		if 0 == i & 1 {
			if ! o.HasType("string", arg, "Type of key arg to Has()", t) {
				return -1
			}
			key = arg.(string)
		} else {
			got := o.Element(gotAny, key, t)
			if ! o.Is(arg, got, desc + ": " + key, t) {
				fails += 1
			}
		}
	}
	return fails
}

// Lacks() takes a list of key strings (each of which can contain multiple
// names separated by ".") and an arbitrary value. Element() is used to verify
// that no key leads to a scalar or to a non-zero non-scalar. For each key
// where Element() does not return 'nil', a test failure is logged. Lacks()
// returns the number of failures.
//
//      got := GetConfig() // Get something to test
//      tutl.Lacks(got, "No security overrides", t,
//          "Override", "Auth.Key", "Cert.Insecure")
//
func Lacks(t TestingT, desc string, got any, keys ...string) int {
	t.Helper()
	return Default.Lacks(t, desc, got, keys...)
}

// See tutl.Lacks() for documentation.
func (o Options) Lacks(t TestingT, desc string, got any, keys ...string) int {
	t.Helper()
	fails := 0
	for _, key := range keys {
		if ! o.Is(nil, o.Element(got, key, t), desc + ": " + key, t) {
			fails++
		}
	}
	return fails
}

// Circa() tests that the 2nd and 3rd arguments are approximately equal to
// each other.  If they are not, then a diagnostic is displayed which also
// causes the unit test to fail.
//
// The diagnostic is similar to "Got {got} not {want} for {desc}.\n" where
// 'want' and 'got' are shown formatted via 'fmt.Sprintf("%.*g", digits, v)'.
// They are considered equal if that formatting produces the same string
// for both values.  That is, 'want' and 'got' are considered roughly equal
// if they are the same to 'digits' significant digits.  Passing 'digits' as
// less than 1 or more than 15 is not useful.
//
// Circa() returns whether the test passed, which is useful for skipping
// tests that would make no sense to run given a prior failure or to display
// extra debug information only when a test fails.
//
func Circa(digits int, want, got float64, desc string, t TestingT) bool {
	t.Helper()
	return Default.Circa(digits, want, got, desc, t)
}

// See tutl.Circa() for documentation.
func (o Options) Circa(
	digits int, want, got float64, desc string, t TestingT,
) bool {
	t.Helper()
	swant := fmt.Sprintf("%.*g", digits, want)
	sgot  := fmt.Sprintf("%.*g", digits, got)
	return o.Is(swant, sgot, desc, t)
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
func Like(got any, desc string, t TestingT, match ...string) int {
	t.Helper()
	return Default.Like(got, desc, t, match...)
}

// See tutl.Like() for documentation.
func (o Options) Like(
	got any, desc string, t TestingT, match ...string,
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
		t.Errorf("No string to check what it is Like(); got %s for %s.",
			empty, desc)
		return len(match)
	}

	failed := 0
	invalid := 0
	lgot := strings.ToLower(sgot)
	and := ""
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
				sMatch := o.ReplaceNewlines(m[1:])
				if negate {
					t.Errorf(and+"Found unwanted <%s>...", sMatch)
				} else {
					t.Errorf(and+"No <%s>...", sMatch)
				}
			}
		} else if re, err := regexp.Compile(m); nil != err {
			invalid++
			t.Errorf(and+"Invalid regexp (%s) in test code: %v", m, err)
		} else if negate == ("" != re.FindString(sgot)) {
			failed++
			if negate {
				t.Errorf(and+"Like unwanted /%s/...", m)
			} else {
				t.Errorf(and+"Not like /%s/...", m)
			}
		}
		if 0 < failed {
			and = "and "
		}
	}
	if 0 < failed {
		t.Errorf("In <%s> for %s.", sgot, desc)
	}
	return failed + invalid
}
