package cloudtoken

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	InitQrcode() gin.HandlerFunc
	CheckQrcode() gin.HandlerFunc
	ModifyName() gin.HandlerFunc
	Delete() gin.HandlerFunc
	List() gin.HandlerFunc
	UsernameLogin() gin.HandlerFunc
}

type service struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:     db,
		logger: logger,
	}
}
