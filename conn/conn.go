package conn

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/snburman/game-server/config"
)

const PING_INTERVAL = 10 * time.Second

var connPool = conns{
	pool: make(map[string]*Conn),
}

type (
	conns struct {
		mu   sync.Mutex
		pool map[string]*Conn
	}
	Conn struct {
		mu            sync.Mutex
		UserID        string
		websocket     *websocket.Conn
		authenticated bool
		LastPing      time.Time
		MapID         string
		Messages      chan []byte
		pingDone      chan bool
		listenDone    chan bool
	}
)

func (c *conns) GetAll() map[string]*Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pool
}

func (c *conns) Get(userID string) (*Conn, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, ok := c.pool[userID]
	return conn, ok
}

func (c *conns) Set(userID string, conn *Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pool[userID] = conn
}

func (c *conns) Delete(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pool, userID)
}

func NewConn(w http.ResponseWriter, r *http.Request, UserID string) (*Conn, error) {
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
		UserID:     UserID,
		Messages:   make(chan []byte, 256),
		LastPing:   time.Now(),
		pingDone:   make(chan bool),
		listenDone: make(chan bool),
	}
	connPool.Set(c.UserID, c)
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
				c.mu.Lock()
				if err := c.websocket.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Println("error writing ping", "error", err)
					c.mu.Unlock()
					c.Close()
					break Ping
				}
				c.LastPing = time.Now()
				c.mu.Unlock()
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

			// Authenticate connection
			if dispatch.Function == Authenicate {
				dispatch := ParseDispatch[map[string][]string](dispatch)
				headers := dispatch.Data
				if len(headers["CLIENT_SECRET"]) == 0 || len(headers["CLIENT_ID"]) == 0 {
					log.Println("missing headers")
					c.Close()
				}
				if headers["CLIENT_SECRET"][0] != config.Env().CLIENT_SECRET ||
					headers["CLIENT_ID"][0] != config.Env().CLIENT_ID {
					log.Println("invalid headers")
					c.Close()
				} else {
					c.authenticated = true
				}
			}
			if !c.authenticated {
				log.Println("unauthenticated connection")
				c.Close()
				break
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
	c.mu.Lock()
	defer c.mu.Unlock()
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
	conn, _ := connPool.Get(c.UserID)
	if conn != c {
		return
	}
	c.Messages <- msg
}

func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.websocket == nil {
		msg := "cannot close nil connection"
		log.Println(msg)
		return errors.New(msg)
	}
	// remove connection from pool
	connPool.Delete(c.UserID)
	// remove player from player pool
	playerPool.Delete(c.UserID)
	// notify online players of player removal
	dispatch := NewDispatch(uuid.NewString(), c, RemoveOnlinePlayer, c.UserID)
	RouteDispatch(dispatch.Marshal())

	c.websocket.Close()
	c.pingDone <- true
	c.listenDone <- true
	log.Println("connection closed: ", c.UserID)
	return nil
}
