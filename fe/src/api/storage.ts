import api from './index'

// 文件扫描统计信息
export interface FileScanStat {
  fileId: number
  waitCount: number
  scannedCount: number
}

// 存储相关接口类型定义
export interface Storage {
  id: number
  parentId: number
  name: string
  osType: string //  subscribe：订阅类型；share：分享类型；person：个人类型；family：家庭类型
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
  fileScanStat?: FileScanStat
}

export interface PreAddStorageRequest {
  protocol: string  // 目前只允许 subscribe、share
  subscribeUser?: string // subscribe 时必填
  shareCode?: string // share 时必填
  shareAccessCode?: string // share 时可选
  cloudToken?: number
}

export interface PreAddStorageResponse {
  name: string
  protocol: string
}

export interface AddStorageRequest {
  localPath: string
  protocol: string  // 目前允许 subscribe、share、person、family
  cloudToken?: number // person、family 时必填
  subscribeUser?: string // subscribe 时必填
  shareCode?: string // share 时必填
  shareAccessCode?: string // share 时可选
  fileId?: string // person、family 时必填
  familyId?: string // family 时必填
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

export interface BatchBindTokenRequest {
  ids: number[]
  cloudToken: number
}

export interface BatchBindTokenResponse {
  successCount: number
  failedCount: number
  failedFiles: number[]
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

export interface ToggleAutoScanRequest {
  id: number
  disableAutoScan: boolean
}

export interface ToggleAutoScanResponse {
  code: number
  msg: string
}

export interface ScanTopResponse {
  message: string
}

// 存储API
export const storageApi = {
  // 预添加存储（获取存储信息）
  preAdd: (data: PreAddStorageRequest): Promise<PreAddStorageResponse> => {
    return api.post('/storage/pre_add', data)
  },

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

  // 批量绑定令牌
  batchBindToken: (data: BatchBindTokenRequest): Promise<BatchBindTokenResponse> => {
    return api.post('/storage/batch_bind_token', data)
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

  // 切换自动扫描设置
  toggleAutoScan: (data: ToggleAutoScanRequest): Promise<ToggleAutoScanResponse> => {
    return api.post('/storage/toggle_auto_scan', data)
  },

  // 扫描顶层文件
  scanTop: (): Promise<ScanTopResponse> => {
    return api.post('/storage/scan_top')
  }
}
