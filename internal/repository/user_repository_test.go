package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_CheckExist(t *testing.T) {
	t.Run("user exists", func(t *testing.T) {
		log, _ := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "123"

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`
			SELECT EXISTS\(SELECT 1 FROM users 
			WHERE id = \$1\)
		`).WithArgs(userID).WillReturnRows(rows)

		exists, err := repo.CheckExist(userID)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user does not exist", func(t *testing.T) {
		log, _ := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{db: db, log: &logger.Logger{log}}
		userID := "456"

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(`SELECT EXISTS`).WithArgs(userID).WillReturnRows(rows)

		exists, err := repo.CheckExist(userID)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		log, _ := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{db: db, log: &logger.Logger{log}}
		userID := "789"

		mock.ExpectQuery(`SELECT EXISTS`).WithArgs(userID).
			WillReturnError(errors.New("connection failed"))

		exists, err := repo.CheckExist(userID)

		assert.Error(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		log, _ := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{db: db, log: &logger.Logger{log}}
		userID := "999"

		rows := sqlmock.NewRows([]string{"exists"}).AddRow("not_a_boolean")
		mock.ExpectQuery(`SELECT EXISTS`).WithArgs(userID).WillReturnRows(rows)

		exists, err := repo.CheckExist(userID)

		assert.Error(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_Create(t *testing.T) {
	t.Run("scan error - type mismatch", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active", "team_name"}).
			AddRow(nil, "john_doe", true, "Avengers")

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(inputUser.ID, inputUser.Username, inputUser.IsActive, inputUser.TeamName).
			WillReturnRows(rows)

		result, err := repo.Create(inputUser)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())

		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
	})

	t.Run("scan error - wrong number of columns", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active"}).
			AddRow("user-123", "john_doe", true)

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(inputUser.ID, inputUser.Username, inputUser.IsActive, inputUser.TeamName).
			WillReturnRows(rows)

		result, err := repo.Create(inputUser)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())

		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
	})

	t.Run("empty user ID", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		// Test data
		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		expectedError := errors.New("null value in column \"id\" violates not-null constraint")
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(inputUser.ID, inputUser.Username, inputUser.IsActive, inputUser.TeamName).
			WillReturnError(expectedError)

		result, err := repo.Create(inputUser)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})

	t.Run("successful user creation", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		expectedUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active", "team_name"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.IsActive, expectedUser.TeamName)

		mock.ExpectQuery(`
			INSERT INTO users \(id, username, is_active, team_name\)
			VALUES \(\$1, \$2, \$3, \$4\)
			RETURNING id, username, is_active, team_name
		`).
			WithArgs(inputUser.ID, inputUser.Username, inputUser.IsActive, inputUser.TeamName).
			WillReturnRows(rows)

		result, err := repo.Create(inputUser)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("database error on insert", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		expectedError := errors.New("unique constraint violation")
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(inputUser.ID, inputUser.Username, inputUser.IsActive, inputUser.TeamName).
			WillReturnError(expectedError)

		result, err := repo.Create(inputUser)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		assert.NoError(t, mock.ExpectationsWereMet())

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
		hook.Reset()
	})

	t.Run("context timeout", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log}, // ваш мок логгера
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		expectedError := errors.New("context deadline exceeded")
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(inputUser.ID, inputUser.Username, inputUser.IsActive, inputUser.TeamName).
			WillReturnError(expectedError)

		result, err := repo.Create(inputUser)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "deadline exceeded")
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})
}

func TestUserRepo_Update(t *testing.T) {
	t.Run("successful user update", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe_updated",
				IsActive: false,
			},
			TeamName: "Justice League",
		}

		expectedUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe_updated",
				IsActive: false,
			},
			TeamName: "Justice League",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active", "team_name"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.IsActive, expectedUser.TeamName)

		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users 
            SET
                username = $1,
                is_active = $2,
                team_name = $3
            WHERE id = $4
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(inputUser.Username, inputUser.IsActive, inputUser.TeamName, inputUser.ID).
			WillReturnRows(rows)

		result, err := repo.Update(inputUser)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("user not found - no rows affected", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "non-existent-id",
				Username: "ghost_user",
				IsActive: true,
			},
			TeamName: "Ghost Team",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users 
            SET
                username = $1,
                is_active = $2,
                team_name = $3
            WHERE id = $4
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(inputUser.Username, inputUser.IsActive, inputUser.TeamName, inputUser.ID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.Update(inputUser)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
		hook.Reset()
	})

	t.Run("database connection error", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		expectedError := errors.New("connection refused")
		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users 
            SET
                username = $1,
                is_active = $2,
                team_name = $3
            WHERE id = $4
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(inputUser.Username, inputUser.IsActive, inputUser.TeamName, inputUser.ID).
			WillReturnError(expectedError)

		result, err := repo.Update(inputUser)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())

		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
	})

	t.Run("unique constraint violation - duplicate username", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "existing_username",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		expectedError := errors.New("duplicate key value violates unique constraint \"users_username_key\"")
		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users 
            SET
                username = $1,
                is_active = $2,
                team_name = $3
            WHERE id = $4
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(inputUser.Username, inputUser.IsActive, inputUser.TeamName, inputUser.ID).
			WillReturnError(expectedError)

		result, err := repo.Update(inputUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate key")
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})

	t.Run("scan error - wrong data types", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active", "team_name"}).
			AddRow("user-123", "john_doe", "not_a_boolean", "Avengers")

		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users 
            SET
                username = $1,
                is_active = $2,
                team_name = $3
            WHERE id = $4
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(inputUser.Username, inputUser.IsActive, inputUser.TeamName, inputUser.ID).
			WillReturnRows(rows)

		result, err := repo.Update(inputUser)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
	})
}

func TestUserRepo_SetUserActive(t *testing.T) {
	t.Run("successfully activate user", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-123"
		status := true

		expectedUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: true,
			},
			TeamName: "Avengers",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active", "team_name"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.IsActive, expectedUser.TeamName)

		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users
            SET
                is_active = $1
            WHERE id = $2
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(status, userID).
			WillReturnRows(rows)

		result, err := repo.SetUserActive(userID, status)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("successfully deactivate user", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-123"
		status := false

		expectedUser := &domain.User{
			Member: domain.Member{
				ID:       "user-123",
				Username: "john_doe",
				IsActive: false,
			},
			TeamName: "Avengers",
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active", "team_name"}).
			AddRow(expectedUser.ID, expectedUser.Username, expectedUser.IsActive, expectedUser.TeamName)

		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users
            SET
                is_active = $1
            WHERE id = $2
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(status, userID).
			WillReturnRows(rows)

		result, err := repo.SetUserActive(userID, status)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("user not found", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "non-existent-id"
		status := true

		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users
            SET
                is_active = $1
            WHERE id = $2
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(status, userID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.SetUserActive(userID, status)

		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
	})

	t.Run("database error", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-123"
		status := true

		expectedError := errors.New("connection failed")
		mock.ExpectQuery(regexp.QuoteMeta(`
            UPDATE users
            SET
                is_active = $1
            WHERE id = $2
            RETURNING id, username, is_active, team_name
        `)).
			WithArgs(status, userID).
			WillReturnError(expectedError)

		result, err := repo.SetUserActive(userID, status)

		// Assertions
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})
}

func TestUserRepo_GetReview(t *testing.T) {
	t.Run("successfully get review list", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-123"

		expectedPRs := []*domain.PullRequestShort{
			{
				ID:       "pr-1",
				Name:     "Feature A",
				AuthorID: "author-1",
				Status:   "open",
			},
			{
				ID:       "pr-2",
				Name:     "Bugfix B",
				AuthorID: "author-2",
				Status:   "closed",
			},
		}

		rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status"}).
			AddRow("pr-1", "Feature A", "author-1", "open").
			AddRow("pr-2", "Bugfix B", "author-2", "closed")

		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT id, name, author_id, status
            FROM pull_requests pr
            JOIN pr_reviewrs pr_rev ON pr_rev.pr_id = pr.id
            WHERE pr_rev.user_id = $1
        `)).
			WithArgs(userID).
			WillReturnRows(rows)

		result, err := repo.GetReview(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedPRs, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("empty review list", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-with-no-reviews"

		rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status"})

		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT id, name, author_id, status
            FROM pull_requests pr
            JOIN pr_reviewrs pr_rev ON pr_rev.pr_id = pr.id
            WHERE pr_rev.user_id = $1
        `)).
			WithArgs(userID).
			WillReturnRows(rows)

		result, err := repo.GetReview(userID)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Len(t, result, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("database query error", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-123"

		expectedError := errors.New("syntax error")
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT id, name, author_id, status
            FROM pull_requests pr
            JOIN pr_reviewrs pr_rev ON pr_rev.pr_id = pr.id
            WHERE pr_rev.user_id = $1
        `)).
			WithArgs(userID).
			WillReturnError(expectedError)

		result, err := repo.GetReview(userID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "failed to exec query")
	})

	t.Run("scan error during row processing", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-123"

		rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status"}).
			AddRow(nil, "Feature A", "author-1", "open").
			AddRow("pr-2", "Bugfix B", "author-2", "closed")

		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT id, name, author_id, status
            FROM pull_requests pr
            JOIN pr_reviewrs pr_rev ON pr_rev.pr_id = pr.id
            WHERE pr_rev.user_id = $1
        `)).
			WithArgs(userID).
			WillReturnRows(rows)

		result, err := repo.GetReview(userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "failed scan")
	})

	t.Run("rows close error", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &userRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		userID := "user-123"

		rows := sqlmock.NewRows([]string{"id", "name", "author_id", "status"}).
			AddRow("pr-1", "Feature A", "author-1", "open").
			AddRow("pr-2", "Bugfix B", "author-2", "closed").
			CloseError(errors.New("close error"))

		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT id, name, author_id, status
            FROM pull_requests pr
            JOIN pr_reviewrs pr_rev ON pr_rev.pr_id = pr.id
            WHERE pr_rev.user_id = $1
        `)).
			WithArgs(userID).
			WillReturnRows(rows)

		result, err := repo.GetReview(userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})
}
