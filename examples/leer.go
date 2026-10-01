package main

import "fmt"

func main() {
	var nombre string
	var edad int

	fmt.Print("Ingresa tu nombre: ")
	// Lee la entrada hasta un espacio o salto de línea
	fmt.Scanln(&nombre)

	fmt.Print("Ingresa tu edad: ")
	// Lee la edad
	fmt.Scanln(&edad)

	fmt.Printf("Hola, %s. Tienes %d años.\n", nombre, edad)
}
