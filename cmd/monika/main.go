package main

import (
	"fmt"

	monika "hyperjumptech/monika/internal/monika"
)

func main() {
	fmt.Println(" __  __          _ _        ")
	fmt.Println("|  \\/  |___ _ _ (_) |____ _ ")
	fmt.Println("| |\\/| / _ \\ ' \\| | / / _` |")
	fmt.Printf("%s %s\n", "|_|  |_\\___/_||_|_|_\\_\\__,_|", "v0.0.1")

	monika.Init()
	select {}
}
