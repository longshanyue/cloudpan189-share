package fs

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

func (f *fs) Create(ctx context.Context, pid int64, file File) (id int64, err error) {
	if file == nil {
		return 0, FileNil
	}

	if pid == 0 {
		return 0, RootDirProhibitsCreateFile
	}

	file.ParentId = pid

	if err = f.getDB(ctx).Create(file).Error; err != nil {
		f.logger.Error("创建文件失败", zap.Error(err))

		return 0, err
	}

	if strmFile, ok := GetStrm(file); ok {
		_, _ = f.Create(ctx, pid, strmFile)
	}

	return file.ID, nil
}

func (f *fs) BatchCreate(ctx context.Context, pid int64, files ...File) (count int64, err error) {
	if len(files) == 0 {
		return 0, FileNil
	}

	if pid == 0 {
		return 0, RootDirProhibitsCreateFile
	}

	if result := f.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "parent_id"}, {Name: "name"}},
		DoNothing: true,
	}).CreateInBatches(files, len(files)); result.Error != nil {
		f.logger.Error("批量创建文件失败", zap.Error(result.Error))

		return 0, result.Error
	} else {
		strmFiles := make([]File, 0)

		for _, file := range files {
			file.ParentId = pid

			if strmFile, ok := GetStrm(file); ok {
				strmFiles = append(strmFiles, strmFile)
			}
		}

		_, _ = f.BatchCreate(ctx, pid, strmFiles...)

		return result.RowsAffected, nil
	}
}

func GetStrm(f File) (File, bool) {
	if f.IsFolder == 1 {
		return nil, false
	}

	if !shared.StrmFileEnable || f.OsType != models.OsTypeFile {
		return nil, false
	}

	extName := strings.TrimPrefix(filepath.Ext(f.Name), ".")
	if len(shared.StrmSupportFileExtList) > 0 && lo.IndexOf(shared.StrmSupportFileExtList, extName) == -1 {
		return nil, false
	}

	now := time.Now()

	// 计算 size
	size := len(fmt.Sprintf("%s/api/file_download?id=%d&random=%s&sign=12345678123456781234567812345678&timestamp=-1", shared.Setting.BaseURL, f.ID, uuid.NewString()))

	strmFile := &models.VirtualFile{
		ParentId:   f.ParentId,
		LinkId:     f.ID,
		Name:       strings.TrimSuffix(f.Name, filepath.Ext(f.Name)) + ".strm",
		IsTop:      f.IsTop,
		Size:       int64(size),
		IsFolder:   f.IsFolder,
		Hash:       "-",
		CreateDate: now.Format(time.DateTime),
		ModifyDate: now.Format(time.DateTime),
		OsType:     models.OsTypeStrmFile,
		Addition:   make(datatypes.JSONMap),
		Rev:        now.Format("20060102150405"),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return strmFile, true
}
