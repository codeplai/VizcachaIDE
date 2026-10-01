package main

import "fmt"

func greet(name string) string {
	return "hola " + name
}

func main() {
	x := 1
	fmt.Println(greet("vizcacha"))
}
