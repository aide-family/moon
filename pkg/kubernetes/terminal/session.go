package terminal

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"time"
)

var _ Handler = &Session{}

const (
	STDIN  = "stdin"
	STDOUT = "stdout"
	RESIZE = "resize"
	PING   = "ping"

	WebSocketReadBufSize  = 4096
	WebSocketWriteBufSize = 4096
)

type Session struct {
	wsConn   *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
}

type Message struct {
	Operation string `json:"operation"`
	Data      string `json:"data"`
	Rows      uint16 `json:"rows"`
	Cols      uint16 `json:"cols"`
}

func (s *Session) Read(p []byte) (n int, err error) {
	_, message, err := s.wsConn.ReadMessage()
	if err != nil {
		return copy(p, EndOfTransmission), err
	}

	var msg Message
	if err = json.Unmarshal(message, &msg); err != nil {
		return copy(p, "\u0004"), err
	}
	switch msg.Operation {
	case STDIN:
		return copy(p, msg.Data), nil
	case RESIZE:
		s.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	case PING:
		return 0, nil
	default:
		return copy(p, EndOfTransmission), fmt.Errorf("unknown message type")
	}
}

func (s *Session) Write(p []byte) (n int, err error) {
	msg, err := json.Marshal(Message{
		Operation: STDOUT,
		Data:      string(p),
	})
	if err != nil {
		return 0, err
	}
	if err = s.wsConn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *Session) Next() *remotecommand.TerminalSize {
	select {
	case size := <-s.sizeChan:
		return &size
	case <-s.doneChan:
		return nil
	}
}

func (s *Session) Done() {
	close(s.doneChan)
}

func (s *Session) Close() error {
	return s.wsConn.Close()
}

func NewTerminalSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	u := websocket.Upgrader{
		HandshakeTimeout: time.Second * 2,
		ReadBufferSize:   WebSocketReadBufSize,
		WriteBufferSize:  WebSocketWriteBufSize,
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			// don't return errors to maintain backwards compatibility
		},
		CheckOrigin: func(r *http.Request) bool {
			// allow all connections by default
			return true
		},
		Subprotocols: []string{r.Header.Get("Sec-WebSocket-Protocol")},
	}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	session := &Session{
		wsConn:   conn,
		sizeChan: make(chan remotecommand.TerminalSize),
		doneChan: make(chan struct{}),
	}
	return session, nil
}
