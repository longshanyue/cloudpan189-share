package bridge

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/types"
)

type getPersonNodesRequest struct {
	ID          string `form:"id,default=-11"`
	CloudToken  int64  `binding:"required" form:"cloudToken"`
	CurrentPage int    `binding:"omitempty,min=1" form:"currentPage,default=1"`
	PageSize    int    `binding:"omitempty,min=1" form:"pageSize,default=30"`
}

type getPersonNodesResponse struct {
	Data        []*FileNode `json:"data"`
	Total       int64       `json:"total"`
	CurrentPage int         `json:"currentPage"`
	PageSize    int         `json:"pageSize"`
}

func (s *service) GetPersonNodes() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req = new(getPersonNodesRequest)

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

		resp, err := client.New().WithToken(client.NewAuthToken(token.AccessToken, token.ExpiresIn)).ListFiles(ctx, client.String(req.ID), func(req2 *client.ListFilesRequest) {
			req2.PageSize = req.PageSize
			req2.PageNum = req.CurrentPage
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.ErrResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})

			return
		}

		list := make([]*FileNode, 0)

		for _, file := range resp.FileListAO.FolderList {
			list = append(list, &FileNode{
				ID:       string(file.Id),
				IsFolder: 1,
				Name:     file.Name,
				ParentId: req.ID,
			})
		}

		for _, file := range resp.FileListAO.FileList {
			list = append(list, &FileNode{
				ID:       string(file.Id),
				IsFolder: 0,
				Name:     file.Name,
				ParentId: req.ID,
			})
		}

		ctx.JSON(http.StatusOK, &getPersonNodesResponse{
			CurrentPage: req.CurrentPage,
			Data:        list,
			PageSize:    req.PageSize,
			Total:       resp.FileListAO.FileListSize,
		})
	}
}
