package services

import (
	"cyberhunt/internal/models"
	"database/sql"
)

type AdminService struct {
	db *sql.DB
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) GetAdminByNameAndPassword(name, password string) (*models.Admin, error) {
	var admin models.Admin
	err := s.db.QueryRow(`
		SELECT id, name, password FROM admins WHERE name = $1 AND password = $2
	`, name, password).Scan(&admin.ID, &admin.Name, &admin.Password)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}
