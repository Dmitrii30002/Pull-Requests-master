package domain

import "time"

type PullRequest struct {
	PullRequestShort
	AssignedReviewers []string  `json:"assigned_reviewers"`
	CreatedAt         time.Time `json:"createdAt"`
	MergedAt          time.Time `json:"mergedAt"`
}

type PullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

type Team struct {
	Name    string    `json:"team_name"`
	Members []*Member `json:"team_members"`
}
type User struct {
	Member
	TeamName string `json:"team_name"`
}

type Member struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
