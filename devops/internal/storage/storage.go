package storage

import (
	"sort"
	"sync"
	"time"
)

type DB struct {
	sync.RWMutex
	counters    map[string]int64
	gauges      map[string]float64
	updateCount int
	timestamp   time.Time
}

func NewDB() *DB {
	return &DB{
		counters:  make(map[string]int64),
		gauges:    make(map[string]float64),
		timestamp: time.Now(),
	}
}

func (db *DB) UpdateCounter(id string, delta int64) int {
	db.Lock()
	defer db.Unlock()
	db.timestamp = time.Now()
	db.counters[id] += delta
	db.updateCount++
	return db.updateCount
}

func (db *DB) UpdateGauge(id string, value float64) int {
	db.Lock()
	defer db.Unlock()
	db.timestamp = time.Now()
	db.gauges[id] = value
	db.updateCount++
	return db.updateCount
}

func (db *DB) Counter(id string) (int64, bool) {
	db.Lock()
	v, ok := db.counters[id]
	db.Unlock()
	return v, ok
}

func (db *DB) Gauge(id string) (float64, bool) {
	db.Lock()
	v, ok := db.gauges[id]
	db.Unlock()
	return v, ok
}

func (db *DB) Timestamp(layout string) string {
	db.RLock()
	defer db.RUnlock()
	return db.timestamp.Format(layout)
}

func (db *DB) UpdateCount() int {
	db.RLock()
	defer db.RUnlock()
	return db.updateCount
}

func (db *DB) MapOrderedCounter(f func(k string, v int64)) {
	db.RLock()
	defer db.RUnlock()
	keys := make([]string, 0, len(db.counters))
	for k := range db.counters {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		f(k, db.counters[k])
	}
}

func (db *DB) MapOrderedGauge(f func(k string, v float64)) {
	db.RLock()
	defer db.RUnlock()
	keys := make([]string, 0, len(db.gauges))
	for k := range db.gauges {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		f(k, db.gauges[k])
	}
}
