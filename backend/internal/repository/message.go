package repository

import (
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/jmoiron/sqlx"
)

// MessageRepository defines the persistence interface for message operations.
type MessageRepository interface {
	Create(chatID, userID int, content string) (*models.MessageResponse, error)
	FindByChatID(chatID, limit, offset int) ([]models.MessageResponse, error)
	MarkChatAsRead(chatID, readerID int) error
	Search(chatID int, query string, limit int) ([]models.MessageResponse, error)
	Update(messageID, userID int, content string) (*models.MessageResponse, error)
	Delete(messageID, userID int) (bool, error)
	GetChatID(messageID int) (int, error)
}

// pgMessageRepository is a PostgreSQL-backed implementation of MessageRepository.
type pgMessageRepository struct {
	db *sqlx.DB
}

// NewMessageRepository initializes and returns a new pgMessageRepository.
func NewMessageRepository(db *sqlx.DB) MessageRepository {
	return &pgMessageRepository{db: db}
}

// Create inserts a new message and returns it with sender info.
func (r *pgMessageRepository) Create(chatID, userID int, content string) (*models.MessageResponse, error) {
	var msg models.MessageResponse
	err := r.db.QueryRowx(`
		INSERT INTO messages (chat_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, chat_id, user_id, content, is_read, created_at
	`, chatID, userID, content).StructScan(&msg.Message)
	if err != nil {
		return nil, err
	}

	if err := r.db.Get(&msg.SenderUsername,
		"SELECT username FROM users WHERE id=$1", userID); err != nil {
		return nil, err
	}

	if err := r.db.Get(&msg.SenderDisplayName,
		"SELECT display_name FROM users WHERE id=$1", userID); err != nil {
		return nil, err
	}

	return &msg, nil
}

// FindByChatID retrieves paginated messages for a chat, ordered newest-first.
func (r *pgMessageRepository) FindByChatID(chatID, limit, offset int) ([]models.MessageResponse, error) {
	rows, err := r.db.Queryx(`
		SELECT m.id, m.chat_id, m.user_id, m.content, m.is_read, m.created_at,
			   u.username AS sender_username, u.display_name AS sender_display_name
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.chat_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var messages []models.MessageResponse
	for rows.Next() {
		var msg models.MessageResponse
		if err := rows.Scan(
			&msg.ID, &msg.ChatID, &msg.UserID, &msg.Content,
			&msg.IsRead, &msg.CreatedAt,
			&msg.SenderUsername, &msg.SenderDisplayName,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if messages == nil {
		messages = []models.MessageResponse{}
	}
	return messages, nil
}

// MarkChatAsRead marks all unread messages in a chat as read, excluding the reader's own.
func (r *pgMessageRepository) MarkChatAsRead(chatID, readerID int) error {
	_, err := r.db.Exec(
		"UPDATE messages SET is_read=true WHERE chat_id=$1 AND user_id != $2 AND is_read=false",
		chatID, readerID,
	)
	return err
}

// Search performs a case-insensitive full-text search over messages in a chat.
func (r *pgMessageRepository) Search(chatID int, query string, limit int) ([]models.MessageResponse, error) {
	rows, err := r.db.Queryx(`
		SELECT m.id, m.chat_id, m.user_id, m.content, m.is_read, m.created_at,
		       u.username AS sender_username, u.display_name AS sender_display_name
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.chat_id = $1
		  AND m.content ILIKE $2
		ORDER BY m.created_at DESC
		LIMIT $3
	`, chatID, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var messages []models.MessageResponse
	for rows.Next() {
		var msg models.MessageResponse
		if err := rows.Scan(
			&msg.ID, &msg.ChatID, &msg.UserID, &msg.Content,
			&msg.IsRead, &msg.CreatedAt,
			&msg.SenderUsername, &msg.SenderDisplayName,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if messages == nil {
		messages = []models.MessageResponse{}
	}
	return messages, nil
}

// Update edits the content of a message owned by the given user.
func (r *pgMessageRepository) Update(messageID, userID int, content string) (*models.MessageResponse, error) {
	var msg models.MessageResponse
	err := r.db.QueryRowx(`
		UPDATE messages SET content = $1
		WHERE id = $2 AND user_id = $3
		RETURNING id, chat_id, user_id, content, is_read, created_at
	`, content, messageID, userID).StructScan(&msg.Message)
	if err != nil {
		return nil, err
	}

	_ = r.db.Get(&msg.SenderUsername, "SELECT username FROM users WHERE id=$1", userID)
	_ = r.db.Get(&msg.SenderDisplayName, "SELECT display_name FROM users WHERE id=$1", userID)

	return &msg, nil
}

// Delete removes a message owned by the given user.
func (r *pgMessageRepository) Delete(messageID, userID int) (bool, error) {
	res, err := r.db.Exec(
		"DELETE FROM messages WHERE id=$1 AND user_id=$2",
		messageID, userID,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// GetChatID returns the chat_id for a given message.
func (r *pgMessageRepository) GetChatID(messageID int) (int, error) {
	var chatID int
	err := r.db.Get(&chatID, "SELECT chat_id FROM messages WHERE id=$1", messageID)
	return chatID, err
}
