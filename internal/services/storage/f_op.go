package storage

import (
	"context"
	"errors"
	"path"
	"slices"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// checkExist 检查路径是否存在
func (s *service) checkExist(ctx context.Context, path string) (bool, error) {
	paths, err := utils.SplitPath(path)
	if err != nil {
		return false, err
	}

	var pid int64

	for _, name := range paths {
		var m = new(models.VirtualFile)

		if err := s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("name", name).Where("parent_id", pid).First(m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, nil
			}

			return false, err
		}

		pid = m.ID
	}

	return true, nil
}

// createAncestors 创建祖先路径（本级不创建）
func (s *service) createAncestors(ctx context.Context, path string) error {
	var (
		paths, err = utils.SplitPath(path)
		pid        int64
	)

	if err != nil {
		return err
	}

	// 去掉最后一级路径
	if len(paths) > 0 {
		paths = paths[:len(paths)-1]
	}

	for _, name := range paths {
		// 检查当前层级是否存在
		var m = new(models.VirtualFile)
		err := s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("name", name).Where("parent_id", pid).First(m).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不存在则创建
				now := time.Now()
				m = &models.VirtualFile{
					ParentId:   pid,
					Name:       name,
					IsFolder:   1,
					Size:       0,
					OsType:     models.OsTypeFolder,
					CreateDate: now.Format(time.DateTime),
					ModifyDate: now.Format(time.DateTime),
					Rev:        now.Format("20060102150405"),
					Addition:   datatypes.JSONMap{},
				}

				if err = s.db.WithContext(ctx).Create(m).Error; err != nil {
					return err
				}

				pid = m.ID

				continue
			}

			return err
		}

		pid = m.ID
	}

	return nil
}

// findOrCreateAncestors 查找或创建所有祖先路径（本级不创建），返回最后一级的父ID
func (s *service) findOrCreateAncestors(ctx context.Context, path string) (int64, error) {
	var (
		paths, err = utils.SplitPath(path)
		pid        int64
	)

	if err != nil {
		return 0, err
	}

	// 去掉最后一级路径
	if len(paths) > 0 {
		paths = paths[:len(paths)-1]
	}

	for _, name := range paths {
		var m = new(models.VirtualFile)
		err := s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("name", name).Where("parent_id", pid).First(m).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不存在则创建
				now := time.Now()
				m = &models.VirtualFile{
					ParentId:   pid,
					Name:       name,
					IsFolder:   1,
					Size:       0,
					OsType:     models.OsTypeFolder,
					CreateDate: now.Format(time.DateTime),
					ModifyDate: now.Format(time.DateTime),
					Rev:        now.Format("20060102150405"),
					Addition:   datatypes.JSONMap{},
				}

				if err = s.db.WithContext(ctx).Create(m).Error; err != nil {
					return 0, err
				}

				pid = m.ID

				continue
			}
			return 0, err
		}

		pid = m.ID
	}

	return pid, nil
}

func (s *service) getFullPath(ctx context.Context, file *models.VirtualFile) (string, error) {
	var paths = []string{file.Name}

	pid := file.ParentId

	for pid > 0 {
		var m = new(models.VirtualFile)
		if err := s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("id", pid).First(m).Error; err != nil {
			return "", err
		}
		paths = append(paths, m.Name)
		pid = m.ParentId
	}

	paths = append(paths, "/")

	slices.Reverse(paths)

	return path.Join(paths...), nil
}
