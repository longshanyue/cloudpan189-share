<template>
  <div class="bus-status" @mouseenter="handleMouseEnter" @mouseleave="handleMouseLeave">
    <!-- 状态指示器 -->
    <div class="status-indicator" :class="statusClass" @click="toggleDetail">
      <Icons name="activity" size="1rem" />
      <span class="status-text">{{ statusText }}</span>
    </div>

    <!-- 详细信息弹窗 -->
    <div v-if="showDetail" class="status-detail" @click.stop>
      <div class="detail-header">
        <h3>后台任务状态</h3>
        <button @click="refresh" class="refresh-btn" :disabled="loading">
          <Icons name="refresh" size="0.875rem" :class="{ 'animate-spin': loading }" />
        </button>
      </div>

      <!-- 统计信息 -->
      <div class="stats-grid">
        <div class="stat-item">
          <span class="stat-label">运行中</span>
          <span class="stat-value running">{{ busDetail?.stats.runningCount || 0 }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">等待中</span>
          <span class="stat-value pending">{{ busDetail?.stats.pendingCount || 0 }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">已完成</span>
          <span class="stat-value completed">{{ busDetail?.stats.completedCount || 0 }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">任务种类</span>
          <span class="stat-value">{{ busDetail?.stats.totalSubscribers || 0 }}</span>
        </div>
      </div>

      <!-- 运行中的任务 -->
      <div v-if="busDetail?.runningTasks.length" class="task-section">
        <h4>运行中的任务</h4>
        <div class="task-list">
          <div v-for="task in busDetail.runningTasks" :key="task.id" class="task-item running">
            <div class="task-info">
              <span class="task-topic">{{ getTopicName(task.topic) }}</span>
              <span class="task-id">{{ task.id }}</span>
            </div>
            <span class="task-time">{{ formatTime(task.startTime) }}</span>
          </div>
        </div>
      </div>

      <!-- 等待中的任务 -->
      <div v-if="busDetail?.pendingTasks.length" class="task-section">
        <h4>等待中的任务</h4>
        <div class="task-list">
          <div v-for="task in busDetail.pendingTasks" :key="task.id" class="task-item pending">
            <div class="task-info">
              <span class="task-topic">{{ getTopicName(task.topic) }}</span>
              <span class="task-id">{{ task.id }}</span>
            </div>
            <span class="task-time">{{ formatTime(task.startTime) }}</span>
          </div>
        </div>
      </div>

      <!-- 无任务时的提示 -->
      <div v-if="!busDetail?.runningTasks.length && !busDetail?.pendingTasks.length" class="no-tasks">
        <Icons name="check-circle" size="2rem" class="no-tasks-icon" />
        <span>当前没有运行中或等待中的任务</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { advancedOpsApi, type BusDetailInfo, BusTopicNames, type BusTopic } from '@/api/advancedops'
import Icons from './Icons.vue'

const showDetail = ref(false)
const loading = ref(false)
const busDetail = ref<BusDetailInfo | null>(null)
const isClickedOpen = ref(false) // 是否通过点击打开
let refreshTimer: number | null = null

// 状态类名
const statusClass = computed(() => {
  if (!busDetail.value) return 'idle'
  
  const { runningCount, pendingCount } = busDetail.value.stats
  
  if (runningCount > 0) return 'running'
  if (pendingCount > 0) return 'pending'
  return 'idle'
})

// 状态文本
const statusText = computed(() => {
  if (!busDetail.value) return '空闲'
  
  const { runningCount, pendingCount } = busDetail.value.stats
  
  if (runningCount > 0) return `运行中 ${runningCount}`
  if (pendingCount > 0) return `等待中 ${pendingCount}`
  return '空闲'
})

// 获取topic显示名称
const getTopicName = (topic: string): string => {
  return BusTopicNames[topic as BusTopic] || topic
}

// 格式化时间
const formatTime = (timeStr: string): string => {
  try {
    const time = new Date(timeStr)
    
    // 检查是否是有效日期
    if (isNaN(time.getTime())) {
      return '无效时间'
    }
    
    const now = new Date()
    const diff = now.getTime() - time.getTime()
    
    if (diff < 60000) return '刚刚'
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
    
    // 超过一天显示具体时间
    return time.toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (error) {
    console.error('时间格式化错误:', error, timeStr)
    return '时间错误'
  }
}

// 刷新数据
const refresh = async () => {
  if (loading.value) return
  
  loading.value = true
  try {
    busDetail.value = await advancedOpsApi.getBusDetail()
  } catch (error) {
    console.error('获取总线状态失败:', error)
  } finally {
    loading.value = false
  }
}

// 处理鼠标进入
const handleMouseEnter = () => {
  if (!isClickedOpen.value) {
    showDetail.value = true
  }
}

// 处理鼠标离开
const handleMouseLeave = () => {
  if (!isClickedOpen.value) {
    showDetail.value = false
  }
}

// 切换详情显示
const toggleDetail = () => {
  isClickedOpen.value = !isClickedOpen.value
  showDetail.value = isClickedOpen.value
  
  if (isClickedOpen.value) {
    // 点击打开时，下一帧添加全局点击监听
    nextTick(() => {
      document.addEventListener('click', handleClickOutside)
    })
  } else {
    // 点击关闭时，移除全局点击监听
    document.removeEventListener('click', handleClickOutside)
  }
}

// 处理点击外部区域
const handleClickOutside = (event: Event) => {
  const target = event.target as Element
  const busStatus = document.querySelector('.bus-status')
  
  if (busStatus && !busStatus.contains(target)) {
    isClickedOpen.value = false
    showDetail.value = false
    document.removeEventListener('click', handleClickOutside)
  }
}

// 自动刷新
const startAutoRefresh = () => {
  refresh()
  refreshTimer = window.setInterval(refresh, 3000) // 每3秒刷新一次
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

onMounted(() => {
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
  // 清理全局事件监听
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.bus-status {
  position: relative;
  display: flex;
  align-items: center;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
  font-weight: 500;
}

.status-indicator.idle {
  background: #f3f4f6;
  color: #6b7280;
}

.status-indicator.pending {
  background: #fef3c7;
  color: #d97706;
}

.status-indicator.running {
  background: #dbeafe;
  color: #2563eb;
}

.status-indicator:hover {
  opacity: 0.8;
}

.status-text {
  white-space: nowrap;
}

.status-detail {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 0.5rem;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  padding: 1rem;
  min-width: 320px;
  max-width: 400px;
  z-index: 1000;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.detail-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: #1f2937;
}

.refresh-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: #f3f4f6;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  color: #6b7280;
  transition: all 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: #e5e7eb;
  color: #374151;
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background: #f9fafb;
  border-radius: 4px;
}

.stat-label {
  font-size: 0.75rem;
  color: #6b7280;
}

.stat-value {
  font-weight: 600;
  font-size: 0.875rem;
}

.stat-value.running {
  color: #2563eb;
}

.stat-value.pending {
  color: #d97706;
}

.stat-value.completed {
  color: #059669;
}

.task-section {
  margin-bottom: 1rem;
}

.task-section:last-child {
  margin-bottom: 0;
}

.task-section h4 {
  margin: 0 0 0.5rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.task-list {
  max-height: 150px;
  overflow-y: auto;
}

.task-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  border-radius: 4px;
  margin-bottom: 0.25rem;
}

.task-item:last-child {
  margin-bottom: 0;
}

.task-item.running {
  background: #eff6ff;
  border-left: 3px solid #2563eb;
}

.task-item.pending {
  background: #fffbeb;
  border-left: 3px solid #d97706;
}

.task-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.task-topic {
  font-size: 0.75rem;
  font-weight: 500;
  color: #374151;
}

.task-id {
  font-size: 0.625rem;
  color: #6b7280;
  font-family: monospace;
}

.task-time {
  font-size: 0.625rem;
  color: #9ca3af;
}

.no-tasks {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem 1rem;
  color: #6b7280;
  text-align: center;
}

.no-tasks-icon {
  color: #10b981;
}
</style>