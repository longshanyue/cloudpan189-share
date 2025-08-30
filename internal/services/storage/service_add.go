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
	Protocol        string `json:"protocol" binding:"required,oneof=subscribe share person family subscribe_share"`
	SubscribeUser   string `json:"subscribeUser"`
	ShareCode       string `json:"shareCode"`
	ShareAccessCode string `json:"shareAccessCode"`
	CloudToken      int64  `json:"cloudToken"`
	FileId          string `json:"fileId"`
	FamilyId        string `json:"familyId"`
}

type addResponse struct {
	ID int64 `json:"id"`
}

func (s *service) Add() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(addRequest)

		if err := ctx.ShouldBindJSON(req); err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		paths, err := utils.SplitPath(req.LocalPath)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		if len(paths) == 0 {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "不允许挂载根路径",
			})

			return
		}

		if !utils.CheckIsPath(req.LocalPath) {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "路径不合法，需要 / 开头的路径",
			})

			return
		}

		if exits, err := s.checkExist(ctx, req.LocalPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "添加失败",
			})

			return
		} else if exits {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "路径已存在",
			})

			return
		}

		if req.Protocol == "subscribe" && req.SubscribeUser == "" {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "subscribeUser is required",
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

		if req.Protocol == "share" && req.ShareCode == "" {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "shareCode is required",
			})

			return
		}

		if req.Protocol == "person" && (req.CloudToken == 0 ||
			req.FileId == "") {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "cloudToken and fileId are required for person protocol",
			})

			return
		}

		if req.Protocol == "family" && (req.CloudToken == 0 ||
			req.FamilyId == "" || req.FileId == "") {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: "cloudToken, familyId and fileId are required for family protocol",
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
				ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
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

			m.Addition[consts.FileAdditionKeyShareId] = resp.ShareId
			m.Addition[consts.FileAdditionKeyShareCode] = req.ShareCode
			m.Addition[consts.FileAdditionKeyAccessCode] = req.ShareAccessCode
			m.Addition[consts.FileAdditionKeyShareMode] = resp.ShareMode
			m.Addition[consts.FileAdditionKeyShareType] = resp.ShareType
			m.Addition[consts.FileAdditionKeyFileId] = resp.FileId
			m.Addition[consts.FileAdditionKeyIsFolder] = resp.IsFolder
		} else if req.Protocol == "person" || req.Protocol == "family" {
			token, err := s.getCloudToken(ctx, req.CloudToken)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				})

				return
			}

			m.Addition[consts.FileAdditionKeyIsFolder] = true
			m.Addition[consts.FileAdditionKeyFileId] = req.FileId

			ct := client.New().WithToken(client.NewAuthToken(token.AccessToken, token.ExpiresIn))

			if req.Protocol == "person" {
				_, err = ct.ListFiles(ctx, client.String(req.FileId))
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
						Code:    http.StatusInternalServerError,
						Message: err.Error(),
					})

					return
				}

				m.OsType = models.OsTypeCloudFolder
			} else {
				_, err = ct.FamilyListFiles(ctx, client.String(req.FamilyId), client.String(req.FileId))
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
						Code:    http.StatusInternalServerError,
						Message: err.Error(),
					})

					return
				}

				m.Addition[consts.FileAdditionKeyFamilyId] = req.FamilyId
				m.OsType = models.OsTypeCloudFamilyFolder
			}
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

			m.Addition[consts.FileAdditionKeySubscribeUser] = req.SubscribeUser
			m.Addition[consts.FileAdditionKeyShareId] = resp.ShareId
			m.Addition[consts.FileAdditionKeyFileId] = resp.FileId
			m.Addition[consts.FileAdditionKeyIsFolder] = resp.IsFolder
		}

		pid, err := s.findOrCreateAncestors(ctx, req.LocalPath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})

			return
		}

		m.ParentId = pid
		if err = s.db.WithContext(ctx).Create(m).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrResponse{
				Code:    http.StatusInternalServerError,
				Message: "创建失败",
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
