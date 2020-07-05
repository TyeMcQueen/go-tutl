// Use internal/test_int.go to test ShowStackOnInterrupt() and AtInterrupt().
package tutl_test

import (
	"bytes"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/TyeMcQueen/go-tutl"
)


func TestInt(t *testing.T) {
	// Compile internal/test_int.go w/ race condition checking enabled:
	u := tutl.New(t)
	cmd := exec.Command("go", "build", "-race", "./internal/test_int.go")
	if ! u.Is(nil, cmd.Run(), "go-build works") {
		return
	}

	// Run test_int and then interrupt it before it finishes:
	func() {
		cmd = exec.Command("./test_int")
		out := new(bytes.Buffer)
		cmd.Stdout = out
		err := new(bytes.Buffer)
		cmd.Stderr = err
		if ! u.Is(nil, cmd.Start(), "spawn ./test_int") {
			return
		}
		time.Sleep(3*time.Second/4)
		if ! u.Is(nil, cmd.Process.Signal(syscall.SIGINT), "kill INT works") {
			return
		}
		exit := cmd.Wait()
		ee, ok := exit.(*exec.ExitError)
		if ! u.Is(true, ok, "./test_int got exit error") {
			u.Is("not this", exit, "how ./test_int failed wrong")
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
		if ! u.Is(nil, cmd.Run(), "2nd ./test_int succeeded") {
			u.IsNot("", err, "2nd ./test_int stderr")
			return
		}
		u.Like(out, "got output from AtInterrupt calls",
			`^Loaded,,,\s+Counting,,,\s+$`,
		)
		u.Is("", err, "got stack traces")
	}()
}
