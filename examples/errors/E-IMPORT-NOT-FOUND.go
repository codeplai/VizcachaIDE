// E-IMPORT-NOT-FOUND: the imported package does not exist (the right name is "math").
package main

import (
	"fmt"
	"maths"
)

func main() {
	fmt.Println(maths.Sqrt(16))
}
