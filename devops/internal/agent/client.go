package agent

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

func NewClient(addr string) (*Client, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}

	header := http.Header{}
	header["jwt"] = []string{"my_best_jwt"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn: c,
	}

	c.SetPongHandler(func(appData string) error {
		log.Println("pong message:", appData)
		return nil
	})

	return client, nil
}

func (c *Client) Do(name, value, mType string) error {
	err := c.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
	if err != nil {
		log.Println("write:", err)
		return err
	}

	err = c.conn.WriteMessage(websocket.TextMessage, []byte(name+"|"+value+"|"+mType))
	if err != nil {
		log.Println("write:", err)
		return err
	}
	return nil
}
