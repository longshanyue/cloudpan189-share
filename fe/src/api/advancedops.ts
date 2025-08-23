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

// Topic 枚举
export enum BusTopic {
  FILE_REFRESH_FILE = 'topic::file::refresh::file',
  FILE_DELETE_FILE = 'topic::file::delete::file',
  FILE_SCAN_TOP = 'topic::file::scan::top',
  FILE_REBUILD_MEDIA_FILE = 'file::rebuild::media::file',
  MEDIA_ADD_STRM_FILE = 'topic::media::add::strm::file',
  MEDIA_DELETE_LINK_FILE = 'topic::media::delete::link::file',
  MEDIA_CLEAR_EMPTY_DIR = 'topic::media::clear::empty::dir',
  MEDIA_CLEAR_ALL_MEDIA = 'topic::media::clear::all::media'
}

// Topic 显示名称映射
export const BusTopicNames: Record<BusTopic, string> = {
  [BusTopic.FILE_REFRESH_FILE]: '刷新文件',
  [BusTopic.FILE_DELETE_FILE]: '删除文件',
  [BusTopic.FILE_SCAN_TOP]: '扫描所有文件',
  [BusTopic.FILE_REBUILD_MEDIA_FILE]: '重建媒体文件',
  [BusTopic.MEDIA_ADD_STRM_FILE]: '添加STRM文件',
  [BusTopic.MEDIA_DELETE_LINK_FILE]: '删除链接文件',
  [BusTopic.MEDIA_CLEAR_EMPTY_DIR]: '清理空目录',
  [BusTopic.MEDIA_CLEAR_ALL_MEDIA]: '清理所有媒体'
}

// 任务状态枚举
export enum TaskStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed'
}

// 任务状态显示名称映射
export const TaskStatusNames: Record<TaskStatus, string> = {
  [TaskStatus.PENDING]: '等待中',
  [TaskStatus.RUNNING]: '运行中',
  [TaskStatus.COMPLETED]: '已完成'
}

// 任务信息
export interface TaskInfo {
  id: string
  topic: BusTopic
  status: TaskStatus
  startTime: string
  data?: any
}

// 总线统计信息
export interface BusStats {
  runningCount: number
  pendingCount: number
  completedCount: number
  queueLength: number
  activeWorkers: number
  totalSubscribers: number
}

// 总线详细信息
export interface BusDetailInfo {
  runningTasks: TaskInfo[]
  pendingTasks: TaskInfo[]
  stats: BusStats
}

// 获取总线详情响应
export interface BusDetailResponse extends BusDetailInfo {}

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

  // 获取总线详情
  getBusDetail: (): Promise<BusDetailResponse> => {
    return api.get('/advanced_ops/bus_detail')
  },

}
