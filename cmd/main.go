package main

import (
	"Pull-Requests-master/internal/handlers"
	"Pull-Requests-master/internal/migration"
	"Pull-Requests-master/package/config"
	"Pull-Requests-master/package/database"
	"Pull-Requests-master/package/logger"
	"fmt"

	"github.com/labstack/echo/v4"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Config wasn't created %v", err)
		return
	}

	log, err := logger.New(config)
	if err != nil {
		fmt.Printf("logger wasn't created %v", err)
		return
	}
	log.Info("logger was created")

	db, err := database.New(config)
	if err != nil {
		fmt.Printf("data base wasn't created %v", err)
		return
	}
	log.Info("data base was connected")

	err = migration.Migrate(db, "migrations")
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Info("migration completed")

	handler := handlers.NewHandler(db, log)
	e := echo.New()

	teams := e.Group("/team")
	{
		teams.POST("/add", handler.AddTeam)
		teams.GET("/get", handler.GetTeam)
	}

	users := e.Group("/users")
	{
		users.POST("/setIsActive", handler.SetUserActive)
		users.GET("/getReview", handler.GetUserReview)
	}

	pullRequests := e.Group("/pullRequest")
	{
		pullRequests.POST("/create", handler.CreatePR)
		pullRequests.POST("/merge", handler.MergePR)
		pullRequests.POST("/reassign", handler.ReassignReviewersPR)
	}
	e.Start(":8080")

	//TODO 10: Допы - под сомнением
}
