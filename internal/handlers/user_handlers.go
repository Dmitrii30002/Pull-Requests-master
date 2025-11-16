package handlers

import (
	"Pull-Requests-master/internal/domain"
	"Pull-Requests-master/internal/repository"
	"Pull-Requests-master/package/logger"
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

/*



	Обработка ошибок


*/

var (
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	prRepo   repository.PullRequestRepository
)

func SetService(db *sql.DB, log *logger.Logger) {
	userRepo = repository.NewUserRepository(db, log)
	teamRepo = repository.NewTeamRepository(db, log)
	prRepo = repository.NewPullRequestRepository(db, log)
}

func AddTeam(c echo.Context) error {
	var team domain.Team
	if err := c.Bind(&team); err != nil {
		return c.String(http.StatusBadRequest, "team is uncorrected")
	}

	newTeam, err := teamRepo.Create(&team)
	if err != nil {
		//Обработка ошибок
		c.Error(err)
		return err
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Команда создана",
		"team":    newTeam,
	})
}

func GetTeam(c echo.Context) error {
	teamName := c.Param("team_name")
	if teamName == "" {
		return c.String(http.StatusBadRequest, "team name is empty")
	}

	team, err := teamRepo.GetByName(teamName)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, team)
}
