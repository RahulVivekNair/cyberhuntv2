package services

import (
	"cyberhunt/internal/models"
	"database/sql"
	"time"
)

type GroupService struct {
	db *sql.DB
}

func NewGroupService(db *sql.DB) *GroupService {
	return &GroupService{db: db}
}

func (s *GroupService) AddGroup(name, pathway, password string) error {
	_, err := s.db.Exec(`
		INSERT INTO groups (name, pathway, password)
		VALUES ($1, $2, $3)
	`, name, pathway, password)
	return err
}

func (s *GroupService) DeleteGroup(id int) error {
	_, err := s.db.Exec("DELETE FROM groups WHERE id = $1", id)
	return err
}

func (s *GroupService) GetGroupByID(id int) (*models.Group, error) {
	var group models.Group
	err := s.db.QueryRow(`
		SELECT id, name, pathway, current_clue_idx, completed, end_time
		FROM groups WHERE id = $1
	`, id).Scan(
		&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx,
		&group.Completed, &group.EndTime,
	)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *GroupService) GetGroupByNameAndPassword(name, password string) (*models.Group, error) {
	var group models.Group
	err := s.db.QueryRow(`
		SELECT id, name, pathway, current_clue_idx, completed, end_time, password
		FROM groups WHERE name = $1 AND password = $2
	`, name, password).Scan(
		&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx,
		&group.Completed, &group.EndTime, &group.Password,
	)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (s *GroupService) UpdateGroupProgress(id, newClueIdx int, completed bool, endTime *time.Time) error {
	_, err := s.db.Exec(`
		UPDATE groups
		SET current_clue_idx = $1, completed = $2, end_time = $3
		WHERE id = $4
	`, newClueIdx, completed, endTime, id)
	return err
}

func (s *GroupService) ResetGroups() error {
	_, err := s.db.Exec(`
		UPDATE groups
		SET current_clue_idx = 0, completed = FALSE, end_time = NULL
	`)
	return err
}

func (s *GroupService) GetGroupsForLeaderboard() ([]models.Group, error) {
	rows, err := s.db.Query(`
		SELECT id, name, pathway, current_clue_idx, completed, end_time
		FROM groups
		ORDER BY completed DESC, current_clue_idx DESC, end_time ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		var endTime sql.NullTime
		err := rows.Scan(
			&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx,
			&group.Completed, &endTime,
		)
		if err != nil {
			continue
		}
		if endTime.Valid {
			group.EndTime = &endTime.Time
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (s *GroupService) GetStats() (int, int, int, error) {
	var totalGroups, completedGroups int

	err := s.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&totalGroups)
	if err != nil {
		return 0, 0, 0, err
	}

	err = s.db.QueryRow("SELECT COUNT(*) FROM groups WHERE completed = TRUE").Scan(&completedGroups)
	if err != nil {
		return 0, 0, 0, err
	}

	inProgressGroups := totalGroups - completedGroups
	return totalGroups, completedGroups, inProgressGroups, nil
}
