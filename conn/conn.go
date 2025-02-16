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

const PING_INTERVAL = 5 * time.Second

var connPool = conns{
	pool: make(map[string]*Conn),
}

type (
	conns struct {
		mu   sync.Mutex
		pool map[string]*Conn
	}
	Conn struct {
		ID         string
		websocket  *websocket.Conn
		LastPing   time.Time
		Messages   chan []byte
		pingDone   chan bool
		listenDone chan bool
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

func NewConn(w http.ResponseWriter, r *http.Request, ID string) (*Conn, error) {
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
		websocket:  websocket,
		ID:         ID,
		Messages:   make(chan []byte, 256),
		LastPing:   time.Now(),
		pingDone:   make(chan bool),
		listenDone: make(chan bool),
	}
	connPool.Set(c.ID, c)
	return c, nil
}

func (c *Conn) Listen() {
	// ping
	go func(c *Conn) {
		ticker := time.NewTicker(PING_INTERVAL)
		defer ticker.Stop()
	Ping:
		for {
			select {
			case <-ticker.C:
				if time.Since(c.LastPing) > PING_INTERVAL*3 {
					if c.websocket != nil {
						log.Println("ping timeout", "time since last ping", time.Since(c.LastPing))
						log.Println("wait time for ping", PING_INTERVAL*3)
						c.Close()
						break Ping
					}
				}
				if err := c.websocket.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("error writing ping", "error", err)
					c.Close()
					break Ping
				}
				c.LastPing = time.Now()
			case <-c.pingDone:
				log.Println("ping done")
				break Ping
			}
		}
	}(c)

	// incoming messages
	go func(c *Conn) {
		defer c.Close()
		var dispatch Dispatch[[]byte]
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
			// Route dispatch to appropriate function
			RouteDispatch(dispatch)
		}
	}(c)

	// outgoing messages
	for {
		select {
		case msg, ok := <-c.Messages:
			if !ok {
				log.Println("channel closed")
				c.Close()
				break
			}
			if c.websocket == nil {
				break
			}

			if err := c.websocket.WriteMessage(1, msg); err != nil {
				log.Println("error writing message", "error", err)
				c.Close()
			}
		case <-c.listenDone:
			log.Println("listen done")
			break
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

func (c *Conn) Close() error {
	if c.websocket == nil {
		msg := "cannot close nil connection"
		log.Println(msg)
		return errors.New(msg)
	}
	connPool.Delete(c.ID)
	c.websocket.Close()
	c.pingDone <- true
	c.listenDone <- true
	log.Println("connection closed: ", c.ID)
	return nil
}
