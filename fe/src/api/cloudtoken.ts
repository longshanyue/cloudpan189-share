import api from './index'

// 云盘令牌相关接口类型定义
export interface CloudToken {
  id: number
  name: string
  accessToken: string
  expiresIn: number
  status: number // 状态 1:正常 2: 失败
  loginType: number // 1: 扫码登录 2: 密码登录
  username: string
  createdAt: string
  updatedAt: string
}

export interface InitQrcodeResponse {
  uuid: string
}

export interface CheckQrcodeRequest {
  id?: number
  uuid: string
}

export interface ModifyNameRequest {
  id: number
  name: string
}

export interface DeleteRequest {
  id: number
}

export interface ListRequest {
  name?: string
}

export interface CheckQrcodeResponse {
  code: number
  msg: string
}

export interface UsernameLoginRequest {
  id?: number // 新建时不传
  username: string
  password: string
}

export interface UsernameLoginResponse {
  id?: number
  rowsAffected?: number
}

// 云盘令牌API
export const cloudTokenApi = {
  // 初始化二维码
  initQrcode: (): Promise<InitQrcodeResponse> => {
    return api.post('/cloud_token/init_qrcode')
  },

  // 检查二维码状态
  checkQrcode: (data: CheckQrcodeRequest): Promise<CheckQrcodeResponse> => {
    return api.post('/cloud_token/check_qrcode', data)
  },

  // 修改令牌名称
  modifyName: (data: ModifyNameRequest): Promise<void> => {
    return api.post('/cloud_token/modify_name', data)
  },

  // 删除令牌
  delete: (data: DeleteRequest): Promise<void> => {
    return api.post('/cloud_token/delete', data)
  },

  // 获取令牌列表
  list: (params?: ListRequest): Promise<CloudToken[]> => {
    return api.get('/cloud_token/list', { params })
  },

  // 密码登录
  usernameLogin: (data: UsernameLoginRequest): Promise<UsernameLoginResponse> => {
    return api.post('/cloud_token/username_login', data)
  }
}