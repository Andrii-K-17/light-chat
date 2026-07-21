package repository_test

import (
	"testing"

	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageRepository_CreateAndFindByChatID(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	user, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, user.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, user.ID))

	msg, err := msgRepo.Create(chat.ID, user.ID, "hello world")
	require.NoError(t, err)
	assert.Equal(t, "hello world", msg.Content)
	assert.Equal(t, "user1", msg.SenderUsername)

	messages, err := msgRepo.FindByChatID(chat.ID, 50, 0)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, "hello world", messages[0].Content)
}

func TestMessageRepository_MarkChatAsRead(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	sender, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)
	reader, err := userRepo.Create("b@test.com", "user2", "User B", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, sender.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, sender.ID))
	require.NoError(t, chatRepo.AddMember(chat.ID, reader.ID))

	_, err = msgRepo.Create(chat.ID, sender.ID, "unread message")
	require.NoError(t, err)

	err = msgRepo.MarkChatAsRead(chat.ID, reader.ID)
	require.NoError(t, err)

	messages, err := msgRepo.FindByChatID(chat.ID, 50, 0)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.True(t, messages[0].IsRead)
}

func TestMessageRepository_Search(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	user, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, user.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, user.ID))

	_, err = msgRepo.Create(chat.ID, user.ID, "hello there")
	require.NoError(t, err)
	_, err = msgRepo.Create(chat.ID, user.ID, "goodbye")
	require.NoError(t, err)

	results, err := msgRepo.Search(chat.ID, "hello", 50)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "hello there", results[0].Content)
}

func TestMessageRepository_Update(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	user, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, user.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, user.ID))

	msg, err := msgRepo.Create(chat.ID, user.ID, "original")
	require.NoError(t, err)

	updated, err := msgRepo.Update(msg.ID, user.ID, "edited")
	require.NoError(t, err)
	assert.Equal(t, "edited", updated.Content)
}

func TestMessageRepository_Update_ForbiddenForOtherUser(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	owner, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)
	other, err := userRepo.Create("b@test.com", "user2", "User B", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, owner.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, owner.ID))
	require.NoError(t, chatRepo.AddMember(chat.ID, other.ID))

	msg, err := msgRepo.Create(chat.ID, owner.ID, "original")
	require.NoError(t, err)

	_, err = msgRepo.Update(msg.ID, other.ID, "hijacked")
	assert.Error(t, err)
}

func TestMessageRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	user, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, user.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, user.ID))

	msg, err := msgRepo.Create(chat.ID, user.ID, "to delete")
	require.NoError(t, err)

	deleted, err := msgRepo.Delete(msg.ID, user.ID)
	require.NoError(t, err)
	assert.True(t, deleted)

	messages, err := msgRepo.FindByChatID(chat.ID, 50, 0)
	require.NoError(t, err)
	assert.Empty(t, messages)
}

func TestMessageRepository_GetChatID(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	msgRepo := repository.NewMessageRepository(db)

	user, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, user.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, user.ID))

	msg, err := msgRepo.Create(chat.ID, user.ID, "hello")
	require.NoError(t, err)

	chatID, err := msgRepo.GetChatID(msg.ID)
	require.NoError(t, err)
	assert.Equal(t, chat.ID, chatID)
}
