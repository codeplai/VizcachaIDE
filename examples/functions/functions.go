package main

import "fmt"

// Add two numbers
func add(a, b int) int {
	return a + b
}

// Multiply two numbers
func multiply(a, b int) int {
	result := a * b
	return result
}

// Greet a person
func greet(name string) {
	message := fmt.Sprintf("Hello, %s! Welcome to VizcachaIDE.", name)
	fmt.Println(message)
}

// Calculate factorial
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func main() {
	// Test addition
	sum := add(10, 20)
	fmt.Println("10 + 20 =", sum)

	// Test multiplication
	product := multiply(5, 7)
	fmt.Println("5 * 7 =", product)

	// Test greeting
	greet("Developer")

	// Test factorial
	fact := factorial(5)
	fmt.Println("5! =", fact)

	// You can set breakpoints and step through these functions
	// to see how the call stack changes!
}
