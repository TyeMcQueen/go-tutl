package tutl_test

import (
    "fmt"
    "os"
    "testing"

    u "github.com/TyeMcQueen/go-tutl"
)


func TestMain(m *testing.M) {
    go u.ShowStackOnInterrupt()
    os.Exit(m.Run())
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

    u.Is("AB", u.S("A","B"), "simple concat", t)
    u.Is("'A''B'", u.S("A"[0],"B"[0]), "simple concat not strings", t)
    u.Is("\\xA0", u.S("\xA0",""), "0xA0 binary string", t)
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

    u.Is(`'\n'`, u.Rune(10),   `\n rune`, t)
    u.Is(`'\r'`, u.Rune('\r'), `\r rune`, t)
    u.Is(`'\t'`, u.Rune('\t'), `\t rune`, t)

    u.Is(`'\x00'`, u.Rune(0),      "0 rune", t)
    u.Is(`'\x07'`, u.Rune('\a'),   `\\a rune`, t)
    u.Is(`'\x08'`, u.Rune('\b'),   `\\b rune`, t)
    u.Is(`'\x0B'`, u.Rune('\v'),   `\\v rune`, t)
    u.Is(`'\x0C'`, u.Rune('\f'),   `\\f rune`, t)
    u.Is(`'\x1B'`, u.Rune('\x1b'), "0x1B rune", t)
    u.Is(`'\x1F'`, u.Rune('\x1f'), "0x1F rune", t)
    u.Is(`'\x7F'`, u.Rune('\x7F'), "del rune", t)
    u.Is(`'\u009F'`, u.Rune('\x9f'), "0x9F rune is escaped", t)
    u.Is("'\u00FE'", u.Rune('\xFE'), "0xFE rune is not escaped", t)

    u.Is(`'\xFE'`, u.Char('\xFE'), "0xFE byte", t)
}
