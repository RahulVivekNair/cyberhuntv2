package services

import (
	"cyberhunt/internal/models"
	"database/sql"
)

type ClueService struct {
	db *sql.DB
}

func NewClueService(db *sql.DB) *ClueService {
	return &ClueService{db: db}
}

func (s *ClueService) GetClueByPathwayAndIndex(pathway string, index int) (*models.Clue, error) {
	var clue models.Clue
	err := s.db.QueryRow(`
		SELECT id, pathway, index_num, content, qrcode
		FROM clues WHERE pathway = $1 AND index_num = $2
	`, pathway, index).Scan(&clue.ID, &clue.Pathway, &clue.Index, &clue.Content, &clue.QRCode)
	if err != nil {
		return nil, err
	}
	return &clue, nil
}

func (s *ClueService) ClearClues() error {
	_, err := s.db.Exec("DELETE FROM clues")
	return err
}

func (s *ClueService) AddClue(pathway string, index int, content, qrCode string) error {
	_, err := s.db.Exec(`
		INSERT INTO clues (pathway, index_num, content, qrcode)
		VALUES ($1, $2, $3, $4)
	`, pathway, index, content, qrCode)
	return err
}
