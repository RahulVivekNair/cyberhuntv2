package handlers

import (
	"context"
	"cyberhunt/internal/services"
	"database/sql"
	"log"
)

type Handler struct {
	groupService   *services.GroupService
	gameService    *services.GameService
	clueService    *services.ClueService
	adminService   *services.AdminService
	LeaderboardHub *LeaderboardHub
	jwtSecret      string
}

func NewHandler(db *sql.DB, jwtSecret string) *Handler {
	h := &Handler{
		groupService:   services.NewGroupService(db),
		gameService:    services.NewGameService(db),
		clueService:    services.NewClueService(db),
		adminService:   services.NewAdminService(db),
		jwtSecret:      jwtSecret,
		LeaderboardHub: NewLeaderboardHub(),
	}

	// Prime the cache once at startup so new SSE clients see data immediately
	go func() {
		if err := h.BroadcastLeaderboard(context.Background()); err != nil {
			log.Printf("initial leaderboard broadcast failed: %v", err)
		}
	}()

	return h
}
