// P-TYPE-ASSERTION: asserting the wrong type on an interface value.
package main

import "fmt"

func main() {
	var value interface{} = "hello"
	number := value.(int)
	fmt.Println(number)
}
