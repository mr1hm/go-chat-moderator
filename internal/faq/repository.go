package faq

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/mr1hm/go-chat-moderator/internal/shared/sqlite"
)

type FAQRepository interface {
	Create(faq *FAQ) error
	List() ([]*FAQ, error)
	FindByID(id string) (*FAQ, error)
	Delete(id string) error
}

var ErrFAQNotFound = errors.New("faq not found")

type sqliteFAQRepo struct{}

func NewFAQRepository() FAQRepository {
	return &sqliteFAQRepo{}
}

func (r *sqliteFAQRepo) Create(faq *FAQ) error {
	faq.ID = uuid.New().String()

	embeddingJSON, err := json.Marshal(faq.Embedding)
	if err != nil {
		return err
	}

	_, err = sqlite.DB.Exec(
		`INSERT INTO faqs (id, question, answer, embedding) VALUES (?, ?, ?, ?)`,
		faq.ID, faq.Question, faq.Answer, string(embeddingJSON),
	)
	return err
}

func (r *sqliteFAQRepo) List() ([]*FAQ, error) {
	rows, err := sqlite.DB.Query(
		`SELECT id, question, answer, embedding, created_at FROM faqs`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var faqs []*FAQ
	for rows.Next() {
		f := &FAQ{}
		var embeddingJSON string
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &embeddingJSON, &f.CreatedAt); err != nil {
			return nil, err
		}

		if embeddingJSON != "" {
			json.Unmarshal([]byte(embeddingJSON), &f.Embedding)
		}
		faqs = append(faqs, f)
	}

	return faqs, nil
}

func (r *sqliteFAQRepo) FindByID(id string) (*FAQ, error) {
	f := &FAQ{}
	var embeddingJSON string

	err := sqlite.DB.QueryRow(
		`SELECT id, question, answer, embedding, created_at FROM faqs WHERE id = ?`,
		id,
	).Scan(&f.ID, &f.Question, &f.Answer, &embeddingJSON, &f.CreatedAt)
	if err != nil {
		return nil, ErrFAQNotFound
	}

	if embeddingJSON != "" {
		json.Unmarshal([]byte(embeddingJSON), &f.Embedding)
	}
	return f, nil
}

func (r *sqliteFAQRepo) Delete(id string) error {
	result, err := sqlite.DB.Exec(`DELETE FROM faqs WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrFAQNotFound
	}

	return nil
}
