package conn

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var connPool = conns{
	pool: make(map[string]*Conn),
}

type (
	conns struct {
		mu   sync.Mutex
		pool map[string]*Conn
	}
	Conn struct {
		websocket *websocket.Conn
		ID        string
		LastPing  time.Time
		Messages  chan []byte
	}
)

func (c *conns) Get(id string) (*Conn, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, ok := c.pool[id]
	return conn, ok
}

func (c *conns) Set(id string, conn *Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pool[id] = conn
}

func (c *conns) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pool, id)
}

func newConn(w http.ResponseWriter, r *http.Request, ID string) (*Conn, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	websocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, errors.New("failed to upgrade connection")
	}

	c := &Conn{
		websocket: websocket,
		ID:        ID,
		Messages:  make(chan []byte, 16),
	}
	connPool.Set(c.ID, c)
	return c, nil
}

func (c *Conn) close() error {
	if c == nil {
		return errors.New("cannot close nil connection")
	}
	connPool.Delete(c.ID)
	c.websocket.Close()
	return nil
}

func (c *Conn) listen() {
	go func(c *Conn) {
		defer c.close()
		var dispatch Dispatch[any]
		for {
			_, message, err := c.websocket.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure,
				) {
					log.Printf("error: %v", err)
				}
				close(c.Messages)
				break
			}
			// Parse dispatch from websocket message
			err = json.Unmarshal(message, &dispatch)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			// Set conn on dispatch
			dispatch.conn = c
			// Dispatch to message handler
			// handler.in <- dispatch
		}
	}(c)

	// outgoing messages
	for {
		msg, ok := <-c.Messages
		if !ok {
			c.close()
			break
		}
		if c.websocket == nil {
			break
		}

		if err := c.websocket.WriteMessage(1, msg); err != nil {
			log.Println("error writing message", "error", err)
			c.close()
		}
	}
}

func (c *Conn) Publish(msg []byte) {
	// if msg is not json encodable, return
	_, err := json.Marshal(msg)
	if err != nil {
		log.Println("message not json encodable", "error", err)
		return
	}
	if c == nil {
		log.Println("connection severed, message not sent")
		return
	}
	conn, _ := connPool.Get(c.ID)
	if conn != c {
		return
	}
	c.Messages <- msg
}

func (c *Conn) Write(p []byte) (n int, err error) {
	c.Messages <- p
	return len(p), nil
}
