package main

import (
	"Pull-Requests-master/package/config"
	"Pull-Requests-master/package/database"
	"Pull-Requests-master/package/logger"
	"fmt"
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

	db, err := database.New(config)
	if err != nil {
		fmt.Printf("data base wasn't created %v", err)
		return
	}

	fmt.Println(db, log)
	fmt.Print("Hallo world")

	//TODO 4.1: Реализация репозеториев и бизнес-логики
	//TODO 4.2: Расписать ошибки

	//TODO 5: Реализация хендлеров

	//TODO 6: Докер образ + докер композе +

	//TODO 7: Юнит тесты + моки

	//TODO 8: Нагрузочное тестирование

	//TODO 9: Интеграционные тесты

	//TODO 10: Допы
}
