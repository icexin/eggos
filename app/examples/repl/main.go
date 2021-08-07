package main

import (
	"crypto/sha1"
	"fmt"
)

func main() {
	var s string
	for {
		fmt.Print(">>> ")
		fmt.Scan(&s)
		sum := sha1.Sum([]byte(s))
		fmt.Printf("sha1(%s) = %x\n", s, sum)
	}
}
