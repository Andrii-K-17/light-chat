package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// Event represents a WebSocket message envelope.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Client represents a single active WebSocket connection.
type Client struct {
	UserID int
	ChatID int
	Send   chan []byte
	Hub    *Hub
}

// Hub manages all active WebSocket clients and routes messages between them.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// NewHub initializes and returns a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[c] = struct{}{}
	slog.Info("ws client registered", "user_id", c.UserID, "chat_id", c.ChatID)
}

// Unregister removes a client from the hub and closes its send channel.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.Send)
		slog.Info("ws client unregistered", "user_id", c.UserID)
	}
}

// BroadcastToChat sends a raw message to all clients subscribed to the given chat.
func (h *Hub) BroadcastToChat(chatID int, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		if c.ChatID == chatID {
			select {
			case c.Send <- message:
			default:
				slog.Warn("ws send buffer full, dropping message", "user_id", c.UserID)
			}
		}
	}
}

// BroadcastReadReceipt notifies all chat members that messages were read by a user.
func (h *Hub) BroadcastReadReceipt(chatID, readerID int) {
	payload, _ := json.Marshal(map[string]any{
		"chat_id":   chatID,
		"reader_id": readerID,
	})

	event, _ := json.Marshal(Event{Type: "read_receipt", Payload: payload})
	h.BroadcastToChat(chatID, event)
}
