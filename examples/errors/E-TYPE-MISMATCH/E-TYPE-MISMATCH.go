// E-TYPE-MISMATCH: a value of one type is stored where another type is expected.
package main

import "fmt"

func main() {
	var age int = "twenty"
	fmt.Println(age)
}
