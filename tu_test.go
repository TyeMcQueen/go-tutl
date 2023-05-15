package tutl_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	u "github.com/Unity-Technologies/go-tutl-internal"
)

func TestMain(m *testing.M) {
	go u.ShowStackOnInterrupt()
	os.Exit(m.Run())
}

func TestOptions(t *testing.T) {
	o := u.New(t)

	u.Is(u.S("hi"), o.S("hi"), "o.S", t)
	o.Is(u.V(byte(32)), o.V(byte(32)), "o.V")
	o.Is(u.DoubleQuote("hi"), o.DoubleQuote("hi"), "o.DoubleQuote")
	o.Is(o.Escape('\t'), u.Escape('\t'), "o.Escape")
	o.Is(o.Rune('\r'), u.Rune('\r'), "o.Rune")
	o.Is(o.Char('\n'), u.Char('\n'), "o.Char")

	u.EscapeNewline(true)
	defer u.EscapeNewline(false)
	p := u.New(t)
	u.Is("\"\\n\"", u.S("\n"), "u escapes", t)
	u.Is("\"\\n\"", p.S("\n"), "p inherits", t)
	u.Is("\"\n\"", o.S("\n"), "o default", t)
	p.EscapeNewline(false)
	u.Is("\"\\n\"", u.S("\n"), "u unchanged", t)
	u.Is("\"\n\"", p.S("\n"), "p changed", t)
	u.Is("\"\n\"", o.S("\n"), "o unchanged", t)
}

func TestS(t *testing.T) {
	if u.V(true) == u.V(false) {
		t.Fatalf("Too broken to even test itself.  true=(%s) false=(%s)\n",
			u.V(true), u.V(false))
	}
	u.Is("true", true, "true", t)
	u.Is(0, "0", "zero", t)
	u.Is("1.2", 1.20, "1.20", t)
	u.Is("\r", []byte("\r"), "V []byte like string", t)
	u.Is(10, rune('\n'), `rune '\n'`, t)
	u.Is(120, 'x', `rune 'x' is number`, t)

	u.Is("1.23456789012", u.V(1.234567890123456789), "V float64", t)
	u.Is("0,1.23456789012", u.V([]float64{0, 1.234567890123456789}), "V []float64", t)
	u.Is("1.2346", u.V(float32(1.23456789)), "V float32", t)
	u.Is("1.23456789012", u.S(1.234567890123456789), "S float64", t)

	u.Is(`>"hi"`, u.S(">", []byte("hi")), `"hi" []byte`, t)
	u.Is(`>"Oops"`, u.S(">", fmt.Errorf("Oops")), `"Oops" error`, t)
	u.Is(`>str`, u.S(">", "str"), `not alone "str" string is not quoted`, t)
	u.Is(`"str"`, u.S("str"), `lonely "str" string is quoted`, t)

	u.Is(`>"\"h\\i\""`, u.S(">", []byte(`"h\i"`)), "`\"h\\i\"` []byte", t)
	u.Is(`>"\"\\"`, u.S(">", fmt.Errorf(`"\`)), "`\"\\` error", t)
	u.Is(`"\\\""`, u.S(`\"`), "lonely `\"` string is quoted", t)

	u.Is("'x'", u.S("x"[0]), "S 'x' byte", t)
	u.Is("'\\x00'", u.S(byte(0)), "S 0 byte", t)
	u.Is("'\\n'", u.S("\n"[0]), "S '\n' byte is escaped", t)
	u.Is(`'\xFF'`, u.S(byte(0xFF)), "S '\xFF' byte is escaped", t)

	u.Is(`"\u009B"`, u.S("\u009B"), "S \u009B string is escaped", t)
	u.Is("\"\u00FF\"", u.S("\u00FF"), "S \u00FF string is not escaped", t)
	u.Is(`"\xFF"`, u.S("\xFF"), "S \xFF string is escaped", t)

	u.Is(1, len(u.V("\n")), "V no esc lf string", t)
	u.Is(`'\n'`, u.S("\n"[0]), "S esc lf byte", t)
	u.Is("\"\n\"", u.S("\n"), "S default no esc lf string", t)
	u.EscapeNewline(true)
	u.Is(`"\n"`, u.S("\n"), "S requested esc lf string", t)
	u.EscapeNewline(false)

	u.Is("AB", u.S("A", "B"), "simple concat", t)
	u.Is("'A''B'", u.S("A"[0], "B"[0]), "simple concat not strings", t)
	u.Is("\\xA0", u.S("\xA0", ""), "0xA0 binary string", t)
}

func TestRune(t *testing.T) {
	u.Is(`' '`, u.Rune(32), "' ' rune", t)
	u.Is(`'~'`, u.Rune('~'), "~ rune", t)
	u.Is(`'''`, u.Rune('\''), "' rune", t)
	u.Is(`'\'`, u.Rune('\\'), "' rune", t)

	u.Is("'\u00A0'", u.Rune(0xA0), "NBSp rune is not escaped", t)
	u.Is(`'\xA0'`, u.Char(0xA0), "0xA0 byte", t)
	u.Is(`'\xC2'`, u.Char("\u00A0"[0]), "NBSp 1st byte", t)
	u.Is(`'\xA0'`, u.Char("\u00A0"[1]), "NBSp 2nd byte", t)

	u.Is(`'\n'`, u.Rune(10), `\n rune`, t)
	u.Is(`'\r'`, u.Rune('\r'), `\r rune`, t)
	u.Is(`'\t'`, u.Rune('\t'), `\t rune`, t)

	u.Is(`'\x00'`, u.Rune(0), "0 rune", t)
	u.Is(`'\x07'`, u.Rune('\a'), `\\a rune`, t)
	u.Is(`'\x08'`, u.Rune('\b'), `\\b rune`, t)
	u.Is(`'\x0B'`, u.Rune('\v'), `\\v rune`, t)
	u.Is(`'\x0C'`, u.Rune('\f'), `\\f rune`, t)
	u.Is(`'\x1B'`, u.Rune('\x1b'), "0x1B rune", t)
	u.Is(`'\x1F'`, u.Rune('\x1f'), "0x1F rune", t)
	u.Is(`'\x7F'`, u.Rune('\x7F'), "del rune", t)
	u.Is(`'\u009F'`, u.Rune('\x9f'), "0x9F rune is escaped", t)
	u.Is("'\u00FE'", u.Rune('\xFE'), "0xFE rune is not escaped", t)

	u.Is(`'\xFE'`, u.Char('\xFE'), "0xFE byte", t)
}

type mock struct {
	fails  int
	output []string
}

func (m *mock) Failed() bool { return false }
func (m *mock) Helper()      {}
func (m *mock) clear()       { m.output = m.output[:0]; m.fails = 0 }

func (m *mock) Error(args ...interface{}) {
	m.fails++
	m.Log(args...)
}

func (m *mock) Errorf(format string, args ...interface{}) {
	m.fails++
	m.Logf(format, args...)
}

func (m *mock) Log(args ...interface{}) {
	m.output = append(m.output, fmt.Sprintln(args...))
}

func (m *mock) Logf(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(line, "\n") {
		line = line + "\n"
	}
	m.output = append(m.output, line)
}

func (m *mock) isOutput(desc string, t *testing.T, want ...string) {
	t.Helper()
	if u.Is(len(want), len(m.output), desc+" count", t) {
		for i, o := range want {
			if strings.HasSuffix(m.output[i], "\n") {
				m.output[i] = m.output[i][:len(m.output[i])-1]
			}
			u.Is(o, m.output[i], u.S(desc, " ", i), t)
		}
	} else {
		t.Log("Surprise output:\n", m.output)
	}
	m.clear()
}

func (m *mock) likeOutput(desc string, t *testing.T, likes ...string) {
	t.Helper()
	if u.Is(1, len(m.output), desc+" count", t) {
		u.Like(m.output[0], desc, t, likes...)
	} else {
		t.Log("Surprise output:\n", m.output)
	}
	m.clear()
}

func TestOutput(t *testing.T) {
	m := new(mock) // Mock controller
	s := u.New(m)  // Simulated tester

	u.Is(false, s.Is(true, false, "anti-tautology"), "Is false", t)
	m.isOutput("simple out", t, "Got false not true for anti-tautology.")

	s.Is("longish stuff", "longer stuff", "were stuff longer or longish")
	m.isOutput("longish out", t,
		"\n"+`Got "longer stuff" not "longish stuff" for `+
			`were stuff longer or longish.`)

	s.Is("longish stuff", "longer stuffy", "were stuff longer or longish")
	m.isOutput("longer out", t,
		"\nGot \"longer stuffy\""+
			"\nnot \"longish stuff\""+
			"\nfor were stuff longer or longish.")

	s.Is("two\nlines", "one line", "multi-line")
	m.isOutput("longer out", t,
		"\nGot \"one line\""+
			"\nnot \"two"+
			"\n....lines\""+
			"\nfor multi-line.")

	if h, err := os.Open("go.mod"); err != nil {
		t.Error("Failed to read go.mod file: ", err)
	} else {
		got := io.Reader(h)
		u.Is(true, s.HasType("*os.File", got, "underlying type"), "hastype", t)
		u.Is(true, s.HasType("*io.Reader", &got, "interface type"), "hastype &", t)
		got = nil
		u.Is(true, s.HasType("nil", got, "nil type"), "hastype nil", t)
		m.isOutput("hastype, no output", t)
		u.Is(false, s.HasType("os.File", got, "not type"), "hastype fail", t)
		m.isOutput("hastype output", t,
			"Got \"nil\" not \"os.File\" for not type.")
	}

	s.Is(true, s.Circa(3, 1.23456, 1.234567, "circa"), "circa")
	m.isOutput("circa, no output", t)

	s.Is(false, s.Circa(3, 1.23456, 1.2356, "circa"), "circa")
	m.isOutput("circa, no output", t,
		"Got 1.24 not 1.23 for circa.")

	u.Is(1, s.Like("", "description"), "1 failure if like no strings", t)
	m.likeOutput("like no strings", t,
		"*called ", " Like[(][)]", "*too few arg", "*in test code")

	u.Is(1, s.Like("foo", "description", "fo+", "+fo"), "1 of 2 regex bad", t)
	m.likeOutput("1 of 2 regex bad out", t,
		"*invalid regexp ", "*(+fo)", "in test code")

	u.Is(2, s.Like("", "empty", "*M", "*T"), "all fail for empty", t)
	m.likeOutput("empty string out", t,
		"*no string ", " Like[(][)]", "* got empty string.")

	u.Is(2, s.Like(error(nil), "no err", "*M", "*T"), "all fail for nil err", t)
	m.likeOutput("empty string out", t,
		"*no string ", " Like[(][)]", "* got nil.")

	u.Is(2, s.Like(nil, "nil", "*M", "*T"), "all fail for nil interface", t)
	m.likeOutput("empty string out", t,
		"*no string ", " Like[(][)]", "* got nil.")

	u.Is(2, s.Like(fmt.Errorf(""), `""`, "*M", "*T"), `all fail for ""`, t)
	m.likeOutput(`became "" out`, t,
		"*no string ", " Like[(][)]", "* got blank.")

	u.Is(0, s.Like("hello", "hello", "l{2,}", "*LL"), "0 for pass", t)
	if !u.Is(0, len(m.output), "no output for success", t) {
		t.Log("Surprise output:\n", m.output)
		m.clear()
	}

	for _, short := range []string{"", "!"} {
		u.Is(1, s.Like("Too short", "short", short), `just "`+short+`"`, t)
		m.likeOutput(`just "`+short+`" output`, t,
			"[Mm]atch strings passed to Like[(][)]",
			"*must not be empty",
			`*nor "!"`,
		)
	}

	u.Is(0, s.Like("Success\n", "success", "*success", "!*error", "!!"),
		"negated", t)
	if !u.Is(0, len(m.output), "no output for success 2", t) {
		t.Log("Surprise output:\n", m.output)
		m.clear()
	}

	u.Is(3, s.Like("Failed!\nError: ...\n", "success",
		"*success", "!*error", "!!"), "negated failure", t)
	m.isOutput("negated failure output", t,
		"No <success>...",
		"and Found unwanted <error>...",
		"and Like unwanted /!/...",
		"In <Failed!\nError: ...\n> for success.",
	)

	u.Is(2, s.Like("good bye", "bye", "o{2,}", "*db", "Bye"), "2 of 3 fail", t)
	m.isOutput("2 of 3 not like out", t,
		"No <db>...",
		"and Not like /Bye/...",
		"In <good bye> for bye.")

	u.Is(false, s.Is("hi\n", "high\n", "newlines"), "false newlines", t)
	m.isOutput("newlines out", t,
		"\nGot \"high\n....\"\nnot \"hi\n....\"\nfor newlines.")

	u.Is(2, s.Like("hi\n", "like lf", "*high", "Hi"), "2 of 2 newlines", t)
	m.isOutput("newlines out", t,
		"No <high>...",
		"and Not like /Hi/...",
		"In <hi\n> for like lf.")

	s.SetLineWidth(0)
	u.Is(false, s.Is(5, 2+2, "math joke"), "joke is false", t)
	m.isOutput("joke out", t, "\nGot 4\nnot 5\nfor math joke.")
}
