import api from './index'

export interface UserGroup {
  id: number
  name: string
  createdAt: string
  updatedAt: string
}

export interface AddUserGroupRequest {
  name: string
}

export interface AddUserGroupResponse {
  id: number
}

export interface DeleteUserGroupRequest {
  id: number
}

export interface DeleteUserGroupResponse {
  rowsAffected: number
}

export interface ModifyUserGroupNameRequest {
  id: number
  name: string
}

export interface ModifyUserGroupNameResponse {
  rowsAffected: number
}

export interface BatchBindFilesRequest {
  groupId: number
  fileIds: number[]
}

export interface BatchBindFilesResponse {
  groupId: number
  bind_count: number
  deletedCount: number
}

export interface UserGroupListRequest {
  currentPage?: number
  pageSize?: number
  noPaginate?: boolean
  name?: string
}

export interface UserGroupListResponse {
  total: number
  currentPage: number
  pageSize: number
  data: UserGroup[]
}

export interface GetBindFilesRequest {
  groupId: number
}

export interface BindFileInfo {
  id?: number
  fileId: number
  groupId: number
  createdAt?: string
  updatedAt?: string
  name?: string
}

export interface GetBindFilesResponse {
  groupId: number
  fileCount: number
  files: BindFileInfo[]
}

// 用户组API
export const userGroupApi = {
  // 添加用户组
  addUserGroup: (data: AddUserGroupRequest): Promise<AddUserGroupResponse> => {
    return api.post('/user_group/add', data)
  },

  // 删除用户组
  deleteUserGroup: (data: DeleteUserGroupRequest): Promise<DeleteUserGroupResponse> => {
    return api.post('/user_group/delete', data)
  },

  // 修改用户组名称
  modifyUserGroupName: (data: ModifyUserGroupNameRequest): Promise<ModifyUserGroupNameResponse> => {
    return api.post('/user_group/modify_name', data)
  },

  // 批量绑定文件
  batchBindFiles: (data: BatchBindFilesRequest): Promise<BatchBindFilesResponse> => {
    return api.post('/user_group/batch_bind_files', data)
  },

  // 获取用户组列表
  getUserGroupList: (params: UserGroupListRequest): Promise<UserGroupListResponse> => {
    return api.post('/user_group/list', params)
  },

  // 获取用户组绑定的文件列表
  getBindFiles: (params: GetBindFilesRequest): Promise<GetBindFilesResponse> => {
    return api.get('/user_group/bind_files', { params })
  },
}
