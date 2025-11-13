package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"database/sql"
)

type TeamRepository interface {
	Add(team *domain.Team) (*domain.Team, error)
	GetByName(teamName string) (*domain.Team, error)
}

type Team struct {
	db  *sql.DB
	log *logger.Logger
}
