package conn

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/snburman/game_server/db"
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
func NewConnection(user db.User) *Connection {
	c := &Connection{
		User: user,
	}
	ConnPool.Set(user.ID.Hex(), c)
	return c
}

func SetClient(id string, conn *websocket.Conn) {
	_conn := ConnPool.Get(id)
	_conn.Client = conn
	ConnPool.Set(id, &_conn)
}

func SetWasm(id string, conn *websocket.Conn) {
	_conn := ConnPool.Get(id)
	_conn.Wasm = conn
	ConnPool.Set(id, &_conn)
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

func (cp *ConnectionPool) Set(key string, conn *Connection) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.connections[key] = conn
}

func (cp *ConnectionPool) Get(key string) Connection {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	return *cp.connections[key]
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
