package main

import (
	"fmt"
	"os"
)

func main() {
	// Пытаемся прочитать переменную окружения NAME
	name, exists := os.LookupEnv("NAME")

	// Если переменная не существует, то прерываем выполнение программы
	if !exists {
		panic("NAME is not set")
	}
	// Выводим значение переменной окружения NAME
	fmt.Print(name)
}
