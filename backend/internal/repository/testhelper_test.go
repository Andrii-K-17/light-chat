package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// newTestDB spins up a disposable Postgres container and returns a ready *sqlx.DB.
func newTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(uuid.NewString()),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	require.NoError(t, err)

	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)

	migrate(t, db)
	return db
}

// migrate applies the minimal schema required by the repository tests.
func migrate(t *testing.T, db *sqlx.DB) {
	t.Helper()

	query := (`
		CREATE TABLE IF NOT EXISTS users (
			id            SERIAL       PRIMARY KEY,
			email         TEXT         UNIQUE NOT NULL,
			username      TEXT         UNIQUE NOT NULL,
			display_name  TEXT         NOT NULL,
			password_hash TEXT         NOT NULL,
			status        TEXT         NOT NULL DEFAULT '',
			created_at    TIMESTAMPTZ  DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id         SERIAL      PRIMARY KEY,
			user_id    INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT        NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS chats (
			id         SERIAL      PRIMARY KEY,
			name       TEXT,
			is_group   BOOLEAN     NOT NULL DEFAULT FALSE,
			created_by INT         REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS chat_members (
			chat_id   INT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
			user_id   INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			joined_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (chat_id, user_id)
		);
		CREATE TABLE IF NOT EXISTS messages (
			id         SERIAL      PRIMARY KEY,
			chat_id    INT         NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
			user_id    INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content    TEXT        NOT NULL,
			is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	
	_, err := db.Exec(query)
	require.NoError(t, err)
}
