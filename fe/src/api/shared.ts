export type JobType = 'del' | 'refresh' | 'deep_refresh';

export const JobStatusMap = {
  del: '删除',
  refresh: '文件扫描中',
  deep_refresh: '文件重建索引'
}

type JobStatus = 0 | 1; // 0: waiting, 1: running

export interface JobStat {
    fileId: number
    jobType: JobType
    status: JobStatus
    generateTime: string
    startTime: string
    waitCount?: number
    scannedCount?: number
}