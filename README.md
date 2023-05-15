# go-tutl

go-tutl is a Tiny Unit Testing Library (for Go).  "Tutl" is also the
Faroese word for "whisper", a nod to it's light-weight nature.

TUTL provides a few helper routines that make simple unit testing in Go
much easier and that encourage you to write tests that, when they fail,
it is easy to figure out what broke.

It started as the 2 small functions that I have felt worth writing every
time I needed to do unit tests in Go, Is() and S().  Plus 2 other small
functions that I usually eventually end up writing when I get further
along with a project, Like() and ShowStackOnInterrupt().  And it has been
collecting small improvements over the years.

Example usage:

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

Sample output from a failing run of the above tests:

    dur_test.go:10: Got "1m 61s" not "1m 1s" for 61.
    dur_test.go:14: Got 3600 not 3605 for '1h 5s'.
    dur_test.go:17:
        Not like /(Unknown|Invalid) unit/...
        and No <fortnight>...
        In <"Bad unit (ortnight) in duration.
        "> for Error from '3 fortnight'.

It also provides a special module to deal with infinite loops in your
code.  If you include:

    import (
        _ "github.com/Unity-Technologies/go-tutl-internal/hang" // ^C gives stack dumps.
    )

in just one of your *_test.go files, then you can interrupt (such as
via typing Ctrl-C) an infinite loop or otherwise hanging test run and be
shown, in response, the stack traces of everything that is running.
