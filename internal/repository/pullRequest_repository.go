package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"database/sql"
)

type PullRequestRepository interface {
	Create(pr domain.PullRequestShort) (*domain.PullRequest, error)
	Merge(id string) (*domain.PullRequest, error)
	Reassign(id string, oldRevID string) (*domain.PullRequest, error)
}

type PullRequestRepo struct {
	db  *sql.DB
	log *logger.Logger
}
