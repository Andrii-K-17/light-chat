package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrii-K-17/light-chat/internal/middleware"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/Andrii-K-17/light-chat/internal/response"
	"github.com/Andrii-K-17/light-chat/internal/ws"
	"github.com/go-chi/chi/v5"
)

// ChatHandler manages chat-related HTTP endpoints.
type ChatHandler struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	hub         *ws.Hub
}

// NewChatHandler initializes and returns a new ChatHandler.
func NewChatHandler(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	hub *ws.Hub,
) *ChatHandler {
	return &ChatHandler{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		hub:         hub,
	}
}

// GetChats returns all chats the authenticated user is a member of.
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chats, err := h.chatRepo.FindAllByUser(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, chats)
}

// createChatRequest represents the payload for creating a new chat.
type createChatRequest struct {
	Name      *string `json:"name"`
	IsGroup   bool    `json:"is_group"`
	MemberIDs []int   `json:"member_ids"`
}

// CreateChat creates a direct or group chat and adds all specified members.
func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req createChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !req.IsGroup && len(req.MemberIDs) == 1 {
		existing, err := h.chatRepo.FindDirectChat(userID, req.MemberIDs[0])
		if err == nil {
			members, _ := h.chatRepo.GetMembers(existing.ID)
			response.JSON(w, http.StatusOK, map[string]any{
				"id":       existing.ID,
				"name":     existing.Name,
				"is_group": existing.IsGroup,
				"members":  members,
			})
			return
		}
	}

	chat, err := h.chatRepo.Create(req.Name, req.IsGroup, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := h.chatRepo.AddMember(chat.ID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	for _, memberID := range req.MemberIDs {
		if memberID == userID {
			continue
		}
		_ = h.chatRepo.AddMember(chat.ID, memberID)
	}

	members, _ := h.chatRepo.GetMembers(chat.ID)

	response.JSON(w, http.StatusCreated, map[string]any{
		"id":       chat.ID,
		"name":     chat.Name,
		"is_group": chat.IsGroup,
		"members":  members,
	})
}

// GetMessages returns paginated message history for a chat.
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	isMember, err := h.chatRepo.IsMember(chatID, userID)
	if err != nil || !isMember {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	}

	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	messages, err := h.messageRepo.FindByChatID(chatID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	_ = h.messageRepo.MarkChatAsRead(chatID, userID)

	response.JSON(w, http.StatusOK, messages)
}

// SearchMessages performs a full-text search over messages in a chat.
func (h *ChatHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	isMember, err := h.chatRepo.IsMember(chatID, userID)
	if err != nil || !isMember {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		response.Error(w, http.StatusBadRequest, "q param required")
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	messages, err := h.messageRepo.Search(chatID, query, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, messages)
}

// DeleteChat removes a chat if the requesting user has permission.
func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	deleted, err := h.chatRepo.Delete(chatID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !deleted {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// updateMessageRequest represents the message edit payload.
type updateMessageRequest struct {
	Content string `json:"content"`
}

// UpdateMessage edits a message owned by the authenticated user.
func (h *ChatHandler) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	messageID, err := strconv.Atoi(chi.URLParam(r, "messageId"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid message id")
		return
	}

	var req updateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		response.Error(w, http.StatusUnprocessableEntity, "content is required")
		return
	}

	msg, err := h.messageRepo.Update(messageID, userID, content)
	if err != nil {
		response.Error(w, http.StatusForbidden, "forbidden or message not found")
		return
	}

	response.JSON(w, http.StatusOK, msg)

	event, _ := json.Marshal(ws.Event{
		Type:    "message_updated",
		Payload: mustMarshalChat(msg),
	})
	h.hub.BroadcastToChat(msg.ChatID, event)
}

// DeleteMessage removes a message owned by the authenticated user.
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	messageID, err := strconv.Atoi(chi.URLParam(r, "messageId"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid message id")
		return
	}

	deleted, err := h.messageRepo.Delete(messageID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !deleted {
		response.Error(w, http.StatusForbidden, "forbidden or message not found")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"deleted": true})

	payload, _ := json.Marshal(map[string]int{
		"message_id": messageID,
		"chat_id":    chatID,
	})

	event, _ := json.Marshal(ws.Event{
		Type:    "message_deleted",
		Payload: payload,
	})

	h.hub.BroadcastToChat(chatID, event)
}

// GetMembers returns all members of a group chat.
func (h *ChatHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	isMember, err := h.chatRepo.IsMember(chatID, userID)
	if err != nil || !isMember {
		response.Error(w, http.StatusForbidden, "forbidden")
		return
	}

	members, err := h.chatRepo.GetMembers(chatID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, members)
}

// addMemberRequest represents the payload for adding a member by username.
type addMemberRequest struct {
	Username string `json:"username"`
}

// AddMember adds a user to a group chat by username (creator only).
func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	chat, err := h.chatRepo.FindByID(chatID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "chat not found")
		return
	}
	if !chat.IsGroup || (chat.CreatedBy == nil || *chat.CreatedBy != userID) {
		response.Error(w, http.StatusForbidden, "only group creator can add members")
		return
	}

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Username) == "" {
		response.Error(w, http.StatusBadRequest, "username required")
		return
	}

	member, err := h.chatRepo.AddMemberByUsername(chatID, strings.TrimSpace(req.Username))
	if err != nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	response.JSON(w, http.StatusOK, member)
}

// RemoveMember removes a user from a group chat.
func (h *ChatHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	memberID, err := strconv.Atoi(chi.URLParam(r, "memberId"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid member id")
		return
	}

	chat, err := h.chatRepo.FindByID(chatID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "chat not found")
		return
	}
	if !chat.IsGroup || (chat.CreatedBy == nil || *chat.CreatedBy != userID) {
		response.Error(w, http.StatusForbidden, "only group creator can remove members")
		return
	}
	if memberID == userID {
		response.Error(w, http.StatusBadRequest, "cannot remove yourself")
		return
	}

	removed, err := h.chatRepo.RemoveMember(chatID, memberID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"removed": removed})
}

func mustMarshalChat(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
