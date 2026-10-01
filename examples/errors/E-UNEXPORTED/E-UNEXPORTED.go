// E-UNEXPORTED: a lowercase (unexported) name from another package is used.
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.toUpper("hello"))
}
