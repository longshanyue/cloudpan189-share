package fs

import (
	"context"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
)

func (f *fs) ClearOsType(ctx context.Context, osTypes ...string) (int64, error) {
	if len(osTypes) == 0 {
		return 0, WhereParamInvalid
	}

	// 检查是否需要删除真实文件
	needDeleteReal := false
	for _, osType := range osTypes {
		if osType == models.OsTypeRealFile {
			needDeleteReal = true

			break
		}
	}

	// 删除真实文件夹
	if needDeleteReal {
		f.clearAllRealFileFolders()
	}

	// 删除数据库记录
	result := f.getDB(ctx).Where("os_type IN ?", osTypes).Delete(&models.VirtualFile{})
	if result.Error != nil {
		f.logger.Error("删除文件失败", zap.Error(result.Error))
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (f *fs) clearAllRealFileFolders() {
	baseDir := configs.GetConfig().FileDir

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		f.logger.Error("读取根目录失败", zap.Error(err))
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if _, err = strconv.ParseInt(entry.Name(), 10, 64); err == nil {
				folderPath := filepath.Join(baseDir, entry.Name())
				if err := os.RemoveAll(folderPath); err != nil {
					f.logger.Error("删除文件夹失败", zap.String("path", folderPath), zap.Error(err))
				} else {
					f.logger.Debug("删除文件夹", zap.String("path", folderPath))
				}
			}
		}
	}
}
