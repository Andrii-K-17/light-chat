package repository_test

import (
	"testing"

	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	user, err := repo.Create("user1@test.com", "user1", "User One", "hash123")
	require.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
	assert.Positive(t, user.ID)

	found, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "User One", found.DisplayName)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	_, err := repo.Create("user1@test.com", "user1", "User One", "hash")
	require.NoError(t, err)

	found, err := repo.FindByEmail("user1@test.com")
	require.NoError(t, err)
	assert.Equal(t, "user1", found.Username)
}

func TestUserRepository_ExistsByEmailAndUsername(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	_, err := repo.Create("user1@test.com", "user1", "User One", "hash")
	require.NoError(t, err)

	existsEmail, err := repo.ExistsByEmail("user1@test.com")
	require.NoError(t, err)
	assert.True(t, existsEmail)

	existsUsername, err := repo.ExistsByUsername("user1")
	require.NoError(t, err)
	assert.True(t, existsUsername)

	existsGhost, err := repo.ExistsByUsername("ghost")
	require.NoError(t, err)
	assert.False(t, existsGhost)
}

func TestUserRepository_Update(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	user, err := repo.Create("user1@test.com", "user1", "User One", "hash")
	require.NoError(t, err)

	newName := "Updated Name"
	newStatus := "Busy"
	err = repo.Update(user.ID, repository.UserUpdateParams{
		DisplayName: &newName,
		Status:      &newStatus,
	})
	require.NoError(t, err)

	found, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, newName, found.DisplayName)
	assert.Equal(t, newStatus, found.Status)
}

func TestUserRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)

	user, err := repo.Create("user1@test.com", "user1", "User One", "hash")
	require.NoError(t, err)

	err = repo.Delete(user.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(user.ID)
	assert.Error(t, err)
}
