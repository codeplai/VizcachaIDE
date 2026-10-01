// E-TOO-MANY-ARGS: a function is called with more arguments than it accepts.
package main

import "fmt"

func add(a int, b int) int {
	return a + b
}

func main() {
	fmt.Println(add(1, 2, 3))
}
