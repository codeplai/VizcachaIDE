package main

import "fmt"

func main() {
	// Simple loop example
	for i := 1; i <= 5; i++ {
		fmt.Printf("Count: %d\n", i)
	}

	// Array iteration
	fruits := []string{"apple", "banana", "orange"}
	for index, fruit := range fruits {
		fmt.Printf("%d: %s\n", index, fruit)
	}
}
