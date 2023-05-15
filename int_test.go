// Use internal/test_int.go to test ShowStackOnInterrupt() and AtInterrupt().
package tutl_test

import (
	"bytes"
	"io"
	"os/exec"
	"syscall"
	"testing"

	"github.com/Unity-Technologies/go-tutl-internal"
)

type waiter struct {
	out io.Writer
	ch  chan<- bool
}

func (w *waiter) Write(b []byte) (int, error) {
	if nil != w.ch {
		w.ch <- true
		w.ch = nil
	}
	return w.out.Write(b)
}

type responder struct {
	response string
	ch       <-chan bool
}

func (r *responder) Read(b []byte) (int, error) {
	if "" == r.response {
		return 0, io.EOF
	}
	<-r.ch
	n := copy(b, r.response)
	r.response = ""
	return n, nil
}

func TestInt(t *testing.T) {
	// Compile internal/test_int.go w/ race condition checking enabled:
	u := tutl.New(t)
	cmd := exec.Command("go", "build", "-race", "./internal/test_int.go")
	if !u.Is(nil, cmd.Run(), "go-build works") {
		return
	}

	// Run test_int and then interrupt it before it finishes:
	func() {
		cmd = exec.Command("./test_int", "100")
		out := new(bytes.Buffer)
		och := make(chan bool, 1)
		cmd.Stdout = &waiter{out, och}
		err := new(bytes.Buffer)
		cmd.Stderr = err
		ich := make(chan bool, 1)
		cmd.Stdin = &responder{"go\n", ich}
		if !u.Is(nil, cmd.Start(), "spawn ./test_int") {
			return
		}
		sig := <-och
		ich <- sig
		if !u.Is(nil, cmd.Process.Signal(syscall.SIGINT), "kill INT works") {
			return
		}
		exit := cmd.Wait()
		ee, ok := exit.(*exec.ExitError)
		if !u.Is(true, ok, "./test_int got exit error") {
			t.Log("How ./test_int failed: ", exit)
			return
		}
		u.Is("exit status 2", ee, "./test_int failed right")
		u.Like(out, "got output from AtInterrupt calls",
			"*AtInterrupt(Third)",
			"*AtInterrupt(Second)",
			"Ran [0-9]+ extras",
		)
		u.Like(out, "ran AtInterrupt in right order",
			"Third[^_]*Second[^_]*extras")
		u.Like(err, "got stack traces",
			"panic: Interrupted",
			`goroutine [0-9]+ \[running\]`,
		)
		u.Like(err, "no race conditions", "!WARNING: DATA RACE")
	}()

	// Run test_int but don't interrupt it:
	func() {
		cmd = exec.Command("./test_int")
		out := new(bytes.Buffer)
		cmd.Stdout = out
		err := new(bytes.Buffer)
		cmd.Stderr = err
		if !u.Is(nil, cmd.Run(), "2nd ./test_int succeeded") {
			u.IsNot("", err, "2nd ./test_int stderr")
			return
		}
		u.Like(out, "got output from AtInterrupt calls",
			`^Loaded,,,\s+Counting,,,\s+$`,
		)
		u.Is("", err, "got stack traces")
	}()
}
