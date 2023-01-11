package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"practicum-webinars/devops/internal/agent"
)

const (
	reportInterval = 5 * time.Second
	pollInterval   = 2 * time.Second

	address = "localhost:8080"
)

func main() {
	fmt.Println("Start monitor client")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обрабатывем сигналы от системы.
	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Регистируем простейший обработчик для выгрузки репортов.
	reporter, err := agent.NewReporter(address)
	if err != nil {
		panic(err)
	}
	scopeOpt := agent.ScopeOptions{Reporter: reporter}
	scope, closer := agent.NewRootScope(ctx, scopeOpt, reportInterval)
	defer closer.Close()

	// Запускаем процесс мониторинга с заданным интервалом.
	go newMemMonitor(ctx, scope, pollInterval)

	// Ожидаем формирование условий, для завершения приложения.
	sig := <-termSignal
	fmt.Println("Finished, reason:", sig.String())
}

func newMemMonitor(ctx context.Context, scope agent.Scope, pollInterval time.Duration) {
	rPollCount := scope.Counter("PollCount")
	rAlloc := scope.Gauge("Alloc")
	rTotalAlloc := scope.Gauge("TotalAlloc")
	rSys := scope.Gauge("Sys")
	rLookups := scope.Gauge("Lookups")
	rMallocs := scope.Gauge("Mallocs")
	rFrees := scope.Gauge("Frees")
	rHeapAlloc := scope.Gauge("HeapAlloc")
	rHeapSys := scope.Gauge("HeapSys")
	rHeapIdle := scope.Gauge("HeapIdle")
	rHeapInuse := scope.Gauge("HeapInuse")
	rHeapReleased := scope.Gauge("HeapReleased")
	rHeapObjects := scope.Gauge("HeapObjects")
	rStackInuse := scope.Gauge("StackInuse")
	rStackSys := scope.Gauge("StackSys")
	rMSpanInuse := scope.Gauge("MSpanInuse")
	rMSpanSys := scope.Gauge("MSpanSys")
	rMCacheInuse := scope.Gauge("MCacheInuse")
	rMCacheSys := scope.Gauge("MCacheSys")
	rBuckHashSys := scope.Gauge("BuckHashSys")
	rGCSys := scope.Gauge("GCSys")
	rOtherSys := scope.Gauge("OtherSys")
	rNextGC := scope.Gauge("NextGC")
	rLastGC := scope.Gauge("LastGC")
	rPauseTotalNs := scope.Gauge("PauseTotalNs")
	rNumGC := scope.Gauge("NumGC")
	rNumForcedGC := scope.Gauge("NumForcedGC")
	rGCCPUFraction := scope.Gauge("GCCPUFraction")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	var rtm runtime.MemStats

	// Обновляем счетчики с заданной периодичностью.
	// Значения (и описание) получаем из пакета runtime.
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("teminate monitor goroutine, reason: %s\n", ctx.Err())
			return
		case <-ticker.C:
		}

		rPollCount.Inc(1)
		fmt.Printf("update monitor metrics with interval: %s\n", pollInterval)
		// Read full mem stats
		runtime.ReadMemStats(&rtm)

		// General statistics.
		// Alloc is bytes of allocated heap objects.
		rAlloc.Update(float64(rtm.Alloc))
		// TotalAlloc is cumulative bytes allocated for heap objects.
		rTotalAlloc.Update(float64(rtm.TotalAlloc))
		// Sys is the total bytes of memory obtained from the OS.
		rSys.Update(float64(rtm.Sys))
		// Lookups is the number of pointer lookups performed by the runtime.
		rLookups.Update(float64(rtm.Lookups))
		// Mallocs is the cumulative count of heap objects allocated.
		// The number of live objects is Mallocs - Frees.
		rMallocs.Update(float64(rtm.Mallocs))
		// Frees is the cumulative count of heap objects freed.
		rFrees.Update(float64(rtm.Frees))

		// Heap memory statistics.
		// HeapAlloc is bytes of allocated heap objects.
		rHeapAlloc.Update(float64(rtm.HeapAlloc))
		// HeapSys is bytes of heap memory obtained from the OS.
		rHeapSys.Update(float64(rtm.HeapSys))
		// HeapIdle is bytes in idle (unused) spans.
		rHeapIdle.Update(float64(rtm.HeapIdle))
		// HeapInuse is bytes in in-use spans.
		rHeapInuse.Update(float64(rtm.HeapInuse))
		// HeapReleased is bytes of physical memory returned to the OS.
		rHeapReleased.Update(float64(rtm.HeapReleased))
		// HeapObjects is the number of allocated heap objects.
		rHeapObjects.Update(float64(rtm.HeapObjects))

		// Stack memory statistics.
		// StackInuse is bytes in stack spans.
		rStackInuse.Update(float64(rtm.StackInuse))
		// StackSys is bytes of stack memory obtained from the OS.
		rStackSys.Update(float64(rtm.StackSys))

		// Off-heap memory statistics.
		// MSpanInuse is bytes of allocated mspan structures.
		rMSpanInuse.Update(float64(rtm.MSpanInuse))
		// MSpanSys is bytes of memory obtained from the OS for mspan
		// structures.
		rMSpanSys.Update(float64(rtm.MSpanSys))
		// MCacheInuse is bytes of allocated mcache structures.
		rMCacheInuse.Update(float64(rtm.MCacheInuse))
		// MCacheSys is bytes of memory obtained from the OS for mcache structures.
		rMCacheSys.Update(float64(rtm.MCacheSys))
		// BuckHashSys is bytes of memory in profiling bucket hash tables.
		rBuckHashSys.Update(float64(rtm.BuckHashSys))
		// GCSys is bytes of memory in garbage collection metadata.
		rGCSys.Update(float64(rtm.GCSys))
		// OtherSys is bytes of memory in miscellaneous off-heap runtime allocations.
		rOtherSys.Update(float64(rtm.OtherSys))

		// Garbage collector statistics.
		// NextGC is the target heap size of the next GC cycle.
		rNextGC.Update(float64(rtm.NextGC))
		// LastGC is the time the last garbage collection finished, as
		// nanoseconds since 1970 (the UNIX epoch).
		rLastGC.Update(float64(rtm.LastGC))
		// PauseTotalNs is the cumulative nanoseconds in GC
		// stop-the-world pauses since the program started.
		rPauseTotalNs.Update(float64(rtm.PauseTotalNs))
		// NumGC is the number of completed GC cycles.
		rNumGC.Update(float64(rtm.NumGC))
		// NumForcedGC is the number of GC cycles that were forced by
		// the application calling the GC function.
		rNumForcedGC.Update(float64(rtm.NumForcedGC))
		// GCCPUFraction is the fraction of this program's available
		// CPU time used by the GC since the program started.
		rGCCPUFraction.Update(rtm.GCCPUFraction)
	}
}
