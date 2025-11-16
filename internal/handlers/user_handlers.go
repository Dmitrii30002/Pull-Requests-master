package handlers

import (
	"Pull-Requests-master/internal/errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) SetUserActive(c echo.Context) error {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
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

	if req.UserID == "" {
		h.log.Debug("invalid data")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "invalid data",
			},
		})
	}

	user, err := h.s.SetUserActive(req.UserID, req.IsActive)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			h.log.Debugf("user with id: %s nott found", req.UserID)
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": errors.ErrNotFound,
			})
		default:
			h.log.Debugf("failed to set user active: %v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GetUserReview(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		h.log.Debugf("not correct user id")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "not correct user id",
			},
		})
	}

	reviews, err := h.s.GetUserReviews(userID)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			h.log.Debugf("user with id: %s nott found", userID)
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": errors.ErrNotFound,
			})
		default:
			h.log.Debugf("failed to get user reviews: %v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": reviews,
	})
}
