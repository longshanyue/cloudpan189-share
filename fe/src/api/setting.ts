// api/setting.ts
import api from './index'

// 设置相关接口类型定义
export interface Setting {
  id: number
  title: string
  enableAuth: boolean
  localProxy: boolean
  multipleStream: boolean
  runTimes: number  // 已经运行的时间
  createdAt: string
  updatedAt: string
  baseURL: string
  enableTopFileAutoRefresh: boolean
  initialized: boolean // 系统是否初始化完成
  jobThreadCount: number // 任务线程数
  autoRefreshMinutes: number // 自动刷新间隔（分钟）
  multipleStreamThreadCount: number // 多线程流线程数
  multipleStreamChunkSize: number // 多线程流块大小
  strmFileEnable: boolean // STRM文件启用状态
  strmSupportFileExtList: string[] // STRM支持的文件扩展名列表
  fileWritable: boolean // 文件可写状态
}

export interface InitSystemRequest {
  title: string
  enableAuth: boolean
  baseURL: string
  superUsername: string
  superPassword: string
}

export interface ModifyNameRequest {
  name: string
}

export interface ToggleAuthRequest {
  disable: boolean
}

export interface ToggleLocalProxyRequest {
  disable: boolean
}

export interface ToggleMultipleStreamRequest {
  disable: boolean
}

export interface ToggleEnableTopFileAutoRefreshRequest {
  disable: boolean
}

export interface ModifyBaseURLRequest {
  baseURL: string
}

export interface ModifyJobThreadCountRequest {
  threadCount: number
}

export interface ModifyAutoRefreshMinutesRequest {
  autoRefreshMinutes: number
}

// 修改多线程流线程数请求
export interface ModifyMultipleStreamThreadCountRequest {
  multipleStreamThreadCount: number // 1-64之间
}

// 修改多线程流块大小请求
export interface ModifyMultipleStreamChunkSizeRequest {
  multipleStreamChunkSize: number // 512KB-32MB之间
}

// 切换STRM文件启用状态请求
export interface ToggleStrmFileEnableRequest {
  strmFileEnable: boolean
}

// 修改STRM支持文件扩展名列表请求
export interface ModifyStrmSupportFileExtListRequest {
  strmSupportFileExtList: string[] // 可选，不传或空数组表示清空列表
}

// 新增：切换文件可写状态请求
export interface ToggleFileWritableRequest {
  fileWritable: boolean
}

// 修改操作的通用响应
export interface ModifyResponse {
  rowsAffected: number
}

// 设置API
export const settingApi = {
  // 获取设置
  getSetting: (): Promise<Setting> => {
    return api.get('/setting/get')
  },

  // 修改网站名称
  modifyName: (data: ModifyNameRequest): Promise<void> => {
    return api.post('/setting/modify_name', data)
  },

  // 刷新密钥
  refreshKey: (): Promise<void> => {
    return api.post('/setting/refresh_key')
  },

  // 切换认证状态
  toggleAuth: (data: ToggleAuthRequest): Promise<void> => {
    return api.post('/setting/toggle_auth', data)
  },

  // 切换本地代理状态
  toggleLocalProxy: (data: ToggleLocalProxyRequest): Promise<void> => {
    return api.post('/setting/toggle_local_proxy', data)
  },

  // 切换多线程流加速状态
  toggleMultipleStream: (data: ToggleMultipleStreamRequest): Promise<void> => {
    return api.post('/setting/toggle_multiple_stream', data)
  },

  // 修改基础URL
  modifyBaseURL: (data: ModifyBaseURLRequest): Promise<void> => {
    return api.post('/setting/modify_base_url', data)
  },

  // 切换挂载文件自动刷新状态
  toggleEnableTopFileAutoRefresh: (data: ToggleEnableTopFileAutoRefreshRequest): Promise<void> => {
    return api.post('/setting/toggle_enable_top_file_auto_refresh', data)
  },

  // 修改任务线程数
  modifyJobThreadCount: (data: ModifyJobThreadCountRequest): Promise<void> => {
    return api.post('/setting/modify_job_thread_count', data)
  },

  // 修改自动刷新间隔
  modifyAutoRefreshMinutes: (data: ModifyAutoRefreshMinutesRequest): Promise<void> => {
    return api.post('/setting/modify_auto_refresh_minutes', data)
  },

  // 修改多线程流线程数
  modifyMultipleStreamThreadCount: (data: ModifyMultipleStreamThreadCountRequest): Promise<ModifyResponse> => {
    return api.post('/setting/modify_multiple_stream_thread_count', data)
  },

  // 修改多线程流块大小
  modifyMultipleStreamChunkSize: (data: ModifyMultipleStreamChunkSizeRequest): Promise<ModifyResponse> => {
    return api.post('/setting/modify_multiple_stream_chunk_size', data)
  },

  // 切换STRM文件启用状态
  toggleStrmFileEnable: (data: ToggleStrmFileEnableRequest): Promise<ModifyResponse> => {
    return api.post('/setting/toggle_strm_file_enable', data)
  },

  // 修改STRM支持文件扩展名列表
  modifyStrmSupportFileExtList: (data: ModifyStrmSupportFileExtListRequest): Promise<ModifyResponse> => {
    return api.post('/setting/modify_strm_support_file_ext_list', data)
  },

  // 新增：切换文件可写状态
  toggleFileWritable: (data: ToggleFileWritableRequest): Promise<ModifyResponse> => {
    return api.post('/setting/toggle_file_writable', data)
  },

  // 初始化系统
  initSystem: (data: InitSystemRequest): Promise<void> => {
    return api.post('/setting/init_system', data)
  }
}