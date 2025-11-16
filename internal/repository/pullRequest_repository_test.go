package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestRepo_Create(t *testing.T) {
	t.Run("successful PR creation without reviewers", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputPR := &domain.PullRequestShort{
			ID:       "pr-1",
			Name:     "Feature A",
			AuthorID: "author-1",
			Status:   "open",
		}

		rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at"}).
			AddRow("pr-1", "Feature A", "author-1", "open", time.Now())
		mock.ExpectQuery(regexp.QuoteMeta(`
            INSERT INTO pull_requests (id, name, author_id, status)
            VALUES ($1, $2, $3, $4)
            RETURNING id, name, author_id, status, created_at
        `)).WithArgs("pr-1", "Feature A", "author-1", "open").WillReturnRows(rows)

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT COUNT(*)
            FROM users
            WHERE team_name IN(
                SELECT team_name
                FROM users
                WHERE id = $1
            ) AND is_active = TRUE
        `)).WithArgs("author-1").WillReturnRows(countRows)

		result, err := repo.Create(inputPR)

		assert.NoError(t, err)
		assert.Equal(t, "pr-1", result.ID)
		assert.Empty(t, result.AssignedReviewers)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("database error on PR creation", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputPR := &domain.PullRequestShort{
			ID:       "pr-1",
			Name:     "Feature A",
			AuthorID: "author-1",
			Status:   "open",
		}

		expectedError := errors.New("unique constraint violation")
		mock.ExpectQuery(regexp.QuoteMeta(`
            INSERT INTO pull_requests (id, name, author_id, status)
            VALUES ($1, $2, $3, $4)
            RETURNING id, name, author_id, status, created_at
        `)).WithArgs("pr-1", "Feature A", "author-1", "open").WillReturnError(expectedError)

		result, err := repo.Create(inputPR)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})
}

func TestPullRequestRepo_Merge(t *testing.T) {
	t.Run("successful PR merge", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "pr-1"

		rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at", "merged_at"}).
			AddRow("pr-1", "Feature A", "author-1", "MERGED", time.Now(), time.Now())
		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE pull_requests
            SET
                status = 'MERGED',
                merged_at = CURRENT_TIMESTAMP
            WHERE id = $1
            RETURNING id, name, author_id, status, created_at, merged_at
        `)).WithArgs("pr-1").WillReturnRows(rows)

		result, err := repo.Merge(prID)

		assert.NoError(t, err)
		assert.Equal(t, "MERGED", result.Status)
		assert.NotNil(t, result.MergedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("PR not found for merge", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "non-existent-pr"

		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE pull_requests
            SET
                status = 'MERGED',
                merged_at = CURRENT_TIMESTAMP
            WHERE id = $1
            RETURNING id, name, author_id, status, created_at, merged_at
        `)).WithArgs("non-existent-pr").WillReturnError(sql.ErrNoRows)

		result, err := repo.Merge(prID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})
}

func TestPullRequestRepo_GetByID(t *testing.T) {
	t.Run("successfully get PR by ID", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "pr-1"

		rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at", "merged_at"}).
			AddRow("pr-1", "Feature A", "author-1", "open", time.Now(), nil)
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT id, name, author_id, status, created_at, merged_at
            FROM pull_requests
            WHERE id = $1
        `)).WithArgs("pr-1").WillReturnRows(rows)

		reviewerRows := sqlmock.NewRows([]string{"user_id"}).
			AddRow("reviewer-1").
			AddRow("reviewer-2")
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT user_id
            FROM pr_reviewrs
            WHERE pr_id = $1
        `)).WithArgs("pr-1").WillReturnRows(reviewerRows)

		result, err := repo.GetByID(prID)

		assert.NoError(t, err)
		assert.Equal(t, "pr-1", result.ID)
		assert.Len(t, result.AssignedReviewers, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("PR not found", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "non-existent-pr"

		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT id, name, author_id, status, created_at, merged_at
            FROM pull_requests
            WHERE id = $1
        `)).WithArgs("non-existent-pr").WillReturnError(sql.ErrNoRows)

		result, err := repo.GetByID(prID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})
}

func TestPullRequestRepo_RemoveReviewer(t *testing.T) {
	t.Run("successfully remove reviewer", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "pr-1"
		reviewerID := "reviewer-1"

		mock.ExpectExec(regexp.QuoteMeta(`
            DELETE FROM pr_reviewrs 
            WHERE pr_id = $1 AND user_id = $2
        `)).WithArgs("pr-1", "reviewer-1").WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.RemoveReviewer(prID, reviewerID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("database error on remove reviewer", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "pr-1"
		reviewerID := "reviewer-1"

		expectedError := errors.New("connection failed")
		mock.ExpectExec(regexp.QuoteMeta(`
            DELETE FROM pr_reviewrs 
            WHERE pr_id = $1 AND user_id = $2
        `)).WithArgs("pr-1", "reviewer-1").WillReturnError(expectedError)

		err = repo.RemoveReviewer(prID, reviewerID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})
}

func TestPullRequestRepo_GetReviewrs(t *testing.T) {
	t.Run("successfully get reviewers", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "pr-1"

		rows := sqlmock.NewRows([]string{"user_id"}).
			AddRow("reviewer-1").
			AddRow("reviewer-2")
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT user_id
            FROM pr_reviewrs
            WHERE pr_id = $1
        `)).WithArgs("pr-1").WillReturnRows(rows)

		result, err := repo.GetReviewrs(prID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, []string{"reviewer-1", "reviewer-2"}, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("no reviewers found", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &pullRequestRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		prID := "pr-1"

		rows := sqlmock.NewRows([]string{"user_id"})
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT user_id
            FROM pr_reviewrs
            WHERE pr_id = $1
        `)).WithArgs("pr-1").WillReturnRows(rows)

		result, err := repo.GetReviewrs(prID)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Len(t, result, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})
}
