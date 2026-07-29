// Пример 3: направление лога в несколько приёмников через io.MultiWriter.
//
// Запуск:
//
//	go run ./01_log/03_output.go
//
// Что демонстрирует:
//   - os.OpenFile для открытия лог-файла в режиме дописывания (append)
//   - io.MultiWriter — один и тот же лог идёт и в консоль, и в файл
//   - log.SetOutput для переключения стандартного логгера
//
// После запуска рядом появится файл demo.log с теми же сообщениями.
package main

import (
	"io"
	"log"
	"os"
)

func main() {
	// 1) Открываем (или создаём) файл в режиме дописывания.
	//    O_APPEND  — писать в конец
	//    O_CREATE  — создать, если не существует
	//    O_WRONLY  — только для записи
	//    0644      — права доступа: владелец rw, остальные r
	logFile, err := os.OpenFile("demo.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("не могу открыть лог-файл: %v", err)
	}
	defer logFile.Close() // всегда закрываем по выходу

	// 2) io.MultiWriter возвращает Writer, который дублирует запись
	//    во все переданные writer'ы.
	multi := io.MultiWriter(os.Stdout, logFile)

	// 3) Переключаем стандартный логгер на MultiWriter.
	//    Теперь любой log.Print* пойдёт и в консоль, и в файл.
	log.SetOutput(multi)

	// Эти строки появятся ОДНОВРЕМЕННО в терминале и в demo.log.
	log.Println("первое сообщение после настройки лога")
	log.Println("второе сообщение — оба приёмника работают")

	// 4) Создадим отдельный логгер с собственным префиксом.
	//    Он унаследует всё ту же настройку MultiWriter.
	special := log.New(multi, "[SPECIAL] ", log.LstdFlags|log.Lshortfile)
	special.Println("сообщение от именного логгера")

	log.Println("завершаем работу — проверьте файл demo.log")
}
