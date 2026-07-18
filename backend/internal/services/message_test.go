package services_test

import (
	"testing"

	"github.com/Andrii-K-17/light-chat/internal/mocks"
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageService_GetHistory_NotMember(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	chatRepo.On("IsMember", 5, 1).Return(false, nil)

	_, err := svc.GetHistory(5, 1, 50, 0)

	assert.ErrorIs(t, err, services.ErrNotChatMember)
	chatRepo.AssertExpectations(t)
}

func TestMessageService_GetHistory_Success(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	expected := []models.MessageResponse{{Message: models.Message{ID: 1, Content: "hi"}}}

	chatRepo.On("IsMember", 5, 1).Return(true, nil)
	msgRepo.On("FindByChatID", 5, 50, 0).Return(expected, nil)
	msgRepo.On("MarkChatAsRead", 5, 1).Return(nil)

	messages, err := svc.GetHistory(5, 1, 50, 0)

	require.NoError(t, err)
	assert.Len(t, messages, 1)
	chatRepo.AssertExpectations(t)
	msgRepo.AssertExpectations(t)
}

func TestMessageService_Search_NotMember(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	chatRepo.On("IsMember", 5, 1).Return(false, nil)

	_, err := svc.Search(5, 1, "hello", 50)

	assert.ErrorIs(t, err, services.ErrNotChatMember)
	chatRepo.AssertExpectations(t)
}

func TestMessageService_Create_EmptyContent(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	_, err := svc.Create(5, 1, "   ")

	assert.ErrorIs(t, err, services.ErrEmptyContent)
	msgRepo.AssertNotCalled(t, "Create")
}

func TestMessageService_Create_Success(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	expected := &models.MessageResponse{Message: models.Message{ID: 1, Content: "hello"}}
	msgRepo.On("Create", 5, 1, "hello").Return(expected, nil)

	msg, err := svc.Create(5, 1, "  hello  ")

	require.NoError(t, err)
	assert.Equal(t, "hello", msg.Content)
	msgRepo.AssertExpectations(t)
}

func TestMessageService_Update_EmptyContent(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	_, err := svc.Update(1, 1, "")

	assert.ErrorIs(t, err, services.ErrEmptyContent)
	msgRepo.AssertNotCalled(t, "Update")
}

func TestMessageService_Update_Forbidden(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	msgRepo.On("Update", 1, 2, "new content").Return(nil, assert.AnError)

	_, err := svc.Update(1, 2, "new content")

	assert.ErrorIs(t, err, services.ErrMessageNotFound)
	msgRepo.AssertExpectations(t)
}

func TestMessageService_Update_Success(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	expected := &models.MessageResponse{Message: models.Message{ID: 1, Content: "edited"}}
	msgRepo.On("Update", 1, 1, "edited").Return(expected, nil)

	msg, err := svc.Update(1, 1, "edited")

	require.NoError(t, err)
	assert.Equal(t, "edited", msg.Content)
	msgRepo.AssertExpectations(t)
}

func TestMessageService_Delete_MessageNotFound(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	msgRepo.On("GetChatID", 1).Return(0, assert.AnError)

	_, err := svc.Delete(1, 1)

	assert.ErrorIs(t, err, services.ErrMessageNotFound)
	msgRepo.AssertExpectations(t)
}

func TestMessageService_Delete_Forbidden(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	msgRepo.On("GetChatID", 1).Return(5, nil)
	msgRepo.On("Delete", 1, 2).Return(false, nil)

	_, err := svc.Delete(1, 2)

	assert.ErrorIs(t, err, services.ErrMessageNotFound)
	msgRepo.AssertExpectations(t)
}

func TestMessageService_Delete_Success(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	msgRepo.On("GetChatID", 1).Return(5, nil)
	msgRepo.On("Delete", 1, 1).Return(true, nil)

	chatID, err := svc.Delete(1, 1)

	require.NoError(t, err)
	assert.Equal(t, 5, chatID)
	msgRepo.AssertExpectations(t)
}

func TestMessageService_IsMember(t *testing.T) {
	msgRepo := new(mocks.MessageRepository)
	chatRepo := new(mocks.ChatRepository)
	svc := services.NewMessageService(msgRepo, chatRepo)

	chatRepo.On("IsMember", 5, 1).Return(true, nil)

	isMember, err := svc.IsMember(5, 1)

	require.NoError(t, err)
	assert.True(t, isMember)
	chatRepo.AssertExpectations(t)
}
