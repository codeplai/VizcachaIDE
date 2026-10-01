package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	// Try autocomplete by pressing Ctrl+Space after typing:

	// 1. Type "fm" then Ctrl+Space to see fmt package
	// 2. Type "fmt." then Ctrl+Space to see all fmt methods

	name := "VizcachaIDE"
	fmt.Println("Welcome to", name)

	// Try typing "str" then Ctrl+Space to see string functions and strconv
	text := "Hello, World!"

	// Type "strings." then Ctrl+Space to see string manipulation functions
	upper := strings.ToUpper(text)
	fmt.Println(upper)

	// Type "ap" then Ctrl+Space to see append function
	numbers := []int{1, 2, 3}
	numbers = append(numbers, 4, 5)

	fmt.Println(numbers)

	// Type "le" then Ctrl+Space to see len function
	length := len(numbers)
	fmt.Printf("Length: %d\n", length)

	// Type "strconv." then Ctrl+Space to see conversion functions
	numStr := strconv.Itoa(42)
	fmt.Println("Number as string:", numStr)

	// Try typing keywords like "fo" for "for", "fu" for "func", etc.
	for i := 0; i < 5; i++ {
		fmt.Printf("Loop %d\n", i)
	}
}
