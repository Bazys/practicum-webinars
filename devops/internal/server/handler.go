package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"practicum-webinars/devops/internal/storage"
	"practicum-webinars/devops/internal/ws"
)

var (
	upgrader = websocket.Upgrader{}
	clients  = sync.Map{}
	counter  atomic.Int32
)

type handler struct {
	sync.RWMutex
	db *storage.DB
}

func NewRouter() http.Handler {
	r := chi.NewRouter()

	h := &handler{
		db: storage.NewDB(),
	}

	r.Post("/update/{type}/{id}/{value}", h.update)
	r.Get("/value/{type}/{id}", h.get)
	r.Get("/", h.info)
	r.Get("/ws", h.ws)

	return r
}

func StopClients() {
	clients.Range(func(key, value interface{}) bool {
		log.Println("stop client:", key)
		value.(*ws.Client).Close()
		return true
	})
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "undefined field 'id'", http.StatusBadRequest)
		return
	}

	rawValue := chi.URLParam(r, "value")
	if rawValue == "" {
		http.Error(w, "undefined field 'value'", http.StatusBadRequest)
		return
	}

	var count int
	reqType := chi.URLParam(r, "type")

	h.Lock()
	defer h.Unlock()
	switch reqType {
	case "counter":
		delta, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			http.Error(w, "wrong type of counter value", http.StatusBadRequest)
			return
		}
		count = h.db.UpdateCounter(id, delta)
	case "gauge":
		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			http.Error(w, "wrong type of gauge value", http.StatusBadRequest)
			return
		}
		count = h.db.UpdateGauge(id, value)
	default:
		http.Error(w, "unknown type of metrics", http.StatusNotImplemented)
		return
	}

	fmt.Printf("update %s: %s=%s, %d\n", reqType, id, rawValue, count)
	_, _ = w.Write([]byte("Updated: " + fmt.Sprintf("%d\n", count)))
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "undefined field 'id'", http.StatusBadRequest)
		return
	}

	reqType := chi.URLParam(r, "type")

	h.Lock()
	defer h.Unlock()
	switch reqType {
	case "counter":
		if v, ok := h.db.Counter(id); ok {
			_, _ = w.Write([]byte(fmt.Sprintf("%d", v)))
			return
		}
	case "gauge":
		if v, ok := h.db.Gauge(id); ok {
			_, _ = w.Write([]byte(fmt.Sprintf("%.3f", v)))
			return
		}
	default:
		http.Error(w, "unknown type of metrics", http.StatusNotImplemented)
		return
	}
	http.NotFound(w, r)
}

func (h *handler) info(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	h.RLock()
	defer h.RUnlock()
	_, _ = io.WriteString(w, `<html>
<head>
<title>Metrics, MustHave.DevOps by Yandex-Practicum</title>
<meta http-equiv="refresh" content="5" />
</head>
<body><h1>Metrics values</h1><h3>Main</h3>`)
	_, _ = io.WriteString(w, `Gen: `+fmt.Sprintf("%d", h.db.UpdateCount())+"<br>\n")
	_, _ = io.WriteString(w, `Timestamp: `+h.db.Timestamp(time.StampMilli)+"<br>\n")
	_, _ = io.WriteString(w, `<h3>Counters</h3>`)
	h.db.MapOrderedCounter(func(k string, v int64) {
		_, _ = io.WriteString(w, k+": "+fmt.Sprintf("%d", v)+"<br>\n")
	})
	_, _ = io.WriteString(w, `<h3>Gauges</h3>`)
	h.db.MapOrderedGauge(func(k string, v float64) {
		_, _ = io.WriteString(w, k+": "+fmt.Sprintf("%.3f", v)+"<br>\n")
	})
	_, _ = io.WriteString(w, `<html></body></html>`)
}

func (h *handler) ws(w http.ResponseWriter, r *http.Request) {
	log.Print("Header", r.Header)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	client := ws.NewClient(c, &counter, &clients)
	clients.Store(client.Name(), client)
	go client.Read()
	go client.Write()
}
