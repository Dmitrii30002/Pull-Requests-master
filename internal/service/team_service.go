package service

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/errors"
)

func (s *Service) CreateTeam(team *domain.Team) (*domain.Team, error) {
	exists, err := s.teamRepo.CheckExist(team.Name)
	if err != nil {
		s.log.Errorf("failed to check exist of team: %v", err)
		return nil, err
	}

	if exists {
		s.log.Debugf("team with name: %s exist", team.Name)
		return nil, errors.ErrTeamExists
	}

	newTeam, err := s.teamRepo.Create(team)
	if err != nil {
		s.log.Errorf("failed to create team: %v", err)
		return nil, err
	}

	for i := 0; i < len(team.Members); i++ {
		user := domain.User{
			Member:   *team.Members[i],
			TeamName: team.Name,
		}
		newUser, err := s.CreateUser(&user)
		if err != nil {
			s.log.Errorf("failed to create user: %v", err)
			return nil, err
		}
		newTeam.Members = append(newTeam.Members, &newUser.Member)
	}

	return newTeam, nil
}

func (s *Service) GetTeamByName(teamName string) (*domain.Team, error) {
	exists, err := s.teamRepo.CheckExist(teamName)
	if err != nil {
		s.log.Errorf("failed to check exist of user: %v", err)
		return nil, err
	}

	if !exists {
		s.log.Debugf("team with name: %s dosn't exist", teamName)
		return nil, errors.ErrNotFound
	}

	newTeam, err := s.teamRepo.GetByName(teamName)
	if err != nil {
		s.log.Errorf("failed to get team by name: %v", err)
		return nil, err
	}

	return newTeam, nil
}
