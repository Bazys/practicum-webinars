// Пример 2: запись файлов — три способа.
//
// Запуск:
//
//	go run ./03_os/02_write.go
//
// Что демонстрирует:
//   - os.WriteFile: целиком перезаписывает файл
//   - os.Create: создаёт/очищает файл, возвращает *os.File
//   - *os.File.WriteString, Write, Sync
//   - почему НЕЛЬЗЯ использовать Create для логов (O_TRUNC по умолчанию)
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// 1) os.WriteFile — целиком перезаписывает файл.
	//    Идеально для разовой записи маленьких файлов (конфиги, отчёты).
	//    ВАЖНО: если файл существовал, он будет ПЕРЕЗАПИСАН (не дописан).
	err := os.WriteFile("out_writefile.txt", []byte("hello от WriteFile\n"), 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("WriteFile: создан out_writefile.txt")
	defer os.Remove("out_writefile.txt")

	// 2) os.Create — создаёт файл (или ОЧИЩАЕТ существующий), открывает на запись.
	//    Возвращает *os.File. Удобно для многократной записи в НОВЫЙ файл.
	//    ВНИМАНИЕ: эквивалент OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666).
	//    Поэтому НЕ подходит для логов — каждый раз очищает!
	f, err := os.Create("out_create.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close() // всегда закрываем
	defer os.Remove("out_create.txt")

	// 3) Методы *os.File для записи.
	if _, err := f.WriteString("первая строка\n"); err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte{'b', 'y', 't', 'e', 's', '\n'}); err != nil {
		log.Fatal(err)
	}

	// 4) Sync — сбросить буферы ОС на диск.
	//    По умолчанию ОС буферизует запись. Если важно, чтобы данные
	//    точно попали на диск (например, перед подтверждением транзакции),
	//    вызываем Sync. Это относительно дорогая операция.
	if err := f.Sync(); err != nil {
		log.Printf("Sync: %v", err)
	}
	fmt.Println("Create + WriteString + Sync: создан out_create.txt")

	// 5) Проверим, что Create действительно очищает файл.
	//    Допишем строку через Append (правильный способ для логов — см. пример 03).
	//    А потом покажем, что Create очистит всё.
	fmt.Println("\n--- демонстрация O_TRUNC в Create ---")
	os.WriteFile("demo_trunc.txt", []byte("это содержимое будет стёрто\n"), 0644)
	defer os.Remove("demo_trunc.txt")

	f2, _ := os.Create("demo_trunc.txt") // откроет существующий и ОЧИСТИТ
	f2.WriteString("новое содержимое\n")
	f2.Close()

	data, _ := os.ReadFile("demo_trunc.txt")
	fmt.Printf("после Create: %q (старое содержимое потеряно)\n", string(data))
}
