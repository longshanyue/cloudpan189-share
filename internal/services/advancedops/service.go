package advancedops

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	RebuildStrm() gin.HandlerFunc
	ClearMedia() gin.HandlerFunc
	BusDetail() gin.HandlerFunc
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
