package services

import (
	"context"
	"cyberhunt/internal/models"
	"database/sql"
	"time"
)

type GameService struct {
	db *sql.DB
}

func NewGameService(db *sql.DB) *GameService {
	return &GameService{db: db}
}

func (s *GameService) StartGame(ctx context.Context) error {
	// Try to start the game only if it hasn't already started
	res, err := s.db.ExecContext(ctx, `
		UPDATE game_settings
		SET game_started = TRUE, start_time = $1
		WHERE id = 1 AND game_started = FALSE
	`, time.Now().UTC())
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrGameAlreadyStarted
	}

	return nil
}

func (s *GameService) EndGame(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var gameStarted, gameEnded bool
	err = tx.QueryRowContext(ctx, `
        SELECT game_started, game_ended 
        FROM game_settings 
        WHERE id = 1 
        FOR UPDATE
    `).Scan(&gameStarted, &gameEnded)

	if err != nil {
		return err
	}

	if !gameStarted {
		return ErrGameNotStarted
	}
	if gameEnded {
		return ErrGameAlreadyEnded
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE game_settings SET game_ended = TRUE WHERE id = 1
    `)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// In GameService
func (s *GameService) ClearAllState(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Take exclusive advisory lock for reset
	_, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(12345)`)
	if err != nil {
		return err
	}

	// Put game into locked state (not started, not ended)
	_, err = tx.ExecContext(ctx, `
		UPDATE game_settings
		SET game_started = FALSE,
		    game_ended = FALSE,
		    start_time = NULL
	`)
	if err != nil {
		return err
	}

	// Reset all groups
	_, err = tx.ExecContext(ctx, `
		UPDATE groups
		SET current_clue_idx = 0,
		    completed = FALSE,
		    end_time = NULL
	`)
	if err != nil {
		return err
	}

	// Commit (advisory lock auto-released)
	return tx.Commit()
}

func (s *GameService) GetGameStatus(ctx context.Context) (*models.GameSettings, error) {
	var settings models.GameSettings
	var startTime sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT id, total_clues, start_time, game_started, game_ended
		FROM game_settings
	`).Scan(&settings.ID, &settings.TotalClues, &startTime, &settings.GameStarted, &settings.GameEnded)
	if err != nil {
		return nil, err
	}
	if startTime.Valid {
		settings.StartTime = &startTime.Time
	}
	return &settings, nil
}

func (s *GameService) GetTotalClues(ctx context.Context) (int, error) {
	var totalClues int
	err := s.db.QueryRowContext(ctx, `SELECT total_clues FROM game_settings`).Scan(&totalClues)
	if err != nil {
		// default to 1 if no value exists
		totalClues = 1
	}
	return totalClues, nil
}
