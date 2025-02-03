package conn

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Socket struct {
	ws           *websocket.Conn
	connectionID string
	msg          chan []byte
	done         chan bool
}

type Message[T any] struct {
	Data T `json:"data"`
}

func NewSocket(connectionID string, conn *websocket.Conn) *Socket {
	return &Socket{
		ws:           conn,
		connectionID: connectionID,
		msg:          make(chan []byte),
		done:         make(chan bool),
	}
}

func (s Socket) Read() ([]byte, error) {
	_, msg, err := s.ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (s Socket) Write(msg any) error {
	// convert to json
	m := Message[any]{Data: msg}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = s.ws.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		return err
	}
	return nil
}

func (s Socket) Ping() error {
	err := s.ws.WriteMessage(websocket.PingMessage, nil)
	if err != nil {
		log.Println("error pinging: ", err)
		return err
	}
	return nil
}

func (s Socket) Close() {
	s.done <- true
}

func (s Socket) Listen() {
	go func() {
		for {
			select {
			case <-s.done:
				s.ws.Close()
				ConnPool.Remove(s.connectionID)
				return
			case msg := <-s.msg:
				fmt.Println("msg", msg)
			}
		}
	}()
}
