package services

import (
	"context"
	"cyberhunt/internal/models"
	"database/sql"
)

type AdminService struct {
	db *sql.DB
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) GetAdminByNameAndPassword(ctx context.Context, name, password string) (*models.Admin, error) {
	var admin models.Admin
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM admins WHERE name = $1 AND password = $2
	`, name, password).Scan(&admin.ID)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *AdminService) UpdateTotalClues(ctx context.Context, totalClues int) error {
	// Update existing settings
	result, err := s.db.ExecContext(ctx, `
		UPDATE game_settings SET total_clues = $1 WHERE id = 1
	`, totalClues)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows affected, insert new settings
	if rowsAffected == 0 {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO game_settings (id, total_clues, game_started, game_ended) VALUES (1, $1, false, false)
		`, totalClues)
		if err != nil {
			return err
		}
	}

	return nil
}
