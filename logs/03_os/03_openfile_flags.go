// Пример 3: os.OpenFile и флаги — КЛЮЧЕВОЙ паттерн для логирования.
//
// Запуск:
//
//	go run ./03_os/03_openfile_flags.go
//
// Что демонстрирует:
//   - os.OpenFile с разными комбинациями флагов
//   - O_APPEND|O_CREATE|O_WRONLY — стандартный набор для логов
//   - O_EXCL — создать «только если не существует» (для lock-файлов)
//   - влияние флагов на поведение при повторных запусках
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	logFile := "demo_app.log"
	defer os.Remove(logFile)

	// 1) Стандартный паттерн для лог-файла:
	//    O_APPEND  — писать в конец файла
	//    O_CREATE  — создать, если не существует
	//    O_WRONLY  — только для записи
	//    0644      — владелец rw, группа/остальные r
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("OpenFile: %v", err)
	}

	// Пишем несколько строк.
	for i := 1; i <= 3; i++ {
		if _, err := fmt.Fprintf(f, "запись #%d\n", i); err != nil {
			log.Fatal(err)
		}
	}
	f.Close()

	// Прочитаем, что получилось.
	data, _ := os.ReadFile(logFile)
	fmt.Printf("после первого запуска (3 строки):\n%s\n", data)

	// 2) Запустим ТОТ ЖЕ код ещё раз — и увидим, что лог ДОПОЛНИЛСЯ, не стёрся.
	//    Это и есть смысл O_APPEND.
	f, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	for i := 4; i <= 5; i++ {
		fmt.Fprintf(f, "запись #%d\n", i)
	}
	f.Close()

	data, _ = os.ReadFile(logFile)
	fmt.Printf("после второго запуска (5 строк — append работает):\n%s\n", data)

	// 3) O_EXCL + O_CREATE — «создать, только если НЕ существует».
	//    Классический паттерн для lock-файла: если файл уже есть,
	//    значит, другой процесс уже работает — выходим.
	fmt.Println("--- O_EXCL: lock-файл ---")
	lockFile := "demo.lock"
	defer os.Remove(lockFile)

	// Первый процесс создаёт lock.
	lock1, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("первый процесс не смог создать lock: %v\n", err)
	} else {
		fmt.Println("первый процесс создал lock")
		lock1.WriteString("pid 12345\n")
		lock1.Close()
	}

	// Второй процесс пытается создать тот же lock — должно упасть с ErrExist.
	_, err = os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("второй процесс получил ошибку (как и ожидалось): %v\n", err)
	}
}
