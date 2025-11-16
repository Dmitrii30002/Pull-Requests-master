package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRepo_Create(t *testing.T) {
	t.Run("successful team creation", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &teamRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputTeam := &domain.Team{Name: "Avengers"}
		expectedTeam := &domain.Team{Name: "Avengers"}

		rows := sqlmock.NewRows([]string{"name"}).AddRow("Avengers")
		mock.ExpectQuery(regexp.QuoteMeta(`
            INSERT INTO teams (name)
            VALUES ($1)
            RETURNING name
        `)).WithArgs("Avengers").WillReturnRows(rows)

		result, err := repo.Create(inputTeam)

		assert.NoError(t, err)
		assert.Equal(t, expectedTeam, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("database error on team creation", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &teamRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		inputTeam := &domain.Team{Name: "Avengers"}

		expectedError := errors.New("unique constraint violation")
		mock.ExpectQuery(regexp.QuoteMeta(`
            INSERT INTO teams (name)
            VALUES ($1)
            RETURNING name
        `)).WithArgs("Avengers").WillReturnError(expectedError)

		result, err := repo.Create(inputTeam)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
		require.Len(t, hook.AllEntries(), 1)
	})
}

func TestTeamRepo_GetByName(t *testing.T) {
	t.Run("successfully get team with members", func(t *testing.T) {
		log, _ := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &teamRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		teamName := "Avengers"
		expectedTeam := &domain.Team{
			Name: "Avengers",
			Members: []*domain.Member{
				{ID: "user-1", Username: "tony_stark", IsActive: true},
				{ID: "user-2", Username: "steve_rogers", IsActive: true},
			},
		}

		rows := sqlmock.NewRows([]string{"id", "username", "is_active"}).
			AddRow("user-1", "tony_stark", true).
			AddRow("user-2", "steve_rogers", true)
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT u.id, u.username, u.is_active
            FROM teams t
            JOIN users u ON t.name = u.team_name
            WHERE t.name = $1
        `)).WithArgs("Avengers").WillReturnRows(rows)

		result, err := repo.GetByName(teamName)

		assert.NoError(t, err)
		assert.Equal(t, expectedTeam, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database query error", func(t *testing.T) {
		log, _ := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &teamRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		teamName := "Avengers"

		expectedError := errors.New("connection failed")
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT u.id, u.username, u.is_active
            FROM teams t
            JOIN users u ON t.name = u.team_name
            WHERE t.name = $1
        `)).WithArgs("Avengers").WillReturnError(expectedError)

		result, err := repo.GetByName(teamName)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to exec query")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTeamRepo_CheckExist(t *testing.T) {
	t.Run("team exists", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &teamRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		teamName := "Avengers"

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT EXISTS(SELECT 1 FROM teams 
            WHERE name = $1)
        `)).WithArgs("Avengers").WillReturnRows(rows)

		exists, err := repo.CheckExist(teamName)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})

	t.Run("team does not exist", func(t *testing.T) {
		log, hook := test.NewNullLogger()
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := &teamRepo{
			db:  db,
			log: &logger.Logger{log},
		}

		teamName := "NonExistentTeam"

		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
		mock.ExpectQuery(regexp.QuoteMeta(`
            SELECT EXISTS(SELECT 1 FROM teams 
            WHERE name = $1)
        `)).WithArgs("NonExistentTeam").WillReturnRows(rows)

		exists, err := repo.CheckExist(teamName)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, hook.AllEntries(), 0)
	})
}
