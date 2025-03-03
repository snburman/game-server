package conn

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()
}

func NewMockWebsocket() *websocket.Conn {
	s := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	return c
}

func TestNewConn(t *testing.T) {
	ws := NewMockWebsocket()
	fmt.Println(ws)
}
