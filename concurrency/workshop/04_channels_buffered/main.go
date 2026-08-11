// Workshop 04. Каналы: buffered vs unbuffered.
//
// Запуск:
//
//	go run ./04_channels_buffered
//
// Что показать:
//  1. Небуферизованный канал синхронен — отправитель ждёт читателя.
//  2. Буферизованный — асинхронен, пока буфер не заполнится.
//  3. Запись в полный канал / чтение из закрытого — паника/zero-value.
//  4. Закрытие канала — это сигнал «больше данных не будет» для range.
package main

import "fmt"

func main() {
	// Unbuffered — handoff: отправитель и получатель «встречаются» в момент передачи.
	unbuf := make(chan string)
	go func() {
		fmt.Println("отправитель: готовлю сообщение")
		unbuf <- "hello" // блокируется, пока главный не прочитает
		fmt.Println("отправитель: сообщение прочитано")
	}()
	msg := <-unbuf
	fmt.Println("получено:", msg)

	// Buffered — отправитель не блокируется, пока буфер не полон.
	buf := make(chan int, 3)
	buf <- 1
	buf <- 2
	buf <- 3
	// buf <- 4 // ← заблокируется, буфер заполнен

	fmt.Println("из буфера:", <-buf, <-buf, <-buf)

	// Close + range — идиома «закрыл источник, потребитель сам выйдет».
	src := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		src <- i
	}
	close(src)
	for v := range src {
		fmt.Print(v, " ")
	}
	fmt.Println()
}
