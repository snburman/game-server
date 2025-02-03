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
	User     db.User
	lastPing time.Duration
	Client   *Socket
	Wasm     *Socket
}

// Connection is initialized with a user and added to the connection pool by key user.ID.Hex()
//
// Client and Wasm are added subequently during handshake
func NewConnection(user db.User) error {
	c := &Connection{
		User: user,
	}
	ConnPool.set(user.ID.Hex(), c)
	return nil
}

func SetClient(id string, conn *websocket.Conn) error {
	_conn, err := ConnPool.Get(id)
	if err != nil {
		log.Println(err.Error() + ": " + id)
		return err
	}
	_conn.Client = NewSocket(id, conn)
	ConnPool.set(id, &_conn)
	return nil
}

func SetWasm(id string, conn *websocket.Conn) error {
	_conn, err := ConnPool.Get(id)
	if err != nil {
		log.Println(err.Error() + ": " + id)
		return err
	}
	_conn.Wasm = NewSocket(id, conn)
	ConnPool.set(id, &_conn)
	return nil
}

func (c *Connection) Close() {
	ConnPool.Remove(c.User.ID.Hex())
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
	// for {
	// 	cp.mu.Lock()
	// 	for key, conn := range cp.connections {
	// 		conn.lastPing += PING_FREQUENCY
	// 		if conn.lastPing > PING_TIMEOUT {
	// 			cp.Remove(key)
	// 			continue
	// 		}
	// 		if cp.connections[key].Client == nil || cp.connections[key].Wasm == nil {
	// 			return
	// 		}
	// 		cp.connections[key].Client.Ping()
	// 		cp.connections[key].Wasm.Ping()
	// 		conn.lastPing = 0
	// 	}
	// 	cp.mu.Unlock()
	// 	time.Sleep(PING_FREQUENCY)
	// }
}
