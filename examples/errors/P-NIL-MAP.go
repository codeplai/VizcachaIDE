// P-NIL-MAP: writing into a map that was never created with make.
package main

func main() {
	var ages map[string]int
	ages["Ana"] = 30
}
