package sse

import (
	"fmt"
	"net/http"

	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/go-kratos/kratos/v2/log"
)

// NewClient creates a new client
func NewClient(id uint32) *Client {
	return &Client{
		ID:   id,
		Send: make(chan []byte, 10),
	}
}

// Client is a client of the SSE server
type Client struct {
	ID   uint32
	Send chan []byte
}

// Close closes the client
func (c *Client) Close() {
	close(c.Send)
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: safety.NewMap[uint32, *Client](),
	}
}

// ClientManager is a manager of the S
type ClientManager struct {
	clients *safety.Map[uint32, *Client]
}

// AddClient adds a client to the manager
func (cm *ClientManager) AddClient(client *Client) {
	if _, ok := cm.clients.Get(client.ID); ok {
		return
	}
	cm.clients.Set(client.ID, client)
}

// RemoveClient removes a client from the manager
func (cm *ClientManager) RemoveClient(id uint32) {
	if client, ok := cm.clients.Get(id); ok {
		client.Close()
		cm.clients.Delete(id)
	}
}

// GetClient returns a client by id
func (cm *ClientManager) GetClient(id uint32) (*Client, bool) {
	return cm.clients.Get(id)
}

// Close closes the client
func (cm *ClientManager) Close() {
	list := cm.clients.List()
	for _, client := range list {
		close(client.Send)
	}
}

// WriteSSE writes data to the client
func (c *Client) WriteSSE(w http.ResponseWriter) {
	defer after.RecoverX()
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Errorw("err", "Streaming unsupported!")
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	log.Debugw("msg", "listen sse client")
	for data := range c.Send {
		log.Debugw("WriteSSE", string(data))
		_, _ = fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}
	log.Debugw("WriteSSE", "data: [DONE]")
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("send message error: %v", r)
		}
	}()
	c.Send <- []byte(message)
	return
}
