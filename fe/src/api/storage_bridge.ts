import api from './index'

// 文件节点类型定义
export interface FileNode {
  parentId: string
  id: string
  name: string
  isFolder: number // 0: 文件, 1: 文件夹
}

// 获取个人文件节点请求参数
export interface GetPersonNodesRequest {
  id?: string // 默认为 -11
  cloudToken: number
  currentPage?: number // 默认为 1
  pageSize?: number // 默认为 30
}

// 获取个人文件节点响应
export interface GetPersonNodesResponse {
  data: FileNode[]
  total: number
  currentPage: number
  pageSize: number
}

// 获取家庭文件节点请求参数
export interface GetFamilyNodesRequest {
  id?: string
  familyId?: string
  cloudToken: number
  currentPage?: number // 默认为 1
  pageSize?: number // 默认为 30
}

// 获取家庭文件节点响应
export interface GetFamilyNodesResponse {
  data: FileNode[]
  total: number
  currentPage: number
  pageSize: number
}

// 家庭列表请求参数
export interface FamilyListRequest {
  cloudToken: number
}

// 家庭信息
export interface FamilyInfo {
  count: number
  createTime: string
  familyId: string
  remarkName: string
  type: number
  useFlag: number
  userRole: number
  expireTime?: string
}

// 家庭列表响应
export type FamilyListResponse = FamilyInfo[]

// 存储桥接 API
export const storageBridgeApi = {
  // 获取个人文件节点
  getPersonNodes: (params: GetPersonNodesRequest): Promise<GetPersonNodesResponse> => {
    return api.get('/storage/bridge/get_person_nodes', { params })
  },

  // 获取家庭文件节点
  getFamilyNodes: (params: GetFamilyNodesRequest): Promise<GetFamilyNodesResponse> => {
    return api.get('/storage/bridge/get_family_nodes', { params })
  },

  // 获取家庭列表
  getFamilyList: (params: FamilyListRequest): Promise<FamilyListResponse> => {
    return api.get('/storage/bridge/family_file', { params })
  }
}