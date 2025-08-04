// api/setting.ts
import api from './index'

// 设置相关接口类型定义
export interface Setting {
  id: number
  title: string
  enableAuth: boolean
  localProxy: boolean
  multipleStream: boolean
  runTimes: number  // 后端实际返回的字段名
  createdAt: string
  updatedAt: string
  baseURL: string
  enableTopFileAutoRefresh: boolean
  initialized: boolean // 系统是否初始化完成
  jobThreadCount: number // 任务线程数
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

  // 初始化系统
  initSystem: (data: InitSystemRequest): Promise<void> => {
    return api.post('/setting/init_system', data)
  }
}