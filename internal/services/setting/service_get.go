package setting

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
)

type getResponse struct {
	*models.Setting
	RunTimes                  int64 `json:"runTimes"` // 已经运行的时间
	MultipleStreamThreadCount int   `json:"multipleStreamThreadCount"`
	MultipleStreamChunkSize   int64 `json:"multipleStreamChunkSize"`
}

func (s *service) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, &getResponse{
			Setting:                   shared.Setting,
			RunTimes:                  time.Now().Unix() - s.starTime.Unix(),
			MultipleStreamThreadCount: shared.MultipleStreamThreadCount,
			MultipleStreamChunkSize:   shared.MultipleStreamChunkSize,
		})
	}
}
