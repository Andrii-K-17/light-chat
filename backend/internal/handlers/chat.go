package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrii-K-17/light-chat/internal/middleware"
	"github.com/Andrii-K-17/light-chat/internal/response"
	"github.com/Andrii-K-17/light-chat/internal/services"
	"github.com/go-chi/chi/v5"
)

// ChatHandler manages chat-related HTTP endpoints.
type ChatHandler struct {
	chatSvc *services.ChatService
	msgSvc  *services.MessageService
}

// NewChatHandler initializes and returns a new ChatHandler.
func NewChatHandler(
	chatSvc *services.ChatService,
	msgSvc *services.MessageService,
) *ChatHandler {
	return &ChatHandler{
		chatSvc: chatSvc,
		msgSvc:  msgSvc,
	}
}

// GetChats returns all chats the authenticated user is a member of.
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chats, err := h.chatSvc.GetAll(userID)
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

	chat, err := h.chatSvc.Create(userID, services.CreateChatPayload{
		Name:      req.Name,
		IsGroup:   req.IsGroup,
		MemberIDs: req.MemberIDs,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusCreated, chat)
}

// DeleteChat removes a chat if the requesting user has permission.
func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	if err := h.chatSvc.Delete(chatID, userID); err != nil {
		if errors.Is(err, services.ErrNotChatMember) {
			response.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// GetMessages returns paginated message history for a chat.
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
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

	messages, err := h.msgSvc.GetHistory(chatID, userID, limit, offset)
	if err != nil {
		if errors.Is(err, services.ErrNotChatMember) {
			response.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

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

	messages, err := h.msgSvc.Search(chatID, userID, query, limit)
	if err != nil {
		if errors.Is(err, services.ErrNotChatMember) {
			response.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, messages)
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

	msg, err := h.msgSvc.Update(messageID, userID, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmptyContent):
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, services.ErrMessageNotFound):
			response.Error(w, http.StatusForbidden, "forbidden or message not found")
		default:
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	response.JSON(w, http.StatusOK, msg)
}

// DeleteMessage removes a message owned by the authenticated user.
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	messageID, err := strconv.Atoi(chi.URLParam(r, "messageId"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid message id")
		return
	}

	if err := h.msgSvc.Delete(messageID, userID); err != nil {
		if errors.Is(err, services.ErrMessageNotFound) {
			response.Error(w, http.StatusForbidden, "forbidden or message not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// GetMembers returns all members of a group chat.
func (h *ChatHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	chatID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	members, err := h.chatSvc.GetMembers(chatID, userID)
	if err != nil {
		if errors.Is(err, services.ErrNotChatMember) {
			response.Error(w, http.StatusForbidden, "forbidden")
			return
		}
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

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Username) == "" {
		response.Error(w, http.StatusBadRequest, "username required")
		return
	}

	member, err := h.chatSvc.AddMember(chatID, userID, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrChatNotFound):
			response.Error(w, http.StatusNotFound, "chat not found")
		case errors.Is(err, services.ErrNotGroupCreator):
			response.Error(w, http.StatusForbidden, "only group creator can add members")
		case errors.Is(err, services.ErrUserNotFound):
			response.Error(w, http.StatusNotFound, "user not found")
		default:
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
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

	removed, err := h.chatSvc.RemoveMember(chatID, userID, memberID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrChatNotFound):
			response.Error(w, http.StatusNotFound, "chat not found")
		case errors.Is(err, services.ErrNotGroupCreator):
			response.Error(w, http.StatusForbidden, "only group creator can remove members")
		case errors.Is(err, services.ErrCannotRemoveSelf):
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"removed": removed})
}
