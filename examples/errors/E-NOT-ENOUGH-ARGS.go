// E-NOT-ENOUGH-ARGS: a function is called with fewer arguments than it needs.
package main

import "fmt"

func add(a int, b int) int {
	return a + b
}

func main() {
	fmt.Println(add(1))
}
