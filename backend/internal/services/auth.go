package services

import (
	"errors"
	"time"

	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// ErrEmailTaken is returned when a registration email is already in use.
var ErrEmailTaken = errors.New("this email is already taken")

// ErrUsernameTaken is returned when a username is already in use.
var ErrUsernameTaken = errors.New("this username is already taken")

// ErrInvalidCredentials is returned on a failed login attempt.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrInvalidRefreshToken is returned when a refresh token is missing or expired.
var ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

// ErrRefreshTokenReused is returned when a rotated refresh token is reused.
var ErrRefreshTokenReused = errors.New("refresh token reuse detected")

// TokenPair holds both the access JWT and the opaque refresh token.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// AuthService handles user registration, login, and session management.
type AuthService struct {
	repo        repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
}

// NewAuthService initializes and returns a new AuthService.
func NewAuthService(repo repository.UserRepository, refreshRepo repository.RefreshTokenRepository) *AuthService {
	return &AuthService{repo: repo, refreshRepo: refreshRepo}
}

// issueTokenPair creates an access JWT and a refresh token for the given user.
func (s *AuthService) issueTokenPair(userID int, jwtSecret string, jwtExpiry, refreshExpiry time.Duration) (*TokenPair, error) {
	accessToken, err := IssueJWT(userID, jwtSecret, jwtExpiry)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenHash := HashRefreshToken(rawRefresh)
	expiresAt := time.Now().Add(refreshExpiry)

	if _, err := s.refreshRepo.Create(userID, tokenHash, expiresAt); err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

// Register validates uniqueness, hashes the password, and creates a new user.
func (s *AuthService) Register(email, username, displayName, password, jwtSecret string, jwtExpiry, refreshExpiry time.Duration) (*models.User, *TokenPair, error) {
	if exists, err := s.repo.ExistsByEmail(email); err != nil || exists {
		if exists {
			return nil, nil, ErrEmailTaken
		}
		return nil, nil, err
	}

	if exists, err := s.repo.ExistsByUsername(username); err != nil || exists {
		if exists {
			return nil, nil, ErrUsernameTaken
		}
		return nil, nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.repo.Create(email, username, displayName, string(hash))
	if err != nil {
		return nil, nil, err
	}

	pair, err := s.issueTokenPair(user.ID, jwtSecret, jwtExpiry, refreshExpiry)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

// Login verifies credentials and returns the authenticated user with a token pair.
func (s *AuthService) Login(email, password, jwtSecret string, jwtExpiry, refreshExpiry time.Duration) (*models.User, *TokenPair, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(user.ID, jwtSecret, jwtExpiry, refreshExpiry)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

// Refresh validates the incoming refresh token, rotates it, and returns a new token pair.
func (s *AuthService) Refresh(rawToken, jwtSecret string, jwtExpiry, refreshExpiry time.Duration) (*TokenPair, error) {
	tokenHash := HashRefreshToken(rawToken)

	stored, err := s.refreshRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if stored.RevokedAt != nil {
		_ = s.refreshRepo.DeleteAllByUserID(stored.UserID)
		return nil, ErrRefreshTokenReused
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	if err := s.refreshRepo.Revoke(tokenHash); err != nil {
		return nil, err
	}

	return s.issueTokenPair(stored.UserID, jwtSecret, jwtExpiry, refreshExpiry)
}

// Logout removes the given refresh token from the store.
func (s *AuthService) Logout(rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return s.refreshRepo.DeleteByTokenHash(HashRefreshToken(rawToken))
}

// GetByID fetches a user by their primary key.
func (s *AuthService) GetByID(id int) (*models.User, error) {
	return s.repo.FindByID(id)
}

// DeleteAccount removes a user and all their refresh tokens from the database.
func (s *AuthService) DeleteAccount(userID int) error {
	if err := s.refreshRepo.DeleteAllByUserID(userID); err != nil {
		return err
	}
	return s.repo.Delete(userID)
}

// UpdateProfile applies a partial patch to a user's profile and returns the updated user.
func (s *AuthService) UpdateProfile(userID int, p repository.UserUpdateParams) (*models.User, error) {
	if err := s.repo.Update(userID, p); err != nil {
		return nil, err
	}
	return s.repo.FindByID(userID)
}

// FindByUsername looks up a user by their exact username.
func (s *AuthService) FindByUsername(username string) (*models.User, error) {
	return s.repo.FindByUsername(username)
}
