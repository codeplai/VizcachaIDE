// E-ASSIGN-MISMATCH: strconv.Atoi returns two values but only one variable receives them.
package main

import (
	"fmt"
	"strconv"
)

func main() {
	number := strconv.Atoi("42")
	fmt.Println(number)
}
