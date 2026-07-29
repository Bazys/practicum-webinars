// Пример 4: os.Stat — проверка существования файла и FileInfo.
//
// Запуск:
//
//	go run ./03_os/04_stat.go
//
// Что демонстрирует:
//   - современная проверка существования через errors.Is(err, os.ErrNotExist)
//   - почему НЕ использовать устаревший os.IsNotExist
//   - что даёт os.FileInfo: Size, ModTime, IsDir, Mode
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

func main() {
	// Создадим файл для демонстрации.
	if err := os.WriteFile("stat_demo.txt", []byte("hello world"), 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove("stat_demo.txt")

	// 1) Современный паттерн проверки существования файла.
	//    errors.Is правильно работает с обёрнутыми ошибками (%w),
	//    os.IsNotExist — нет.
	info, err := os.Stat("stat_demo.txt")
	switch {
	case errors.Is(err, os.ErrNotExist):
		fmt.Println("файл не существует")
	case err != nil:
		// Другая ошибка: нет прав, обрыв диска и т.п.
		log.Fatalf("Stat: %v", err)
	default:
		fmt.Println("файл существует")
	}

	// 2) FileInfo даёт много полезного.
	fmt.Printf("\nИнформация о файле:\n")
	fmt.Printf("  Name():    %q\n", info.Name())
	fmt.Printf("  Size():    %d байт\n", info.Size())
	fmt.Printf("  Mode():    %s (octal: %o)\n", info.Mode(), info.Mode())
	fmt.Printf("  ModTime(): %v\n", info.ModTime())
	fmt.Printf("  IsDir():   %v\n", info.IsDir())

	// 3) Тот же подход для несуществующего файла.
	fmt.Println("\n--- проверка несуществующего файла ---")
	_, err = os.Stat("нет_такого_файла.txt")
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("правильно определили: файл не существует")
	}

	// 4) НЕ используйте устаревший os.IsNotExist в новом коде.
	//    Он не работает с обёрнутыми ошибками:
	//
	//     wrapped := fmt.Errorf("wrap: %w", os.ErrNotExist)
	//     os.IsNotExist(wrapped)   // может дать false!
	//     errors.Is(wrapped, os.ErrNotExist) // true — правильно

	// 5) Проверка директории.
	if err := os.Mkdir("stat_demo_dir", 0755); err == nil {
		defer os.Remove("stat_demo_dir")
		info, _ := os.Stat("stat_demo_dir")
		fmt.Printf("\nДиректория IsDir(): %v, Mode(): %s\n", info.IsDir(), info.Mode())
	}
}
