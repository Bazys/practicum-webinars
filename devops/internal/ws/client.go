package ws

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn      *websocket.Conn
	writeChan chan message
	name      string

	clients *sync.Map
	counter *atomic.Int32
	once    sync.Once
}
type data struct {
	name   string
	value  string
	metric string
}

type message struct {
	code int
	data string
}

func NewClient(conn *websocket.Conn, counter *atomic.Int32, clients *sync.Map) *Client {
	client := &Client{
		conn:      conn,
		writeChan: make(chan message, 1),
		name:      uuid.New().String(),

		clients: clients,
		counter: counter,
		once:    sync.Once{},
	}

	conn.SetPingHandler(func(appData string) error {
		client.writeMessage("pong", websocket.PongMessage)
		return nil
	})

	counter.Add(1)
	log.Println("create new client:", client.name, counter.Load())

	return client
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) close() error {
	c.counter.Add(-1)
	log.Println("close client:", c.name, c.counter.Load())
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.clients.Delete(c.name)
	close(c.writeChan)
	return c.conn.Close()
}

func (c *Client) Close() (err error) {
	c.once.Do(func() {
		err = c.close()
	})

	return
}

func (c *Client) writeMessage(data string, messageType int) {
	c.writeChan <- message{messageType, data}
}

func (c *Client) Write() {
	for msg := range c.writeChan {
		err := c.conn.WriteMessage(msg.code, []byte(msg.data))
		if err != nil {
			log.Println("write:", err)
			c.Close()
			return
		}
	}
}

func (c *Client) Read() {
	defer c.Close()
	for {
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err, c.name)
			return
		}

		log.Printf("recv: %d %s %s", mt, message, c.name)
	}
}
