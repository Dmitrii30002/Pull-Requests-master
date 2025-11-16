package service

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/errors"
	"Pull-Requests-master/internal/repository"
)

type service struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	prRepo   repository.PullRequestRepository
}

func (s *service) CreateUser(user *domain.User) (*domain.User, error) {
	var newUser *domain.User
	exists, err := s.userRepo.CheckExist(user.ID)
	if err != nil {
		return nil, err
	}
	if !exists {
		newUser, err = s.userRepo.Create(user)
		if err != nil {
			return nil, err
		}
	} else {
		newUser, err = s.userRepo.Update(user)
		if err != nil {
			return nil, err
		}
	}

	return newUser, nil
}

func (s *service) SetUserActive(id string, status bool) (*domain.User, error) {
	var newUser *domain.User
	exists, err := s.userRepo.CheckExist(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.ErrNotFound
	}

	newUser, err = s.userRepo.SetUserActive(id, status)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (s *service) GetUserReview(id string) ([]*domain.PullRequestShort, error) {
	exists, err := s.userRepo.CheckExist(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		//r.log.Debugf("user with id: %s not found", id)
		return nil, errors.ErrNotFound
	}

	pullRequests, err := s.userRepo.GetReview(id)
	if err != nil {
		return nil, err
	}

	return pullRequests, nil
}
