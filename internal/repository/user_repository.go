package repository

import (
	"Pull-Requests-master/internal/domain"
	myErrors "Pull-Requests-master/internal/errors"
	"Pull-Requests-master/package/logger"
	"context"
	"database/sql"
	"fmt"
)

type UserRepository interface {
	Post(user *domain.User) (*domain.User, error)
	SetUserActive(id string) (*domain.User, error)
	GetReview(id string) ([]*domain.PullRequestShort, error)
}

type userRepo struct {
	db  *sql.DB
	log *logger.Logger
}

func (r *userRepo) Post(user *domain.User) (*domain.User, error) {
	ctx := context.Background()
	exists, err := checkUserExist(ctx, r.db, user.ID)
	if err != nil {
		r.log.Debugf("failed to check exist %v", err)
		return nil, err
	}
	if !exists {
		query := `
			INSERT INTO users (id, username, is_active, team_name)
			VALUES ($1, $2, $3, $4)
		`
		_, err = r.db.ExecContext(ctx, query,
			user.ID, user.Username, user.IsActive, user.TeamName,
		)
		if err != nil {
			r.log.Debug("failed to insert user %v", err)
			return nil, err
		}
	} else {
		query := `
			UPDATE users 
			SET
				username = $1,
				is_active = $2,
				team_name = $3
			WHERE id = $4
		`
		_, err = r.db.ExecContext(ctx, query,
			user.Username, user.IsActive, user.TeamName, user.ID,
		)
		if err != nil {
			r.log.Debug("failed to update user %v", err)
			return nil, err
		}
	}

	return user, nil
}

func (r *userRepo) SetUserActive(id string, status bool) (*domain.User, error) {
	ctx := context.Background()
	exists, err := checkUserExist(ctx, r.db, id)
	if err != nil {
		r.log.Debugf("failed to check exist %v", err)
		return nil, err
	}
	if !exists {
		r.log.Debugf("user with id: %s dosn't exist", id)
		err = myErrors.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: fmt.Sprintf("user with id: %s doesn't exists", id)}
		return nil, err
	}

	query := `
		UPDATE users
		SET
			is_active = $1
		WHERE id = $2
		RETURNING id, username, is_active, team_name
	`
	var user domain.User
	err = r.db.QueryRowContext(ctx, query, id, status).Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		r.log.Debugf("failed update: %v", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetReview(id string) ([]*domain.PullRequestShort, error) {
	ctx := context.Background()
	exists, err := checkUserExist(ctx, r.db, id)
	if err != nil {
		r.log.Debugf("failed to check exist %v", err)
		return nil, err
	}
	if !exists {
		r.log.Debugf("user with id: %s dosn't exist", id)
		err = myErrors.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: fmt.Sprintf("user with id: %s doesn't exists", id)}
		return nil, err
	}

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

func checkUserExist(ctx context.Context, db *sql.DB, id string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM users 
		WHERE id = $1)
	`
	err := db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return exists, err
	}
	return exists, nil
}
