package errors

import "fmt"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrTeamExists = APIError{
		Code:    "TEAM_EXISTS",
		Message: "team_name already exists",
	}

	ErrPRExists = APIError{
		Code:    "PR_EXISTS",
		Message: "PR id already exists",
	}

	ErrPRMerged = APIError{
		Code:    "PR_MERGED",
		Message: "cannot reassign on merged PR",
	}

	ErrNotFound = APIError{
		Code:    "NOT_FOUND",
		Message: "resource not found",
	}
)
