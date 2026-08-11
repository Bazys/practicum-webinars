// Пример 6: переменные окружения, аргументы, выход из процесса.
//
// Запуск:
//
//	go run ./03_os/06_env_exit.go arg1 arg2
//	PORT=9090 go run ./03_os/06_env_exit.go
//
// Что демонстрирует:
//   - os.Getenv vs os.LookupEnv — в чём разница
//   - os.Setenv, os.Environ
//   - os.Args — аргументы командной строки
//   - os.Exit — выход из процесса (defer НЕ выполняется!)
//   - os.Stdin/Stdout/Stderr — стандартные потоки
package main

import (
	"fmt"
	"os"
)

func main() {
	// 1) os.Args — аргументы командной строки.
	//    os.Args[0] — имя исполняемого файла.
	//    os.Args[1:] — реальные аргументы.
	fmt.Printf("Args: %v\n", os.Args)
	if len(os.Args) > 1 {
		fmt.Printf("первый аргумент: %q\n", os.Args[1])
	}

	// 2) os.Getenv — простое получение переменной окружения.
	//    Если переменной нет, возвращает пустую строку.
	//    НЕЛЬЗЯ различить «переменная = пусто» и «переменной нет».
	port := os.Getenv("PORT")
	fmt.Printf("PORT из GetEnv: %q (пусто может означать и нет, и =\"\")\n", port)

	// 3) os.LookupEnv — правильный способ.
	//    Возвращает (value, ok), где ok=false, если переменной нет.
	port2, ok := os.LookupEnv("PORT")
	if !ok {
		fmt.Println("PORT не задан — используем значение по умолчанию 8080")
		port2 = "8080"
	}
	fmt.Printf("итоговый PORT: %s\n", port2)

	// 4) os.Setenv — установка переменной для текущего процесса
	//    (и его потомков, которые запустит текущий).
	os.Setenv("MYAPP_DEBUG", "1")
	fmt.Printf("MYAPP_DEBUG = %q\n", os.Getenv("MYAPP_DEBUG"))

	// 5) os.Environ — все переменные в формате []string{"KEY=VALUE", ...}.
	//    Удобно для передачи в дочерний процесс.
	envs := os.Environ()
	fmt.Printf("всего переменных окружения: %d\n", len(envs))

	// 6) os.Stdin/Stdout/Stderr — стандартные потоки.
	//    Реализуют io.Reader (Stdin) и io.Writer (Stdout/Stderr).
	//    fmt.Println пишет в os.Stdout. log.* — в os.Stderr.
	fmt.Fprintln(os.Stdout, "это идёт в stdout")
	fmt.Fprintln(os.Stderr, "это идёт в stderr")

	// 7) os.Exit — немедленный выход из процесса.
	//    ВАЖНО: deferred-функции НЕ выполняются!
	//    Использовать только в конце main(), не в обработчиках/воркерах.
	//
	// Раскомментируйте, чтобы увидеть выход с кодом 42.
	// fmt.Println("до Exit — напечатается")
	// defer fmt.Println("этот defer НЕ выполнится при Exit")
	// os.Exit(42)
	// fmt.Println("после Exit — НЕ напечатается")
}
