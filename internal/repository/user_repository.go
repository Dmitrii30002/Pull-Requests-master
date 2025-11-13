package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"database/sql"
)

type UserRepository interface {
	SetUserActive(id string) (*domain.User, error)
	GetReview(id string) (*[]domain.PullRequestShort, error)
}

type UserRepo struct {
	db  *sql.DB
	log *logger.Logger
}

func (repo *UserRepo) SetUserActive(id string) (*domain.User, error) {

	return nil, nil
}

func (repo *UserRepo) GetReview(id string) (*[]domain.PullRequestShort, error) {
	return nil, nil
}
