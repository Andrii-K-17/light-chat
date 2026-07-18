package mocks

import (
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/stretchr/testify/mock"
)

type ChatRepository struct {
	mock.Mock
}

func (m *ChatRepository) Create(name *string, isGroup bool, createdBy int) (*models.Chat, error) {
	args := m.Called(name, isGroup, createdBy)
	if v := args.Get(0); v != nil {
		return v.(*models.Chat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ChatRepository) FindByID(id int) (*models.Chat, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*models.Chat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ChatRepository) FindAllByUser(userID int) ([]models.ChatResponse, error) {
	args := m.Called(userID)
	if v := args.Get(0); v != nil {
		return v.([]models.ChatResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ChatRepository) FindDirectChat(userA, userB int) (*models.Chat, error) {
	args := m.Called(userA, userB)
	if v := args.Get(0); v != nil {
		return v.(*models.Chat), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ChatRepository) AddMember(chatID, userID int) error {
	args := m.Called(chatID, userID)
	return args.Error(0)
}

func (m *ChatRepository) IsMember(chatID, userID int) (bool, error) {
	args := m.Called(chatID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *ChatRepository) GetMembers(chatID int) ([]models.ChatMember, error) {
	args := m.Called(chatID)
	if v := args.Get(0); v != nil {
		return v.([]models.ChatMember), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ChatRepository) Delete(chatID, userID int) (bool, error) {
	args := m.Called(chatID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *ChatRepository) AddMemberByUsername(chatID int, username string) (*models.ChatMember, error) {
	args := m.Called(chatID, username)
	if v := args.Get(0); v != nil {
		return v.(*models.ChatMember), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *ChatRepository) RemoveMember(chatID, userID int) (bool, error) {
	args := m.Called(chatID, userID)
	return args.Bool(0), args.Error(1)
}
