// P-SLICE-BOUNDS: taking a sub-slice that goes past the end.
package main

import "fmt"

func main() {
	letters := []string{"a", "b", "c"}
	end := 5
	fmt.Println(letters[:end])
}
