package cloudtoken

import (
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"net/http"
)

type initQrcodeResponse struct {
	UUID string `json:"uuid"`
}

func (s *service) InitQrcode() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := client.LoginInit()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  "登录初始化失败",
			})

			return
		}

		ctx.JSON(http.StatusOK, &initQrcodeResponse{
			UUID: resp.UUID,
		})
	}
}
