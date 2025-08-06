package usergroup

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	Add() gin.HandlerFunc
	Delete() gin.HandlerFunc
	List() gin.HandlerFunc
	ModifyName() gin.HandlerFunc
	BatchBindFiles() gin.HandlerFunc
	GetBindFiles() gin.HandlerFunc
}

type service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建用户组服务
func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:     db,
		logger: logger,
	}
}
