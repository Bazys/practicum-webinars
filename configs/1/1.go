package main

import (
	"fmt"
	"os"
)

func main() {
	// Читаем переменные окружения NAME & BURROW
	name := os.Getenv("NAME")
	burrow := os.Getenv("BURROW")
	// Выводим их значения
	fmt.Printf("%s lives in %s.\n", name, burrow) // Output:  lives in .

	// Теперь установим переменные окружения
	_ = os.Setenv("NAME", "gopher")
	_ = os.Setenv("BURROW", "/usr/gopher")

	fmt.Println("--------------") // отбивка

	name = os.Getenv("NAME")
	burrow = os.Getenv("BURROW")
	fmt.Printf("%s lives in %s.\n", name, burrow) // Output: gopher lives in /usr/gopher.
}
