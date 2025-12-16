# Pull-Requests-master
 Микросервис, позволяющий автоматически назначать ревьюеров на Pull Request’ы (PR)


## Описание проекта 
Данный проект является микросервисом, позволяющим управлять pull requests и командами. Данный сервис использует REST API. Схема api находится в каталоге docs в файле openapi.yaml. 
### Основной стэк:
* golang 1.25 (echo)
* postgreSQL
* docker(docker-compose)

## Копирование репозитория
Используйте данную команду, чтобы начать работать с репозиторием:
``` bash
		git clone https://github.com/Dmitrii30002/Pull-Requests-master.git
    	cd  Pull-Requests-master
```

## Запуск проекта
Для запуска приложения имеется makefile. чтобы запустить файл необходимо ввести следующую команду:
``` bash
		make docker-up
```
Чтобы опустить проект используйте:
``` bash
		make docker-down
```
Для запуска тестов пропишите:
``` bash
		make test
```

## Контейнеризация проекта
Проект использует docker для контейнеризации. База данных поднимется в контейнере. Для проекта был написан docker файл, а также файл docker-compose.
Чтобы поднять контейнер с проектом пропишите:
``` bash
		docker build -t go-app . 
    	docker run -d -p 8080:8080 go-app
```
Чтобы поднять весь проект с базой данных используйте:
``` bash
	docker-compose up
```

## Тестирование
API было протестированно с помощью Postman. Также для репозиториев были прописаны Unit-тесты. Для запуска тестов пропишите:
``` bash
		make test
```
Или
``` bash
		go test -v ./...
```
## Конфигураци
Конфигурация сервера прописывается в файле config.yaml. В данном файле присутсвуют следующие значения:
* server.Host
* server.Port
* loggger.level
* logger.out
Для конфигурации подключенияк БД используюся переменные окружения. Их можно передать в контейнер во время запуска, а можно изменить в файле docker-compose.

## Логирование
Для логирования используется логгер библиотеки logrus.

## Структура проекта
```
project/
├── cmd/
│   └── main.go
├── docs/
│   └── openapi.yaml
├── internal/
│   ├── errors/
│   │	└── errors.go
│   ├── domain/
│   │	└── domain.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── team_repository.go
│   │   └── pull_request_repository.go
│   ├── service/
│   │   ├── user_service.go
│   │   ├── team_service.go
│   │   └── pr_service.go
│   ├── migration/
│   │	└── migration.go
│   └── handler/
│       ├── user_handler.go
│       ├── team_handler.go
│       └── pr_handler.go
│──package/
│  ├── database/
│  │   └── database.go
│  ├── logger/
│  │   └── logger.go
│  └── config/
│      └── config.go
├── migrations/
│   └── *.sql
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── config.yaml
├── go.mod
└── go.sum
```

## Будущее проекта
Для данного проекта предусматривается нагрузочное и интеграционное тестирование. Также необходимо добавить ручки для получения статистики. К тому же, необходимо реализовать graceful shutdown. 
