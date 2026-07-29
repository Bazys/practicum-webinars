// Пример 5: работа с директориями.
//
// Запуск:
//
//	go run ./03_os/05_dirs.go
//
// Что демонстрирует:
//   - os.MkdirAll — создание всей цепочки каталогов (как mkdir -p)
//   - os.ReadDir — список содержимого (возвращает []DirEntry, Go 1.16+)
//   - os.Remove / RemoveAll — удаление файла и директории рекурсивно
//   - os.Rename — переименование/перемещение
//   - os.MkdirTemp — временная директория (для тестов)
//   - path/filepath.Join, filepath.WalkDir — сопутствующий пакет
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 1) MkdirAll — создаёт всю цепочку каталогов.
	//    Аналог mkdir -p. НЕ падает, если уже существует.
	tree := filepath.Join("data", "cache", "thumbs")
	if err := os.MkdirAll(tree, 0755); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("создано дерево: %s\n", tree)
	defer os.RemoveAll("data") // уберём за собой в конце

	// Положим пару файлов, чтобы было что перечислять.
	os.WriteFile(filepath.Join("data", "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join("data", "cache", "file2.txt"), []byte("2"), 0644)

	// 2) ReadDir — список содержимого каталога.
	//    Возвращает []os.DirEntry — лёгкие объекты, НЕ открывающие файлы.
	//    Это быстрее устаревшего ioutil.ReadDir, который делал stat по каждому.
	entries, err := os.ReadDir("data")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nсодержимое data/:")
	for _, e := range entries {
		// DirEntry даёт имя и флаг «это директория?» без открытия файла.
		// За полной FileInfo — e.Info().
		typeMark := "F"
		if e.IsDir() {
			typeMark = "D"
		}
		fmt.Printf("  [%s] %s\n", typeMark, e.Name())
	}

	// 3) filepath.WalkDir — обход всего дерева каталогов (рекурсивно).
	//    Появился в Go 1.16, быстрее старого filepath.Walk.
	fmt.Println("\nобход дерева через WalkDir:")
	err = filepath.WalkDir("data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("  %s\n", path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// 4) Rename — переименование или перемещение.
	old := filepath.Join("data", "file1.txt")
	new := filepath.Join("data", "renamed.txt")
	if err := os.Rename(old, new); err != nil {
		log.Printf("Rename: %v", err)
	}

	// 5) Remove — удаляет ОДИН файл или ПУСТУЮ директорию.
	os.Remove(filepath.Join("data", "cache", "file2.txt"))

	// 6) MkdirTemp — создаёт временную директорию со случайным именем.
	//    Идеально для тестов и промежуточных файлов.
	tmpDir, err := os.MkdirTemp("", "webinar-demo-*")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nвременная директория: %s\n", tmpDir)
	// ОБЯЗАТЕЛЬНО убираем за собой. В тестах можно t.TempDir() —
	// он сам очистится после завершения теста.
	defer os.RemoveAll(tmpDir)

	// 7) os.TempDir() — системная папка для временных файлов.
	fmt.Printf("системный temp: %s\n", os.TempDir())
}
