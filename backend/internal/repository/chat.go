package repository

import (
	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/jmoiron/sqlx"
)

// ChatRepository defines the persistence interface for chat operations.
type ChatRepository interface {
	Create(name *string, isGroup bool, createdBy int) (*models.Chat, error)
	FindByID(id int) (*models.Chat, error)
	FindAllByUser(userID int) ([]models.ChatResponse, error)
	FindDirectChat(userA, userB int) (*models.Chat, error)
	AddMember(chatID, userID int) error
	IsMember(chatID, userID int) (bool, error)
	GetMembers(chatID int) ([]models.ChatMember, error)
	Delete(chatID, userID int) (bool, error)
	AddMemberByUsername(chatID int, username string) (*models.ChatMember, error)
	RemoveMember(chatID, userID int) (bool, error)
}

// pgChatRepository is a PostgreSQL-backed implementation of ChatRepository.
type pgChatRepository struct {
	db *sqlx.DB
}

// NewChatRepository initializes and returns a new pgChatRepository.
func NewChatRepository(db *sqlx.DB) ChatRepository {
	return &pgChatRepository{db: db}
}

// Create inserts a new chat record and returns it.
func (r *pgChatRepository) Create(name *string, isGroup bool, createdBy int) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.QueryRowx(
		`INSERT INTO chats (name, is_group, created_by)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, is_group, created_by, created_at`,
		name, isGroup, createdBy,
	).StructScan(&chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// FindByID retrieves a chat by its primary key.
func (r *pgChatRepository) FindByID(id int) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Get(&chat,
		"SELECT id, name, is_group, created_by, created_at FROM chats WHERE id=$1", id,
	)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// FindAllByUser retrieves all chats the given user is a member of, with last message and unread count.
func (r *pgChatRepository) FindAllByUser(userID int) ([]models.ChatResponse, error) {
	query := `
		SELECT c.id, c.name, c.is_group, c.created_by, c.created_at
		FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = $1
		ORDER BY (
			SELECT created_at FROM messages
			WHERE chat_id = c.id
			ORDER BY created_at DESC LIMIT 1
		) DESC NULLS LAST
	`
	var chats []models.Chat
	if err := r.db.Select(&chats, query, userID); err != nil {
		return nil, err
	}

	result := make([]models.ChatResponse, 0, len(chats))
	for _, chat := range chats {
		members, err := r.GetMembers(chat.ID)
		if err != nil {
			return nil, err
		}

		var lastMsg *models.Message
		var msg models.Message
		err = r.db.Get(&msg,
			`SELECT id, chat_id, user_id, content, is_read, created_at
			 FROM messages WHERE chat_id=$1 ORDER BY created_at DESC LIMIT 1`,
			chat.ID,
		)
		if err == nil {
			lastMsg = &msg
		}

		var unread int
		_ = r.db.Get(&unread,
			`SELECT COUNT(*) FROM messages
			 WHERE chat_id=$1 AND user_id != $2 AND is_read=false`,
			chat.ID, userID,
		)

		result = append(result, models.ChatResponse{
			Chat:        chat,
			Members:     members,
			LastMessage: lastMsg,
			UnreadCount: unread,
		})
	}

	return result, nil
}

// FindDirectChat finds an existing private chat between two users.
func (r *pgChatRepository) FindDirectChat(userA, userB int) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Get(&chat, `
		SELECT c.id, c.name, c.is_group, c.created_by, c.created_at
		FROM chats c
		JOIN chat_members a ON c.id = a.chat_id AND a.user_id = $1
		JOIN chat_members b ON c.id = b.chat_id AND b.user_id = $2
		WHERE c.is_group = false
		LIMIT 1
	`, userA, userB)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// AddMember inserts a user into a chat, ignoring duplicate entries.
func (r *pgChatRepository) AddMember(chatID, userID int) error {
	_, err := r.db.Exec(
		"INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		chatID, userID,
	)
	return err
}

// IsMember checks whether a user belongs to the given chat.
func (r *pgChatRepository) IsMember(chatID, userID int) (bool, error) {
	var exists bool
	err := r.db.Get(&exists,
		"SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id=$1 AND user_id=$2)",
		chatID, userID,
	)
	return exists, err
}

// GetMembers returns all member profiles for the given chat.
func (r *pgChatRepository) GetMembers(chatID int) ([]models.ChatMember, error) {
	var members []models.ChatMember
	err := r.db.Select(&members, `
		SELECT u.id, u.username, u.display_name, u.status
		FROM users u
		JOIN chat_members cm ON u.id = cm.user_id
		WHERE cm.chat_id = $1
	`, chatID)
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []models.ChatMember{}
	}
	return members, nil
}

// Delete removes a chat if the user is its creator or a member.
func (r *pgChatRepository) Delete(chatID, userID int) (bool, error) {
	res, err := r.db.Exec(`
		DELETE FROM chats
		WHERE id = $1
		  AND (
		    created_by = $2
		    OR (
		      is_group = false
		      AND EXISTS (
		        SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2
		      )
		    )
		  )
	`, chatID, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// AddMemberByUsername resolves a username to a user ID and adds them to the chat.
func (r *pgChatRepository) AddMemberByUsername(chatID int, username string) (*models.ChatMember, error) {
	var member models.ChatMember
	err := r.db.QueryRowx(`
		INSERT INTO chat_members (chat_id, user_id)
		SELECT $1, id FROM users WHERE username = $2
		ON CONFLICT DO NOTHING
		RETURNING (SELECT id FROM users WHERE username = $2),
		          (SELECT username FROM users WHERE username = $2),
		          (SELECT display_name FROM users WHERE username = $2),
		          (SELECT status FROM users WHERE username = $2)
	`, chatID, username).Scan(
		&member.ID, &member.Username, &member.DisplayName, &member.Status,
	)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// RemoveMember removes a user from a chat.
func (r *pgChatRepository) RemoveMember(chatID, userID int) (bool, error) {
	res, err := r.db.Exec(
		"DELETE FROM chat_members WHERE chat_id=$1 AND user_id=$2",
		chatID, userID,
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
