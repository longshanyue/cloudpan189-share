package advancedops

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/bus"
	"net/http"
)

type busDetailResponse = bus.DetailInfo

func (s *service) BusDetail() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var detail busDetailResponse = bus.Detail()

		ctx.JSON(http.StatusOK, detail)
	}
}
