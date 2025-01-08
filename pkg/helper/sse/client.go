package sse

import (
	"fmt"
	"net/http"
	"sync"

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

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[uint32]*Client),
	}
}

type ClientManager struct {
	clients map[uint32]*Client
	mu      sync.RWMutex
}

// AddClient adds a client to the manager
func (cm *ClientManager) AddClient(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.clients[client.ID]; ok {
		return
	}
	cm.clients[client.ID] = client
}

// RemoveClient removes a client from the manager
func (cm *ClientManager) RemoveClient(id uint32) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, id)
}

// GetClient returns a client by id
func (cm *ClientManager) GetClient(id uint32) (*Client, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, ok := cm.clients[id]
	return client, ok
}

// WriteSSE writes data to the client
func (c *Client) WriteSSE(w http.ResponseWriter) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Errorw("err", "Streaming unsupported!")
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	log.Debugw("msg", "listen sse client")
	for data := range c.Send {
		log.Debugw("WriteSSE", string(data))
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}
	fmt.Fprintf(w, "data: [DONE]\n\n")
	log.Debugw("WriteSSE", "data: [DONE]")
}
