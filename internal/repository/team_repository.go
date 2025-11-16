package repository

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/package/logger"
	"context"
	"database/sql"
	"fmt"
)

type TeamRepository interface {
	Create(team *domain.Team) (*domain.Team, error)
	GetByName(teamName string) (*domain.Team, error)
	CheckExist(teamName string) (bool, error)
}

type teamRepo struct {
	db  *sql.DB
	log *logger.Logger
}

func NewTeamRepository(db *sql.DB, log *logger.Logger) TeamRepository {
	return &teamRepo{db: db, log: log}
}

func (r *teamRepo) Create(team *domain.Team) (*domain.Team, error) {
	ctx := context.Background()
	query := `
		INSERT INTO teams (name)
		VALUES ($1)
		RETURNING name
	`
	var newTeam domain.Team
	err := r.db.QueryRowContext(ctx, query, team.Name).Scan(&newTeam.Name)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return nil, err
	}

	return &newTeam, nil
}

func (r *teamRepo) GetByName(teamName string) (*domain.Team, error) {
	ctx := context.Background()
	query := `
		SELECT u.id, u.username, u.is_active
		FROM teams t
		JOIN users u ON t.name = u.team_name
		WHERE t.name = $1
	`

	team := &domain.Team{
		Name:    teamName,
		Members: []*domain.Member{},
	}
	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to exec query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member domain.Member
		err = rows.Scan(&member.ID, &member.Username, &member.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %v", err)
		}
		team.Members = append(team.Members, &member)
	}

	return team, nil
}

func (r *teamRepo) CheckExist(teamName string) (bool, error) {
	ctx := context.Background()
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM teams 
		WHERE name = $1)
	`
	err := r.db.QueryRowContext(ctx, query, teamName).Scan(&exists)
	if err != nil {
		r.log.Errorf("failed to exec query: %v", err)
		return exists, err
	}
	return exists, nil
}
