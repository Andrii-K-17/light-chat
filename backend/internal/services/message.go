package services

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/Andrii-K-17/light-chat/internal/ws"
)

// ErrMessageNotFound is returned when a message does not exist or the user lacks permission.
var ErrMessageNotFound = errors.New("message not found or forbidden")

// ErrEmptyContent is returned when a message edit payload has no content.
var ErrEmptyContent = errors.New("content is required")

// MessageService handles message history, search, editing, and deletion
type MessageService struct {
	repo     repository.MessageRepository
	chatRepo repository.ChatRepository
	hub      *ws.Hub
}

// NewMessageService initializes and returns a new MessageService.
func NewMessageService(
	repo repository.MessageRepository,
	chatRepo repository.ChatRepository,
	hub *ws.Hub,
) *MessageService {
	return &MessageService{
		repo:     repo,
		chatRepo: chatRepo,
		hub:      hub,
	}
}

// GetHistory returns paginated message history and marks the chat as read for the requester.
func (s *MessageService) GetHistory(chatID, userID, limit, offset int) ([]models.MessageResponse, error) {
	isMember, err := s.chatRepo.IsMember(chatID, userID)
	if err != nil || !isMember {
		return nil, ErrNotChatMember
	}

	messages, err := s.repo.FindByChatID(chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	if err := s.repo.MarkChatAsRead(chatID, userID); err != nil {
		return nil, err
	}

	return messages, nil
}

// Search performs a full-text search over messages in a chat the requester belongs to.
func (s *MessageService) Search(chatID, userID int, query string, limit int) ([]models.MessageResponse, error) {
	isMember, err := s.chatRepo.IsMember(chatID, userID)
	if err != nil || !isMember {
		return nil, ErrNotChatMember
	}

	return s.repo.Search(chatID, query, limit)
}

// Update edits a message owned by the given user and broadcasts the change to the chat.
func (s *MessageService) Update(messageID, userID int, content string) (*models.MessageResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyContent
	}

	msg, err := s.repo.Update(messageID, userID, content)
	if err != nil {
		return nil, ErrMessageNotFound
	}

	event, _ := json.Marshal(ws.Event{
		Type:    "message_updated",
		Payload: mustMarshal(msg),
	})
	s.hub.BroadcastToChat(msg.ChatID, event)

	return msg, nil
}

// Delete removes a message owned by the given user and broadcasts the deletion to the chat.
func (s *MessageService) Delete(messageID, userID int) error {
	chatID, err := s.repo.GetChatID(messageID)
	if err != nil {
		return ErrMessageNotFound
	}

	deleted, err := s.repo.Delete(messageID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrMessageNotFound
	}

	payload, _ := json.Marshal(map[string]int{
		"message_id": messageID,
		"chat_id":    chatID,
	})
	event, _ := json.Marshal(ws.Event{
		Type:    "message_deleted",
		Payload: payload,
	})
	s.hub.BroadcastToChat(chatID, event)

	return nil
}

// mustMarshal marshals v to JSON, returning nil on error.
func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
