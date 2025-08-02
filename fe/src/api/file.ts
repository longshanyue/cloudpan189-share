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

// 文件API
export const fileApi = {
  // 获取文件/文件夹信息
  getFile: (path: string = ''): Promise<FileItem> => {
    // 处理路径，确保正确的API调用
    if (path) {
      // 对路径进行编码，但保留路径分隔符
      const encodedPath = path.split('/').map(segment => encodeURIComponent(segment)).join('/')
      return api.get(`/open_file/${encodedPath}`)
    } else {
      return api.get('/open_file')
    }
  }
}