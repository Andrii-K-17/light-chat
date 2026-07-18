package services_test

import (
	"testing"

	"github.com/Andrii-K-17/light-chat/internal/mocks"
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatService_GetAll(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	expected := []models.ChatResponse{{Chat: models.Chat{ID: 1}}}
	repo.On("FindAllByUser", 1).Return(expected, nil)

	chats, err := svc.GetAll(1)

	require.NoError(t, err)
	assert.Len(t, chats, 1)
	repo.AssertExpectations(t)
}

func TestChatService_Create_ReusesExistingDirectChat(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	existing := &models.Chat{ID: 5, IsGroup: false}
	members := []models.ChatMember{{ID: 1}, {ID: 2}}

	repo.On("FindDirectChat", 1, 2).Return(existing, nil)
	repo.On("GetMembers", 5).Return(members, nil)

	chat, err := svc.Create(1, services.CreateChatPayload{
		IsGroup:   false,
		MemberIDs: []int{2},
	})

	require.NoError(t, err)
	assert.Equal(t, 5, chat.ID)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Create")
}

func TestChatService_Create_NewGroupChat(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	name := "Team"
	created := &models.Chat{ID: 10, Name: &name, IsGroup: true}
	members := []models.ChatMember{{ID: 1}, {ID: 2}, {ID: 3}}

	repo.On("Create", &name, true, 1).Return(created, nil)
	repo.On("AddMember", 10, 1).Return(nil)
	repo.On("AddMember", 10, 2).Return(nil)
	repo.On("AddMember", 10, 3).Return(nil)
	repo.On("GetMembers", 10).Return(members, nil)

	chat, err := svc.Create(1, services.CreateChatPayload{
		Name:      &name,
		IsGroup:   true,
		MemberIDs: []int{1, 2, 3},
	})

	require.NoError(t, err)
	assert.Equal(t, 10, chat.ID)
	assert.Len(t, chat.Members, 3)
	repo.AssertExpectations(t)
}

func TestChatService_Delete_NotMember(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	repo.On("Delete", 5, 1).Return(false, nil)

	err := svc.Delete(5, 1)

	assert.ErrorIs(t, err, services.ErrNotChatMember)
	repo.AssertExpectations(t)
}

func TestChatService_Delete_Success(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	repo.On("Delete", 5, 1).Return(true, nil)

	err := svc.Delete(5, 1)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestChatService_GetMembers_Forbidden(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	repo.On("IsMember", 5, 1).Return(false, nil)

	_, err := svc.GetMembers(5, 1)

	assert.ErrorIs(t, err, services.ErrNotChatMember)
	repo.AssertExpectations(t)
}

func TestChatService_AddMember_NotCreator(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	creatorID := 99
	chat := &models.Chat{ID: 5, IsGroup: true, CreatedBy: &creatorID}
	repo.On("FindByID", 5).Return(chat, nil)

	_, err := svc.AddMember(5, 1, "someuser")

	assert.ErrorIs(t, err, services.ErrNotGroupCreator)
	repo.AssertExpectations(t)
}

func TestChatService_AddMember_UserNotFound(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	creatorID := 1
	chat := &models.Chat{ID: 5, IsGroup: true, CreatedBy: &creatorID}
	repo.On("FindByID", 5).Return(chat, nil)
	repo.On("AddMemberByUsername", 5, "ghost").Return(nil, assert.AnError)

	_, err := svc.AddMember(5, 1, "ghost")

	assert.ErrorIs(t, err, services.ErrUserNotFound)
	repo.AssertExpectations(t)
}

func TestChatService_AddMember_Success(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	creatorID := 1
	chat := &models.Chat{ID: 5, IsGroup: true, CreatedBy: &creatorID}
	newMember := &models.ChatMember{ID: 2, Username: "user2"}

	repo.On("FindByID", 5).Return(chat, nil)
	repo.On("AddMemberByUsername", 5, "user2").Return(newMember, nil)

	member, err := svc.AddMember(5, 1, "user2")

	require.NoError(t, err)
	assert.Equal(t, "user2", member.Username)
	repo.AssertExpectations(t)
}

func TestChatService_RemoveMember_CannotRemoveSelf(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	creatorID := 1
	chat := &models.Chat{ID: 5, IsGroup: true, CreatedBy: &creatorID}
	repo.On("FindByID", 5).Return(chat, nil)

	_, err := svc.RemoveMember(5, 1, 1)

	assert.ErrorIs(t, err, services.ErrCannotRemoveSelf)
	repo.AssertExpectations(t)
}

func TestChatService_RemoveMember_Success(t *testing.T) {
	repo := new(mocks.ChatRepository)
	svc := services.NewChatService(repo)

	creatorID := 1
	chat := &models.Chat{ID: 5, IsGroup: true, CreatedBy: &creatorID}
	repo.On("FindByID", 5).Return(chat, nil)
	repo.On("RemoveMember", 5, 2).Return(true, nil)

	removed, err := svc.RemoveMember(5, 1, 2)

	require.NoError(t, err)
	assert.True(t, removed)
	repo.AssertExpectations(t)
}
