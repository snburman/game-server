package conn

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/snburman/game_server/db"
	"github.com/snburman/game_server/errors"
)

const PING_FREQUENCY = 30 * time.Second
const PING_TIMEOUT = 3 * PING_FREQUENCY

var ConnPool = NewConnectionPool()

type Connection struct {
	Client   *websocket.Conn
	Wasm     *websocket.Conn
	User     db.User
	lastPing time.Duration
}

// Connection is initialized with a user and added to the connection pool by key user.ID.Hex()
//
// Client and Wasm are added subequently during handshake
func NewConnection(user db.User) error {
	c := &Connection{
		User: user,
	}
	_, err := ConnPool.Get(user.ID.Hex())
	if err == nil {
		err = errors.ErrConnectionExists
		log.Println(err.Error() + ": " + user.ID.Hex())
		return err
	}
	ConnPool.set(user.ID.Hex(), c)
	return nil
}

func SetClient(id string, conn *websocket.Conn) {
	_conn, err := ConnPool.Get(id)
	if err != nil {
		log.Println(err.Error() + ": " + id)
		return
	}
	_conn.Client = conn
	ConnPool.set(id, &_conn)
}

func SetWasm(id string, conn *websocket.Conn) {
	_conn, err := ConnPool.Get(id)
	if err != nil {
		log.Println(err.Error() + ": " + id)
		return
	}
	_conn.Wasm = conn
	ConnPool.set(id, &_conn)
}

func (c *Connection) Close() {
	ConnPool.Remove(c.User.ID.Hex())
}

func (c Connection) MessageClient(m []byte) {
	if c.Client != nil {
		err := c.Client.WriteMessage(websocket.TextMessage, m)
		if err != nil {
			log.Println(err)
			if err == websocket.ErrCloseSent {
				c.Client.Close()
			}
		}
	}
}

func (c Connection) MessageWasm(m []byte) {
	if c.Wasm != nil {
		c.Wasm.WriteMessage(websocket.TextMessage, m)
	}
}

type ConnectionPool struct {
	mu          sync.Mutex
	connections map[string]*Connection
}

func NewConnectionPool() *ConnectionPool {
	c := &ConnectionPool{
		connections: make(map[string]*Connection),
	}
	go c.ping()
	return c
}

func (cp *ConnectionPool) set(key string, conn *Connection) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.connections[key] = conn
}

func (cp *ConnectionPool) Get(key string) (Connection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	c := cp.connections[key]
	if c == nil {
		return Connection{}, errors.ErrConnectionNotFound
	}
	return *cp.connections[key], nil
}

func (cp *ConnectionPool) Remove(key string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.connections, key)
}

func (cp *ConnectionPool) ping() {
	for {
		cp.mu.Lock()
		for key, conn := range cp.connections {
			conn.lastPing += PING_FREQUENCY
			if conn.lastPing > PING_TIMEOUT {
				cp.Remove(key)
				continue
			}
			if cp.connections[key].Client == nil || cp.connections[key].Wasm == nil {
				return
			}
			cp.connections[key].Client.WriteMessage(websocket.PingMessage, []byte{})
			cp.connections[key].Wasm.WriteMessage(websocket.PingMessage, []byte{})
			conn.lastPing = 0
		}
		cp.mu.Unlock()
		time.Sleep(PING_FREQUENCY)
	}
}
