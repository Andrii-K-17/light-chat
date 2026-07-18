package mocks

import (
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/stretchr/testify/mock"
)

type MessageRepository struct {
	mock.Mock
}

func (m *MessageRepository) Create(chatID, userID int, content string) (*models.MessageResponse, error) {
	args := m.Called(chatID, userID, content)
	if v := args.Get(0); v != nil {
		return v.(*models.MessageResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MessageRepository) FindByChatID(chatID, limit, offset int) ([]models.MessageResponse, error) {
	args := m.Called(chatID, limit, offset)
	if v := args.Get(0); v != nil {
		return v.([]models.MessageResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MessageRepository) MarkChatAsRead(chatID, readerID int) error {
	args := m.Called(chatID, readerID)
	return args.Error(0)
}

func (m *MessageRepository) Search(chatID int, query string, limit int) ([]models.MessageResponse, error) {
	args := m.Called(chatID, query, limit)
	if v := args.Get(0); v != nil {
		return v.([]models.MessageResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MessageRepository) Update(messageID, userID int, content string) (*models.MessageResponse, error) {
	args := m.Called(messageID, userID, content)
	if v := args.Get(0); v != nil {
		return v.(*models.MessageResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MessageRepository) Delete(messageID, userID int) (bool, error) {
	args := m.Called(messageID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MessageRepository) GetChatID(messageID int) (int, error) {
	args := m.Called(messageID)
	return args.Int(0), args.Error(1)
}
