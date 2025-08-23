package bridge

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
)

type familyListRequest struct {
	CloudToken int64 `binding:"required" form:"cloudToken"`
}

func (s *service) FamilyList() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(familyListRequest)

		if err := ctx.ShouldBindQuery(req); err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		token, err := s.getCloudToken(ctx, req.CloudToken)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		resp, err := client.New().
			WithToken(client.NewAuthToken(token.AccessToken, token.ExpiresIn)).
			GetFamilyList(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		ctx.JSON(http.StatusOK, resp.FamilyInfoResp)
	}
}
