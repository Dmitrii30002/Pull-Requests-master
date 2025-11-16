package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"context"
	"database/sql"
)

type PullRequestRepository interface {
	Create(pr *domain.PullRequestShort) (*domain.PullRequest, error)
	Merge(id string) (*domain.PullRequest, error)
	Reassign(id string, oldRevID string) (*domain.PullRequest, error)
	GetByID(id string) (*domain.PullRequest, error)
	GetReviewrs(id string) ([]string, error)
	RemoveReviewer(id string, revID string) error
	CheckPRExist(id string) (bool, error)
}

type pullRequestRepo struct {
	db  *sql.DB
	log *logger.Logger
}

func NewPullRequestRepository(db *sql.DB, log *logger.Logger) PullRequestRepository {
	return &pullRequestRepo{db: db, log: log}
}

func (r *pullRequestRepo) Create(pr *domain.PullRequestShort) (*domain.PullRequest, error) {
	ctx := context.Background()
	query := `
		INSERT INTO pull_requests (id, name, author_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, author_id, status, created_at
	`
	newPR := domain.PullRequest{AssignedReviewers: []string{}}
	err := r.db.QueryRowContext(ctx, query, pr.ID, pr.Name, pr.AuthorID, pr.Status).Scan(&newPR.ID, &newPR.Name, &newPR.AuthorID, &newPR.Status, &newPR.CreatedAt)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return nil, err
	}

	var cnt int
	query = `
		SELECT COUNT(*)
		FROM users
		WHERE team_name IN(
			SELECT team_name
			FROM users
			WHERE id = $1
		) AND is_active = TRUE
	`
	err = r.db.QueryRowContext(ctx, query, newPR.AuthorID).Scan(&cnt)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return nil, err
	}
	if cnt > 1 {
		places := 1
		if cnt >= 3 {
			places = 2
		}
		for true {
			id, err := r.getRandomUserID(ctx)
			if err != nil {
				r.log.Errorf("failed to get random user id: %v", err)
				return nil, err
			}
			query = `
				INSERT INTO pr_reviewrs (user_id, pr_id)
				VALUES ($1, $2)
			`
			if id != newPR.AuthorID && len(newPR.AssignedReviewers) == 0 {
				newPR.AssignedReviewers = append(newPR.AssignedReviewers, id)
				_, err = r.db.ExecContext(ctx, query, id, newPR.ID)
				if err != nil {
					r.log.Errorf("failed to exec query: %v", err)
					return nil, err
				}
			} else {
				if id != newPR.AuthorID && id != newPR.AssignedReviewers[0] {
					newPR.AssignedReviewers = append(newPR.AssignedReviewers, id)
					_, err = r.db.ExecContext(ctx, query, id, newPR.ID)
					if err != nil {
						r.log.Errorf("failed to exec query: %v", err)
						return nil, err
					}
				}
			}
			if len(newPR.AssignedReviewers) == places {
				break
			}
		}
	}

	return &newPR, nil
}

func (r *pullRequestRepo) Merge(id string) (*domain.PullRequest, error) {
	ctx := context.Background()
	query := `
		UPDATE pull_requests
		SET
			status = 'MERGED',
			merged_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, author_id, status, created_at, merged_at
	`
	newPR := domain.PullRequest{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&newPR.ID, &newPR.Name, &newPR.AuthorID, &newPR.Status, &newPR.CreatedAt, &newPR.MergedAt)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return nil, err
	}

	return &newPR, nil
}

func (r *pullRequestRepo) Reassign(id string, oldRevID string) (*domain.PullRequest, error) {
	ctx := context.Background()
	pr, err := r.GetByID(id)
	if err != nil {
		r.log.Errorf("failed to get pr by id: %v", err)
		return nil, err
	}

	usersID, err := r.GetReviewrs(id)
	if err != nil {
		r.log.Errorf("failed to get reviewers: %v", err)
		return nil, err
	}

	cnt, err := r.countReviewrs(ctx, pr)
	if err != nil {
		r.log.Errorf("failed to count reviewers: %v", err)
		return nil, err
	}

	if cnt-1-len(usersID) >= 1 {
		for true {
			newRevID, err := r.getRandomUserID(ctx)
			if err != nil {
				r.log.Errorf("failed to get random user id: %v", err)
				return nil, err
			}
			query := `
				INSERT INTO pr_reviewrs (user_id, pr_id)
				VALUES ($1, $2)
			`
			if newRevID != pr.AuthorID {
				flag := true
				for i := 0; i < len(usersID); i++ {
					if newRevID == usersID[i] {
						flag = false
						break
					}
				}
				if flag {
					_, err = r.db.ExecContext(ctx, query, newRevID, id)
					if err != nil {
						r.log.Errorf("failed to exec query: %v", err)
						return nil, err
					}
					err = r.RemoveReviewer(id, oldRevID)
					if err != nil {
						r.log.Errorf("failed to remove reviewer: %v", err)
						return nil, err
					}
					break
				}
			}
		}
	}

	usersID, err = r.GetReviewrs(id)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = usersID

	return pr, nil
}

func (r *pullRequestRepo) RemoveReviewer(id string, revID string) error {
	ctx := context.Background()
	query := `
		DELETE FROM pr_reviewrs 
		WHERE pr_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, id, revID)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return err
	}

	return nil
}

func (r *pullRequestRepo) GetByID(id string) (*domain.PullRequest, error) {
	ctx := context.Background()
	query := `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`
	newPR := domain.PullRequest{AssignedReviewers: []string{}}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&newPR.ID, &newPR.Name, &newPR.AuthorID, &newPR.Status, &newPR.CreatedAt, &newPR.MergedAt)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return nil, err
	}

	newPR.AssignedReviewers, err = r.GetReviewrs(id)
	if err != nil {
		r.log.Errorf("failed to get reviewers: %v", err)
		return nil, err
	}

	return &newPR, nil
}

func (r *pullRequestRepo) GetReviewrs(id string) ([]string, error) {
	ctx := context.Background()
	query := `
		SELECT user_id
		FROM pr_reviewrs
		WHERE pr_id = $1
	`
	var usersID []string
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			r.log.Errorf("failed to scan user_id: %v", err)
			return nil, err
		}
		usersID = append(usersID, userID)
	}

	return usersID, nil
}

func (r *pullRequestRepo) getRandomUserID(ctx context.Context) (string, error) {
	query := `
        SELECT id
        FROM users
		WHERE is_active = true 
        ORDER BY RANDOM() 
        LIMIT 1
    `

	var id string
	err := r.db.QueryRowContext(ctx, query).Scan(&id)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return "", err
	}

	return id, nil
}

func (r *pullRequestRepo) CheckPRExist(id string) (bool, error) {
	ctx := context.Background()
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM pull_requests 
		WHERE id = $1)
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		r.log.Errorf("failed to check exist %v", err)
		return exists, err
	}
	return exists, nil
}

func (r *pullRequestRepo) countReviewrs(ctx context.Context, rp *domain.PullRequest) (int, error) {
	var cnt int
	query := `
		SELECT COUNT(*)
		FROM users
		WHERE team_name IN(
			SELECT team_name
			FROM users
			WHERE id = $1
		) AND is_active = TRUE
	`
	err := r.db.QueryRowContext(ctx, query, rp.AuthorID).Scan(&cnt)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return 0, err
	}

	return cnt, nil
}
