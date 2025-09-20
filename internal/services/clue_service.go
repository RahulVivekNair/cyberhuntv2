package services

import (
	"context"
	"cyberhunt/internal/models"
	"database/sql"
	"errors"
)

type ClueService struct {
	db *sql.DB
}

func NewClueService(db *sql.DB) *ClueService {
	return &ClueService{db: db}
}

func (s *ClueService) GetClueByPathwayAndIndex(ctx context.Context, pathway string, index int) (*models.Clue, error) {
	var clue models.Clue

	err := s.db.QueryRowContext(ctx, `
        SELECT id, pathway, index_num, content, qrcode
        FROM clues
        WHERE pathway = $1 AND index_num = $2
    `, pathway, index).Scan(
		&clue.ID,
		&clue.Pathway,
		&clue.Index,
		&clue.Content,
		&clue.QRCode,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("clue not found")
	}
	if err != nil {
		return nil, errors.New("failed to fetch clue")
	}

	return &clue, nil
}

func (s *ClueService) ClearClues(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM clues")
	return err
}

func (s *ClueService) AddClue(ctx context.Context, pathway string, index int, content, qrCode string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO clues (pathway, index_num, content, qrcode)
		VALUES ($1, $2, $3, $4)
	`, pathway, index, content, qrCode)
	return err
}
