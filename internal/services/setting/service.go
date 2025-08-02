package setting

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Service interface {
	Get() gin.HandlerFunc
	RefreshKey() gin.HandlerFunc
	ModifyName() gin.HandlerFunc
	ToggleAuth() gin.HandlerFunc
	ToggleLocalProxy() gin.HandlerFunc
	ToggleMultipleStream() gin.HandlerFunc
	ModifyBaseURL() gin.HandlerFunc
	ToggleEnableTopFileAutoRefresh() gin.HandlerFunc
	InitSystem() gin.HandlerFunc
}

type service struct {
	db       *gorm.DB
	logger   *zap.Logger
	starTime time.Time
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:       db,
		logger:   logger,
		starTime: time.Now(),
	}
}

func (s *service) get(ctx context.Context) (*models.Setting, error) {
	return shared.Setting, nil
}
