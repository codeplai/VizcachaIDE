// E-NO-NEW-VARS: := is used again for a variable that already exists.
package main

import "fmt"

func main() {
	score := 1
	score := 2
	fmt.Println(score)
}
