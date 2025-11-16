package service

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/errors"
)

func (s *Service) CreateUser(user *domain.User) (*domain.User, error) {
	var newUser *domain.User
	exists, err := s.userRepo.CheckExist(user.ID)
	if err != nil {
		s.log.Errorf("failed to check exist of user: %v", err)
		return nil, err
	}
	if !exists {
		newUser, err = s.userRepo.Create(user)
		if err != nil {
			s.log.Errorf("failed to create user: %v", err)
			return nil, err
		}
	} else {
		newUser, err = s.userRepo.Update(user)
		if err != nil {
			s.log.Errorf("failed to update user: %v", err)
			return nil, err
		}
	}

	return newUser, nil
}

func (s *Service) SetUserActive(id string, status bool) (*domain.User, error) {
	var newUser *domain.User
	exists, err := s.userRepo.CheckExist(id)
	if err != nil {
		s.log.Errorf("failed to check exist of user: %v", err)
		return nil, err
	}
	if !exists {
		s.log.Debugf("user with id: %s not found", id)
		return nil, errors.ErrNotFound
	}

	newUser, err = s.userRepo.SetUserActive(id, status)
	if err != nil {
		s.log.Errorf("failed to set user active: %v", err)
		return nil, err
	}
	return newUser, nil
}

func (s *Service) GetUserReviews(id string) ([]*domain.PullRequestShort, error) {
	exists, err := s.userRepo.CheckExist(id)
	if err != nil {
		s.log.Errorf("failed to check exist of user: %v", err)
		return nil, err
	}
	if !exists {
		s.log.Debugf("user with id: %s not found", id)
		return nil, errors.ErrNotFound
	}

	pullRequests, err := s.userRepo.GetReview(id)
	if err != nil {
		s.log.Errorf("failed to get review: %v", err)
		return nil, err
	}

	return pullRequests, nil
}
