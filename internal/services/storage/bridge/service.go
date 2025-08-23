package bridge

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service interface {
	GetPersonNodes() gin.HandlerFunc
	GetFamilyNodes() gin.HandlerFunc
	FamilyList() gin.HandlerFunc
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

type FileNode struct {
	ParentId string `json:"parentId"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsFolder int64  `json:"isFolder"`
}

func (s *service) getCloudToken(ctx context.Context, id int64) (*models.CloudToken, error) {
	m := new(models.CloudToken)

	err := s.db.WithContext(ctx).Where("id = ?", id).First(m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("令牌不存在")
		}

		return nil, err
	}

	return m, nil
}
