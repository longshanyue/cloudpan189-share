import api from './index'

// 高级操作相关接口类型定义

// 通用响应类型
export interface AdvancedOpsResponse {
  code: number
  message: string
}

// 重建 STRM 文件响应
export interface RebuildStrmResponse extends AdvancedOpsResponse {}

// 清理媒体文件响应
export interface ClearMediaResponse extends AdvancedOpsResponse {}

// 高级操作 API
export const advancedOpsApi = {
  // 重建 STRM 文件
  rebuildStrm: (): Promise<RebuildStrmResponse> => {
    return api.post('/advanced_ops/rebuild_strm')
  },

  // 清理媒体文件
  clearMedia: (): Promise<ClearMediaResponse> => {
    return api.post('/advanced_ops/clear_media')
  },

}
