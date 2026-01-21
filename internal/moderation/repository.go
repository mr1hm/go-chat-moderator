package moderation

import (
	"github.com/google/uuid"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

type ModerationLogRepository interface {
	Create(log *ModerationLog) error
}
type sqliteModerationLogRepo struct{}

func NewModerationLogRepository() ModerationLogRepository {
	return &sqliteModerationLogRepo{}
}

func (r *sqliteModerationLogRepo) Create(log *ModerationLog) error {
	log.ID = uuid.New().String()

	// SQLite uses 0 and 1 for booleans
	flagged := 0
	if log.IsFlagged {
		flagged = 1
	}

	_, err := sqlite.DB.Exec(
		`INSERT INTO moderation_logs (id, message_id, toxicity_score, is_flagged) VALUES (?, ?, ?, ?)`,
		log.ID, log.MessageID, log.ToxicityScore, flagged,
	)

	return err
}
