package fs

import (
	"context"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type File = *models.VirtualFile

type FS interface {
	Create(ctx context.Context, pid int64, file File) (id int64, err error)
	BatchCreate(ctx context.Context, pid int64, files ...File) (count int64, err error)
	Delete(ctx context.Context, id int64) error
	Children(ctx context.Context, id int64) ([]File, error)
	Update(ctx context.Context, id int64, mp map[string]interface{}) error
	// ClearOsType 清空指定osType的所有文件
	ClearOsType(ctx context.Context, osTypes ...string) (int64, error)
}

type fs struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewFS(db *gorm.DB, logger *zap.Logger) FS {
	return &fs{
		db:     db,
		logger: logger,
	}
}

func (f *fs) getDB(ctx context.Context) *gorm.DB {
	return f.db.WithContext(ctx).Model(&models.VirtualFile{})
}

func (f *fs) Children(ctx context.Context, id int64) ([]File, error) {
	var children []*models.VirtualFile

	if err := f.getDB(ctx).Where("parent_id = ?", id).Find(&children).Error; err != nil {
		f.logger.Error("获取文件子节点失败", zap.Error(err))

		return nil, err
	}

	return children, nil
}

func (f *fs) Update(ctx context.Context, id int64, mp map[string]interface{}) error {
	if mp == nil {
		return nil
	}

	if err := f.getDB(ctx).Where("id = ?", id).Updates(mp).Error; err != nil {
		f.logger.Error("更新文件失败", zap.Error(err))

		return err
	}

	return nil
}
