// V-PRINTF-ARGS: the format has a %d verb but no value is passed (found by: go vet).
package main

import "fmt"

func main() {
	fmt.Printf("You are %d years old\n")
}
