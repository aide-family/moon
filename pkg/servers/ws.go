package servers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/util/hash"
)

var _ transport.Server = (*WebsocketServer)(nil)

type WebsocketServer struct {
	addr      string
	server    *http.Server
	wsMap     map[string]*websocket.Conn
	msgHandle func(*Message)
	StopCh    chan struct{}
}

func NewWebsocketServer(addr string) *WebsocketServer {
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	return &WebsocketServer{
		addr:   addr,
		server: server,
		wsMap:  make(map[string]*websocket.Conn),
		StopCh: make(chan struct{}),
	}
}

func (l *WebsocketServer) pumpStdin(source string, ws *websocket.Conn) {
	log.Info("new websocket, ", "source: ", source)
	defer func() {
		log.Info("close websocket, ", "source: ", source)
		ws.Close()
		delete(l.wsMap, source)
	}()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		log.Infow("source", source, "message", string(message))
	}
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (l *WebsocketServer) serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade:", err)
		return
	}
	origin := hash.MD5(strconv.FormatInt(time.Now().UnixMicro(), 10) + uuid.NewString())
	l.wsMap[origin] = ws

	go l.pumpStdin(origin, ws)
}

func (l *WebsocketServer) Start(_ context.Context) error {
	http.HandleFunc("/ws", l.serveWs)
	log.Info("[websocket] server started: ", l.addr)
	go func() {
		defer after.RecoverX()
		log.Error(l.server.ListenAndServe())
	}()

	return nil
}

func (l *WebsocketServer) Stop(_ context.Context) error {
	defer log.Info("[websocket] server stopped")
	close(l.StopCh)
	return l.server.Close()
}

// RegisterMessageHandler 注册消息处理器
func (l *WebsocketServer) RegisterMessageHandler(handler func(msg *Message)) {
	l.msgHandle = handler
}

type MessageType int32

type Message struct {
	// 消息表现形式， 弹窗、提醒、通知
	MsgType MessageType `json:"msgType"`
	// 消息内容json
	Content any `json:"content"`
	// 消息标题
	Title string `json:"title"`
	// 消息业务类型
	Biz string `json:"biz"`
}

// Bytes 返回消息的字节
func (m *Message) Bytes() []byte {
	bs, _ := json.Marshal(m)
	return bs
}

// SendMessage 发送消息
func (l *WebsocketServer) SendMessage(message *Message) {
	for _, ws := range l.wsMap {
		wsTmp := ws
		go func() {
			defer after.RecoverX()
			if err := wsTmp.WriteMessage(websocket.TextMessage, message.Bytes()); err != nil {
				log.Errorw("write:", err)
			}
		}()
	}
}
