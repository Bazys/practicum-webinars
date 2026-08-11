// Пример 2: кастомный log.Logger и флаги форматирования.
//
// Запуск:
//
//	go run ./01_log/02_logger_flags.go
//
// Что демонстрирует:
//   - создание собственного логгера через log.New(writer, prefix, flags)
//   - влияние каждого флага на формат вывода
//   - метод Output() с управлением calldepth
package main

import (
	"bytes"
	"fmt"
	"log"
)

func main() {
	// Базовый случай: логгер с выводом в os.Stderr, без префикса, с дефолтными флагами.
	std := log.Default()
	std.Println("сообщение от стандартного логгера")

	// log.New принимает: writer, prefix, flags.
	// Здесь используем bytes.Buffer как writer — чтобы увидеть вывод в виде строки.
	var buf bytes.Buffer
	logger := log.New(&buf, "[myapp] ", log.LstdFlags|log.Lshortfile)
	logger.Print("запись в буфер")
	fmt.Printf("из буфера: %q\n", buf.String())

	// 2) Демонстрация флагов форматирования.
	//    Каждый флаг добавляет к строке что-то своё.
	demoFlags()

	// 3) Префикс и Lmsgprefix: по умолчанию префикс идёт В НАЧАЛЕ строки (до даты).
	//    Lmsgprefix — переносит префикс перед сообщением (после даты/файла).
	fmt.Println("\n--- Lmsgprefix ---")
	var b2 bytes.Buffer
	l2 := log.New(&b2, "PREFIX ", log.LstdFlags|log.Lmsgprefix)
	l2.Print("hello")
	fmt.Println(b2.String())

	// 4) Метод Output — ручное управление глубиной стека вызовов.
	//    Полезно, когда оборачиваете логгер в свою функцию и хотите,
	//    чтобы Lshortfile показывал вызывающего, а не вашу обёртку.
	fmt.Println("\n--- Output с calldepth ---")
	var b3 bytes.Buffer
	l3 := log.New(&b3, "", log.Lshortfile)
	logViaWrapper(l3, "сообщение через обёртку")
	fmt.Println(b3.String())
}

// demoFlags печатает, как разные комбинации флагов влияют на вывод.
func demoFlags() {
	cases := []struct {
		name  string
		flags int
	}{
		{"LstdFlags (по умолчанию)", log.LstdFlags},
		{"Ldate | Ltime | Lmicroseconds", log.Ldate | log.Ltime | log.Lmicroseconds},
		{"Lshortfile", Lshortflags()}, // см. helper ниже
		{"LUTC + Ldate + Ltime", log.Ldate | log.Ltime | log.LUTC},
		{"без флагов (0)", 0},
	}

	fmt.Println("\n--- Влияние флагов на формат ---")
	for _, c := range cases {
		var b bytes.Buffer
		l := log.New(&b, "", c.flags)
		l.Print("test")
		fmt.Printf("%-35s => %s", c.name, b.String())
	}
}

// logViaWrapper демонстрирует calldepth в Logger.Output.
// calldepth=1 указал бы на сам Output, calldepth=2 — на эту функцию-обёртку
// (logViaWrapper), calldepth=3 — на её вызывающего (main).
// calldepth=4 и выше уходит в рантайм (proc.go) — слишком высоко.
func logViaWrapper(l *log.Logger, msg string) {
	// calldepth=2 — увидим logViaWrapper как источник строки.
	_ = l.Output(2, msg)
}

// Lshortflags возвращает Lshortfile — отдельная функция нужна только для того,
// чтобы строка вызова была на отдельной строке и demoFlags красиво печаталась.
// По сути это просто log.Lshortfile.
func Lshortflags() int { return log.Lshortfile }
