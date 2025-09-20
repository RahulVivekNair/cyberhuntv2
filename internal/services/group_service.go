package services

import (
	"context"
	"crypto/subtle"
	"cyberhunt/internal/models"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type GroupService struct {
	db *sql.DB
}

func NewGroupService(db *sql.DB) *GroupService {
	return &GroupService{db: db}
}

func (s *GroupService) GetGroupByNameAndPassword(ctx context.Context, name, password string) (*models.Group, error) {
	var group models.Group
	err := s.db.QueryRowContext(ctx, `
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

func (s *GroupService) AddGroup(ctx context.Context, name, pathway, password string) error {
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO groups (name, pathway, password)
        VALUES ($1, $2, $3)
    `, name, pathway, password)

	if err != nil {
		// Check for Postgres unique violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrGroupExists // <-- return the custom error
		}
		return err
	}

	return nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM groups WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *GroupService) GetGroupByID(ctx context.Context, id int) (*models.Group, error) {
	var group models.Group
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, pathway, current_clue_idx, completed, end_time
		FROM groups
		WHERE id = $1
	`, id).Scan(
		&group.ID, &group.Name, &group.Pathway, &group.CurrentClueIdx,
		&group.Completed, &group.EndTime,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group %d not found", id) // clean error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group %d: %w", id, err)
	}
	return &group, nil
}

func (s *GroupService) UpdateGroupProgress(ctx context.Context, id, expectedClueIdx int, totalClues int) error {
	// Move to the *next* clue only if the group is still at expectedClueIdx
	// This prevents races / duplicate scans.
	query := `
		UPDATE groups
		SET 
			current_clue_idx = current_clue_idx + 1,
			completed = (current_clue_idx + 1) >= $2,
			end_time = CASE 
				WHEN (current_clue_idx + 1) >= $2 AND end_time IS NULL 
				THEN NOW() AT TIME ZONE 'UTC' 
				ELSE end_time 
			END
		WHERE id = $1 AND current_clue_idx = $3 AND completed = FALSE
	`
	res, err := s.db.ExecContext(ctx, query, id, totalClues, expectedClueIdx)
	if err != nil {
		return fmt.Errorf("failed to update progress for group %d: %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check progress update for group %d: %w", id, err)
	}
	if rows == 0 {
		// Either wrong clue index, already completed, or race condition lost
		return errors.New("progress not updated (stale or already completed)")
	}

	return nil
}

func (s *GroupService) ResetGroups(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE groups
		SET current_clue_idx = 0, completed = FALSE, end_time = NULL
	`)
	return err
}

func (s *GroupService) GetGroupsForLeaderboard(ctx context.Context) ([]models.Group, error) {
	rows, err := s.db.QueryContext(ctx, `
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

func (s *GroupService) GetLeaderboardData(ctx context.Context) (int, []models.Group, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	var totalClues int
	err = tx.QueryRowContext(ctx, `
        SELECT total_clues
        FROM game_settings
        WHERE id = 1
    `).Scan(&totalClues)
	if err == sql.ErrNoRows {
		totalClues = 1
	} else if err != nil {
		return 0, nil, err
	}

	rows, err := tx.QueryContext(ctx, `
        SELECT id, name, pathway, current_clue_idx, completed, end_time
        FROM groups
        ORDER BY completed DESC, current_clue_idx DESC, end_time ASC, id ASC
    `)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		var endTime sql.NullTime
		if err := rows.Scan(
			&g.ID, &g.Name, &g.Pathway, &g.CurrentClueIdx,
			&g.Completed, &endTime,
		); err != nil {
			return 0, nil, err
		}
		if endTime.Valid {
			g.EndTime = &endTime.Time
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}

	return totalClues, groups, nil
}

// In GroupService or a new "GameService"
func (s *GroupService) ScanAndUpdateProgress(
	ctx context.Context,
	groupID int,
	scannedCode string,
	totalClues int,
) (*models.Group, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // rollback if not committed

	// 1. Lock the group row
	var g models.Group
	err = tx.QueryRowContext(ctx, `
        SELECT id, name, pathway, current_clue_idx, completed, end_time
        FROM groups
        WHERE id = $1
        FOR UPDATE
    `, groupID).Scan(&g.ID, &g.Name, &g.Pathway, &g.CurrentClueIdx, &g.Completed, &g.EndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group %d not found", groupID)
		}
		return nil, fmt.Errorf("query group: %w", err)
	}
	if g.Completed {
		return &g, fmt.Errorf("group already completed")
	}

	// 2. Load expected clue
	var expectedCode string
	err = tx.QueryRowContext(ctx, `
        SELECT qrcode FROM clues WHERE pathway = $1 AND index_num = $2
    `, g.Pathway, g.CurrentClueIdx).Scan(&expectedCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("clue not found for pathway=%s index=%d", g.Pathway, g.CurrentClueIdx)
		}
		return nil, fmt.Errorf("query clue: %w", err)
	}

	// 3. Validate scanned QR
	if subtle.ConstantTimeCompare([]byte(scannedCode), []byte(expectedCode)) != 1 {
		return nil, fmt.Errorf("invalid QR code")
	}

	// 4. Update progress atomically and fetch new state
	err = tx.QueryRowContext(ctx, `
        UPDATE groups
        SET current_clue_idx = current_clue_idx + 1,
            completed = (current_clue_idx + 1) >= $2,
            end_time = CASE
                WHEN (current_clue_idx + 1) >= $2 AND end_time IS NULL
                THEN NOW() AT TIME ZONE 'UTC'
                ELSE end_time
            END
        WHERE id = $1
        RETURNING id, name, pathway, current_clue_idx, completed, end_time
    `, g.ID, totalClues).Scan(
		&g.ID, &g.Name, &g.Pathway, &g.CurrentClueIdx, &g.Completed, &g.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("update group progress: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &g, nil
}
