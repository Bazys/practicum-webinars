// Пример 1: чтение файлов — три способа.
//
// Запуск:
//
//	go run ./03_os/01_read.go
//
// Что демонстрирует:
//   - os.ReadFile: целиком в память (Go 1.16+) — для маленьких файлов
//   - os.Open + io.Reader: потоковое чтение через буфер — для больших
//   - bufio.Scanner: построчное чтение — для логов/текстовых файлов
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// Создадим тестовый файл, чтобы было что читать.
	if err := os.WriteFile("sample.txt", []byte("первая строка\nвторая строка\nтретья\n"), 0644); err != nil {
		log.Fatal(err)
	}
	defer os.Remove("sample.txt")

	// 1) os.ReadFile — целиком в []byte. Самый простой способ.
	//    Идеально для маленьких файлов: конфигов, сертификатов, шаблонов.
	data, err := os.ReadFile("sample.txt")
	if err != nil {
		log.Fatalf("ReadFile: %v", err)
	}
	fmt.Printf("ReadFile: %q\n", string(data))

	// 2) os.Open + io.ReadFull — потоковое чтение фиксированным буфером.
	//    Подходит для больших файлов: не загружаем всё в память.
	f, err := os.Open("sample.txt")
	if err != nil {
		log.Fatalf("Open: %v", err)
	}

	buf := make([]byte, 8) // маленький буфер для демо — в реальности 4-32 КБ
	for {
		n, err := f.Read(buf)
		if n > 0 {
			fmt.Printf("прочитано %d байт: %q\n", n, buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Read: %v", err)
		}
	}
	f.Close() // закрыли явно (не через defer — чтобы показать порядок)

	// 3) bufio.Scanner — построчное чтение. Удобно для текстовых файлов и логов.
	f2, err := os.Open("sample.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close() // правильный паттерн: defer Close сразу после успешного Open

	scanner := bufio.NewScanner(f2)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		fmt.Printf("строка %d: %q\n", lineNum, scanner.Text())
	}
	// ОБЯЗАТЕЛЬНО проверяем ошибку после цикла: если чтение упало,
	// цикл просто закончится без паники.
	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner: %v", err)
	}

	// Подводный камень: по умолчанию Scanner ограничивает длину строки 64 КБ.
	// Если у вас длинные строки (например, весь JSON в одну строку), увеличивайте:
	//
	//   scanner := bufio.NewScanner(f)
	//   scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // до 1 МБ
}
