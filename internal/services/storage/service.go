package storage

import (
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Service interface {
	Add() gin.HandlerFunc
	Delete() gin.HandlerFunc
	List() gin.HandlerFunc
	ModifyToken() gin.HandlerFunc
	DeepRefreshFile() gin.HandlerFunc
	Search() gin.HandlerFunc
}

type service struct {
	db     *gorm.DB
	logger *zap.Logger
	cache  *cache.Cache
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:     db,
		logger: logger,
		cache:  cache.New(time.Minute, time.Minute),
	}
}
