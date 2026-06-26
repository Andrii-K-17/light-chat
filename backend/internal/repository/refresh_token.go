package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/Andrii-K-17/light-chat/internal/models"
)

// RefreshTokenRepository defines the persistence interface for refresh token operations.
type RefreshTokenRepository interface {
	Create(userID int, tokenHash string, expiresAt time.Time) (*models.RefreshToken, error)
	FindByTokenHash(tokenHash string) (*models.RefreshToken, error)
	Revoke(tokenHash string) error
	DeleteByTokenHash(tokenHash string) error
	DeleteAllByUserID(userID int) error
}

// pgRefreshTokenRepository is a PostgreSQL-backed implementation of RefreshTokenRepository.
type pgRefreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository initializes and returns a new pgRefreshTokenRepository.
func NewRefreshTokenRepository(db *sqlx.DB) RefreshTokenRepository {
	return &pgRefreshTokenRepository{db: db}
}

// Create inserts a new refresh token record and returns it.
func (r *pgRefreshTokenRepository) Create(userID int, tokenHash string, expiresAt time.Time) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.QueryRowx(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at`,
		userID, tokenHash, expiresAt,
	).StructScan(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// FindByTokenHash retrieves a refresh token record by its hash.
func (r *pgRefreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Get(&token,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		 FROM refresh_tokens WHERE token_hash=$1`,
		tokenHash,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Revoke marks a refresh token as used so a later replay can be detected.
func (r *pgRefreshTokenRepository) Revoke(tokenHash string) error {
	_, err := r.db.Exec(
		"UPDATE refresh_tokens SET revoked_at=NOW() WHERE token_hash=$1", tokenHash,
	)
	return err
}

// DeleteByTokenHash removes a refresh token record by its hash.
func (r *pgRefreshTokenRepository) DeleteByTokenHash(tokenHash string) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE token_hash=$1", tokenHash)
	return err
}

// DeleteAllByUserID removes all refresh tokens belonging to the given user.
func (r *pgRefreshTokenRepository) DeleteAllByUserID(userID int) error {
	_, err := r.db.Exec("DELETE FROM refresh_tokens WHERE user_id=$1", userID)
	return err
}
