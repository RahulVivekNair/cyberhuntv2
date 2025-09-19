package handlers

import (
	"cyberhunt/internal/services"
	"database/sql"
)

type Handler struct {
	groupService *services.GroupService
	gameService  *services.GameService
	clueService  *services.ClueService
	adminService *services.AdminService
	jwtSecret    string
}

func NewHandler(db *sql.DB, jwtSecret string) *Handler {
	return &Handler{
		groupService: services.NewGroupService(db),
		gameService:  services.NewGameService(db),
		clueService:  services.NewClueService(db),
		adminService: services.NewAdminService(db),
		jwtSecret:    jwtSecret,
	}
}
