<template>
  <Layout>
    <!-- 系统设置主卡片 -->
    <PageCard title="系统设置" subtitle="管理系统配置和参数设置">
      <SectionDivider />
      
      <SubsectionTitle title="基本设置" />
        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">网站名称</span>
            <span class="label-desc">设置网站的显示名称</span>
          </div>
          <div class="setting-control">
            <input 
              v-model="websiteName" 
              type="text" 
              class="setting-input"
              placeholder="请输入网站名称"
              :disabled="loading"
            >
            <button 
              @click="handleModifyName" 
              class="btn btn-primary btn-sm"
              :disabled="loading || !websiteName.trim() || websiteName === originalWebsiteName"
            >
              {{ loading ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">基础URL</span>
            <span class="label-desc">设置系统访问的基础URL，用于生成分享链接等</span>
          </div>
          <div class="setting-control">
            <input 
              v-model="baseURL" 
              type="text" 
              class="setting-input"
              placeholder="请输入基础URL"
              :disabled="loading"
            >
            <button 
              @click="handleAutoFillBaseURL" 
              class="btn btn-secondary btn-sm"
              :disabled="loading"
            >
              自动获取
            </button>
            <button 
              @click="handleModifyBaseURL" 
              class="btn btn-primary btn-sm"
              :disabled="loading || !baseURL.trim() || baseURL === originalBaseURL"
            >
              {{ loading ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
        
        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">系统运行时间</span>
            <span class="label-desc">当前系统已运行时间</span>
          </div>
          <div class="setting-value">{{ formatRunTime(settingStore.setting?.runTimes || 0) }}</div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">更新时间</span>
            <span class="label-desc">配置上一次更新的时间</span>
          </div>
          <div class="setting-value">{{ formatDate(settingStore.setting?.updatedAt) }}</div>
        </div>
        
        <SectionDivider />
        
        <SubsectionTitle title="功能设置" />
        <div class="setting-item">
            <div class="setting-label">
              <span class="label-text">用户认证</span>
              <span class="label-desc">开启后访问WebDAV需要用户登录认证</span>
            </div>
          <div class="setting-control">
            <div class="custom-switch" @click="handleToggleAuth">
              <input 
                type="checkbox" 
                :checked="settingStore.setting?.enableAuth" 
                :disabled="loading"
                class="switch-input"
              >
              <span class="switch-slider"></span>
            </div>
          </div>
        </div>
        
        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">本地代理</span>
            <span class="label-desc">系统默认通过302跳转方式提供资源，开启后服务器代为获取资源再转发给用户</span>
          </div>
          <div class="setting-control">
            <div class="custom-switch" @click="handleToggleLocalProxy">
              <input 
                type="checkbox" 
                :checked="settingStore.setting?.localProxy" 
                :disabled="loading"
                class="switch-input"
              >
              <span class="switch-slider"></span>
            </div>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">多线程流式下载</span>
            <span class="label-desc">使用多个连接并发下载视频片段，提升下载速度和播放流畅度，优先级高于本地代理设置</span>
          </div>
          <div class="setting-control">
            <div class="custom-switch" @click="handleToggleMultipleStream">
              <input 
                type="checkbox" 
                :checked="settingStore.setting?.multipleStream" 
                :disabled="loading"
                class="switch-input"
              >
              <span class="switch-slider"></span>
            </div>
          </div>
        </div>

        <!-- 多线程流线程数设置 -->
        <div class="setting-item" v-if="settingStore.setting?.multipleStream">
          <div class="setting-label">
            <span class="label-text">多线程流线程数</span>
            <span class="label-desc">设置多线程流式下载的并发线程数，范围：1-64，太多的线程数可能会导致加载时间过长甚至程序崩溃</span>
          </div>
          <div class="setting-control">
            <div class="thread-count-control">
              <input
                v-model.number="multipleStreamThreadCount"
                type="range"
                min="1"
                max="64"
                step="1"
                class="thread-slider"
                :disabled="loading"
                @input="handleMultipleStreamThreadCountChange"
              >
              <span class="thread-count-value">{{ multipleStreamThreadCount }}</span>
            </div>
            <button
              @click="handleModifyMultipleStreamThreadCount"
              class="btn btn-primary btn-sm"
              :disabled="loading || multipleStreamThreadCount === originalMultipleStreamThreadCount"
            >
              {{ loading ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>

        <!-- 多线程流块大小设置 -->
        <div class="setting-item" v-if="settingStore.setting?.multipleStream">
          <div class="setting-label">
            <span class="label-text">多线程流块大小</span>
            <span class="label-desc">设置每个下载块的大小，范围：512KB-8MB，较大的块大小可能提升性能但会增加内存使用</span>
          </div>
          <div class="setting-control">
            <div class="thread-count-control">
              <input
                v-model.number="multipleStreamChunkSize"
                type="range"
                min="512"
                max="8192"
                step="512"
                class="thread-slider"
                :disabled="loading"
                @input="handleMultipleStreamChunkSizeChange"
              >
              <span class="thread-count-value">{{ formatChunkSize(multipleStreamChunkSize) }}</span>
            </div>
            <button
              @click="handleModifyMultipleStreamChunkSize"
              class="btn btn-primary btn-sm"
              :disabled="loading || multipleStreamChunkSize === originalMultipleStreamChunkSize"
            >
              {{ loading ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">挂载文件自动刷新</span>
            <span class="label-desc">开启后，系统将自动刷新挂载的文件列表，此功能不能保证文件列表的实时性，可在前台文件浏览页面使用<b>刷新索引</b>功能手动刷新</span>
          </div>
          <div class="setting-control">
            <div class="custom-switch" @click="handleToggleEnableTopFileAutoRefresh">
              <input
                type="checkbox"
                :checked="settingStore.setting?.enableTopFileAutoRefresh"
                :disabled="loading"
                class="switch-input"
              >
              <span class="switch-slider"></span>
            </div>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">自动刷新间隔</span>
            <span class="label-desc">设置挂载文件自动刷新的时间间隔，范围：5-120分钟</span>
          </div>
          <div class="setting-control">
            <div class="thread-count-control">
              <input
                v-model.number="autoRefreshMinutes"
                type="range"
                min="5"
                max="120"
                step="5"
                class="thread-slider"
                :disabled="loading"
                @input="handleAutoRefreshMinutesChange"
              >
              <span class="thread-count-value">{{ autoRefreshMinutes }}分钟</span>
            </div>
            <button
              @click="handleModifyAutoRefreshMinutes"
              class="btn btn-primary btn-sm"
              :disabled="loading || autoRefreshMinutes === originalAutoRefreshMinutes"
            >
              {{ loading ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">任务线程数</span>
            <span class="label-desc">当添加挂载文件或刷新扫描文件时，最大的并发线程数，为了保证服务可靠性，目前最大值为8（任务执行还是单任务按序执行，只是对某个任务的执行速度加快了）</span>
          </div>
          <div class="setting-control">
            <div class="thread-count-control">
              <input
                v-model.number="jobThreadCount"
                type="range"
                min="1"
                max="8"
                step="1"
                class="thread-slider"
                :disabled="loading"
                @input="handleThreadCountChange"
              >
              <span class="thread-count-value">{{ jobThreadCount }}</span>
            </div>
            <button
              @click="handleModifyJobThreadCount"
              class="btn btn-primary btn-sm"
              :disabled="loading || jobThreadCount === originalJobThreadCount"
            >
              {{ loading ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <span class="label-text">API密钥</span>
            <span class="label-desc">刷新后所有已登录账号都将重新登录系统（包括当前账号）</span>
          </div>
          <div class="setting-control">
            <button
              @click="handleRefreshKey"
              class="btn btn-danger btn-sm"
              :disabled="loading"
            >
              {{ loading ? '刷新中...' : '刷新密钥' }}
            </button>
          </div>
        </div>

        <SectionDivider />
    </PageCard>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useSettingStore } from '@/stores/setting'
import Layout from '@/components/Layout.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'
import { toast } from '@/utils/toast'
import { confirmDialog } from '@/utils/confirm'

const settingStore = useSettingStore()

// 响应式数据
const loading = ref(false)
const websiteName = ref('')
const originalWebsiteName = ref('') // 用于存储原始网站名称
const baseURL = ref('')
const originalBaseURL = ref('') // 用于存储原始基础URL
const jobThreadCount = ref(1)
const originalJobThreadCount = ref(1) // 用于存储原始任务线程数
const autoRefreshMinutes = ref(10)
const originalAutoRefreshMinutes = ref(10) // 用于存储原始自动刷新间隔

// 新增：多线程流相关参数
const multipleStreamThreadCount = ref(6)
const originalMultipleStreamThreadCount = ref(6) // 用于存储原始多线程流线程数
const multipleStreamChunkSize = ref(4096) // 默认4MB，以KB为单位
const originalMultipleStreamChunkSize = ref(4096) // 用于存储原始多线程流块大小

// 定时器引用
const timer = ref<NodeJS.Timeout | null>(null)

// 格式化运行时间
const formatRunTime = (seconds: number): string => {
  if (seconds < 60) {
    return `${seconds}秒`
  }

  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) {
    const remainingHours = hours % 24
    const remainingMinutes = minutes % 60
    return `${days}天${remainingHours}小时${remainingMinutes}分钟`
  } else if (hours > 0) {
    const remainingMinutes = minutes % 60
    return `${hours}小时${remainingMinutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}

// 格式化日期
const formatDate = (dateString?: string): string => {
  if (!dateString) return '未知'

  try {
    const date = new Date(dateString)
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (error) {
    return '格式错误'
  }
}

// 新增：格式化块大小显示
const formatChunkSize = (sizeKB: number): string => {
  if (sizeKB >= 1024) {
    const sizeMB = sizeKB / 1024
    return `${sizeMB}MB`
  }
  return `${sizeKB}KB`
}

// 获取设置数据
const fetchSettingData = async () => {
  try {
    loading.value = true
    const data = await settingStore.fetchSetting()
    if (data) {
      websiteName.value = data.title
      originalWebsiteName.value = data.title
      baseURL.value = data.baseURL
      originalBaseURL.value = data.baseURL
      jobThreadCount.value = data.jobThreadCount || 1
      originalJobThreadCount.value = data.jobThreadCount || 1
      autoRefreshMinutes.value = data.autoRefreshMinutes || 10
      originalAutoRefreshMinutes.value = data.autoRefreshMinutes || 10

      // 新增：设置多线程流相关参数
      multipleStreamThreadCount.value = data.multipleStreamThreadCount || 6
      originalMultipleStreamThreadCount.value = data.multipleStreamThreadCount || 6
      // 后端返回的是字节，转换为KB
      multipleStreamChunkSize.value = Math.round((data.multipleStreamChunkSize || 4194304) / 1024)
      originalMultipleStreamChunkSize.value = Math.round((data.multipleStreamChunkSize || 4194304) / 1024)
    }
  } catch (error) {
    toast.error('获取设置失败')
    console.error('获取设置失败:', error)
  } finally {
    loading.value = false
  }
}

// 切换挂载文件自动刷新状态
const handleToggleEnableTopFileAutoRefresh = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const currentStatus = settingStore.setting?.enableTopFileAutoRefresh || false
    await settingStore.toggleEnableTopFileAutoRefresh(currentStatus)
    toast.success('挂载文件自动刷新状态已更新')
  } catch (error) {
    toast.error('更新挂载文件自动刷新状态失败')
    console.error('更新挂载文件自动刷新状态失败:', error)
  } finally {
    loading.value = false
  }
}

// 修改网站名称
const handleModifyName = async () => {
  if (!websiteName.value.trim()) {
    toast.warning('网站名称不能为空')
    return
  }

  try {
    loading.value = true
    await settingStore.modifyName(websiteName.value.trim())
    originalWebsiteName.value = websiteName.value.trim()
    toast.success('网站名称修改成功')
  } catch (error) {
    console.error('修改网站名称失败:', error)
    toast.error('修改网站名称失败')
  } finally {
    loading.value = false
  }
}

// 切换用户认证
const handleToggleAuth = async () => {
  const currentAuth = settingStore.setting?.enableAuth
  const action = currentAuth ? '关闭' : '开启'

  const confirmed = await confirmDialog({
    title: `${action}用户认证`,
    message: `确定要${action}用户认证功能吗？${currentAuth ? '关闭后任何人都可以访问WebDAV。' : '开启后访问WebDAV需要用户登录认证'}`,
    confirmText: '确认',
    cancelText: '取消',
    isDanger: false
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    await settingStore.toggleAuth(!!currentAuth)
    toast.success(`用户认证已${action}`)
  } catch (error) {
    console.error('切换用户认证失败:', error)
    toast.error('切换用户认证失败')
  } finally {
    loading.value = false
  }
}

// 刷新API密钥
const handleRefreshKey = async () => {
  const confirmed = await confirmDialog({
    title: '刷新API密钥',
    message: '确定要刷新API密钥吗？刷新后需要重新配置所有客户端的密钥。',
    confirmText: '确认',
    cancelText: '取消',
    isDanger: true
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    await settingStore.refreshKey()
    toast.success('API密钥刷新成功')
  } catch (error) {
    console.error('刷新API密钥失败:', error)
    toast.error('刷新API密钥失败')
  } finally {
    loading.value = false
  }
}

// 切换本地代理
const handleToggleLocalProxy = async () => {
  const currentLocalProxy = settingStore.setting?.localProxy
  const action = currentLocalProxy ? '关闭' : '开启'

  const confirmed = await confirmDialog({
    title: `${action}本地代理`,
    message: `确定要${action}本地代理功能吗？`,
    confirmText: '确认',
    cancelText: '取消',
    isDanger: false
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    await settingStore.toggleLocalProxy(!!currentLocalProxy)
    toast.success(`本地代理已${action}`)
  } catch (error) {
    console.error('切换本地代理失败:', error)
    toast.error('切换本地代理失败')
  } finally {
    loading.value = false
  }
}

// 切换多线程流式下载
const handleToggleMultipleStream = async () => {
  const currentMultipleStream = settingStore.setting?.multipleStream
  const action = currentMultipleStream ? '关闭' : '开启'

  const confirmed = await confirmDialog({
    title: `${action}多线程流式下载`,
    message: `确定要${action}多线程流式下载功能吗？`,
    confirmText: '确认',
    cancelText: '取消',
    isDanger: false
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    await settingStore.toggleMultipleStream(!!currentMultipleStream)
    toast.success(`多线程流式下载已${action}`)
  } catch (error) {
    console.error('切换多线程流式下载失败:', error)
    toast.error('切换多线程流式下载失败')
  } finally {
    loading.value = false
  }
}

// 修改基础URL
const handleModifyBaseURL = async () => {
  if (!baseURL.value.trim()) {
    toast.warning('基础URL不能为空')
    return
  }

  try {
    loading.value = true
    await settingStore.modifyBaseURL(baseURL.value.trim())
    originalBaseURL.value = baseURL.value.trim()
    toast.success('基础URL修改成功')
  } catch (error) {
    console.error('修改基础URL失败:', error)
    toast.error('修改基础URL失败')
  } finally {
    loading.value = false
  }
}

// 自动获取基础URL
const handleAutoFillBaseURL = () => {
  baseURL.value = window.location.origin
  toast.info('已自动获取当前URL')
}

// 处理任务线程数变化
const handleThreadCountChange = () => {
  // 实时更新显示值，但不保存
}

// 修改任务线程数
const handleModifyJobThreadCount = async () => {
  if (jobThreadCount.value < 1 || jobThreadCount.value > 8) {
    toast.warning('任务线程数必须在1-8之间')
    return
  }

  try {
    loading.value = true
    await settingStore.modifyJobThreadCount(jobThreadCount.value)
    originalJobThreadCount.value = jobThreadCount.value
    toast.success('任务线程数修改成功')
  } catch (error) {
    console.error('修改任务线程数失败:', error)
    toast.error('修改任务线程数失败')
  } finally {
    loading.value = false
  }
}

// 处理自动刷新间隔变化
const handleAutoRefreshMinutesChange = () => {
  // 实时更新显示值，但不保存
}

// 修改自动刷新间隔
const handleModifyAutoRefreshMinutes = async () => {
  if (autoRefreshMinutes.value < 5 || autoRefreshMinutes.value > 120) {
    toast.warning('自动刷新间隔必须在5-120分钟之间')
    return
  }

  try {
    loading.value = true
    await settingStore.modifyAutoRefreshMinutes(autoRefreshMinutes.value)
    originalAutoRefreshMinutes.value = autoRefreshMinutes.value
    toast.success('自动刷新间隔修改成功')
  } catch (error) {
    console.error('修改自动刷新间隔失败:', error)
    toast.error('修改自动刷新间隔失败')
  } finally {
    loading.value = false
  }
}

// 新增：处理多线程流线程数变化
const handleMultipleStreamThreadCountChange = () => {
  // 实时更新显示值，但不保存
}

// 新增：修改多线程流线程数
const handleModifyMultipleStreamThreadCount = async () => {
  if (multipleStreamThreadCount.value < 1 || multipleStreamThreadCount.value > 64) {
    toast.warning('多线程流线程数必须在1-64之间')
    return
  }

  try {
    loading.value = true
    await settingStore.modifyMultipleStreamThreadCount(multipleStreamThreadCount.value)
    originalMultipleStreamThreadCount.value = multipleStreamThreadCount.value
    toast.success('多线程流线程数修改成功')
  } catch (error) {
    console.error('修改多线程流线程数失败:', error)
    toast.error('修改多线程流线程数失败')
  } finally {
    loading.value = false
  }
}

// 新增：处理多线程流块大小变化
const handleMultipleStreamChunkSizeChange = () => {
  // 实时更新显示值，但不保存
}

// 新增：修改多线程流块大小
const handleModifyMultipleStreamChunkSize = async () => {
  if (multipleStreamChunkSize.value < 512 || multipleStreamChunkSize.value > 8192) {
    toast.warning('多线程流块大小必须在512KB-8MB之间')
    return
  }

  try {
    loading.value = true
    // 转换为字节发送给后端：KB * 1024
    await settingStore.modifyMultipleStreamChunkSize(multipleStreamChunkSize.value * 1024)
    originalMultipleStreamChunkSize.value = multipleStreamChunkSize.value
    toast.success('多线程流块大小修改成功')
  } catch (error) {
    console.error('修改多线程流块大小失败:', error)
    toast.error('修改多线程流块大小失败')
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取设置并启动定时器
onMounted(async () => {
  // 初始获取数据
  await fetchSettingData()

  // 每30秒刷新一次数据，让运行时间自动增长
  timer.value = setInterval(fetchSettingData, 30000)
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (timer.value) {
    clearInterval(timer.value)
    timer.value = null
  }
})
</script>

<style scoped>
/* 设置项样式 */
.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 2rem;
  border-bottom: 1px solid #f3f4f6;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.label-text {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.label-desc {
  font-size: 0.75rem;
  color: #6b7280;
  line-height: 1.4;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.setting-input {
  padding: 0.75rem 1rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  min-width: 200px;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}

.setting-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.setting-input:disabled {
  background-color: #f9fafb;
  color: #6b7280;
  cursor: not-allowed;
}

.setting-value {
  font-size: 0.875rem;
  color: #1f2937;
  font-weight: 500;
}

/* 按钮样式 */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
}

.btn-sm {
  padding: 0.5rem 1rem; /* 更小的内边距 */
  font-size: 0.75rem; /* 更小的字体 */
  border-radius: 6px; /* 更小的圆角 */
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #4b5563;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

/* 自定义开关 */
.custom-switch {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 26px;
  cursor: pointer;
}

.switch-input {
  opacity: 0;
  width: 0;
  height: 0;
}

.switch-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #d1d5db;
  transition: all 0.3s ease;
  border-radius: 26px;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.1);
}

.switch-slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: all 0.3s ease;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2), 0 1px 2px rgba(0, 0, 0, 0.1);
}

.switch-input:checked + .switch-slider {
  background: #3b82f6;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.1), 0 0 8px rgba(59, 130, 246, 0.3);
}

.switch-input:checked + .switch-slider:before {
  transform: translateX(22px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.25), 0 1px 3px rgba(0, 0, 0, 0.15);
}

/* 任务线程数控件样式 */
.thread-count-control {
  display: flex;
  align-items: center;
  gap: 1rem;
  min-width: 200px;
}

.thread-slider {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: #e5e7eb;
  outline: none;
  cursor: pointer;
  transition: all 0.2s;
}

.thread-slider::-webkit-slider-thumb {
  appearance: none;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #3b82f6;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s;
}

.thread-slider::-webkit-slider-thumb:hover {
  background: #2563eb;
  transform: scale(1.1);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}

.thread-slider::-moz-range-thumb {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #3b82f6;
  cursor: pointer;
  border: none;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s;
}

.thread-slider::-moz-range-thumb:hover {
  background: #2563eb;
  transform: scale(1.1);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}

.thread-slider:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.thread-count-value {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
  min-width: 60px;
  text-align: center;
  padding: 0.25rem 0.5rem;
  background: #f3f4f6;
  border-radius: 4px;
  border: 1px solid #d1d5db;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
  }
  
  .setting-control {
    width: 100%;
    justify-content: flex-start;
  }
  
  .setting-input {
    min-width: auto;
    flex: 1;
    width: 100%;
  }
  
  .thread-count-control {
    min-width: auto;
    width: 100%;
  }
}
</style>
