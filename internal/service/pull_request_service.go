package service

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/errors"
)

func (s *service) Create(pr *domain.PullRequestShort) (*domain.PullRequest, error) {
	exists, err := s.prRepo.CheckPRExist(pr.ID)
	if err != nil {
		return nil, err
	}

	if exists {
		//r.log.Debugf("pull request with id: %s exist", pr.ID)
		return nil, errors.ErrPRExists
	}

	newPR, err := s.prRepo.Create(pr)
	if err != nil {
		return nil, err
	}

	return newPR, nil
}

func (s *service) Merge(id string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.CheckPRExist(id)
	if err != nil {
		return nil, err
	}

	if !exists {
		//r.log.Debugf("pull request with id: %s doesn't exist", id)
		return nil, errors.ErrNotFound
	}

	newPR, err := s.prRepo.Merge(id)
	if err != nil {
		return nil, err
	}

	return newPR, nil
}

func (s *service) Reassign(id string, oldRevID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if pr.Status == "MERGED" {
		return nil, errors.ErrPRMerged
	}

	newPR, err := s.prRepo.Reassign(id, oldRevID)
	if err != nil {
		return nil, err
	}

	return newPR, nil
}
