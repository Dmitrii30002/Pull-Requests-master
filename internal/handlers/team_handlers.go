package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

/*



	Обработка ошибок


*/

func SetUserActive(c echo.Context) error {
	userID := c.QueryParam("user_name")
	isActive := c.QueryParam("is_active")

	status, err := strconv.ParseBool(isActive)
	if err != nil {
		return err
	}
	user, err := userRepo.SetUserActive(userID, status)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}
