package mocks

import (
	"time"

	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/stretchr/testify/mock"
)

type RefreshTokenRepository struct {
	mock.Mock
}

func (m *RefreshTokenRepository) Create(userID int, tokenHash string, expiresAt time.Time) (*models.RefreshToken, error) {
	args := m.Called(userID, tokenHash, expiresAt)
	if v := args.Get(0); v != nil {
		return v.(*models.RefreshToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *RefreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	args := m.Called(tokenHash)
	if v := args.Get(0); v != nil {
		return v.(*models.RefreshToken), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *RefreshTokenRepository) Revoke(tokenHash string) error {
	args := m.Called(tokenHash)
	return args.Error(0)
}

func (m *RefreshTokenRepository) DeleteByTokenHash(tokenHash string) error {
	args := m.Called(tokenHash)
	return args.Error(0)
}

func (m *RefreshTokenRepository) DeleteAllByUserID(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}
