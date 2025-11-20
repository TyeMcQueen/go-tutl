package tutl

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

// Stupid test to improve coverage stats since int_test.go doesn't know
// about the code being tested in internal/test_int.go.
func TestSignal(t *testing.T) {
	AtInterrupt(func() {})
	go ShowStackOnInterrupt()
	go ShowStackOnInterrupt()
	defer func() {
		p := recover()
		if str, _ := p.(string); str == "Interrupted" {
			// Let test pass
			fmt.Printf("Caught interrupt-caused panic.\n")
		} else if nil != p {
			fmt.Printf("Caught other panic.\n")
			panic(p)
		}
	}()
	_sigs <- syscall.SIGKILL
	time.Sleep(10 * time.Millisecond)
}
