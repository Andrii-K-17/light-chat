package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// sendMessagePayload is the structure expected when a client sends a message.
type sendMessagePayload struct {
	ChatID  int    `json:"chat_id"`
	Content string `json:"content"`
}

// MessageSender is the subset of MessageService behavior the WS handler needs.
type MessageSender interface {
	IsMember(chatID, userID int) (bool, error)
	Create(chatID, userID int, content string) (*models.MessageResponse, error)
}

// Handler manages the WebSocket upgrade and per-connection pump loops.
type Handler struct {
	hub       *Hub
	msgSvc    MessageSender
	jwtSecret string
}

// NewHandler initializes and returns a new WebSocket Handler.
func NewHandler(
	hub *Hub,
	msgSvc MessageSender,
	jwtSecret string,
) *Handler {
	return &Handler{
		hub:       hub,
		msgSvc:    msgSvc,
		jwtSecret: jwtSecret,
	}
}

// ServeWS upgrades the HTTP connection to WebSocket and starts read/write pumps.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID, err := h.extractUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	chatIDStr := r.URL.Query().Get("chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil || chatID == 0 {
		http.Error(w, "chat_id required", http.StatusBadRequest)
		return
	}

	isMember, err := h.msgSvc.IsMember(chatID, userID)
	if err != nil || !isMember {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws upgrade failed", "error", err)
		return
	}

	client := &Client{
		UserID: userID,
		ChatID: chatID,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}
	h.hub.Register(client)

	go h.writePump(client, conn)
	h.readPump(client, conn)
}

// readPump listens for incoming messages from the WebSocket connection.
func (h *Handler) readPump(c *Client, conn *websocket.Conn) {
	defer func() {
		h.hub.Unregister(c)
		_ = conn.Close()
	}()

	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("ws read error", "error", err)
			}
			return
		}

		var event Event
		if err := json.Unmarshal(raw, &event); err != nil {
			continue
		}

		switch event.Type {
		case "send_message":
			h.handleSendMessage(c, event.Payload)
		case "read_receipt":
			h.hub.BroadcastReadReceipt(c.ChatID, c.UserID)
		}
	}
}

// writePump forwards outgoing messages from the send channel to the WebSocket connection.
func (h *Handler) writePump(c *Client, conn *websocket.Conn) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSendMessage delegates persistence to MessageService and broadcasts the result.
func (h *Handler) handleSendMessage(c *Client, payload json.RawMessage) {
	var p sendMessagePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	if p.ChatID != c.ChatID {
		return
	}

	msg, err := h.msgSvc.Create(c.ChatID, c.UserID, p.Content)
	if err != nil {
		slog.Error("ws failed to save message", "error", err)
		return
	}

	event, _ := json.Marshal(Event{
		Type:    "new_message",
		Payload: mustMarshal(msg),
	})
	h.hub.BroadcastToChat(c.ChatID, event)
}

// extractUserID parses the JWT from the token query param or cookie and returns the user ID.
func (h *Handler) extractUserID(r *http.Request) (int, error) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		cookie, err := r.Cookie("token")
		if err != nil {
			return 0, err
		}
		tokenStr = cookie.Value
	}

	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !parsed.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrSignatureInvalid
	}

	return int(claims["user_id"].(float64)), nil
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
