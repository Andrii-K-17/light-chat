package services

import (
	"errors"
	"strings"

	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/repository"
)

// ErrMessageNotFound is returned when a message does not exist or the user lacks permission.
var ErrMessageNotFound = errors.New("message not found or forbidden")

// ErrEmptyContent is returned when a message edit payload has no content.
var ErrEmptyContent = errors.New("content is required")

// MessageService handles message history, search, creation, editing, and deletion.
type MessageService struct {
	repo     repository.MessageRepository
	chatRepo repository.ChatRepository
}

// NewMessageService initializes and returns a new MessageService.
func NewMessageService(
	repo repository.MessageRepository,
	chatRepo repository.ChatRepository,
) *MessageService {
	return &MessageService{
		repo:     repo,
		chatRepo: chatRepo,
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

// Create persists a new message. The caller is responsible for broadcasting it.
func (s *MessageService) Create(chatID, userID int, content string) (*models.MessageResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyContent
	}
	return s.repo.Create(chatID, userID, content)
}

// Update edits a message owned by the given user. The caller is responsible for broadcasting it.
func (s *MessageService) Update(messageID, userID int, content string) (*models.MessageResponse, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyContent
	}

	msg, err := s.repo.Update(messageID, userID, content)
	if err != nil {
		return nil, ErrMessageNotFound
	}
	
	return msg, nil
}

// Delete removes a message owned by the given user and returns the chat ID for broadcasting.
func (s *MessageService) Delete(messageID, userID int) (chatID int, err error) {
	chatID, err = s.repo.GetChatID(messageID)
	if err != nil {
		return 0, ErrMessageNotFound
	}

	deleted, err := s.repo.Delete(messageID, userID)
	if err != nil {
		return 0, err
	}
	if !deleted {
		return 0, ErrMessageNotFound
	}

	return chatID, nil
}

// IsMember checks whether a user belongs to the given chat.
func (s *MessageService) IsMember(chatID, userID int) (bool, error) {
	return s.chatRepo.IsMember(chatID, userID)
}
