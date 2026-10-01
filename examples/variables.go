package main

import "fmt"

func main() {
	// Basic types
	var name string = "VizcachaIDE"
	var version float64 = 1.0
	var isReady bool = true
	count := 42

	fmt.Println("IDE Name:", name)
	fmt.Println("Version:", version)
	fmt.Println("Ready:", isReady)
	fmt.Println("Count:", count)

	// Slice
	numbers := []int{1, 2, 3, 4, 5}
	fmt.Println("Numbers:", numbers)

	// Map
	features := map[string]bool{
		"syntax_highlighting": true,
		"debugging":           true,
		"auto_complete":       false,
	}
	fmt.Println("Features:", features)

	// Struct
	type Person struct {
		Name string
		Age  int
	}

	person := Person{
		Name: "Go Developer",
		Age:  25,
	}
	fmt.Printf("Person: %+v\n", person)
}
