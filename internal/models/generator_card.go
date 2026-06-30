package models

import (
	"database/sql"
	"fmt"
	"time"
)

type GeneratorCard struct {
	ID           int
	RequestID    int
	Filename     string
	Size         int64
	Content      []byte
	UploadedBy   int
	UploaderName string
	CreatedAt    time.Time
}

func (c *GeneratorCard) ContentString() string {
	return string(c.Content)
}

type GeneratorCardStore struct {
	db *sql.DB
	dbHelper
}

func NewGeneratorCardStore(db *sql.DB, driver string) *GeneratorCardStore {
	return &GeneratorCardStore{db: db, dbHelper: newHelper(driver)}
}

func (s *GeneratorCardStore) Add(card *GeneratorCard) (int64, error) {
	var uploadedBy interface{}
	if card.UploadedBy != 0 {
		uploadedBy = card.UploadedBy
	}
	var id int64
	err := s.db.QueryRow(s.rebind(`
		INSERT INTO generator_cards (request_id, filename, size, content, uploaded_by)
		VALUES (?, ?, ?, ?, ?) RETURNING id`),
		card.RequestID, card.Filename, card.Size, card.Content, uploadedBy,
	).Scan(&id)
	return id, err
}

func (s *GeneratorCardStore) GetByRequestID(requestID int) ([]*GeneratorCard, error) {
	rows, err := s.db.Query(s.rebind(`
		SELECT gc.id, gc.request_id, gc.filename, gc.size, COALESCE(gc.uploaded_by, 0),
		       COALESCE(u.display_name, ''), gc.created_at
		FROM generator_cards gc
		LEFT JOIN users u ON u.id = gc.uploaded_by
		WHERE gc.request_id = ?
		ORDER BY gc.created_at ASC`), requestID)
	if err != nil {
		return nil, fmt.Errorf("query generator cards: %w", err)
	}
	defer rows.Close()

	var cards []*GeneratorCard
	for rows.Next() {
		var c GeneratorCard
		if err := rows.Scan(&c.ID, &c.RequestID, &c.Filename, &c.Size, &c.UploadedBy, &c.UploaderName, timeVal{&c.CreatedAt}); err != nil {
			return nil, err
		}
		cards = append(cards, &c)
	}
	return cards, rows.Err()
}

func (s *GeneratorCardStore) GetByID(id int) (*GeneratorCard, error) {
	var c GeneratorCard
	err := s.db.QueryRow(s.rebind(`
		SELECT id, request_id, filename, size, content, COALESCE(uploaded_by, 0), created_at
		FROM generator_cards WHERE id = ?`), id).
		Scan(&c.ID, &c.RequestID, &c.Filename, &c.Size, &c.Content, &c.UploadedBy, timeVal{&c.CreatedAt})
	return &c, err
}

func (s *GeneratorCardStore) Delete(id int) error {
	_, err := s.db.Exec(s.rebind("DELETE FROM generator_cards WHERE id = ?"), id)
	return err
}
