import api from './index'

// 文件/文件夹信息接口
export interface FileItem {
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
  path: string
  href: string
  downloadURL?: string // 文件才有
  children?: FileItem[] // 文件夹才有
  addition?: {
    file_id: string
    share_id: number
    subscribe_user: string
  }
}

// 获取文件请求参数接口
export interface GetFileRequest {
  includeAutoGenerateStrmFile?: boolean // 是否包括STRM文件
}

// 删除文件响应接口
export interface DeleteFileResponse {
  code: number
  message: string
}

// 文件API
export const fileApi = {
  // 获取文件/文件夹信息
  getFile: (path: string = '', options?: GetFileRequest): Promise<FileItem> => {
    // 处理路径，确保正确的API调用
    let url = ''
    if (path) {
      // 对路径进行编码，但保留路径分隔符
      const encodedPath = path.split('/').map(segment => encodeURIComponent(segment)).join('/')
      url = `/open_file/${encodedPath}`
    } else {
      url = '/open_file'
    }

    // 如果有查询参数，添加到URL中
    if (options) {
      const params = new URLSearchParams()

      if (options.includeAutoGenerateStrmFile !== undefined) {
        params.append('includeAutoGenerateStrmFile', options.includeAutoGenerateStrmFile.toString())
      }

      const queryString = params.toString()
      if (queryString) {
        url += `?${queryString}`
      }
    }

    return api.get(url)
  },

  // 删除文件/文件夹
  deleteFile: (path: string): Promise<DeleteFileResponse> => {
    if (!path) {
      throw new Error('删除路径不能为空')
    }

    // 对路径进行编码，但保留路径分隔符
    const encodedPath = path.split('/').map(segment => encodeURIComponent(segment)).join('/')
    const url = `/open_file/${encodedPath}`

    return api.delete(url)
  }
}