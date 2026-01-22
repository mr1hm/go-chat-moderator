package chat

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrMessageNotFound = errors.New("message not found")
)

type RoomRepository interface {
	Create(room *Room) error
	FindByID(id string) (*Room, error)
	List() ([]*Room, error)
}

type sqliteRoomRepo struct{}

func NewRoomRepository() RoomRepository {
	return &sqliteRoomRepo{}
}

func (r *sqliteRoomRepo) Create(room *Room) error {
	room.ID = uuid.New().String()
	_, err := sqlite.DB.Exec(
		`INSERT INTO rooms (id, name, created_by) VALUES (?, ?, ?)`,
		room.ID, room.Name, room.CreatedBy,
	)

	return err
}

func (r *sqliteRoomRepo) FindByID(id string) (*Room, error) {
	room := &Room{}
	err := sqlite.DB.QueryRow(
		`SELECT id, name, created_by, created_at FROM rooms WHERE id = ?`, id,
	).Scan(&room.ID, &room.Name, &room.CreatedBy, &room.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoomNotFound
	}

	return room, err
}

func (r *sqliteRoomRepo) List() ([]*Room, error) {
	rows, err := sqlite.DB.Query(`SELECT id, name, created_by, created_at FROM rooms ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("error while querying rooms (list): %w", err)
	}
	defer rows.Close()

	var rooms []*Room
	for rows.Next() {
		room := &Room{}
		if err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.CreatedBy,
			&room.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error while scanning rooms: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

// Message Repository
type MessageRepository interface {
	Create(msg *Message) error
	FindByRoom(roomID string, limit int) ([]*Message, error)
	UpdateStatus(id, status string) error
}

type sqliteMessageRepo struct{}

func NewMessageRepository() MessageRepository {
	return &sqliteMessageRepo{}
}

func (r *sqliteMessageRepo) Create(msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	_, err := sqlite.DB.Exec(
		`INSERT INTO messages (id, room_id, user_id, content, moderation_status) VALUES (?, ?, ?, ?, ?)`,
		msg.ID, msg.RoomID, msg.UserID, msg.Content, "pending",
	)

	return err
}

func (r *sqliteMessageRepo) FindByRoom(roomID string, limit int) ([]*Message, error) {
	rows, err := sqlite.DB.Query(
		`SELECT id, room_id, user_id, content, moderation_status, created_at
		 FROM messages WHERE room_id = ? ORDER BY created_at DESC LIMIT ?`,
		roomID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("error while querying messages by room: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		if err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Content,
			&msg.ModerationStatus,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error while scanning messages: %w", err)
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

func (r *sqliteMessageRepo) UpdateStatus(id, status string) error {
	_, err := sqlite.DB.Exec(`UPDATE messages SET moderation_status = ? WHERE id = ?`, status, id)
	return err
}
