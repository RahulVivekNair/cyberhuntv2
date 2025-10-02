package services

import (
	"context"
	"cyberhunt/internal/models"
	"database/sql"
	"fmt"
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
		&clue.ID, &clue.Pathway, &clue.Index, &clue.Content, &clue.QRCode,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("clue not found for pathway=%s, index=%d", pathway, index)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clue (pathway=%s, index=%d): %w", pathway, index, err)
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

func (s *ClueService) GetAllClues(ctx context.Context) ([]*models.Clue, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, pathway, index_num, content, qrcode
		FROM clues
		ORDER BY pathway, index_num
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clues: %w", err)
	}
	defer rows.Close()

	var clues []*models.Clue
	for rows.Next() {
		var clue models.Clue
		err := rows.Scan(&clue.ID, &clue.Pathway, &clue.Index, &clue.Content, &clue.QRCode)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clue: %w", err)
		}
		clues = append(clues, &clue)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clues: %w", err)
	}

	return clues, nil
}

func (s *ClueService) UpdateClue(ctx context.Context, id int, pathway string, index int, content, qrCode string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE clues
		SET pathway = $2, index_num = $3, content = $4, qrcode = $5
		WHERE id = $1
	`, id, pathway, index, content, qrCode)
	if err != nil {
		return fmt.Errorf("failed to update clue with id %d: %w", id, err)
	}
	return nil
}

func (s *ClueService) DeleteClue(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM clues WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete clue with id %d: %w", id, err)
	}
	return nil
}

func (s *ClueService) GetClueByID(ctx context.Context, id int) (*models.Clue, error) {
	var clue models.Clue
	err := s.db.QueryRowContext(ctx, `
		SELECT id, pathway, index_num, content, qrcode
		FROM clues
		WHERE id = $1
	`, id).Scan(
		&clue.ID, &clue.Pathway, &clue.Index, &clue.Content, &clue.QRCode,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("clue not found with id=%d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clue (id=%d): %w", id, err)
	}

	return &clue, nil
}
