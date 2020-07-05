// This is just a sample program that uses ShowStackOnInterrupt() and
// AtInterrupt().  It is used by int_test.go to check for data races
// and to verify the basic functionality of those two functions.

package main

import (
	"fmt"
	"time"

	u "github.com/TyeMcQueen/go-tutl"
)


func note(s string) {
	u.AtInterrupt(func() {
		fmt.Printf("AtInterrupt(%s)\n", s)
	})
}


func main() {
	go u.ShowStackOnInterrupt()
	fmt.Println("Loaded,,,")
	c := 0
	u.AtInterrupt(func() {
		fmt.Printf("Ran %d extras\n", c)
	})
	note("Second")
	note("Third")
	fmt.Println("Counting,,,")
	for i := 0; i < 100; i++ {
		u.AtInterrupt(func() {
			c++
		})
		time.Sleep(20*time.Millisecond)
	}
}
