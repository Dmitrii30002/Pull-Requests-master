package handlers

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/errors"
	"Pull-Requests-master/internal/service"
	"Pull-Requests-master/package/logger"
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	s   *service.Service
	log *logger.Logger
}

func NewHandler(db *sql.DB, log *logger.Logger) *Handler {
	return &Handler{
		s:   service.NewService(db, log),
		log: log,
	}
}

func (h *Handler) CreatePR(c echo.Context) error {
	var pr domain.PullRequestShort
	err := c.Bind(&pr)
	if err != nil {
		h.log.Debugf("failed to pars json: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid JSON",
			},
		})
	}

	if pr.ID == "" || pr.Status == "" {
		h.log.Debug("invalid data")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "invalid data",
			},
		})
	}

	newPR, err := h.s.CreatePR(&pr)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			h.log.Debugf(err.Error())
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": errors.ErrNotFound,
			})
		case errors.ErrPRExists:
			h.log.Debugf("PR with id: %s exist", pr.ID)
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error": errors.ErrPRExists,
			})
		default:
			h.log.Debugf("failed to create PR: %v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"pr": newPR,
	})
}

func (h *Handler) MergePR(c echo.Context) error {
	var req struct {
		PRID string `json:"pull_request_id"`
	}
	err := c.Bind(&req)
	if err != nil {
		h.log.Debugf("failed to pars json: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid JSON",
			},
		})
	}

	if req.PRID == "" {
		h.log.Debug("invalid data")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "invalid data",
			},
		})
	}

	pr, err := h.s.MergePR(req.PRID)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			h.log.Debugf(err.Error())
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": errors.ErrNotFound,
			})
		default:
			h.log.Debugf("failed to merge PR: %v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"pr": pr,
	})
}

func (h *Handler) ReassignReviewersPR(c echo.Context) error {
	var req struct {
		PRID   string `json:"pull_request_id"`
		OldRev string `json:"old_reviewer_id"`
	}
	err := c.Bind(&req)
	if err != nil {
		h.log.Debugf("failed to pars json: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid JSON",
			},
		})
	}

	if req.PRID == "" || req.OldRev == "" {
		h.log.Debug("invalid data")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "invalid data",
			},
		})
	}

	pr, err := h.s.ReassignReviewersPR(req.PRID, req.OldRev)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			h.log.Debugf(err.Error())
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": errors.ErrNotFound,
			})
		case errors.ErrPRMerged:
			h.log.Debugf("PR with id: %s merged", req.PRID)
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error": errors.ErrPRMerged,
			})
		default:
			h.log.Debugf("failed to create PR: %v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": req.OldRev,
	})
}
