package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Andrii-K-17/light-chat/internal/middleware"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/Andrii-K-17/light-chat/internal/response"
	"github.com/go-chi/chi/v5"
)

// ChatHandler manages chat-related HTTP endpoints.
type ChatHandler struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
}

// NewChatHandler initializes and returns a new ChatHandler.
func NewChatHandler(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
) *ChatHandler {
	return &ChatHandler{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
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
