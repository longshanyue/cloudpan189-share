package configs

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	logger *zap.Logger
)

func DB() *gorm.DB {
	return db
}

func Logger() *zap.Logger {
	return logger
}

func GetConfig() *Config {
	return c
}
