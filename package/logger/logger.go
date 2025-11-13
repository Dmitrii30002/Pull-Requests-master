package logger

import (
	"Pull-Requests-master/package/config"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func New(cfg *config.Config) (*Logger, error) {
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.Logger.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse level: %v", err)
	}
	logger.SetLevel(level)

	if cfg.Logger.Path != "" {
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("file doesn't exist: %v", err)
		}
		logger.SetOutput(file)
	}

	return &Logger{logger}, nil
}
