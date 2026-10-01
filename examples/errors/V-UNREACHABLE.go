// V-UNREACHABLE: a statement after return can never run (found by: go vet).
package main

import "fmt"

func main() {
	fmt.Println("start")
	return
	fmt.Println("never printed")
}
