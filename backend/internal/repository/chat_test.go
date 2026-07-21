package repository_test

import (
	"testing"

	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatRepository_CreateDirectChatAndListForUser(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)

	userA, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)
	userB, err := userRepo.Create("b@test.com", "user2", "User B", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, userA.ID)
	require.NoError(t, err)

	require.NoError(t, chatRepo.AddMember(chat.ID, userA.ID))
	require.NoError(t, chatRepo.AddMember(chat.ID, userB.ID))

	chats, err := chatRepo.FindAllByUser(userA.ID)
	require.NoError(t, err)
	require.Len(t, chats, 1)
	assert.Len(t, chats[0].Members, 2)
}

func TestChatRepository_FindDirectChat(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)

	userA, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)
	userB, err := userRepo.Create("b@test.com", "user2", "User B", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, false, userA.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, userA.ID))
	require.NoError(t, chatRepo.AddMember(chat.ID, userB.ID))

	found, err := chatRepo.FindDirectChat(userA.ID, userB.ID)
	require.NoError(t, err)
	assert.Equal(t, chat.ID, found.ID)
}

func TestChatRepository_IsMember(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)

	user, err := userRepo.Create("a@test.com", "user1", "User A", "hash")
	require.NoError(t, err)

	chat, err := chatRepo.Create(nil, true, user.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, user.ID))

	isMember, err := chatRepo.IsMember(chat.ID, user.ID)
	require.NoError(t, err)
	assert.True(t, isMember)

	isMember, err = chatRepo.IsMember(chat.ID, 9999)
	require.NoError(t, err)
	assert.False(t, isMember)
}

func TestChatRepository_Delete_ByCreator(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)

	creator, err := userRepo.Create("creator@test.com", "creator", "Creator", "hash")
	require.NoError(t, err)

	name := "Group"
	chat, err := chatRepo.Create(&name, true, creator.ID)
	require.NoError(t, err)

	deleted, err := chatRepo.Delete(chat.ID, creator.ID)
	require.NoError(t, err)
	assert.True(t, deleted)

	_, err = chatRepo.FindByID(chat.ID)
	assert.Error(t, err)
}

func TestChatRepository_Delete_ForbiddenForNonCreatorGroupMember(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)

	creator, err := userRepo.Create("creator@test.com", "creator", "Creator", "hash")
	require.NoError(t, err)
	member, err := userRepo.Create("member@test.com", "member", "Member", "hash")
	require.NoError(t, err)

	name := "Group"
	chat, err := chatRepo.Create(&name, true, creator.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, member.ID))

	deleted, err := chatRepo.Delete(chat.ID, member.ID)
	require.NoError(t, err)
	assert.False(t, deleted)
}

func TestChatRepository_AddMemberByUsername(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)

	creator, err := userRepo.Create("creator@test.com", "creator", "Creator", "hash")
	require.NoError(t, err)
	newUser, err := userRepo.Create("new@test.com", "newuser", "New User", "hash")
	require.NoError(t, err)

	name := "Group"
	chat, err := chatRepo.Create(&name, true, creator.ID)
	require.NoError(t, err)

	member, err := chatRepo.AddMemberByUsername(chat.ID, "newuser")
	require.NoError(t, err)
	assert.Equal(t, newUser.ID, member.ID)

	isMember, err := chatRepo.IsMember(chat.ID, newUser.ID)
	require.NoError(t, err)
	assert.True(t, isMember)
}

func TestChatRepository_RemoveMember(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)

	creator, err := userRepo.Create("creator@test.com", "creator", "Creator", "hash")
	require.NoError(t, err)
	member, err := userRepo.Create("member@test.com", "member", "Member", "hash")
	require.NoError(t, err)

	name := "Group"
	chat, err := chatRepo.Create(&name, true, creator.ID)
	require.NoError(t, err)
	require.NoError(t, chatRepo.AddMember(chat.ID, member.ID))

	removed, err := chatRepo.RemoveMember(chat.ID, member.ID)
	require.NoError(t, err)
	assert.True(t, removed)

	isMember, err := chatRepo.IsMember(chat.ID, member.ID)
	require.NoError(t, err)
	assert.False(t, isMember)
}
