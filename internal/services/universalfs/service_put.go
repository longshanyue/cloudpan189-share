package universalfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"

	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
)

func (s *service) Put() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		returnCode, err := s.put(ctx)
		if err != nil {
			s.logger.Error("文件上传失败",
				zap.String("path", ctx.Param("path")),
				zap.Int("code", returnCode),
				zap.Error(err))
			ctx.JSON(returnCode, types.ErrResponse{
				Code:    returnCode,
				Message: err.Error(),
			})
			return
		}

		ctx.Writer.WriteHeader(returnCode)
	}
}

func (s *service) put(ctx *gin.Context) (int, error) {
	rawPath := ctx.Param("path")
	now := time.Now()

	release, status, err := s.confirmLocks(rawPath, "")
	if err != nil {
		return status, err
	}
	defer release()

	var (
		pid      = ctx.GetInt64(consts.CtxKeyParentId)
		fid      = ctx.GetInt64(consts.CtxKeyFileId)
		filename = ctx.GetString(consts.CtxKeyFilename)
	)

	if err = s.ensureParentDirectory(pid); err != nil {
		return s.handleDirectoryError(err)
	}

	if fid == -1 {
		return s.createNewFile(ctx, pid, filename, now)
	}

	return s.updateExistingFile(ctx, fid, now)
}

func (s *service) ensureParentDirectory(pid int64) error {
	dirName := path.Join(configs.GetConfig().FileDir, strconv.FormatInt(pid, 10))
	if err := os.Mkdir(dirName, 0777); err != nil && !os.IsExist(err) {
		s.logger.Error("创建父级目录失败",
			zap.Int64("parentId", pid),
			zap.String("dirName", dirName),
			zap.Error(err))

		return err
	}

	return nil
}

func (s *service) handleDirectoryError(err error) (int, error) {
	if os.IsNotExist(err) {
		return http.StatusConflict, err
	}

	return http.StatusMethodNotAllowed, err
}

func (s *service) createNewFile(ctx *gin.Context, pid int64, filename string, now time.Time) (int, error) {
	writeFilename := uuid.NewString() + ".bin"
	dirName := path.Join(configs.GetConfig().FileDir, strconv.FormatInt(pid, 10))
	realPath := path.Join(dirName, writeFilename)

	file, err := os.OpenFile(realPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		s.logger.Error("创建新文件失败",
			zap.String("realPath", realPath),
			zap.String("filename", filename),
			zap.Error(err))

		return s.handleFileError(err)
	}

	fileInfo, err := s.writeFileContent(file, ctx.Request.Body)
	if err != nil {
		s.logger.Error("写入文件内容失败",
			zap.String("filename", filename),
			zap.Error(err))
		return http.StatusMethodNotAllowed, err
	}

	if err := s.saveNewFileRecord(ctx, pid, filename, fileInfo.Size(), writeFilename, now); err != nil {
		s.logger.Error("保存文件记录失败",
			zap.String("filename", filename),
			zap.Error(err))
		return http.StatusInternalServerError, err
	}

	s.logger.Info("新文件创建成功",
		zap.String("filename", filename),
		zap.Int64("size", fileInfo.Size()),
		zap.String("realPath", realPath))

	return http.StatusCreated, nil
}

func (s *service) updateExistingFile(ctx *gin.Context, fid int64, now time.Time) (int, error) {
	file := &models.VirtualFile{}
	if err := s.db.WithContext(ctx).Where("id=?", fid).First(file).Error; err != nil {
		s.logger.Error("查询文件记录失败", zap.Int64("fileId", fid), zap.Error(err))

		return http.StatusNotFound, err
	}

	if file.OsType != models.OsTypeRealFile {
		s.logger.Warn("尝试更新非真实文件",
			zap.Int64("fileId", fid),
			zap.String("fileName", file.Name),
			zap.String("osType", file.OsType))

		return http.StatusMethodNotAllowed, fmt.Errorf("文件类型不允许更新")
	}

	realPath, err := s.getRealFilePath(file)
	if err != nil {
		s.logger.Error("获取文件真实路径失败",
			zap.Int64("fileId", fid),
			zap.String("fileName", file.Name),
			zap.Error(err))

		return http.StatusMethodNotAllowed, err
	}

	fullPath := path.Join(configs.GetConfig().FileDir, realPath)
	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		s.logger.Error("打开文件失败",
			zap.String("fullPath", fullPath),
			zap.Error(err))

		return s.handleFileError(err)
	}

	fileInfo, err := s.writeFileContent(f, ctx.Request.Body)
	if err != nil {
		s.logger.Error("更新文件内容失败",
			zap.Int64("fileId", fid),
			zap.String("fileName", file.Name),
			zap.Error(err))

		return http.StatusMethodNotAllowed, err
	}

	if err := s.updateFileRecord(ctx, fid, fileInfo.Size(), now); err != nil {
		s.logger.Error("更新文件记录失败",
			zap.Int64("fileId", fid),
			zap.Error(err))

		return http.StatusInternalServerError, err
	}

	s.logger.Info("文件更新成功",
		zap.Int64("fileId", fid),
		zap.String("fileName", file.Name),
		zap.Int64("size", fileInfo.Size()))

	return http.StatusCreated, nil
}

func (s *service) handleFileError(err error) (int, error) {
	if os.IsNotExist(err) {
		return http.StatusConflict, err
	}
	return http.StatusNotFound, err
}

func (s *service) writeFileContent(file *os.File, body io.Reader) (os.FileInfo, error) {
	defer func() {
		_ = file.Close()
	}()

	_, copyErr := io.Copy(file, body)
	if copyErr != nil {
		return nil, copyErr
	}

	fileInfo, statErr := file.Stat()
	if statErr != nil {
		return nil, statErr
	}

	return fileInfo, nil
}

func (s *service) getRealFilePath(file *models.VirtualFile) (string, error) {
	vRealPath, ok := file.Addition[consts.FileAdditionKeyFilePath]
	if !ok {
		return "", fmt.Errorf("文件路径不存在")
	}

	realPath, ok := vRealPath.(string)
	if !ok {
		return "", fmt.Errorf("文件路径格式错误")
	}

	return realPath, nil
}

func (s *service) saveNewFileRecord(ctx *gin.Context, pid int64, filename string, size int64, writeFilename string, now time.Time) error {
	return s.db.WithContext(ctx).Create(&models.VirtualFile{
		ParentId: pid,
		Name:     filename,
		IsTop:    0,
		Size:     size,
		Hash:     "-",
		OsType:   models.OsTypeRealFile,
		Addition: datatypes.JSONMap{
			consts.FileAdditionKeyFilePath: fmt.Sprintf("%d/%s", pid, writeFilename),
		},
		Rev:        now.Format("20060102150405"),
		CreateDate: now.Format(time.DateTime),
		ModifyDate: now.Format(time.DateTime),
	}).Error
}

func (s *service) updateFileRecord(ctx *gin.Context, fid int64, size int64, now time.Time) error {
	return s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("id=?", fid).Updates(map[string]any{
		"size":        size,
		"modify_date": now.Format(time.DateTime),
		"rev":         now.Format("20060102150405"),
	}).Error
}
