package fs

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
)

func (f *fs) Delete(ctx context.Context, id int64) error {
	return f.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		return f.deleteRecursively(ctx, tx, id)
	})
}

func (f *fs) deleteRecursively(ctx context.Context, tx *gorm.DB, id int64) error {
	// 先获取当前节点信息
	var current models.VirtualFile
	if err := tx.Where("id = ?", id).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			f.logger.Warn("文件不存在，跳过删除", zap.Int64("id", id))

			return nil
		}

		return fmt.Errorf("获取文件信息失败 id=%d: %w", id, err)
	}

	// 如果是文件夹，先递归删除所有子节点
	if current.IsFolder == 1 {
		children, err := f.getChildren(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("获取子节点失败: %w", err)
		}

		for _, child := range children {
			if err := f.deleteRecursively(ctx, tx, child.ID); err != nil {
				f.logger.Error("删除子文件失败",
					zap.Int64("parent_id", id),
					zap.Int64("child_id", child.ID),
					zap.Error(err))
				return err
			}
		}
	}

	// 删除真实文件/文件夹
	if err := f.deleteRealFileOrFolder(&current); err != nil {
		f.logger.Error("删除真实文件失败",
			zap.Int64("id", id),
			zap.String("name", current.Name),
			zap.Bool("is_folder", current.IsFolder == 1),
			zap.Error(err))
		// 如果物理文件已经不存在，记录警告但继续删除数据库记录
		if !os.IsNotExist(err) {
			return fmt.Errorf("删除真实文件失败: %w", err)
		}

		f.logger.Warn("物理文件不存在，继续删除数据库记录", zap.Int64("id", id))
	}

	// 删除数据库记录
	if err := tx.Where("id = ?", id).Or("link_id = ?", id).Delete(&models.VirtualFile{}).Error; err != nil {
		f.logger.Error("删除数据库记录失败", zap.Int64("id", id), zap.Error(err))
		return fmt.Errorf("删除数据库记录失败 id=%d: %w", id, err)
	}

	f.logger.Debug("成功删除文件",
		zap.Int64("id", id),
		zap.String("name", current.Name),
		zap.Bool("is_folder", current.IsFolder == 1))

	return nil
}

func (f *fs) deleteRealFileOrFolder(file *models.VirtualFile) error {
	if file.IsFolder == 1 {
		// 删除文件夹
		return f.deleteRealFolder(file.ID)
	} else {
		// 删除文件 - 从 addition.file_path 获取路径
		filePath := f.getFilePathFromAddition(file)
		if filePath == "" {
			f.logger.Warn("文件路径为空，跳过删除物理文件",
				zap.Int64("id", file.ID),
				zap.String("name", file.Name))

			return nil
		}

		return f.deleteRealFile(filePath)
	}
}

// 从 addition 字段中获取文件路径
func (f *fs) getFilePathFromAddition(file *models.VirtualFile) string {
	if file.Addition == nil {
		return ""
	}

	if filePath, ok := file.Addition["file_path"].(string); ok {
		return filePath
	}

	f.logger.Warn("addition中没有file_path字段",
		zap.Int64("id", file.ID),
		zap.String("name", file.Name),
		zap.Any("addition", file.Addition))

	return ""
}

func (f *fs) getChildren(ctx context.Context, tx *gorm.DB, parentID int64) ([]models.VirtualFile, error) {
	var children []models.VirtualFile
	err := tx.Where("parent_id = ?", parentID).Find(&children).Error
	if err != nil {
		return nil, fmt.Errorf("查询子节点失败 parent_id=%d: %w", parentID, err)
	}
	return children, nil
}

func (f *fs) deleteRealFile(filePath string) error {
	// filePath 已经是完整路径，或者是相对于 FileDir 的路径
	var fullPath string
	if filepath.IsAbs(filePath) {
		fullPath = filePath
	} else {
		fullPath = filepath.Join(configs.GetConfig().FileDir, filePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		f.logger.Warn("文件不存在，跳过删除", zap.String("path", fullPath))

		return nil
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("删除文件失败 %s: %w", fullPath, err)
	}

	f.logger.Debug("成功删除文件", zap.String("path", fullPath))

	return nil
}

func (f *fs) deleteRealFolder(folderId int64) error {
	folderPath := filepath.Join(configs.GetConfig().FileDir, strconv.FormatInt(folderId, 10))

	// 检查文件夹是否存在
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		f.logger.Warn("文件夹不存在，跳过删除", zap.String("path", folderPath))

		return nil
	}

	if err := os.RemoveAll(folderPath); err != nil {
		return fmt.Errorf("删除文件夹失败 %s: %w", folderPath, err)
	}

	f.logger.Debug("成功删除文件夹", zap.String("path", folderPath))

	return nil
}
