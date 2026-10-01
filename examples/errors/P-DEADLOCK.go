// P-DEADLOCK: sending on a channel that nobody receives from.
package main

func main() {
	messages := make(chan string)
	messages <- "hello"
}
