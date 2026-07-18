package services_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Andrii-K-17/light-chat/internal/mocks"
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/Andrii-K-17/light-chat/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const (
	testSecret        = "test_secret"
	testExpiry        = time.Hour
	testRefreshExpiry = 7 * 24 * time.Hour
)

func TestAuthService_Register_Success(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	expected := &models.User{
		ID:          1,
		Email:       "user@test.com",
		Username:    "user1",
		DisplayName: "User One",
	}
	stored := &models.RefreshToken{ID: 1, UserID: 1, ExpiresAt: time.Now().Add(testRefreshExpiry)}

	userRepo.On("ExistsByEmail", "user@test.com").Return(false, nil)
	userRepo.On("ExistsByUsername", "user1").Return(false, nil)
	userRepo.On("Create", "user@test.com", "user1", "User One", mock.MatchedBy(func(s string) bool {
		return strings.HasPrefix(s, "$2a$")
	})).Return(expected, nil)
	refreshRepo.On("Create", 1, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(stored, nil)

	user, pair, err := svc.Register(
		"user@test.com",
		"user1",
		"User One",
		"password123",
		testSecret,
		testExpiry,
		testRefreshExpiry,
	)

	require.NoError(t, err)
	assert.Equal(t, expected.Username, user.Username)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	userRepo.AssertExpectations(t)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailTaken(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	userRepo.On("ExistsByEmail", "user@test.com").Return(true, nil)

	_, _, err := svc.Register(
		"user@test.com",
		"user1",
		"User One",
		"password123",
		testSecret,
		testExpiry,
		testRefreshExpiry,
	)

	assert.ErrorIs(t, err, services.ErrEmailTaken)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Register_UsernameTaken(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	userRepo.On("ExistsByEmail", "user@test.com").Return(false, nil)
	userRepo.On("ExistsByUsername", "user1").Return(true, nil)

	_, _, err := svc.Register(
		"user@test.com",
		"user1",
		"User One",
		"password123",
		testSecret,
		testExpiry,
		testRefreshExpiry,
	)

	assert.ErrorIs(t, err, services.ErrUsernameTaken)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	stored := &models.User{ID: 1, Email: "user@test.com", PasswordHash: string(hash)}
	storedToken := &models.RefreshToken{ID: 1, UserID: 1, ExpiresAt: time.Now().Add(testRefreshExpiry)}

	userRepo.On("FindByEmail", "user@test.com").Return(stored, nil)
	refreshRepo.On("Create", 1, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(storedToken, nil)

	user, pair, err := svc.Login(
		"user@test.com",
		"password123",
		testSecret,
		testExpiry,
		testRefreshExpiry,
	)

	require.NoError(t, err)
	assert.Equal(t, stored.ID, user.ID)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	userRepo.AssertExpectations(t)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	stored := &models.User{ID: 1, Email: "user@test.com", PasswordHash: string(hash)}

	userRepo.On("FindByEmail", "user@test.com").Return(stored, nil)

	_, _, err := svc.Login("user@test.com", "wrong", testSecret, testExpiry, testRefreshExpiry)

	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	userRepo.On("FindByEmail", "ghost@test.com").Return(nil, errors.New("not found"))

	_, _, err := svc.Login("ghost@test.com", "password123", testSecret, testExpiry, testRefreshExpiry)

	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	rawToken := "some_raw_token"
	tokenHash := services.HashRefreshToken(rawToken)
	storedToken := &models.RefreshToken{
		ID:        1,
		UserID:    1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	newToken := &models.RefreshToken{ID: 2, UserID: 1, ExpiresAt: time.Now().Add(testRefreshExpiry)}

	refreshRepo.On("FindByTokenHash", tokenHash).Return(storedToken, nil)
	refreshRepo.On("Revoke", tokenHash).Return(nil)
	refreshRepo.On("Create", 1, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(newToken, nil)

	pair, err := svc.Refresh(rawToken, testSecret, testExpiry, testRefreshExpiry)

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_Expired(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	rawToken := "expired_token"
	tokenHash := services.HashRefreshToken(rawToken)
	storedToken := &models.RefreshToken{
		ID:        1,
		UserID:    1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	refreshRepo.On("FindByTokenHash", tokenHash).Return(storedToken, nil)

	_, err := svc.Refresh(rawToken, testSecret, testExpiry, testRefreshExpiry)

	assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	rawToken := "unknown_token"
	tokenHash := services.HashRefreshToken(rawToken)

	refreshRepo.On("FindByTokenHash", tokenHash).Return(nil, errors.New("not found"))

	_, err := svc.Refresh(rawToken, testSecret, testExpiry, testRefreshExpiry)

	assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_Refresh_ReuseDetected(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	rawToken := "reused_token"
	tokenHash := services.HashRefreshToken(rawToken)
	revokedAt := time.Now().Add(-time.Minute)
	storedToken := &models.RefreshToken{
		ID:        1,
		UserID:    1,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
		RevokedAt: &revokedAt,
	}

	refreshRepo.On("FindByTokenHash", tokenHash).Return(storedToken, nil)
	refreshRepo.On("DeleteAllByUserID", 1).Return(nil)

	_, err := svc.Refresh(rawToken, testSecret, testExpiry, testRefreshExpiry)

	assert.ErrorIs(t, err, services.ErrRefreshTokenReused)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_DeleteAccount(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	refreshRepo.On("DeleteAllByUserID", 15).Return(nil)
	userRepo.On("Delete", 15).Return(nil)

	err := svc.DeleteAccount(15)

	require.NoError(t, err)
	userRepo.AssertExpectations(t)
	refreshRepo.AssertExpectations(t)
}

func TestAuthService_UpdateProfile(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	newName := "Updated Name"
	expected := &models.User{ID: 1, DisplayName: newName}

	userRepo.On("Update", 1, mock.Anything).Return(nil)
	userRepo.On("FindByID", 1).Return(expected, nil)

	user, err := svc.UpdateProfile(1, repository.UserUpdateParams{DisplayName: &newName})

	require.NoError(t, err)
	assert.Equal(t, newName, user.DisplayName)
	userRepo.AssertExpectations(t)
}

func TestAuthService_FindByUsername(t *testing.T) {
	userRepo := new(mocks.UserRepository)
	refreshRepo := new(mocks.RefreshTokenRepository)
	svc := services.NewAuthService(userRepo, refreshRepo)

	expected := &models.User{ID: 1, Username: "user1"}
	userRepo.On("FindByUsername", "user1").Return(expected, nil)

	user, err := svc.FindByUsername("user1")

	require.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
	userRepo.AssertExpectations(t)
}
