import api from './index'
import { JobStat } from './shared'

// 存储相关接口类型定义
export interface Storage {
  id: number
  parentId: number
  name: string
  osType: string //  subscribe：订阅类型；share：分享类型
  createdAt: string
  updatedAt: string
  localPath: string
  addition: {
    cloud_token?: number
    subscribe_user?: string
    share_code?: string
    share_access_code?: string
    disable_auto_scan?: boolean
  }
  jobStatus?: JobStat
}

export interface AddStorageRequest {
  localPath: string
  protocol: string  // 目前只允许 subscribe、share
  cloudToken?: number 
  subscribeUser?: string // subscribe 时必填
  shareCode?: string // share 时必填
  shareAccessCode?: string // share 时必填
}

export interface AddStorageResponse {
  id: number
}

export interface DeleteStorageRequest {
  id: number
}

export interface ModifyTokenRequest {
  id: number
  cloudToken: number
}

export interface StorageListRequest {
  currentPage?: number
  pageSize?: number
  name?: string
  noPaginate?: boolean
}

export interface StorageListResponse {
  total: number
  currentPage: number
  pageSize: number
  data: Storage[]
}

export interface DeepRefreshFileRequest {
  id: number
}

export interface DeepRefreshFileResponse {
  code: number
  msg: string
}

export interface VirtualFile {
  id: number
  parentId: number
  name: string
  isTop: number
  size: number
  isFolder: number
  hash: string
  createDate: string
  modifyDate: string
  osType: string
  addition: Record<string, any>
  rev: string
  createdAt: string
  updatedAt: string
}

export interface SearchItem {
  id: number
  parentId: number
  name: string
  isTop: number
  size: number
  isFolder: number
  hash: string
  createDate: string
  modifyDate: string
  osType: string
  rev: string
  createdAt: string
  updatedAt: string
  localPath: string
  path: string
  href: string
}

export interface SearchRequest {
  keyword: string
  pid?: number
  global?: boolean
  pageSize: number
  currentPage: number
}

export interface SearchResponse {
  total: number
  currentPage: number
  pageSize: number
  data: SearchItem[]
}

export interface ClearRealFileResponse {
  message: string
}

export interface ToggleAutoScanRequest {
  id: number
  disableAutoScan: boolean
}

export interface ToggleAutoScanResponse {
  code: number
  msg: string
}

// 存储API
export const storageApi = {
  // 添加存储
  add: (data: AddStorageRequest): Promise<AddStorageResponse> => {
    return api.post('/storage/add', data)
  },

  // 删除存储
  delete: (data: DeleteStorageRequest): Promise<void> => {
    return api.post('/storage/delete', data)
  },

  // 修改令牌绑定
  modifyToken: (data: ModifyTokenRequest): Promise<void> => {
    return api.post('/storage/modify_token', data)
  },

  // 获取存储列表
  list: (params?: StorageListRequest): Promise<StorageListResponse> => {
    return api.get('/storage/list', { params })
  },

  // 刷新文件夹
  deepRefreshFile: (data: DeepRefreshFileRequest): Promise<DeepRefreshFileResponse> => {
    return api.post('/storage/deep_refresh_file', data)
  },

  // 搜索文件
  search: (params: SearchRequest): Promise<SearchResponse> => {
    return api.get('/storage/file/search', { params })
  },

  // 清空本地真实存储
  clearRealFile: (): Promise<ClearRealFileResponse> => {
    return api.post('/storage/clear_real_file')
  },

  // 切换自动扫描设置
  toggleAutoScan: (data: ToggleAutoScanRequest): Promise<ToggleAutoScanResponse> => {
    return api.post('/storage/toggle_auto_scan', data)
  }
}
