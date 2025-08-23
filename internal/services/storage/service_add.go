package storage

import (
	"net/http"
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/bus"
	"github.com/xxcheng123/cloudpan189-share/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"gorm.io/datatypes"
)

type addRequest struct {
	LocalPath       string `json:"localPath" binding:"required"`
	Protocol        string `json:"protocol" binding:"required,oneof=subscribe share"`
	SubscribeUser   string `json:"subscribeUser"`
	ShareCode       string `json:"shareCode"`
	ShareAccessCode string `json:"shareAccessCode"`
	CloudToken      int64  `json:"cloudToken"`
}

type addResponse struct {
	ID int64 `json:"id"`
}

func (s *service) Add() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(addRequest)

		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})
		}

		paths, err := utils.SplitPath(req.LocalPath)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": err.Error(),
			})

			return
		}

		if len(paths) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "不允许挂载根路径",
			})

			return
		}

		if !utils.CheckIsPath(req.LocalPath) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "路径不合法，需要 / 开头的路径",
			})

			return
		}

		if exits, err := s.checkExist(ctx, req.LocalPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "添加失败",
			})

			return
		} else if exits {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "路径已存在",
			})

			return
		}

		if req.Protocol == "subscribe" && req.SubscribeUser == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "subscribeUser is required",
			})

			return
		}

		if req.Protocol == "share" && req.ShareCode == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "shareCode is required",
			})

			return
		}

		now := time.Now()
		m := &models.VirtualFile{
			Name:       paths[len(paths)-1],
			IsTop:      1,
			Size:       0,
			Hash:       "",
			CreateDate: now.Format(time.DateTime),
			ModifyDate: now.Format(time.DateTime),
			Rev:        now.Format("20060102150405"),
			OsType:     req.Protocol,
			IsFolder:   1,
			Addition:   make(datatypes.JSONMap),
		}

		if req.CloudToken != 0 {
			m.Addition[consts.FileAdditionKeyCloudToken] = req.CloudToken
		}

		if req.Protocol == "subscribe" {
			_, err := client.New().GetUpResourceShare(ctx, req.SubscribeUser, 1, 30)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": err.Error(),
				})

				return
			}

			m.Addition[consts.FileAdditionKeySubscribeUser] = req.SubscribeUser
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

			m.Addition[consts.FileAdditionKeyShareId] = resp.ShareId
			m.Addition[consts.FileAdditionKeyShareCode] = req.ShareCode
			m.Addition[consts.FileAdditionKeyAccessCode] = req.ShareAccessCode
			m.Addition[consts.FileAdditionKeyShareMode] = resp.ShareMode
			m.Addition[consts.FileAdditionKeyShareType] = resp.ShareType
			m.Addition[consts.FileAdditionKeyFileId] = resp.FileId
			m.Addition[consts.FileAdditionKeyIsFolder] = resp.IsFolder
		}

		pid, err := s.findOrCreateAncestors(ctx, req.LocalPath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			})

			return
		}

		m.ParentId = pid
		if err = s.db.WithContext(ctx).Create(m).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "创建失败",
			})

			return
		}

		if err = bus.PublishVirtualFileRefresh(ctx, m.ID, false); err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "创建成功，但是刷新文件失败",
			})
		}

		ctx.JSON(http.StatusOK, addResponse{ID: m.ID})
	}
}
