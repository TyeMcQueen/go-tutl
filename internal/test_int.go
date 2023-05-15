// This is just a sample program that uses ShowStackOnInterrupt() and
// AtInterrupt().  It is used by int_test.go to check for data races
// and to verify the basic functionality of those two functions.

package main

import (
	"fmt"
	"os"
	"time"

	u "github.com/Unity-Technologies/go-tutl-internal"
)

func note(s string) {
	u.AtInterrupt(func() {
		fmt.Printf("AtInterrupt(%s)\n", s)
	})
}

func main() {
	go u.ShowStackOnInterrupt()
	go u.ShowStackOnInterrupt(false)
	fmt.Println("Loaded,,,")
	c := 0
	u.AtInterrupt(func() {
		fmt.Printf("Ran %d extras\n", c)
	})
	note("Second")
	note("Third")
	fmt.Println("Counting,,,")
	max := 10
	if 1 < len(os.Args) {
		max = 100
		os.Stdout.WriteString("Ready? ")
		response := make([]byte, 1024)
		_, _ = os.Stdin.Read(response)
	}
	for i := 0; i < max; i++ {
		u.AtInterrupt(func() {
			c++
		})
		time.Sleep(200 * time.Millisecond)
	}
}
