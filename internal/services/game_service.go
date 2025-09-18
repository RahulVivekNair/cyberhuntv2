package services

import (
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

func (s *GameService) StartGame() error {
	startTime := time.Now().UTC()
	_, err := s.db.Exec(`
		UPDATE game_settings
		SET game_started = TRUE, start_time = $1
	`, startTime)
	return err
}

func (s *GameService) EndGame() error {
	_, err := s.db.Exec(`
		UPDATE game_settings SET game_ended = TRUE
	`)
	return err
}

func (s *GameService) ClearState() error {
	_, err := s.db.Exec(`
		UPDATE game_settings
		SET game_started = FALSE, game_ended = FALSE, start_time = NULL
	`)
	return err
}

func (s *GameService) GetGameStatus() (*models.GameSettings, error) {
	var settings models.GameSettings
	var startTime sql.NullTime
	err := s.db.QueryRow(`
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

func (s *GameService) GetTotalClues() (int, error) {
	var totalClues int
	err := s.db.QueryRow(`SELECT total_clues FROM game_settings`).Scan(&totalClues)
	if err != nil {
		totalClues = 1
	}
	return totalClues, nil
}
