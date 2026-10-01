// P-INDEX-RANGE: reading an element past the end of a slice.
package main

import "fmt"

func main() {
	numbers := []int{1, 2, 3}
	position := 5
	fmt.Println(numbers[position])
}
