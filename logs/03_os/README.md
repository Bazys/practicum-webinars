# Примеры: пакет `os`

Операции с файлами, директориями, окружением.

## Запуск

```bash
cd examples
go run ./03_os/01_read.go
go run ./03_os/02_write.go
go run ./03_os/03_openfile_flags.go
go run ./03_os/04_stat.go
go run ./03_os/05_dirs.go
# С аргументами и переменной окружения:
PORT=9090 go run ./03_os/06_env_exit.go myarg
```

> Каждый файл — отдельный `package main`. Запускайте по одному.
> Примеры создают временные файлы/директории рядом и удаляют их через `defer`.

## Что в каждом файле

### `01_read.go` — чтение (3 способа)
- `os.ReadFile` — целиком в память (для маленьких файлов)
- `os.Open` + `io.Reader` — потоковое (для больших)
- `bufio.Scanner` — построчно (для логов/текста)
- подводный камень: ограничение строки 64 КБ

### `02_write.go` — запись (3 способа)
- `os.WriteFile` — целиком перезаписывает
- `os.Create` — создаёт/очищает, возвращает `*os.File`
- `WriteString`, `Write`, `Sync`
- демонстрация `O_TRUNC` в `Create` (почему не для логов)

### `03_openfile_flags.go` — `os.OpenFile` и флаги ⭐
- **ключевой паттерн для логирования:** `O_APPEND|O_CREATE|O_WRONLY`
- демонстрация append при повторных запусках
- `O_EXCL` для lock-файлов

### `04_stat.go` — проверка существования и FileInfo
- современный паттерн: `errors.Is(err, os.ErrNotExist)`
- почему НЕ использовать `os.IsNotExist`
- `Size`, `ModTime`, `IsDir`, `Mode`

### `05_dirs.go` — директории
- `MkdirAll` (как `mkdir -p`)
- `ReadDir` (возвращает `[]DirEntry`, быстрее устаревшего `ioutil.ReadDir`)
- `Remove`, `RemoveAll`, `Rename`, `MkdirTemp`
- `filepath.Join`, `filepath.WalkDir`

### `06_env_exit.go` — окружение и процесс
- `Getenv` vs `LookupEnv` — разница между «пусто» и «нет»
- `Setenv`, `Environ`, `Args`
- `os.Exit` — defer НЕ выполняется
- `Stdin/Stdout/Stderr`

## Подводные камни

1. **Не закрывать `*os.File`** → утечка файловых дескрипторов.
   Всегда `defer f.Close()`.
2. **Игнорировать ошибку `Close` при записи** → можно потерять данные.
   Для надёжности: `f.Sync()` перед закрытием.
3. **`os.IsNotExist` устарел** → `errors.Is(err, os.ErrNotExist)`.
4. **Scanner и 64 КБ** → длинные строки обрежутся, увеличивайте через `scanner.Buffer()`.
5. **`O_TRUNC` в `os.Create`** → очищает существующий файл, не для логов.
6. **TOCTOU:** не проверяйте `os.Stat` перед `os.Open` — открывайте и
   обрабатывайте ошибку (гонка «проверил-открыл»).

## Связанные слайды

21, 22, 23, 24, 25, 26, 27, 28 в `slides/webinar.md`.

## Связанный конспект

`notes/03-os-files.md`.
