package service

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/errors"
	"Pull-Requests-master/internal/repository"
	"Pull-Requests-master/package/logger"
	"database/sql"
)

type Service struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	prRepo   repository.PullRequestRepository
	log      *logger.Logger
}

func NewService(db *sql.DB, logger *logger.Logger) *Service {
	return &Service{
		userRepo: repository.NewUserRepository(db, logger),
		teamRepo: repository.NewTeamRepository(db, logger),
		prRepo:   repository.NewPullRequestRepository(db, logger),
		log:      logger,
	}
}

func (s *Service) CreatePR(pr *domain.PullRequestShort) (*domain.PullRequest, error) {
	exists, err := s.prRepo.CheckPRExist(pr.ID)
	if err != nil {
		s.log.Errorf("failed to check exist of pr: %v", err)
		return nil, err
	}

	if exists {
		s.log.Debugf("pull request with id: %s exist", pr.ID)
		return nil, errors.ErrPRExists
	}

	pr.Status = "OPEN"
	newPR, err := s.prRepo.Create(pr)
	if err != nil {
		s.log.Errorf("failed to create pr: %v", err)
		return nil, err
	}

	return newPR, nil
}

func (s *Service) MergePR(id string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.CheckPRExist(id)
	if err != nil {
		s.log.Errorf("failed to check exist of pr: %v", err)
		return nil, err
	}

	if !exists {
		s.log.Debugf("pull request with id: %s doesn't exist", id)
		return nil, errors.ErrNotFound
	}

	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		s.log.Errorf("failed to get pr by id: %v", err)
		return nil, err
	}

	if pr.Status == "MERGED" {
		s.log.Debugf("pr with id: %s already merged", id)
		return pr, nil
	}

	newPR, err := s.prRepo.Merge(id)
	if err != nil {
		s.log.Errorf("failed to merge pr: %v", err)
		return nil, err
	}

	return newPR, nil
}

func (s *Service) ReassignReviewersPR(id string, oldRevID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		s.log.Errorf("failed to get pr by id: %v", err)
		return nil, err
	}

	if pr.Status == "MERGED" {
		s.log.Debugf("pr with id: %s merged", id)
		return nil, errors.ErrPRMerged
	}

	newPR, err := s.prRepo.Reassign(id, oldRevID)
	if err != nil {
		s.log.Errorf("failed to Reassign reviewer: %v", err)
		return nil, err
	}

	return newPR, nil
}
