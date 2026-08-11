// Workshop 02. sync.Mutex vs sync.RWMutex.
//
// Этот пакет — бенчмарк. Запуск:
//
//	go test -bench=. -benchmem ./02_mutex_vs_rwmutex
//
// Что показать:
//  1. При короткой критической секции RWMutex часто проигрывает обычному Mutex
//     из-за накладных расходов на подсчёт читателей и пробуждение writer'а.
//  2. RLock выгоден, когда чтение «длинное», а запись — редкая.
//  3. Замеры — единственный честный способ разрешить спор «Mutex vs RWMutex».
package main

import (
	"fmt"
	"strconv"
	"sync"
)

type CacheMutex struct {
	mu   sync.Mutex
	data map[string]string
}

func (c *CacheMutex) Get(k string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[k]
}

type CacheRWMutex struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *CacheRWMutex) Get(k string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[k]
}

func makeData() map[string]string {
	m := make(map[string]string, 100)
	for i := 0; i < 100; i++ {
		m[strconv.Itoa(i)] = "value"
	}
	return m
}

func main() {
	// Небольшая демонстрация: оба кэша возвращают одно и то же значение.
	cm := &CacheMutex{data: makeData()}
	cr := &CacheRWMutex{data: makeData()}
	fmt.Println("mutex get:", cm.Get("1"))
	fmt.Println("rwmutex get:", cr.Get("1"))
	fmt.Println()
	fmt.Println("Запустите бенчмарк:")
	fmt.Println("  go test -bench=. -benchmem ./02_mutex_vs_rwmutex")
}
