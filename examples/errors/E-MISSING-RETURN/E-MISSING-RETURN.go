// E-MISSING-RETURN: a function that returns a value can end without returning.
package main

import "fmt"

func sign(n int) string {
	if n >= 0 {
		return "positive"
	}
}

func main() {
	fmt.Println(sign(3))
}
