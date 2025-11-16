package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"context"
	"database/sql"
)

type UserRepository interface {
	SetUserActive(id string, status bool) (*domain.User, error)
	GetReview(id string) ([]*domain.PullRequestShort, error)
	CheckExist(id string) (bool, error)
	Create(user *domain.User) (*domain.User, error)
	Update(user *domain.User) (*domain.User, error)
}

type userRepo struct {
	db  *sql.DB
	log *logger.Logger
}

func NewUserRepository(db *sql.DB, log *logger.Logger) UserRepository {
	return &userRepo{db: db, log: log}
}

func (r *userRepo) CheckExist(id string) (bool, error) {
	ctx := context.Background()
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM users 
		WHERE id = $1)
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		r.log.Debugf("failed to exec query: %v", err)
		return false, err
	}
	return exists, nil
}

func (r *userRepo) Create(user *domain.User) (*domain.User, error) {
	ctx := context.Background()
	var newUser domain.User
	query := `
		INSERT INTO users (id, username, is_active, team_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, is_active, team_name
	`
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.IsActive, user.TeamName,
	).Scan(&newUser.ID, &newUser.Username, &newUser.IsActive, &newUser.TeamName)
	if err != nil {
		r.log.Debugf("failed to exec query: %v", err)
		return nil, err
	}

	return &newUser, nil
}

func (r *userRepo) Update(user *domain.User) (*domain.User, error) {
	ctx := context.Background()
	var newUser domain.User
	query := `
			UPDATE users 
			SET
				username = $1,
				is_active = $2,
				team_name = $3
			WHERE id = $4
			RETURNING (id, username, is_active, team_name)
		`
	_, err := r.db.ExecContext(ctx, query,
		user.Username, user.IsActive, user.TeamName, user.ID,
	)
	if err != nil {
		r.log.Debugf("failed to insert user %v", err)
		return nil, err
	}

	return &newUser, nil
}

func (r *userRepo) SetUserActive(id string, status bool) (*domain.User, error) {
	ctx := context.Background()
	query := `
		UPDATE users
		SET
			is_active = $1
		WHERE id = $2
		RETURNING id, username, is_active, team_name
	`
	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id, status).Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		r.log.Debugf("failed set status: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetReview(id string) ([]*domain.PullRequestShort, error) {
	ctx := context.Background()
	query := `
		SELECT id, name, author_id, status
		FROM pull_requests pr
		JOIN pr_reviewrs pr_rev ON pr_rev.pr_id = pr.id
		WHERE pr_rev.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		r.log.Debugf("failed query: %v", err)
		return nil, err
	}
	defer rows.Close()

	pullRequests := []*domain.PullRequestShort{}
	for rows.Next() {
		var pr domain.PullRequestShort
		err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status)
		if err != nil {
			r.log.Debugf("failed scan: %v", err)
			return nil, err
		}
		pullRequests = append(pullRequests, &pr)
	}

	return pullRequests, nil
}
