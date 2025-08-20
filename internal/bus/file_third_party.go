package bus

import (
	"context"
	"errors"
	"fmt"
	"github.com/xxcheng123/cloudpan189-interface/client"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/models"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
	"strings"
	"sync"
)

func (w *fileBusWorker) getSubscribeUserFiles(ctx context.Context, f File) ([]File, error) {
	_userId, ok := f.Addition[consts.FileAdditionKeySubscribeUser]
	if !ok {
		return nil, errors.New("no subscribe_user")
	}

	userId := utils.String(_userId)

	var (
		pageNum  int64 = 1
		pageSize int64 = 200
		files          = make([]File, 0)
	)

	resp, err := w.client.GetUpResourceShare(ctx, userId, pageNum, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	if resp.Data != nil {
		for _, v := range resp.Data.FileList {
			files = append(files, &models.VirtualFile{
				ParentId:   f.ID,
				Name:       v.Name,
				IsTop:      0,
				Size:       v.Size,
				IsFolder:   int8(v.Folder),
				Hash:       strings.ToLower(v.Md5),
				CreateDate: v.CreateDate,
				ModifyDate: v.LastOpTime,
				OsType:     models.OsTypeSubscribeShare,
				Addition: map[string]any{
					consts.FileAdditionKeySubscribeUser: userId,
					consts.FileAdditionKeyShareId:       v.ShareId,
					consts.FileAdditionKeyFileId:        v.Id,
				},
				Rev: v.Rev,
			})
		}
	}

	if resp.Data != nil && int64(len(files)) < resp.Data.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]File
		)

		totalPages := (resp.Data.Count + pageSize - 1) / pageSize
		allFiles = make([][]File, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := w.client.GetUpResourceShare(ctx, userId, pageNum, pageSize)
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				if subResp.Data != nil {
					var pageFiles []File
					for _, v := range subResp.Data.FileList {
						pageFiles = append(pageFiles, &models.VirtualFile{
							ParentId:   f.ID,
							Name:       v.Name,
							IsTop:      0,
							Size:       v.Size,
							IsFolder:   int8(v.Folder),
							Hash:       strings.ToLower(v.Md5),
							CreateDate: v.CreateDate,
							ModifyDate: v.LastOpTime,
							OsType:     models.OsTypeSubscribeShare,
							Addition: map[string]any{
								consts.FileAdditionKeySubscribeUser: userId,
								consts.FileAdditionKeyShareId:       v.ShareId,
								consts.FileAdditionKeyFileId:        v.Id,
							},
							Rev: v.Rev,
						})
					}
					allFiles[index] = pageFiles
				}
			}(i, int(i-2))
		}

		wg.Wait()

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}

		for _, pageFiles := range allFiles {
			files = append(files, pageFiles...)
		}
	}

	return files, nil
}

func (w *fileBusWorker) getSubscribeShareFiles(ctx context.Context, f File) ([]File, error) {
	_userId, ok := f.Addition[consts.FileAdditionKeySubscribeUser]
	if !ok {
		return nil, errors.New("no subscribe_user")
	}

	_shareId, ok := f.Addition[consts.FileAdditionKeyShareId]
	if !ok {
		return nil, errors.New("no share_id")
	}

	_fileId, ok := f.Addition[consts.FileAdditionKeyFileId]
	if !ok {
		return nil, errors.New("no file_id")
	}

	var (
		userId     = utils.String(_userId)
		shareId, _ = utils.Int64(_shareId)
		fileId     = utils.String(_fileId)
		pageNum    = 1
		pageSize   = 200
		files      = make([]File, 0)
	)

	resp, err := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
		req.PageNum = pageNum
		req.PageSize = pageSize
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	for _, v := range resp.FileListAO.FolderList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       0,
			IsFolder:   1,
			Hash:       "",
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeSubscribeShare,
			Addition: map[string]any{
				consts.FileAdditionKeySubscribeUser: userId,
				consts.FileAdditionKeyShareId:       shareId,
				consts.FileAdditionKeyFileId:        v.Id,
			},
			Rev: v.Rev,
		})
	}

	for _, v := range resp.FileListAO.FileList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       v.Size,
			IsFolder:   0,
			Hash:       strings.ToLower(v.Md5),
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeFile,
			Addition: map[string]any{
				consts.FileAdditionKeySubscribeUser: userId,
				consts.FileAdditionKeyShareId:       shareId,
				consts.FileAdditionKeyFileId:        v.Id,
			},
			Rev: v.Rev,
		})
	}

	if int64(len(files)) < resp.FileListAO.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]File
		)

		totalPages := (resp.FileListAO.Count + int64(pageSize) - 1) / int64(pageSize)
		allFiles = make([][]File, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
					req.PageNum = int(pageNum)
					req.PageSize = pageSize
				})
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				var pageFiles []File
				for _, v := range subResp.FileListAO.FolderList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       0,
						IsFolder:   1,
						Hash:       "",
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeSubscribeShare,
						Addition: map[string]any{
							consts.FileAdditionKeySubscribeUser: userId,
							consts.FileAdditionKeyShareId:       shareId,
							consts.FileAdditionKeyFileId:        v.Id,
						},
						Rev: v.Rev,
					})
				}

				for _, v := range subResp.FileListAO.FileList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       v.Size,
						IsFolder:   0,
						Hash:       strings.ToLower(v.Md5),
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeFile,
						Addition: map[string]any{
							consts.FileAdditionKeySubscribeUser: userId,
							consts.FileAdditionKeyShareId:       shareId,
							consts.FileAdditionKeyFileId:        v.Id,
						},
						Rev: v.Rev,
					})
				}
				allFiles[index] = pageFiles
			}(i, int(i-2))
		}

		wg.Wait()

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}

		for _, pageFiles := range allFiles {
			files = append(files, pageFiles...)
		}
	}

	return files, nil
}

func (w *fileBusWorker) getShareFiles(ctx context.Context, f File) ([]File, error) {
	var vv, ok = f.Addition[consts.FileAdditionKeyShareId]
	if !ok {
		return nil, errors.New("no share_id")
	}

	shareId, _ := utils.Int64(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyFileId]
	if !ok {
		return nil, errors.New("no file_id")
	}

	fileId := utils.String(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyShareMode]
	if !ok {
		return nil, errors.New("no share_mode")
	}

	shareMode, _ := utils.Int(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyAccessCode]
	if !ok {
		return nil, errors.New("no access_code")
	}

	accessCode := utils.String(vv)

	vv, ok = f.Addition[consts.FileAdditionKeyIsFolder]
	if !ok {
		return nil, errors.New("no is_folder")
	}

	var (
		pageNum  = 1
		pageSize = 200
		files    = make([]File, 0)
		addMpFn  = func(mp map[string]any) map[string]any {
			mp[consts.FileAdditionKeyShareId] = shareId
			mp[consts.FileAdditionKeyShareMode] = shareMode
			mp[consts.FileAdditionKeyAccessCode] = accessCode

			return mp
		}
	)

	resp, err := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
		req.PageNum = pageNum
		req.PageSize = pageSize
		req.IsFolder, _ = utils.Bool(vv)
		req.AccessCode = accessCode
		req.ShareMode = shareMode
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}

	for _, v := range resp.FileListAO.FolderList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       0,
			IsFolder:   1,
			Hash:       "",
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeShare,
			Addition: addMpFn(map[string]any{
				consts.FileAdditionKeyFileId:   v.Id,
				consts.FileAdditionKeyIsFolder: true,
			}),
			Rev: v.Rev,
		})
	}

	for _, v := range resp.FileListAO.FileList {
		files = append(files, &models.VirtualFile{
			ParentId:   f.ID,
			Name:       v.Name,
			IsTop:      0,
			Size:       v.Size,
			IsFolder:   0,
			Hash:       strings.ToLower(v.Md5),
			CreateDate: v.CreateDate,
			ModifyDate: v.LastOpTime,
			OsType:     models.OsTypeFile,
			Addition: addMpFn(map[string]any{
				consts.FileAdditionKeyFileId:   v.Id,
				consts.FileAdditionKeyIsFolder: false,
			}),
			Rev: v.Rev,
		})
	}

	if int64(len(files)) < resp.FileListAO.Count {
		var (
			mu       sync.Mutex
			wg       sync.WaitGroup
			errs     []error
			allFiles [][]File
		)

		totalPages := (resp.FileListAO.Count + int64(pageSize) - 1) / int64(pageSize)
		allFiles = make([][]File, totalPages-1)

		for i := int64(2); i <= totalPages; i++ {
			wg.Add(1)
			go func(pageNum int64, index int) {
				defer wg.Done()

				subResp, subErr := w.client.ListShareDir(ctx, shareId, client.String(fileId), func(req *client.ListShareFileRequest) {
					req.PageNum = int(pageNum)
					req.PageSize = pageSize
				})
				if subErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to get page %d: %w", pageNum, subErr))
					mu.Unlock()
					return
				}

				var pageFiles []File
				for _, v := range subResp.FileListAO.FolderList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       0,
						IsFolder:   1,
						Hash:       "",
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeShare,
						Addition: addMpFn(map[string]any{
							consts.FileAdditionKeyFileId:   v.Id,
							consts.FileAdditionKeyIsFolder: true,
						}),
						Rev: v.Rev,
					})
				}

				for _, v := range subResp.FileListAO.FileList {
					pageFiles = append(pageFiles, &models.VirtualFile{
						ParentId:   f.ID,
						Name:       v.Name,
						IsTop:      0,
						Size:       v.Size,
						IsFolder:   0,
						Hash:       strings.ToLower(v.Md5),
						CreateDate: v.CreateDate,
						ModifyDate: v.LastOpTime,
						OsType:     models.OsTypeFile,
						Addition: addMpFn(map[string]any{
							consts.FileAdditionKeyFileId:   v.Id,
							consts.FileAdditionKeyIsFolder: false,
						}),
						Rev: v.Rev,
					})
				}
				allFiles[index] = pageFiles
			}(i, int(i-2))
		}

		wg.Wait()

		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}

		for _, pageFiles := range allFiles {
			files = append(files, pageFiles...)
		}
	}

	return files, nil
}
