package storage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
)

type preAddRequest struct {
	Protocol        string `json:"protocol" binding:"required,oneof=subscribe share subscribe_share"`
	SubscribeUser   string `json:"subscribeUser"`
	ShareCode       string `json:"shareCode"`
	ShareAccessCode string `json:"shareAccessCode"`
	CloudToken      int64  `json:"cloudToken"`
}

type preAddResponse struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

func (s *service) PreAdd() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(preAddRequest)

		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		if req.Protocol == "subscribe" && req.SubscribeUser == "" {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "订阅用户不能为空",
			})

			return
		}

		if req.Protocol == "share" && req.ShareCode == "" {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "分享码不能为空",
			})

			return
		}

		if req.Protocol == "subscribe_share" && (req.SubscribeUser == "" || req.ShareCode == "") {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "subscribeUser or shareCode is required",
			})

			return
		}

		var name string

		if req.Protocol == "subscribe" {
			subscribeUser, err := client.New().SubscribeGetUser(ctx, req.SubscribeUser)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": err.Error(),
				})

				return
			}

			name = subscribeUser.Data.Name
		} else if req.Protocol == "share" {
			var opts []client.GetShareInfoOption

			if req.ShareAccessCode != "" {
				opts = append(opts, func(r *client.GetShareInfoRequest) {
					r.AccessCode = req.ShareAccessCode
				})
			}

			resp, err := client.New().GetShareInfo(ctx, req.ShareCode, opts...)
			if err != nil {
				var clientErr = new(client.RespErr)
				if errors.As(err, &clientErr) {
					if clientErr.ResCode == "ShareAuditWaiting" {
						ctx.JSON(http.StatusBadRequest, gin.H{
							"code":    http.StatusBadRequest,
							"message": "当前分享审核中，请稍后再试",
						})

						return
					}
				}

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": err.Error(),
				})

				return
			}

			if resp.ShareId == 0 {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "分享查询失败",
				})

				return
			}

			name = resp.FileName
		} else if req.Protocol == "subscribe_share" {
			resp, err := client.New().GetShareInfo(ctx, req.ShareCode)
			if err != nil {
				var clientErr = new(client.RespErr)
				if errors.As(err, &clientErr) {
					if clientErr.ResCode == "ShareAuditWaiting" {
						ctx.JSON(http.StatusBadRequest, types.ErrResponse{
							Code:    http.StatusBadRequest,
							Message: "当前分享审核中，请稍后再试",
						})

						return
					}
				}

				ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				})

				return
			}

			if resp.ShareId == 0 {
				ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: "分享查询失败",
				})

				return
			}

			name = resp.FileName
		}

		ctx.JSON(http.StatusOK, preAddResponse{
			Name:     name,
			Protocol: req.Protocol,
		})
	}
}
