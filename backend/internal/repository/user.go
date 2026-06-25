package repository

import (
	"fmt"
	"strings"

	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/jmoiron/sqlx"
)

// UserRepository defines the persistence interface for user operations.
type UserRepository interface {
	Create(email, username, displayName, passwordHash string) (*models.User, error)
	FindByID(id int) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
	Update(id int, p UserUpdateParams) error
	Delete(id int) error
}

// UserUpdateParams holds the optional fields that can be patched on a user.
type UserUpdateParams struct {
	DisplayName *string
	Username    *string
	Email       *string
	Status      *string
}

// pgUserRepository is a PostgreSQL-backed implementation of UserRepository.
type pgUserRepository struct {
	db *sqlx.DB
}

// NewUserRepository initializes and returns a new pgUserRepository.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &pgUserRepository{db: db}
}

// Create inserts a new user record and returns the created user.
func (r *pgUserRepository) Create(email, username, displayName, passwordHash string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowx(
		`INSERT INTO users (email, username, display_name, password_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, username, display_name, status, created_at`,
		email, username, displayName, passwordHash,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by their primary key.
func (r *pgUserRepository) FindByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user,
		"SELECT id, email, username, display_name, status, created_at FROM users WHERE id=$1", id,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail retrieves a user by their email address.
func (r *pgUserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user,
		"SELECT id, email, username, display_name, password_hash, status FROM users WHERE email=$1", email,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername retrieves a user by their unique username.
func (r *pgUserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user,
		"SELECT id, email, username, display_name, status FROM users WHERE username=$1", username,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail checks whether a user with the given email already exists.
func (r *pgUserRepository) ExistsByEmail(email string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email,
	)
	return exists, err
}

// ExistsByUsername checks whether a user with the given username already exists.
func (r *pgUserRepository) ExistsByUsername(username string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists,
		"SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", username,
	)
	return exists, err
}

// Update applies a partial patch to a user record.
func (r *pgUserRepository) Update(id int, p UserUpdateParams) error {
	sets := make([]string, 0)
	args := make([]any, 0)
	idx := 1

	if p.DisplayName != nil {
		sets = append(sets, fmt.Sprintf("display_name=$%d", idx))
		args = append(args, *p.DisplayName)
		idx++
	}
	if p.Username != nil {
		sets = append(sets, fmt.Sprintf("username=$%d", idx))
		args = append(args, *p.Username)
		idx++
	}
	if p.Email != nil {
		sets = append(sets, fmt.Sprintf("email=$%d", idx))
		args = append(args, *p.Email)
		idx++
	}
	if p.Status != nil {
		sets = append(sets, fmt.Sprintf("status=$%d", idx))
		args = append(args, *p.Status)
		idx++
	}

	if len(sets) == 0 {
		return fmt.Errorf("no fields to update")
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id=$%d", strings.Join(sets, ", "), idx)
	_, err := r.db.Exec(query, args...)
	return err
}

// Delete removes a user record by their primary key.
func (r *pgUserRepository) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id=$1", id)
	return err
}
