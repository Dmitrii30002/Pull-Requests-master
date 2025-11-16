package handlers

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) AddTeam(c echo.Context) error {
	var team domain.Team
	err := c.Bind(&team)
	if err != nil {
		h.log.Debugf("failed to pars json: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid JSON",
			},
		})
	}

	newTeam, err := h.s.CreateTeam(&team)
	if err != nil {
		switch err {
		case errors.ErrTeamExists:
			h.log.Debugf("team with name: %s already exist", team.Name)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": errors.ErrTeamExists,
			})
		default:
			h.log.Debugf("failed to create team: %v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusCreated, newTeam)
}

func (h *Handler) GetTeam(c echo.Context) error {
	teamName := c.QueryParam("team_name")

	team, err := h.s.GetTeamByName(teamName)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			h.log.Debugf("team with name: %s doesn't found", team.Name)
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": errors.ErrNotFound,
			})
		default:
			h.log.Debugf("failed to get team: %v", err)
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	return c.JSON(http.StatusOK, team)
}
