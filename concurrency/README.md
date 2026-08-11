# Многопоточность в Golang — конспект вебинара

Вебинар курса «Продвинутый Go-разработчик». Уровень: **хардкор** — внутренности
планировщика, memory model, race detection, паттерны конкурентности.

Этот документ — **теоретический конспект с разбором каждого слайда**. К каждому
разделу даны ссылки на первоисточники, чтобы вы могли прочитать оригинал и
перепроверить сказанное на вебинаре.

---

## Содержание

- [Материалы и как запускать](#материалы-и-как-запускать)
- [Часть 1. История и эволюция](#часть-1-история-и-эволюция)
- [Часть 2. Внутренности планировщика GMP](#часть-2-внутренности-планировщика-gmp)
- [Часть 3. Синхронизация: примитивы и паттерны](#часть-3-синхронизация-примитивы-и-паттерны)
- [Часть 4. Паттерны конкурентности — Workshop](#часть-4-паттерны-конкурентности--workshop)
- [Часть 5. Профилирование и отладка](#часть-5-профилирование-и-отладка)
- [Часть 6. Q\&A и ссылки](#часть-6-qа-и-ссылки)

---

## Материалы и как запускать

| Файл / каталог | Описание |
| --- | --- |
| `Многопоточность_в_Golang.pptx` | Презентация (38 слайдов, 16:9). |
| `workshop/` | 12 runnable-примеров для живого кодинга. |

```bash
# Запустить пример:
go run ./concurrency/workshop/01_data_race

# С race-детектором:
go run -race ./concurrency/workshop/01_data_race

# Бенчмарк (Mutex vs RWMutex):
go test -bench=. -benchmem ./concurrency/workshop/02_mutex_vs_rwmutex
```

---

## Часть 1. История и эволюция

### Эволюция: 45 лет от теории к Go

Конкурентность в Go — не «изобретение с нуля», а результат 40-летней эволюции
идеи **CSP (Communicating Sequential Processes)**.

#### Ключевые вехи

**1978 — CSP Хоара.**
Чарльз Энтони Ричард Хоар (C. A. R. Hoare) публикует статью «Communicating
Sequential Processes». CSP — это формальный язык для описания систем из
нескольких параллельных процессов, которые общаются **передачей сообщений**,
а не через разделяемую память. Процессы не имеют общих переменных; весь обмен
данными идёт через каналы (точки синхронизации). Эта статья стала
теоретической базой для каналов в Go.

> **Источник:** Hoare, C. A. R. «Communicating Sequential Processes.»
> Communications of the ACM 21.8 (1978): 666–677.
> — оригинальная статья (есть в открытом доступе):
> <https://www.cs.cmu.edu/~crary/819-f09/Hoare78.pdf>
>
> - перевод на руский:
> <https://ssd.sscc.ru/old/chair/files/hoar/hoar.pdf>

**1983–1989 — Occam.**
Язык программирования Occam (Inmos, Великобритания) — первое практическое
воплощение CSP. Использовался для транзьютеров (transputer) — ранних
параллельных процессоров. Occam доказал, что CSP-модель реализуема в
железе.

**1995 — Limbo.**
В Bell Labs Роб Пайк, Шон Дорвард и Фил Винтерботтом создают язык **Limbo**
для операционной системы Inferno. В Limbo уже были: горутины (назывались
*threads*), каналы для общения между ними, сборщик мусора. Limbo —
непосредственный предок Go; та же команда позже создаст Go.

> **Источник:** Dorward, Winterbottom, Pike.
> — Wikipedia:
> <https://en.wikipedia.org/wiki/Limbo_(programming_language)>

**2009 — анонс Go.**
Роб Пайк представляет Go: «Go-routines» + каналы — встроены в язык как
первоклассные сущности. Канал — это тип данных (`chan T`), а не библиотечный
конструкт.

**2012 — Go 1.0 + переписывание планировщика.**
Дмитрий Выюков переписывает планировщик с простой модели GM (Goroutine +
Machine) на модель GMP (Goroutine + Machine + Processor). Это решило
проблему глобальной блокировки планировщика.

> **Источник:** Vyukov, Dmitry. «Scalable Go Scheduler Design Doc.» (2012)
> <https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw>

**2015 — Go 1.5: GOMAXPROCS по умолчанию = NumCPU.**
До этого `GOMAXPROCS` по умолчанию был равен **1** — программа использовала
только одно ядро. С Go 1.5 значение по умолчанию стало равно числу
логических процессоров (`runtime.NumCPU()`).

> **Источник:** Go 1.5 Release Notes — Runtime:
> <https://go.dev/doc/go1.5>
> Цитата: *«the runtime now sets the default number of threads to run
> simultaneously, defined by GOMAXPROCS, to the number of cores available
> on the CPU. In prior releases the default was 1.»*

**2020 — Go 1.14: асинхронный preemption.**
Горутины стали вытесняемыми (preemptible) асинхронно — через сигналы.
Тight loop без вызова функций больше не может «захватить» поток навсегда.

> **Источник:** Go 1.14 Release Notes — Runtime:
> <https://go.dev/doc/go1.14>
> Цитата: *«Goroutines are now asynchronously preemptible. As a result,
> loops without function calls no longer potentially deadlock the
> scheduler or significantly delay garbage collection.»*

---

### Эволюция планировщика: GM → GMP

#### Модель GM (до Go 1.0)

Изначально в планировщике было две сущности:

- **G (Goroutine)** — горутина со своим стеком и состоянием.
- **M (Machine)** — поток операционной системы (OS thread).

Все runnable-горутины лежали в **одной глобальной очереди** (global run queue),
защищённой мьютексом. Когда M нужна была работа, она лочила мьютекс,
доставала G из очереди, разлачивала и выполняла.

**Проблемы GM:**

1. **Глобальная блокировка.** Все M спорят за один мьютекс глобальной очереди.
   При росте числа ядер (8, 16, 32+) это становится узким местом.
2. **Плохая локальность кэша.** G, запущенная на одном M, может быть
   продолжена на другом M — её данные уже вытеснены из кэша.
3. **Head-of-line blocking.** Один G в голове очереди блокирует продвижение
   остальных при конкурентном доступе.

#### Модель GMP (Go 1.0+, дизайн Выюкова)

Выюков добавил третью сущность:

- **P (Processor)** — логический процессор, «контекст выполнения». Содержит
  локальную очередь runnable-горутин и кэш ресурсов (memory cache для
  аллокатора).

`GOMAXPROCS` теперь задаёт **число P**, а не число M. M может существовать
больше, чем P (потому что blocking syscalls отсоединяют M от P). Чтобы G
выполнился, нужен тройной кортеж (G, M, P): M должен держать P, а P —
указывать на G.

**Ключевое преимущество:** каждый P имеет свою локальную очередь, поэтому
в обычном случае создание и запуск горутины не требует глобальной блокировки.

---

### Почему именно CSP / горутины?

Сравнение моделей конкурентности в разных языках.

| Подход | Языки | Цена переключения | Модель обмена данными |
| --- | --- | --- | --- |
| OS threads | C, C++, Java (старая) | ~1–10 мкс | Shared memory + locks |
| async/await | JS, Python, C#, Rust | ~0.1–1 мкс | Futures / callbacks |
| actors | Erlang, Akka | ~1 мкс | Сообщения (mailbox) |
| goroutines + CSP | Go | ~0.2–1 мкс | Каналы / shared + sync |

**Почему Go выбрал CSP:**

1. **Горутины очень дёшевы.** Стартовый стек ~2 КБ (растёт по мере надобности).
   Можно запустить сотни тысяч горутин. OS thread стоит ~1–8 МБ стека —
   десятки тысяч уже упрутся в память.
2. **Канал = синхронизация + передача данных одновременно.** Не нужны
   отдельные мьютексы для защиты общих данных — данные передаются через
   канал, и владение ими переходит от отправителя к получателю.
3. **Код читается линейно.** В отличие от async/await,
   горутина запускается одним словом `go` и
   синтаксически не отличается от обычной функции.

> **Источник:** Effective Go — Concurrency:
> <https://go.dev/doc/effective_go#concurrency>

---

### Ключевой тезис Go

> *«Don't communicate by sharing memory; share memory by communicating.»*
> — Go proverb.

Это **главная идеологема** Go. В традиционных языках (C++, Java) принято
создавать общую переменную и защищать её мьютексом («communicate by sharing
memory»). В Go принято передавать владение данными через канал: один
гороутин отправляет объект в канал и теряет к нему доступ, другой —
получает и становится единственным владельцем.

**Практический пример:** вместо того чтобы две гороутины делили `*Buffer`
под мьютексом, один гороутин создаёт буфер, отправляет в канал, и
получатель работает с ним без всяких блокировок.

Это не значит, что мьютексы запрещены — `sync.Mutex` есть и иногда это
правильный выбор. Но **по умолчанию** в Go лучше думать каналами.

> **Источник:** Go Proverbs (Rob Pike):
> <https://go-proverbs.github.io/>
> и Go Blog — «Share Memory By Communicating» (Andrew Gerrand, 2010):
> <https://go.dev/blog/codelab-share>
> (полный codewalk: <https://go.dev/doc/codewalk/sharemem/>)

Дополнительные цифры по стоимости горутины:

- Стартовый размер стека гороутины: 2 КБ (растёт до 1 ГБ максимум).
  См. `runtime/debug.SetMaxStack`.
- Стартовый размер стека OS thread: зависит от ОС, обычно 1–8 МБ.
- Создание гороутины: ~200–500 нс. Создание OS thread: ~10–50 мкс.

> **Источник:**.runtime/debug — SetMaxStack:
> <https://pkg.go.dev/runtime/debug#SetMaxStack>

---

## Часть 2. Внутренности планировщика GMP

### Слайд 9. G · M · P — три сущности

Это «святая троица» планировщика Go. Разбор каждой сущности:

#### G — Goroutine

Структура `runtime.g` в исходниках Go. Содержит:

- **Стек** (`stack`) — стартует с 2 КБ, растёт/уменьшается динамически.
- **m** — указатель на текущий M, выполняющий эту G (или `nil`).
- **atomicstatus** — состояние гороутины (см. слайд 10).
- **sched** — сохранённые регистры (PC, SP) для возобновления.
- **goid** — идентификатор гороутины.

#### M — Machine

Структура `runtime.m`. Это **настоящий поток ОС** (POSIX thread на Linux/macOS).
M создаются runtime по мере надобности и могут существовать в количестве,
превышающем `GOMAXPROCS` (когда часть из них заблокирована в syscall'ах).

#### P — Processor

Структура `runtime.p`. **Логический процессор** — ресурс, необходимый для
выполнения Go-кода. `GOMAXPROCS` определяет, сколько P существует
одновременно. P содержит:

- **Local run queue** — локальная очередь runnable-горутин (до 256 штук).
- **mcache** — кэш аллокатора для маленьких объектов (ускоряет аллокацию).
- **deferpool, sudogcache** — кэши вспомогательных структур.

**Правило:** чтобы выполнить Go-код, M должен «держать» P. Если M
блокируется в syscall, он отпускает P, и другой M может его подобрать.

> **Источник:**
>
> - Vyukov, «Scalable Go Scheduler Design Doc»:
>   <https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw>
> - Исходный код `runtime/proc.go` (определения `g`, `m`, `p`):
>   <https://github.com/golang/go/blob/master/src/runtime/proc.go>
> - `runtime/runtime2.go` (структуры g, m, p):
>   <https://github.com/golang/go/blob/master/src/runtime/runtime2.go>

---

### Work-stealing и очереди

#### Локальные очереди

Каждый P имеет **локальную очередь** runnable-горутин (run queue) ёмкостью
256 элементов. Это **кольцевой буфер** без блокировок (lock-free), потому что
в обычном случае только один P имеет к ней доступ.

Когда вы пишете `go myFunc()`, новая горутина `G` помещается в **локальную
очередь текущего P** (тот, на котором выполняется вызвавший код). Это O(1) без
глобальных блокировок.

#### Global queue и переполнение

Если локальная очередь переполнилась (больше 256), её **половина**
переносится в **глобальную очередь** (global run queue), защищённую
мьютексом. Это редкое событие, поэтому глобальный мьютекс не становится
узким местом.

#### Work-stealing

Когда P опустошает свою локальную очередь, он ищет новую работу по
следующему алгоритму:

1. Проверить локальную очередь. Если пусто —
2. Попробовать забрать из global queue.
3. Опросить network poller (есть ли готовые I/O-операции).
4. **Украсть (steal) половину** горутин из локальной очереди другого P
   (выбирается случайно).

Шаг 4 — **work-stealing**, ключевая идея дизайна Выюкова. Благодаря ему
нагрузка автоматически балансируется между P: занятые P отдают работу
свободным.

#### GOMAXPROCS в рантайме

`runtime.GOMAXPROCS(n)` можно вызвать в любой момент — runtime создаст или
уничтожит P. Это **дешёвая операция** (не создаёт OS threads, только
логические структуры).

**Результат:** средняя стоимость запуска горутины — O(1), без глобальных
блокировок в обычном случае.

> **Источник:** Vyukov, «Scalable Go Scheduler Design Doc» — раздел
> «Work-Stealing»:
> <https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw>
> и `runtime/proc.go`, функция `findRunnable`:
> <https://github.com/golang/go/blob/master/src/runtime/proc.go>

---

### Blocking syscall и handoff M→P

Это один из самых важных механизмов планировщика. Проблема: что делать,
когда горутина делает **блокирующий** системный вызов (например, `read`
из файла или медленного сокета)?

#### Проблема без handoff

Если M просто заблокируется в syscall, то и привязанный к нему P тоже
оказывается заблокирован — а P ценен, потому что `GOMAXPROCS` обычно равен
числу ядер. Один заблокированный syscall «съедает» целое ядро.

#### Решение: handoff

1. Перед входом в potentially-blocking syscall runtime вызывает
   `entersyscall()` — помечает текущую пару (M, P) как «в syscall».
2. Если syscall длится дольше ~20 мкс, монитор `sysmon` (системная
   гороутина) замечает это и вызывает **handoff**: P отсоединяется от
   заблокированного M.
3. Свободный P подбирает **другой M** (из пула или создаётся новый) и
   продолжает выполнять горутины из своей локальной очереди.
4. Заблокированный M всё ещё ждёт завершения syscall в ядре ОС.
5. Когда syscall завершается, бывший M пытается:
   - Вернуть себе P (если он свободен).
   - Найти другой свободный P.
   - Если свободных P нет — положить свою G в global queue и уйти в
     спящий пул (mput).

**Вывод:** число M может быть **больше** `GOMAXPROCS`, потому что
заблокированные в syscall M не считаются. Именно поэтому
`runtime.NumGoroutine()` и число потоков в ОС (`top`, `ps`) могут
превышать `GOMAXPROCS`.

#### Network poller — отдельный путь

Неблокирующий сетевой I/O в Go идёт через **network poller** (epoll на
Linux, kqueue на macOS/BSD). Горутина, ждущая данных из сокета, не
блокирует M: runtime регистрирует сокет в epoll, а G переводится в статус
`Gwaiting`. Когда ядро сообщает, что данные готовы (`epoll_wait`),
соответствующая G возвращается в очередь runnable. **M не блокируется.**

> **Источник:**
>
> - Vyukov, Design Doc, раздел «Blocking Syscalls»:
>   <https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw>
> - `runtime/proc.go`, функции `entersyscall`, `exitsyscall`, `handoffp`:
>   <https://github.com/golang/go/blob/master/src/runtime/proc.go>
> - `runtime/netpoll_epoll.go` (network poller для Linux):
>   <https://github.com/golang/go/blob/master/src/runtime/netpoll_epoll.go>

---

### Preemption: коопертивный → асинхронный (Go 1.14)

#### До Go 1.14: кооперативное вытеснение

Горутина могла быть переключена **только в безопасных точках** (safe
points), где runtime может вклиниться:

- Вызов функции (проверка стека `morestack`).
- Блокировка на канале / мьютексе.
- Системный вызов.

**Проблема:** tight loop без вызова функций — например, `for {}` или
`for i := 0; i < 1e18; i++ { x++ }` — **никогда** не отдаёт управление
планировщику. Такой цикл:

- Захватывает целый P навсегда (другие горутины на этом P не получат время).
- Может задержать сборку мусора (GC нужно остановить все горутины, а эта
  не останавливается).

#### Go 1.14+: асинхронное вытеснение

Runtime посылает потоку сигнал **SIGURG** (через `tgkill` на Linux). В
обработчике сигнала runtime проверяет, не пора ли вытеснить текущую
горутину. Если да — он **асинхронно** сохраняет её контекст и ставит
обратно в очередь runnable, а P берёт следующую горутину.

Системная гороутина `sysmon` опрашивает все P каждые ~10 мс и посылает
SIGURG тем M, которые работают слишком долго без переключения (sustained
> 10 мс).

**Следствие:** tight loop теперь безопасен. GC stop-the-world больше не
ждёт cooperative-точек бесконечно.

**Побочный эффект:** программы на Unix стали получать больше сигналов.
Системные вызовы, обрабатываемые через `syscall` или `golang.org/x/sys/unix`,
могут возвращать `EINTR` — их нужно повторять в цикле.

> **Источник:**
>
> - Go 1.14 Release Notes — Runtime:
>   <https://go.dev/doc/go1.14>
>   Цитата: *«Goroutines are now asynchronously preemptible. As a result,
>   loops without function calls no longer potentially deadlock the
>   scheduler or significantly delay garbage collection.»*
> - A Journey With Go — «Go: Asynchronous Preemption» (подробный разбор):
>   <https://medium.com/a-journey-with-go/go-asynchronous-preemption-b5194227371c>
> - GopherCon 2020 — «Pardon the Interruption: Loop Preemption in Go
>   1.14» (Austin Clements):
>   <https://www.youtube.com/watch?v=1I1WmeSjRSw>

---

## Часть 3. Синхронизация: примитивы и паттерны

### Data race — что это и почему это UB

#### Что такое data race

**Data race (гонка данных)** возникает, когда:

1. Две или более гороутины обращаются к одной переменной.
2. Хотя бы одно обращение — **запись**.
3. Нет синхронизации (happens-before) между обращениями.

В этом случае результат **не определён** (undefined behavior с точки зрения
Go Memory Model). Программа может выдать любое значение, упасть, или
работать «почти всегда правильно, кроме раз в год».

#### Пример: `count++`

```go
count++
```

Эта операция **не атомарна** — она состоит из трёх шагов:

1. Прочитать `count` из памяти.
2. Увеличить на 1 в регистре.
3. Записать обратно.

Если две гороутины сделают это одновременно, оба прочитают одно и то же
значение (например, 5), оба увеличат до 6, и оба запишут 6. Итог: вместо
ожидаемых 7 получаем 6. При тысячах горутин результат будет значительно
меньше ожидаемого и **непредсказуемым** при каждом запуске.

#### Race detector

Флаг `-race` включает встроенный детектор гонок (на базе ThreadSanitizer).

```bash
go run -race main.go
go test -race ./...
go build -race -o myapp
```

Детектор instrumentation'ом отслеживает все обращения к памяти на лету и
сообщает о гонках, которые **реально произошли** во время прогона. Он не
делает статического анализа — если гонка не случилась во время теста,
детектор её не увидит.

**Накладные расходы:** CPU ~2–10x, память ~5–10x. Не включайте в
production под нагрузкой без необходимости.

**Важно:** `go vet` тоже находит часть гонок (например, копирование
мьютекса), но `-race` гораздо мощнее.

> **Источник:**
>
> - Race Detector — официальная документация:
>   <https://go.dev/doc/articles/race_detector>
> - The Go Memory Model (определение гонки):
>   <https://go.dev/ref/mem>

---

### sync.Mutex vs sync.RWMutex

#### sync.Mutex

Простейший примитив взаимного исключения. Два метода:

```go
mu.Lock()   // захватить — другие гороутины ждут
// ... критическая секция ...
mu.Unlock() // освободить
```

**Правила:**

- Только одна гороутина может держать лок одновременно.
- `Unlock()` для незалоченного мьютекса — **паника**.
- Мьютекс **нельзя копировать** после первого использования (`go vet`
  поймает ошибку `lock value copy`). Передавайте по указателю.

#### sync.RWMutex

Read-write мьютекс: различает «читателей» и «писателей».

```go
mu.RLock()    // много читателей могут держать одновременно
// ... чтение ...
mu.RUnlock()

mu.Lock()     // эксклюзивно — никто другой не может ни читать, ни писать
// ... запись ...
mu.Unlock()
```

#### Когда RWMutex проигрывает Mutex

Кажется, что RWMutex всегда лучше — ведь чтение может быть параллельным.
На практике **RWMutex сложнее внутри**:

- Нужно вести счётчик активных читателей.
- Нужно отслеживать, ждёт ли writer (и не давать новым читателям
  «обгонять» writer).
- Каждый RLock/RUnlock — это атомарная операция + потенциально cache
  contention.

Если критическая секция **короткая** (меньше ~50 наносекунд, например,
простое чтение из map), накладные расходыы RWMutex перевешивают
выгоду от параллелизма чтения. Обычный `Mutex` оказывается **быстрее**.

RWMutex выгоден, когда:

- Критическая секция **длинная** (сотни наносекунд или микросекунды).
- Чтений **значительно больше**, чем записей (например, 95% / 5%).
- Конкурентность (много горутин одновременно читают).

**Правило: измеряйте.** Бенчмарк `go test -bench` решает спор.

> **Источник:**
>
> - sync.RWMutex — официальная документация:
>   <https://pkg.go.dev/sync#RWMutex>
> - Bryan C. Mills — «Design Principles for Synchronization» (внутренний
>   документ Go team, на который ссылается документация):
>   см. обсуждение в <https://github.com/golang/go/issues/17973>

#### Дополнительно: defer и hot-path

```go
mu.Lock()
defer mu.Unlock()
```

`defer` безопасен (лок гарантированно освободится при выходе из функции),
но добавляет **один вызов функции** (~10–20 нс). В горячих путях (hot path,
миллионы вызовов в секунду) лучше писать `mu.Unlock()` явно.

> **Источник:** Defer, Panic, and Recover (Go Blog):
> <https://go.dev/blog/defer-panic-and-recover>

---

### sync/atomic — lock-free счётчики

Пакет `sync/atomic` предоставляет **атомарные** операции для примитивных
типов: чтение, запись, добавление, compare-and-swap (CAS).

#### Типы-обёртки (Go 1.19+)

До Go 1.19 приходилось пользоваться функциями вроде `atomic.AddInt64(&n, 1)`,
что было неудобно и небезопасно (можно перепутать типы). С Go 1.19 появились
типы-обёртки:

```go
var counter atomic.Int64
counter.Add(1)
val := counter.Load()
counter.Store(0)
```

Доступные типы: `atomic.Int64`, `atomic.Uint64`, `atomic.Bool`,
`atomic.Pointer[T]`, `atomic.Int32`, `atomic.Uint32`.

#### Compare-and-Swap (CAS)

Основа всех lock-free алгоритмов. CAS проверяет, равно ли текущее значение
ожидаемому `old`, и если да — заменяет на `new`. Атомарно.

```go
for {
    old := state.Load()
    new := compute(old)
    if state.CompareAndSwap(old, new) {
        break // успешно
    }
    // кто-то успел раньше — пробуем снова
}
```

#### Когда atomic, а когда mutex?

- **atomic** — для одной переменной (счётчик, флаг). Быстрее мьютекса в
  5–10 раз (~1–5 нс против ~20–50 нс).
- **mutex** — для защиты **нескольких** связанных переменных или сложной
  критической секции.

**Важно:** atomic даёт только атомарность одной операции. Он **не даёт**
memory ordering для surrounding code без явного барьера. Если нужна
сложная логика — берите mutex.

> **Источник:**
>
> - sync/atomic — документация:
>   <https://pkg.go.dev/sync/atomic>
> - atomic types (Go 1.19 release notes):
>   <https://go.dev/doc/go1.19#atomic_types>

---

### Каналы: buffered vs unbuffered

Канал (`chan T`) — встроенный тип Go для общения между горутинами.

#### Unbuffered (`make(chan T)`)

Размер буфера = 0. **Синхронный** handshake:

- Отправитель `ch <- v` **блокируется**, пока получатель не сделает `<-ch`.
- Получатель `<-ch` **блокируется**, пока отправитель не запишет.

Передача и приём происходят **одновременно** (rendezvous). Используется для:

- Сигналов (например, `done := make(chan struct{})`).
- Гарантии получения данных.
- Синхронизации «point-to-point».

#### Buffered (`make(chan T, n)`)

Размер буфера = n. Отправитель **не блокируется**, пока буфер не полон:

```go
ch := make(chan int, 3)
ch <- 1 // OK
ch <- 2 // OK
ch <- 3 // OK
ch <- 4 // блокируется — буфер полон
```

Используется для:

- Сглаживания burst'ов (producer быстрее consumer'а).
- Batch-обработки.
- Реализации semaphore.

#### Close + range

`close(ch)` сигнализирует, что данных больше не будет. Получатель может
использовать `range`:

```go
for v := range ch {
    fmt.Println(v) // выйдет из цикла при close
}
```

**Правила:**

- Закрывать канал должен **отправитель**, не получатель.
- Запись в закрытый канал — **паника**.
- Чтение из закрытого канала — возвращает zero-value + `ok == false`.
- Повторное закрытие закрытого канала — **паника**.

> **Источник:**.Effective Go — Channels:
> <https://go.dev/doc/effective_go#channels>
> и Go by Example — Channels:
> <https://gobyexample.com/channels>

---

### Сигнальный канал: chan struct{}

`struct{}` — тип без данных (zero-size). Канал `chan struct{}` — это
**сигнальный** канал: он не передаёт значения, только факт события.

```go
done := make(chan struct{})

go func() {
    // ... долгая работа ...
    close(done) // → broadcast: уведомить всех ждущих
}()

<-done // блокируется до close
```

**Почему close, а не запись?**

`close(done)` уведомляет **всех** получателей одновременно (broadcast).
Если бы мы писали `done <- struct{}{}`, уведомился бы только один.

В связке с `select` можно реализовать тайм-аут:

```go
select {
case <-done:
    fmt.Println("done")
case <-time.After(time.Second):
    fmt.Println("timeout")
}
```

> **Источник:** Go by Example — Closing Channels:
> <https://gobyexample.com/closing-channels>
> и Dave Cheney — «Constant errors» (использование `struct{}`):
> <https://dave.cheney.net/2016/04/07/constant-errors>

---

### select — мультиплексирование каналов

`select` — это «switch для каналов». Он ждёт, пока хотя бы один `case`
не будет готов, и выполняет его.

```go
select {
case v := <-ch1:
    // ch1 готов
case ch2 <- 42:
    // запись в ch2 прошла
case <-time.After(time.Second):
    // тайм-аут
default:
    // никто не готов — не блокируемся
}
```

**Правила:**

1. Если готовы **несколько** case — select выбирает **случайный** (не
   первый, не приоритетный — именно случайный). Это prevents starvation.
2. `default` делает select **неблокирующим**: если никто не готов,
   выполняется default.
3. Без default — select блокируется до первого готового case.
4. `nil`-канал в select никогда не готов — можно использовать для
   динамического включения/выключения case (установите канал в nil,
   чтобы перестать его слушать).

> **Источник:** Effective Go — A select:
> <https://go.dev/doc/effective_go#control-structures-select>
> и Go Language Spec — Select statements:
> <https://go.dev/ref/spec#Select_statements>

---

### context.Context — отмена и дедлайны

`context.Context` — интерфейс для передачи **дедлайна, сигнала отмены и
значений** через границу вызова (обычно HTTP/gRPC handler → service → repo).

#### Три функции-конструктора

```go
ctx, cancel := context.WithCancel(parent)        // ручная отмена
ctx, cancel := context.WithTimeout(parent, 5*time.Second) // дедлайн через 5с
ctx       := context.WithValue(parent, key, val) // значение
```

**Контракт:** `cancel()` нужно вызывать **всегда** (через `defer cancel()`),
иначе ресурсы контекста (таймер, гороутина) утечут.

#### ctx.Err() и ctx.Done()

- `ctx.Done()` — возвращает канал, который закрывается при отмене или
  истечении дедлайна. Используется в `select`:

  ```go
  select {
  case <-ctx.Done():
      return ctx.Err() // context.Canceled или context.DeadlineExceeded
  case result := <-work:
      return result, nil
  }
  ```

- `ctx.Err()` — возвращает причину отмены (`nil`, если ещё активен).

#### Антипаттерны

1. **Хранить context.Context в поле структуры.** Контекст предназначен
   для **сквозной передачи** через параметры функции, не для хранения.
   Исключение — когда структура является request-scoped (например,
   request-объект, живущий один запрос).
2. **Использовать context.WithValue как глобал.** Значения в контексте —
   для request-scoped данных (requestID, userID), не для конфигурации
   или инъекции зависимостей.
3. **Передавать nil context.** Если нет родителя — используйте
   `context.Background()` или `context.TODO()`.

> **Источник:**
>
> - context — документация:
>   <https://pkg.go.dev/context>
> - Go Blog — «Context and Structs» (Chloe Yip, 2024):
>   <https://go.dev/blog/context-and-structs>
> - Go Blog — «Go Concurrency Patterns: Context»:
>   <https://go.dev/blog/context>

---

### sync.* — швейцарский нож

| Примитив | Зачем | Подводный камень |
| --- | --- | --- |
| `sync.Pool` | Кеш временных объектов (буферы, декодеры). | Может быть очищен в любой момент (например, при GC). Не подходит как кеш. |
| `sync.Once` | Инициализация ровно один раз. | Не подходит для ленивых перерасчётов — для них есть `sync.OnceFunc` / `OnceValue` (Go 1.21+). |
| `sync.Cond` | Ожидание условия. | Почти всегда лучше канал или atomic. Редко нужен. |
| `sync.Map` | Concurrent map для редких writes. | Медленнее `map + Mutex` при частых writes. |
| `sync.WaitGroup` | Ждать N горутин. | `Add(1)` **до** `go`, не внутри. Копировать нельзя. |

#### sync.Pool

`Pool` — это набор временных объектов, которые можно переиспользовать,
избегая аллокаций. GC может очистить Pool в любой момент — поэтому
**Pool не подходит как кеш** данных. Подходит для:

- `bytes.Buffer` при сериализации.
- Временных слайсов/мап для декодирования JSON.
- Объектов с дорогой инициализацией.

```go
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

buf := bufPool.Get().(*bytes.Buffer)
defer bufPool.Put(buf)
buf.Reset()
// ... используем buf ...
```

#### sync.Once

Гарантирует, что функция `Do(f)` выполнится ровно один раз, даже если
вызвана из N горутин одновременно.

```go
var (
    once   sync.Once
    instance *DB
)

func GetDB() *DB {
    once.Do(func() {
        instance = connectDB() // ровно один раз
    })
    return instance
}
```

#### sync.WaitGroup

Счётчик ожидающих горутин. `Add(n)` увеличивает счётчик, `Done()`
уменьшает (эквивалент `Add(-1)`), `Wait()` блокируется, пока счётчик не
станет нулём.

**Типичная ошибка:** `wg.Add(1)` **после** `go func()` — может привести к
тому, что `wg.Wait()` вернётся до запуска гороутины. Правильно:

```go
wg.Add(1)           // ДО запуска go
go func() {
    defer wg.Done()
    // ...
}()
```

> **Источник:**
>
> - sync — документация:
>   <https://pkg.go.dev/sync>
> - sync.Pool — Go Blog «Profiling Go Programs»:
>   <https://go.dev/blog/pprof>
> - Go 1.21 — OnceFunc / OnceValue:
>   <https://go.dev/doc/go1.21#sync>

---

## Часть 4. Паттерны конкурентности — Workshop

Каждый паттерн — это runnable-пример в `workshop/`. Ниже разбирается теория.

### Workshop 1. Worker Pool

**Проблема:** нужно обработать поток задач, но ограничить число одновременных
обработчиков (чтобы не положить внешнюю систему — БД, API).

**Решение:** фиксированное число горутин-воркеров, читающих из общего канала
задач.

```go
jobs := make(chan Job, 100)
results := make(chan Result, 100)

for w := 0; w < N; w++ {
    go worker(jobs, results)
}

for _, job := range allJobs { jobs <- job }
close(jobs) // сигнал: работы больше нет

for r := 0; r < len(allJobs); r++ {
    <-results // собираем результаты
}
```

**Ключевые моменты:**

1. `close(jobs)` — это сигнал «работы больше нет». Воркеры выходят из
   `for j := range jobs`.
2. Каналы `jobs` и `results` **буферизованы**, чтобы producer не
   блокировался.
3. Число воркеров `N` подбирается: для CPU-bound ≈ `NumCPU`, для I/O-bound
   можно больше (10–100), ограничено лишь внешней системой.

**Пример кода:** `workshop/07_worker_pool/main.go`

> **Источник:** Go by Example — Worker Pools:
> <https://gobyexample.com/worker-pools>
> и Effective Go — Channel-based parallel computation:
> <https://go.dev/doc/effective_go#channels>

---

### Workshop 2. Fan-Out / Fan-In (слайд 24)

**Fan-Out** — раздать работу нескольким параллельным обработчикам.
**Fan-In** — собрать результаты с нескольких каналов в один.

```go
// Fan-Out: несколько stage-горутин читают из одного src
outs := make([]<-chan T, n)
for i := range outs {
    outs[i] = process(src)
}

// Fan-In: merge — собирает все в один канал
out := merge(outs...)
```

**Merge-функция** — стандартная идиома:

```go
func merge[T any](chans ...<-chan T) <-chan T {
    var wg sync.WaitGroup
    out := make(chan T)
    wg.Add(len(chans))
    for _, c := range chans {
        go func(c <-chan T) {
            defer wg.Done()
            for v := range c { out <- v }
        }(c)
    }
    go func() { wg.Wait(); close(out) }()
    return out
}
```

**Важно:** `close(out)` вызывается **после** завершения всех горутин-источников
(через `wg.Wait()`). Это гарантирует, что выходной канал закрывается только
когда все данные собраны.

**Пример кода:** `workshop/08_fan_in_fan_out/main.go`

> **Источник:** Go Concurrency Patterns — Fan-in/Fan-out (Go Blog):
> <https://go.dev/blog/pipelines>
> («Go Concurrency Patterns: Pipelines and cancellation»)

---

### Workshop 3. Pipeline (слайд 25)

**Pipeline** — цепочка стадий, каждая из которых читает из канала и пишет
в канал.

```go
gen → square → filterEven → print
```

Каждая стадия:

- Принимает `<-chan T` (read-only канал).
- Возвращает `<-chan T` (новый канал).
- Внутри запускает гороутину, которая читает вход, обрабатывает, пишет
  выход.
- Обязательно закрывает выходной канал через `defer close(out)`.

```go
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)       // ← ОБЯЗАТЕЛЬНО
        for n := range in {
            out <- n * n
        }
    }()
    return out
}
```

**Проблема:** если consumer прерывается (например, по ctx.Done()),上游
стадии заблокируются на записи (никто не читает out). Решение —
pipeline с cancellation (см. Go Blog ниже).

**Пример кода:** `workshop/09_pipeline/main.go`

> **Источник:** Go Blog — «Go Concurrency Patterns: Pipelines and
> cancellation»:
> <https://go.dev/blog/pipelines>

---

### Workshop 4. errgroup — структурированная конкурентность

**Проблема:** запускаем N горутин, но хотим отменить все, если **хоть одна**
вернёт ошибку.

`golang.org/x/sync/errgroup` — это «умный WaitGroup с обработкой ошибок».

```go
g, ctx := errgroup.WithContext(context.Background())
g.SetLimit(10) // ограничить параллелизм (Go 1.x)

for _, url := range urls {
    url := url
    g.Go(func() error {
        return fetch(ctx, url)
    })
}

if err := g.Wait(); err != nil {
    // первая ошибка из любой гороутины
}
```

**Как это работает:**

1. `errgroup.WithContext(parent)` возвращает `*Group` и **производный
   контекст** `ctx`.
2. При возврате **первой** ошибки из `g.Go(f)` — `ctx` автоматически
   **отменяется** (`cancel()`).
3. Остальные гороутины, если они уважают `ctx`, увидят отмену и завершатся.
4. `g.Wait()` возвращает **первую** ошибку (или `nil`, если все успешно).

**SetLimit(n)** (добавлено в `errgroup` позже) — встроенный семафор:
не более `n` горутин из группы работают одновременно. Эквивалент семафора
из Workshop 5, но встроен.

**Capture loop variable:** До Go 1.22 нужно явно копировать переменную
цикла (`url := url`), иначе все гороутины получат последнее значение.
С Go 1.22 (`go.mod` с `go 1.22`) переменная цикла создаётся заново на
каждой итерации — копирование не нужно.

**Пример кода:** `workshop/10_errgroup/main.go`

> **Источник:**
>
> - errgroup — документация:
>   <https://pkg.go.dev/golang.org/x/sync/errgroup>
> - Go 1.22 — loop variable experiment:
>   <https://go.dev/blog/loopvar-preview>
> - Scott Mansfield — «Goroutine leaks» (Netflix TechBlog):
>   <https://medium.com/@mdwheele/goroutine-leaks-50b45b3c6dde>

---

### Workshop 5. Semaphore на chan struct{} (слайд 27)

**Семафор** ограничивает число одновременных операций. В Go можно построить
его только на stdlib — через буферизованный канал.

```go
type Semaphore chan struct{}

func New(max int) Semaphore {
    return make(Semaphore, max) // буфер = max слотов
}

func (s Semaphore) Acquire() { s <- struct{}{} } // занять слот
func (s Semaphore) Release() { <-s }              // освободить слот
```

- `Acquire()` — запись в канал. Если буфер полон (все слоты заняты) —
  блокируется.
- `Release()` — чтение из канала. Освобождает слот.

**Использование:**

```go
sem := New(3) // не больше 3 одновременных
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        sem.Acquire()
        defer sem.Release()
        // ... не более 3 горотин одновременно тут ...
    }()
}
```

**Для production** лучше использовать `golang.org/x/sync/semaphore.Weighted`,
который:

- Поддерживает взвешенные семафоры (можно занять 2 «билета» сразу).
- Корректно работает с `context.Context` (отмена ожидания).

**Пример кода:** `workshop/11_semaphore_chan/main.go`

> **Источник:**
>
> - golang.org/x/sync/semaphore — документация:
>   <https://pkg.go.dev/golang.org/x/sync/semaphore>
> - Разбор semaphore на каналах (Go by Example):
>   <https://gobyexample.com/rate-limiting>

---

### Workshop 6. Goroutine leak — было/стало (слайд 28)

**Утечка гороутины** — ситуация, когда гороутина заблокирована навсегда и
не может завершиться, занимая ресурсы (стек, память, запись в локе).

#### Типичный сценарий утечки

```go
func fetch() <-chan int {
    ch := make(chan int) // НЕбуферизованный
    go func() {
        ch <- compute() // ← повиснет, если читатель ушёл
    }()
    return ch
}

select {
case v := <-fetch():
    // ...
case <-time.After(time.Second):
    // timeout — читатель ушёл, но отправитель ВИСНЕТ НАВСЕГДА
}
```

Если читатель отвалился по тайм-ауту, отправитель `ch <- compute()` будет
ждать вечно. Гороутина утекла — её стек (~2 КБ), `compute()` результат
всё в памяти, и GC не может это освободить, пока гороутина «достижима» из
корней (а она достижима — через планировщик).

#### Решение: context + буферизованный канал

```go
func fetch(ctx context.Context) (int, error) {
    ch := make(chan int, 1) // ← буфер = 1
    go func() {
        ch <- compute()
    }()

    select {
    case v := <-ch:
        return v, nil
    case <-ctx.Done():
        return 0, ctx.Err()
    }
}
```

**Почему буфер = 1 решает проблему:** даже если читатель ушёл по `ctx.Done()`,
отправитель `ch <- compute()` не блокируется (в буфере есть место) и
гороутина **завершается**. Результат остаётся в канале и будет собран GC.

**Правило:** гороутина должна знать, **как она завершится**. Передавайте
`context.Context` в каждую функцию, которая может быть отменена.

**Пример кода:** `workshop/12_goroutine_leak/main.go`

> **Источник:**
>
> - Dave Cheney — «Never start a goroutine without knowing how it stops»:
>   <https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop>
> - Ardan Labs — «Goroutine Leaks: The Forgotten Sender»:
>   <https://www.ardanlabs.com/blog/2018/11/goroutine-leaks-the-forgotten-sender.html>
> - Scott Mansfield — «Goroutine leaks» (обзор):
>   <https://medium.com/@mdwheele/goroutine-leaks-50b45b3c6dde>

---

## Часть 5. Профилирование и отладка

### Слайд 30. Race detector: -race

Флаг `-race` включает встроенный детектор гонок данных, основанный на
Google ThreadSanitizer (TSan). Он instrumentation'ом отслеживает все
обращения к памяти во время выполнения и сообщает о гонках, которые
**реально произошли**.

```bash
go run -race main.go
go test -race ./...
go build -race -o myapp
```

#### Как это работает

При сборке с `-race` компилятор добавляет instrumentation-код вокруг
каждого чтения/записи памяти. Runtime отслеживает:

- Кто (какая гороутина) обращался к адресу.
- Happens-before отношения (через мьютексы, каналы и т.д.).

Если два обращения без happens-before и хотя бы одно — запись, детектор
печатает предупреждение со стеками обоих горутин.

#### Ограничения

1. **Накладные расходы:** CPU ~2–10x, память ~5–10x.
2. **Гонки, которые не случились, не будут пойманы.** Детектор
   динамический, не статический.
3. **Требуется cgo** (на некоторых платформах).
4. **Не включайте в production под нагрузкой** — overhead слишком велик.
   Но можно включать в staging / канареечных развертываниях.

#### Настройка через GORACE

```bash
GORACE="log_path=/tmp/race strip_path_prefix=/my/sources" go test -race
```

Параметры: `log_path`, `exitcode` (по умолчанию 66),
`history_size`, `halt_on_error`.

> **Источник:**
>
> - Data Race Detector — официальная документация:
>   <https://go.dev/doc/articles/race_detector>
> - ThreadSanitizer (база):
>   <https://research.google/pubs/pub35604/>

---

### Слайд 31. Поиск утечек горутин: pprof

`net/http/pprof` — пакет, который регистрирует debug-эндпоинты на
`DefaultServeMux`. Импортируйте его (blank import), запустите HTTP-сервер —
и получаете живой профиль работающего приложения:

```go
import _ "net/http/pprof"

func main() {
    go http.ListenAndServe("localhost:6060", nil)
    // ... ваше приложение ...
}
```

#### Ключевые эндпоинты

| URL | Что показывает |
| --- | --- |
| `/debug/pprof/` | Индексная страница со всеми профилями |
| `/debug/pprof/goroutine?debug=1` | Текстовый список всех горутин + стеки |
| `/debug/pprof/goroutine?debug=2` | Полная информация (со статусами) |
| `/debug/pprof/heap` | Профиль аллокаций |
| `/debug/pprof/profile?seconds=30` | CPU-профиль за 30 секунд |
| `/debug/pprof/block` | Профиль блокировок |
| `/debug/pprof/mutex` | Профиль contention'а мьютексов |

#### Поиск утечек

```bash
# Сколько горотин сейчас живёт?
curl http://localhost:6060/debug/pprof/goroutine?debug=1 | head -5

# Если число растёт без меры — утекла. Смотрим стеки:
curl http://localhost:6060/debug/pprof/goroutine?debug=1 | grep -A 10 "main."
```

В коде можно просто проверить:

```go
fmt.Println(runtime.NumGoroutine())
```

Если это число растёт со временем — у вас утечка.

#### Интерактивный pprof

```bash
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top
(pprof) list main.*
(pprof) web  # граф в браузере (нужен graphviz)
```

> **Источник:**
>
> - net/http/pprof — документация:
>   <https://pkg.go.dev/net/http/pprof>
> - runtime/pprof — для non-HTTP приложений:
>   <https://pkg.go.dev/runtime/pprof>
> - Diagnostics — официальная документация Go:
>   <https://go.dev/doc/diagnostics>

---

### Слайд 32. go tool trace — рентген планировщика

`go tool trace` — это **визуализатор трейса выполнения** (execution tracer).
В отличие от pprof (агрегированные данные), trace показывает
**таймлайн** каждого события: когда гороутина запустилась, заблокировалась,
какой M/P её выполнял.

#### Запись трейса

```go
import (
    "os"
    "runtime/trace"
)

func main() {
    f, _ := os.Create("trace.out")
    defer f.Close()
    trace.Start(f)
    defer trace.Stop()

    // ... ваш код ...
}
```

Или через HTTP-эндпоинт pprof:

```bash
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5
```

#### Просмотр

```bash
go tool trace trace.out
```

Откроется веб-интерфейс (на `localhost:порт`) со следующими разделами:

- **Goroutine analysis** — для каждой гороутины: сколько времени она была
  в каждом статусе (running, runnable, waiting, blocked).
- **Scheduler latency profile** — где гороутины ждали дольше всего.
- **Network blocking profile** — где ждали I/O.
- **Heap / GC events** — моменты сборки мусора и их длительность.
- **View trace** — интерактивный таймлайн (помогает понять, почему
  приложение не так параллельно, как кажется).

#### Когда использовать

- «Почему моё приложение не утилизирует CPU на 100%?» — посмотреть
  Goroutine analysis, найти блокировки.
- «Почему задержки выросли?» — Scheduler latency profile.
- «Какой GC overhead?» — Heap / GC events.

> **Источник:**
>
> - runtime/trace — документация:
>   <https://pkg.go.dev/runtime/trace>
> - Go Blog — «More powerful Go execution traces» (Michael Knyszek, 2024):
>   <https://go.dev/blog/execution-traces-2024>
> - Diagnostics — официальная документация Go (pprof + trace):
>   <https://go.dev/doc/diagnostics>

---

### Слайд 33. Deadlock-чеклист

| Симптом | Причина | Лекарство |
| --- | --- | --- |
| Все гороутины спят, fatal error | Циклическая блокировка мьютексов | Фиксируйте порядок захвата; не держите лок при вызове чужого кода |
| Тайм-аут в проде, локально ОК | Небуферизованный канал без читателя | Буфер или `select + ctx.Done()` |
| Растёт `NumGoroutine` | Утечка — отправитель без приёмника | `chan` buffered=1 + `context.Cancel()` |
| Sometimes-гонка в проде | Общий map/slice без синхронизации | `sync.Mutex` или `sync.Map`, race-тесты |
| GC stop-the-world > 100 мс | Слишком много горутин / большой heap | Пул воркеров, `sync.Pool`, лимит параллелизма |

#### Классические deadlock'и

**1. Циклическая блокировка мьютексов:**

```go
// Goroutine A: mu1.Lock(); mu2.Lock()
// Goroutine B: mu2.Lock(); mu1.Lock()
// ← дедлок, если A держит mu1 и ждёт mu2, а B держит mu2 и ждёт mu1
```

**Решение:** зафиксировать **глобальный порядок захвата** (всегда сначала
mu1, потом mu2).

**2. Забытый receiver:**

```go
ch := make(chan int)
go func() { ch <- 1 }() // повиснет, если никто не читает
// если main не сделает <-ch — deadlock
```

**Решение:** буферизованный канал или гарантированный reader.

**3. Копирование мьютекса:**

```go
type Cache struct { mu sync.Mutex }
c := Cache{}
c2 := c // ← копирование мьютекса! state corrupt'ится
```

**Решение:** передавайте `*Cache`, не `Cache`. `go vet` ловит это.

> **Источник:**
>
> - Effective Go — Concurrency:
>   <https://go.dev/doc/effective_go#concurrency>
> - Dave Cheney — «Never start a goroutine without knowing how it stops»:
>   <https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop>

---

## Часть 6. Q\&A и ссылки

### Слайд 35. Что почитать и посмотреть

#### Обязательное чтение

1. **Effective Go — Concurrency.** Официальный тур по горутинам и каналам,
   базовые паттерны.
   <https://go.dev/doc/effective_go#concurrency>

2. **The Go Memory Model.** Определяет, когда чтение переменной в одной
   гороутине гарантированно увидит запись из другой. Обязательно для
   понимания happens-before.
   <https://go.dev/ref/mem>

3. **Go Blog — «Go Concurrency Patterns: Pipelines and cancellation».»
   Разбор pipeline, fan-out/in, cancellation через context.
   <https://go.dev/blog/pipelines>

4. **Go Blog — «Context».** Почему и как использовать `context.Context`.
   <https://go.dev/blog/context>

5. **Dmitry Vyukov — «Scalable Go Scheduler Design Doc».** Классика про
   планировщик GMP. Самый авторитетный первоисточник по внутренностям.
   <https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw>

#### Документация пакетов

1. **sync** — Mutex, RWMutex, WaitGroup, Once, Map, Pool, Cond:
   <https://pkg.go.dev/sync>

2. **sync/atomic** — атомарные операции:
   <https://pkg.go.dev/sync/atomic>

3. **context** — отмена, дедлайны, значения:
   <https://pkg.go.dev/context>

4. **golang.org/x/sync/errgroup** — структурированная конкурентность:
   <https://pkg.go.dev/golang.org/x/sync/errgroup>

5. **golang.org/x/sync/semaphore** — взвешенный семафор:
    <https://pkg.go.dev/golang.org/x/sync/semaphore>

6. **net/http/pprof** — HTTP-профайлер:
    <https://pkg.go.dev/net/http/pprof>

7. **runtime/trace** — execution tracer:
    <https://pkg.go.dev/runtime/trace>

#### Блоги и статьи

 1. **Dave Cheney — «Never start a goroutine without knowing how it stops».»
    Главный принцип предотвращения утечек.
    <https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop>

 2. **A Journey With Go (Vincent Blanchon).** Глубокие разборы внутренностей
    runtime — планировщик, GC, preemption.
    <https://medium.com/a-journey-with-go>

 3. **Ardan Labs — «Scheduling In Go».» Серия статей Уильяма Кеннеди
    про планировщик (OS scheduler → Go scheduler → GMP).
    Part 1 (OS scheduler):
    <https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part1.html>
    Part 2 (Go scheduler):
    <https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part2.html>

#### Книги

 1. **Katherine Cox-Buday — «Concurrency in Go».» (O'Reilly, 2017).
    Самая глубокая книга по теме: паттерны, ловушки, инструменты.

 2. **Bryan C. Mills — «Design Principles for Synchronization».»
    Внутренний документ Go team (на него ссылается документация `sync`).
    См. обсуждение в <https://github.com/golang/go/issues/17973>

#### Видео

 1. **GopherCon 2017 — «Concurrency Patterns» (Rob Pike).**
    <https://www.youtube.com/watch?v=f6kdp27TYZs>

 2. **GopherCon 2018 — «Design Principles for Synchronization» (Bryan C. Mills).»
    <https://www.youtube.com/watch?v=K2pxX8ohTG0>

 3. **GopherCon 2020 — «Pardon the Interruption: Loop Preemption in Go
    1.14» (Austin Clements).»
    <https://www.youtube.com/watch?v=1I1WmeSjRSw>

#### Release notes (история эволюции)

 1. **Go 1.5** — GOMAXPROCS default = NumCPU:
    <https://go.dev/doc/go1.5>

 2. **Go 1.14** — асинхронный preemption:
    <https://go.dev/doc/go1.14>

 3. **Go 1.19** — atomic типы-обёртки (`atomic.Int64` и т.д.):
    <https://go.dev/doc/go1.19#atomic_types>

 4. **Go 1.21** — `OnceFunc`, `OnceValue`, `min`/`max`, `clear`:
    <https://go.dev/doc/go1.21#sync>

 5. **Go 1.22** — переменная цикла создаётся заново на каждой итерации:
    <https://go.dev/blog/loopvar-preview>

---

### Слайд 38. Финал

Всем спасибо! Вопросы — в чат или на почту. Код примеров — в репозитории
курса в каталоге `concurrency/workshop/`.
