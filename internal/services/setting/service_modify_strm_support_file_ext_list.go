package setting

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/shared"
	"net/http"
)

type modifyStrmSupportFileExtListRequest struct {
	StrmSupportFileExtList []string `json:"strmSupportFileExtList"`
}

type modifyStrmSupportFileExtListResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

func (s *service) ModifyStrmSupportFileExtList() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(modifyStrmSupportFileExtListRequest)

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "参数错误",
			})

			return
		}

		// 直接使用传入的切片，如果为空就是空切片
		var fileExtList []string = req.StrmSupportFileExtList
		if fileExtList == nil {
			fileExtList = make([]string, 0)
		}

		// 如果传入了具体的扩展名列表，验证每个扩展名不为空
		if len(req.StrmSupportFileExtList) > 0 {
			for _, ext := range req.StrmSupportFileExtList {
				if ext == "" {
					ctx.JSON(http.StatusBadRequest, gin.H{
						"code": http.StatusBadRequest,
						"msg":  "文件扩展名不能为空",
					})
					return
				}
			}
		}

		result := new(models.SettingDict).SetStrmSupportFileExtList(s.db.WithContext(ctx), fileExtList)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  fmt.Sprintf("修改失败：%s", result.Error.Error()),
			})

			return
		}

		shared.StrmSupportFileExtList = fileExtList

		ctx.JSON(http.StatusOK, modifyStrmSupportFileExtListResponse{
			RowsAffected: result.RowsAffected,
		})
	}
}
