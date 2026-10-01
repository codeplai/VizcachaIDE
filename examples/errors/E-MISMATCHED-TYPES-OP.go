// E-MISMATCHED-TYPES-OP: an int and a string are added together.
package main

import "fmt"

func main() {
	apples := 3
	label := " apples"
	fmt.Println(apples + label)
}
