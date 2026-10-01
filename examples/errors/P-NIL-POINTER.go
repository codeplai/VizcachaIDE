// P-NIL-POINTER: reading through a pointer that is nil.
package main

import "fmt"

type Person struct {
	Name string
}

func main() {
	var p *Person
	fmt.Println(p.Name)
}
