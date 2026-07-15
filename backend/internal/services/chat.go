package services

import (
	"errors"
	"strings"

	"github.com/Andrii-K-17/light-chat/internal/models"
	"github.com/Andrii-K-17/light-chat/internal/repository"
)

// ErrChatNotFound is returned when a chat does not exist.
var ErrChatNotFound = errors.New("chat not found")

// ErrNotChatMember is returned when the user is not a member of the chat.
var ErrNotChatMember = errors.New("not a member of this chat")

// ErrNotGroupCreator is returned when a non-creator attempts a creator-only action.
var ErrNotGroupCreator = errors.New("only group creator can perform this action")

// ErrCannotRemoveSelf is returned when the creator tries to remove themselves.
var ErrCannotRemoveSelf = errors.New("cannot remove yourself")

// ErrUserNotFound is returned when a target username does not resolve to a user.
var ErrUserNotFound = errors.New("user not found")

// ChatService handles chat creation, membership, and lifecycle logic.
type ChatService struct {
	repo repository.ChatRepository
}

// NewChatService initializes and returns a new ChatService.
func NewChatService(repo repository.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

// GetAll retrieves all chats the given user belongs to.
func (s *ChatService) GetAll(userID int) ([]models.ChatResponse, error) {
	return s.repo.FindAllByUser(userID)
}

// CreateChatPayload holds the fields required to create a new chat.
type CreateChatPayload struct {
	Name      *string
	IsGroup   bool
	MemberIDs []int
}

// ChatWithMembers is a lightweight chat representation returned right after creation.
type ChatWithMembers struct {
	models.Chat
	Members []models.ChatMember `json:"members"`
}

// Create creates a direct or group chat, reusing an existing direct chat if one exists.
func (s *ChatService) Create(userID int, p CreateChatPayload) (*ChatWithMembers, error) {
	if !p.IsGroup && len(p.MemberIDs) == 1 {
		if existing, err := s.repo.FindDirectChat(userID, p.MemberIDs[0]); err == nil {
			members, _ := s.repo.GetMembers(existing.ID)
			return &ChatWithMembers{Chat: *existing, Members: members}, nil
		}
	}

	chat, err := s.repo.Create(p.Name, p.IsGroup, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AddMember(chat.ID, userID); err != nil {
		return nil, err
	}

	for _, memberID := range p.MemberIDs {
		if memberID == userID {
			continue
		}
		if err := s.repo.AddMember(chat.ID, memberID); err != nil {
			return nil, err
		}
	}

	members, err := s.repo.GetMembers(chat.ID)
	if err != nil {
		return nil, err
	}

	return &ChatWithMembers{Chat: *chat, Members: members}, nil
}

// Delete removes a chat if the user has permission.
func (s *ChatService) Delete(chatID, userID int) error {
	deleted, err := s.repo.Delete(chatID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotChatMember
	}
	return nil
}

// GetMembers returns all members of a chat if the requester belongs to it.
func (s *ChatService) GetMembers(chatID, requesterID int) ([]models.ChatMember, error) {
	isMember, err := s.repo.IsMember(chatID, requesterID)
	if err != nil || !isMember {
		return nil, ErrNotChatMember
	}

	return s.repo.GetMembers(chatID)
}

// AddMember adds a user to a group chat by username; only the creator may do this.
func (s *ChatService) AddMember(chatID, requesterID int, username string) (*models.ChatMember, error) {
	chat, err := s.repo.FindByID(chatID)
	if err != nil {
		return nil, ErrChatNotFound
	}
	if !chat.IsGroup || chat.CreatedBy == nil || *chat.CreatedBy != requesterID {
		return nil, ErrNotGroupCreator
	}

	member, err := s.repo.AddMemberByUsername(chatID, strings.TrimSpace(username))
	if err != nil {
		return nil, ErrUserNotFound
	}
	return member, nil
}

// RemoveMember removes a user from a group chat; only the creator may do this, and not themselves.
func (s *ChatService) RemoveMember(chatID, requesterID, memberID int) (bool, error) {
	chat, err := s.repo.FindByID(chatID)
	if err != nil {
		return false, ErrChatNotFound
	}
	if !chat.IsGroup || chat.CreatedBy == nil || *chat.CreatedBy != requesterID {
		return false, ErrNotGroupCreator
	}
	if memberID == requesterID {
		return false, ErrCannotRemoveSelf
	}

	return s.repo.RemoveMember(chatID, memberID)
}

// IsMember checks whether a user belongs to the given chat.
func (s *ChatService) IsMember(chatID, userID int) (bool, error) {
	return s.repo.IsMember(chatID, userID)
}
