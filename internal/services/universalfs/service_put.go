package universalfs

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xxcheng123/cloudpan189-share/configs"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"gorm.io/datatypes"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

func (s *service) Put() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		returnCode, err := s.put(ctx)

		if err != nil {
			ctx.JSON(returnCode, gin.H{
				"code":    returnCode,
				"message": err.Error(),
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
		pid      = ctx.GetInt64("x_pid")
		fid      = ctx.GetInt64("x_fid")
		filename = ctx.GetString("x_file_name")
	)

	// 创建它的上级文件夹
	dirName := path.Join(configs.GetConfig().FileDir, strconv.FormatInt(pid, 10))

	if err = os.Mkdir(dirName, 0777); err != nil && !os.IsExist(err) {
		if os.IsNotExist(err) {
			return http.StatusConflict, err
		}

		return http.StatusMethodNotAllowed, err
	}

	if fid == -1 {
		writeFilename := uuid.NewString() + ".bin"
		realPath := fmt.Sprintf("%s/%s", dirName, writeFilename)
		f, err := os.OpenFile(realPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			if os.IsNotExist(err) {
				return http.StatusConflict, err
			}

			return http.StatusNotFound, err
		}

		_, copyErr := io.Copy(f, ctx.Request.Body)
		fi, statErr := f.Stat()
		closeErr := f.Close()
		// TODO(rost): Returning 405 Method Not Allowed might not be appropriate.
		if copyErr != nil {
			return http.StatusMethodNotAllowed, copyErr
		}
		if statErr != nil {
			return http.StatusMethodNotAllowed, statErr
		}
		if closeErr != nil {
			return http.StatusMethodNotAllowed, closeErr
		}

		s.db.WithContext(ctx).Create(&models.VirtualFile{
			ParentId: pid,
			Name:     filename,
			IsTop:    0,
			Size:     fi.Size(),
			Hash:     "-",
			OsType:   models.OsTypeRealFile,
			Addition: datatypes.JSONMap{
				"file_path": fmt.Sprintf("%d/%s", pid, writeFilename),
			},
			Rev:        now.Format("20060102150405"),
			CreateDate: now.Format(time.DateTime),
			ModifyDate: now.Format(time.DateTime),
		})

		return http.StatusCreated, nil
	} else {
		file := &models.VirtualFile{}

		if err = s.db.WithContext(ctx).Where("id=?", fid).First(file).Error; err != nil {
			return http.StatusNotFound, err
		}

		if file.OsType != models.OsTypeRealFile {
			return http.StatusMethodNotAllowed, fmt.Errorf("file not allowed to update")
		}

		vRealPath, ok := file.Addition["file_path"]
		if !ok {
			return http.StatusMethodNotAllowed, fmt.Errorf("file not found")
		}
		realPath, ok := vRealPath.(string)
		if !ok {
			return http.StatusMethodNotAllowed, fmt.Errorf("file not found")
		}

		f, err := os.OpenFile(path.Join(configs.GetConfig().FileDir, realPath), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			if os.IsNotExist(err) {
				return http.StatusConflict, err
			}

			return http.StatusNotFound, err
		}

		_, copyErr := io.Copy(f, ctx.Request.Body)
		fi, statErr := f.Stat()
		closeErr := f.Close()
		// TODO(rost): Returning 405 Method Not Allowed might not be appropriate.
		if copyErr != nil {
			return http.StatusMethodNotAllowed, copyErr
		}
		if statErr != nil {
			return http.StatusMethodNotAllowed, statErr
		}
		if closeErr != nil {
			return http.StatusMethodNotAllowed, closeErr
		}

		s.db.WithContext(ctx).Model(new(models.VirtualFile)).Where("id=?", fid).Updates(map[string]any{
			"size":        fi.Size(),
			"modify_date": now.Format(time.DateTime),
			"rev":         now.Format("20060102150405"),
		})

		return http.StatusCreated, nil
	}
}
