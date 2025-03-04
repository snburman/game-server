package conn

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
}

func newMockWebsocket() *websocket.Conn {
	s := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		panic(err)
	}
	return c
}

func NewMockConn() *Conn {
	ws := newMockWebsocket()
	return &Conn{
		websocket:     ws,
		UserID:        "67bfa82f165e6e4169699147",
		Messages:      make(chan []byte, 256),
		pingDone:      make(chan bool),
		listenDone:    make(chan bool),
		connType:      WasmConn,
		authenticated: true,
	}
}

func TestPublishAndClose(t *testing.T) {
	conn := NewMockConn()
	conn.Publish([]byte("test"))
	msg := <-conn.Messages
	assert.Equal(t, "test", string(msg))
	go conn.Close()
	pingDone := <-conn.pingDone
	assert.True(t, pingDone)
	listenDone := <-conn.listenDone
	assert.True(t, listenDone)
}
