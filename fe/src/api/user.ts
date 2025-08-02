import api from './index'

// 用户相关接口类型定义
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
  user: User
}

export interface User {
  id: number
  username: string
  status: number
  permissions: number
  version: number
  createdAt: string
  updatedAt: string
}

export interface RefreshTokenRequest {
  refreshToken: string
}

export interface AddUserRequest {
  username: string
  password: string
  is_super: number
}

export interface UpdateUserRequest {
  id: number
  permissions?: number
}

export interface ModifyPasswordRequest {
  id: number
  password: string
}

export interface ModifyOwnPasswordRequest {
  password: string
  oldPassword: string
}

export interface UserListRequest {
  currentPage?: number
  pageSize?: number
  noPaginate?: boolean
  username?: string
}

export interface UserListResponse {
  total: number
  currentPage: number
  pageSize: number
  data: User[]
}

// 用户API
export const userApi = {
  // 登录
  login: (data: LoginRequest): Promise<LoginResponse> => {
    return api.post('/user/login', data)
  },

  // 刷新token
  refreshToken: (data: RefreshTokenRequest): Promise<LoginResponse> => {
    return api.post('/user/refresh_token', data)
  },

  // 获取用户信息
  getUserInfo: (): Promise<User> => {
    return api.get('/user/info')
  },

  // 添加用户
  addUser: (data: AddUserRequest): Promise<{ id: number }> => {
    return api.post('/user/add', data)
  },

  // 删除用户
  deleteUser: (id: number): Promise<void> => {
    return api.post('/user/del', { id })
  },

  // 更新用户
  updateUser: (data: UpdateUserRequest): Promise<void> => {
    return api.post('/user/update', data)
  },

  // 管理员修改用户密码
  modifyPassword: (data: ModifyPasswordRequest): Promise<void> => {
    return api.post('/user/modify_pass', data)
  },

  // 用户修改自己的密码
  modifyOwnPassword: (data: ModifyOwnPasswordRequest): Promise<void> => {
    return api.post('/user/modify_own_pass', data)
  },

  // 获取用户列表
  getUserList: (params: UserListRequest): Promise<UserListResponse> => {
    return api.get('/user/list', { params })
  },
}