package agent

import (
	"bytes"
	"fmt"
	"net/http"
)

type reporter struct {
	client       *http.Client
	wsClient     *Client
	counterFlush int
	address      string
}

// Reporter интерфейс для репортера.
type Reporter interface {
	Flush()

	// ReportCounter отправляет значения счетчиков.
	ReportCounter(
		name string,
		value int64,
	)

	// ReportGauge отправляет значения датчиков.
	ReportGauge(
		name string,
		value float64,
	)
}

func NewReporter(address string) (Reporter, error) {
	wsClient, err := NewClient(address)
	if err != nil {
		return nil, err
	}
	return &reporter{
		client:   &http.Client{},
		wsClient: wsClient,
		address:  address,
	}, nil
}

// ReportCounter реализация тривиального варианта репортера в текущую консоль.
func (r *reporter) ReportCounter(name string, value int64) {
	fmt.Printf("report counter: '%s', value: %d\n", name, value)
	r.wsClient.Do(name, fmt.Sprintf("%d", value), "counter")
	// r.post(name, fmt.Sprintf("%d", value), "counter", "update counter")
}

func (r *reporter) ReportGauge(name string, value float64) {
	fmt.Printf("report gauge: '%s', value: %.3f\n", name, value)
	r.wsClient.Do(name, fmt.Sprintf("%.3f", value), "gauge")
	// r.post(name, fmt.Sprintf("%.3f", value), "gauge", "update gauge")
}

func (r *reporter) post(name, value, metric, body string) {
	resp, err := r.client.Post("http://"+r.address+"/update/"+metric+"/"+name+"/"+value, "text/plain", bytes.NewReader([]byte(body)))
	if err != nil {
		fmt.Printf("failed get: %s\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("got response, status: %d, proto: %s, value: %s\n", resp.StatusCode, resp.Proto, value)
}

func (r *reporter) Flush() {
	r.counterFlush++
	fmt.Printf("flush, count %d\n", r.counterFlush)
}
