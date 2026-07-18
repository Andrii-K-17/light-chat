package mocks

import (
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"github.com/stretchr/testify/mock"
)

type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) Create(email, username, displayName, passwordHash string) (*models.User, error) {
	args := m.Called(email, username, displayName, passwordHash)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepository) FindByID(id int) (*models.User, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if v := args.Get(0); v != nil {
		return v.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepository) ExistsByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) Update(id int, p repository.UserUpdateParams) error {
	args := m.Called(id, p)
	return args.Error(0)
}

func (m *UserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
