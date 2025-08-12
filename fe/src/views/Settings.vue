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

      <!-- STRM文件生成设置 -->
      <div class="setting-item">
        <div class="setting-label">
          <span class="label-text">STRM文件生成</span>
          <span class="label-desc">开启后系统将为支持的视频文件生成STRM文件，用于媒体服务器播放。关闭时会清空已有的STRM文件</span>
        </div>
        <div class="setting-control">
          <div class="custom-switch" @click="handleToggleStrmFileEnable">
            <input
                type="checkbox"
                :checked="settingStore.setting?.strmFileEnable"
                :disabled="loading"
                class="switch-input"
            >
            <span class="switch-slider"></span>
          </div>
          <button
              v-if="settingStore.setting?.strmFileEnable"
              @click="handleRebuildStrmFiles"
              class="btn btn-danger btn-sm"
              :disabled="loading"
          >
            {{ loading ? '重建中...' : '重建STRM文件' }}
          </button>
        </div>
      </div>

      <!-- STRM支持文件格式设置 -->
      <div class="setting-item">
        <div class="setting-label">
          <span class="label-text">STRM支持文件格式</span>
          <span class="label-desc">设置哪些文件格式会生成STRM文件，支持常见的视频格式。无论是否开启STRM功能都可以预先配置</span>
        </div>
        <div class="setting-control">
          <div class="strm-ext-display">
            <span v-if="strmSupportFileExtList.length === 0" class="no-ext-text">暂无设置任何格式（支持所有格式）</span>
            <span v-else class="ext-count">已设置 {{ strmSupportFileExtList.length }} 种格式</span>
            <div v-if="strmSupportFileExtList.length > 0" class="ext-preview">
              {{ strmSupportFileExtList.slice(0, 5).join(', ') }}
              <span v-if="strmSupportFileExtList.length > 5">...</span>
            </div>
          </div>
          <button
              @click="openStrmExtModal"
              class="btn btn-primary btn-sm"
              :disabled="loading"
          >
            编辑格式
          </button>
        </div>
      </div>

      <!-- 新增：WebDAV写入权限设置 -->
      <div class="setting-item">
        <div class="setting-label">
          <span class="label-text">WebDAV写入权限</span>
          <span class="label-desc">开启后允许通过WebDAV协议写入和删除真实文件，关闭后WebDAV仅提供只读访问。注意：此功能仅影响真实文件，不包括挂载的分享文件以及strm等虚拟文件，如果父级文件夹被删除，则文件也会被删除。</span>
        </div>
        <div class="setting-control">
          <div class="custom-switch" @click="handleToggleFileWritable">
            <input
                type="checkbox"
                :checked="settingStore.setting?.fileWritable"
                :disabled="loading"
                class="switch-input"
            >
            <span class="switch-slider"></span>
          </div>
        </div>
      </div>

      <!-- 清空本地真实存储 -->
      <div class="setting-item">
        <div class="setting-label">
          <span class="label-text">清空本地真实存储</span>
          <span class="label-desc">清空所有本地真实存储的文件，包括 Emby 生成的 nfo 等文件，此操作不可逆，请谨慎使用。此功能不会影响挂载的分享文件和虚拟文件。</span>
        </div>
        <div class="setting-control">
          <button
              @click="handleClearRealFile"
              class="btn btn-danger btn-sm"
              :disabled="loading"
          >
            {{ loading ? '清空中...' : '清空本地真实存储' }}
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

    <!-- STRM文件格式编辑弹窗 -->
    <div v-if="showStrmExtModal" class="modal-overlay" @click="closeStrmExtModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">编辑STRM支持文件格式</h3>
          <button @click="closeStrmExtModal" class="modal-close">×</button>
        </div>
        <div class="modal-body">
          <div class="ext-editor">
            <div class="ext-input-section">
              <label class="ext-label">添加新格式（不包含点号）:</label>
              <div class="ext-input-group">
                <input
                    v-model="newExtension"
                    type="text"
                    placeholder="例如: mp4"
                    class="ext-input"
                    @keyup.enter="addExtension"
                    maxlength="10"
                >
                <button @click="addExtension" class="btn btn-primary btn-sm" :disabled="!newExtension.trim()">
                  添加
                </button>
              </div>
            </div>

            <div class="ext-list-section">
              <div class="ext-list-header">
                <label class="ext-label">当前支持的格式:</label>
                <div class="ext-actions">
                  <button @click="selectAllExtensions" class="btn btn-secondary btn-xs">全选</button>
                  <button @click="clearAllExtensions" class="btn btn-secondary btn-xs">清空</button>
                  <button @click="resetToDefaultExtensions" class="btn btn-secondary btn-xs">恢复默认</button>
                </div>
              </div>

              <div class="ext-tags">
                <div
                    v-for="(ext, index) in tempStrmSupportFileExtList"
                    :key="index"
                    class="ext-tag"
                >
                  <span class="ext-name">{{ ext }}</span>
                  <button @click="removeExtension(index)" class="ext-remove">×</button>
                </div>
                <div v-if="tempStrmSupportFileExtList.length === 0" class="no-ext-message">
                  暂无设置任何格式（支持所有格式）
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeStrmExtModal" class="btn btn-secondary">取消</button>
          <button @click="saveStrmExtensions" class="btn btn-primary" :disabled="modalLoading">
            {{ modalLoading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
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
import { storageApi } from '@/api/storage'

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

// 多线程流相关参数
const multipleStreamThreadCount = ref(6)
const originalMultipleStreamThreadCount = ref(6) // 用于存储原始多线程流线程数
const multipleStreamChunkSize = ref(4096) // 默认4MB，以KB为单位
const originalMultipleStreamChunkSize = ref(4096) // 用于存储原始多线程流块大小

// STRM相关参数
const strmSupportFileExtList = ref<string[]>([])

// STRM文件格式编辑弹窗相关
const showStrmExtModal = ref(false)
const modalLoading = ref(false)
const tempStrmSupportFileExtList = ref<string[]>([])
const newExtension = ref('')

// 默认的文件扩展名列表
const DEFAULT_EXTENSIONS = [
  'mp4', 'mkv', 'avi', 'mov', 'wmv', 'flv', 'webm', 'm4v',
  'mpg', 'mpeg', 'm2v', 'm4p', 'm4b', 'ts', 'mts', 'm2ts', 'm2t',
  'mxf', 'dv', 'dvr-ms', 'asf', '3gp', '3g2', 'f4v', 'f4p', 'f4a', 'f4b',
  'vob', 'ogv', 'ogg', 'divx', 'xvid', 'rm', 'rmvb', 'dat', 'nsv',
  'qt', 'amv', 'mpv', 'm1v', 'svi', 'viv', 'fli', 'flc'
]

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

// 格式化块大小显示
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

      // 设置多线程流相关参数
      multipleStreamThreadCount.value = data.multipleStreamThreadCount || 6
      originalMultipleStreamThreadCount.value = data.multipleStreamThreadCount || 6
      // 后端返回的是字节，转换为KB
      multipleStreamChunkSize.value = Math.round((data.multipleStreamChunkSize || 4194304) / 1024)
      originalMultipleStreamChunkSize.value = Math.round((data.multipleStreamChunkSize || 4194304) / 1024)

      // 设置STRM相关参数
      strmSupportFileExtList.value = data.strmSupportFileExtList || []
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

// 处理多线程流线程数变化
const handleMultipleStreamThreadCountChange = () => {
  // 实时更新显示值，但不保存
}

// 修改多线程流线程数
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

// 处理多线程流块大小变化
const handleMultipleStreamChunkSizeChange = () => {
  // 实时更新显示值，但不保存
}

// 修改多线程流块大小
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

// 切换STRM文件生成
const handleToggleStrmFileEnable = async () => {
  const currentStrmFileEnable = settingStore.setting?.strmFileEnable
  const action = currentStrmFileEnable ? '关闭' : '开启'

  const confirmed = await confirmDialog({
    title: `${action}STRM文件生成`,
    message: `确定要${action}STRM文件生成功能吗？${currentStrmFileEnable ? '关闭后会清空所有已生成的STRM文件。' : '开启后系统将为支持的视频文件生成STRM文件。'}`,
    confirmText: '确认',
    cancelText: '取消',
    isDanger: currentStrmFileEnable
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    await settingStore.toggleStrmFileEnable(!currentStrmFileEnable)
    toast.success(`STRM文件生成已${action}`)
  } catch (error) {
    console.error('切换STRM文件生成失败:', error)
    toast.error('切换STRM文件生成失败')
  } finally {
    loading.value = false
  }
}

// 重建STRM文件
const handleRebuildStrmFiles = async () => {
  const confirmed = await confirmDialog({
    title: '重建STRM文件',
    message: '确定要重建所有STRM文件吗？这将删除现有的STRM文件并重新生成。',
    confirmText: '确认',
    cancelText: '取消',
    isDanger: true
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    // 传入true表示重建
    await settingStore.toggleStrmFileEnable(true)
    toast.success('STRM文件重建已开始')
  } catch (error) {
    console.error('重建STRM文件失败:', error)
    toast.error('重建STRM文件失败')
  } finally {
    loading.value = false
  }
}

// 新增：切换WebDAV写入权限
const handleToggleFileWritable = async () => {
  const currentFileWritable = settingStore.setting?.fileWritable
  const action = currentFileWritable ? '关闭' : '开启'

  const confirmed = await confirmDialog({
    title: `${action}WebDAV写入权限`,
    message: `确定要${action}WebDAV写入权限吗？${currentFileWritable ? '关闭后WebDAV将变为只读模式。' : '开启后允许通过WebDAV写入和删除真实文件。'}`,
    confirmText: '确认',
    cancelText: '取消',
    isDanger: currentFileWritable
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    await settingStore.toggleFileWritable(!currentFileWritable)
    toast.success(`WebDAV写入权限已${action}`)
  } catch (error) {
    console.error('切换WebDAV写入权限失败:', error)
    toast.error('切换WebDAV写入权限失败')
  } finally {
    loading.value = false
  }
}

// 打开STRM文件格式编辑弹窗
const openStrmExtModal = () => {
  tempStrmSupportFileExtList.value = [...strmSupportFileExtList.value]
  newExtension.value = ''
  showStrmExtModal.value = true
  // 阻止页面滚动
  document.body.style.overflow = 'hidden'
}

// 关闭STRM文件格式编辑弹窗
const closeStrmExtModal = () => {
  showStrmExtModal.value = false
  tempStrmSupportFileExtList.value = []
  newExtension.value = ''
  // 恢复页面滚动
  document.body.style.overflow = ''
}

// 添加文件扩展名
const addExtension = () => {
  const ext = newExtension.value.trim().toLowerCase()
  if (!ext) {
    toast.warning('请输入文件扩展名')
    return
  }

  // 验证格式（只允许字母和数字）
  if (!/^[a-z0-9]+$/.test(ext)) {
    toast.warning('文件扩展名只能包含字母和数字')
    return
  }

  if (tempStrmSupportFileExtList.value.includes(ext)) {
    toast.warning('该扩展名已存在')
    return
  }

  tempStrmSupportFileExtList.value.push(ext)
  newExtension.value = ''
  toast.success('扩展名添加成功')
}

// 移除文件扩展名
const removeExtension = (index: number) => {
  tempStrmSupportFileExtList.value.splice(index, 1)
}

// 全选扩展名
const selectAllExtensions = () => {
  tempStrmSupportFileExtList.value = [...DEFAULT_EXTENSIONS]
  toast.info('已选择所有默认格式')
}

// 清空扩展名
const clearAllExtensions = () => {
  tempStrmSupportFileExtList.value = []
  toast.info('已清空所有格式')
}

// 恢复默认扩展名
const resetToDefaultExtensions = () => {
  tempStrmSupportFileExtList.value = [...DEFAULT_EXTENSIONS]
  toast.info('已恢复默认格式')
}

// 保存STRM文件扩展名设置
const saveStrmExtensions = async () => {
  try {
    modalLoading.value = true
    await settingStore.modifyStrmSupportFileExtList(tempStrmSupportFileExtList.value)
    strmSupportFileExtList.value = [...tempStrmSupportFileExtList.value]
    toast.success('STRM支持文件格式保存成功')
    closeStrmExtModal()
  } catch (error) {
    console.error('保存STRM支持文件格式失败:', error)
    toast.error('保存STRM支持文件格式失败')
  } finally {
    modalLoading.value = false
  }
}

// 清空本地真实存储
const handleClearRealFile = async () => {
  const confirmed = await confirmDialog({
    title: '清空本地真实存储',
    message: '确定要清空所有本地真实存储的文件吗？包括 Emby 生成的 nfo 等文件，此操作不可逆，请谨慎使用。此功能不会影响挂载的分享文件和虚拟文件。',
    confirmText: '确认清空',
    cancelText: '取消',
    isDanger: true
  })

  if (!confirmed) {
    return
  }

  try {
    loading.value = true
    const response = await storageApi.clearRealFile()
    toast.success(response.message)
  } catch (error) {
    console.error('清空本地真实存储失败:', error)
    toast.error('清空本地真实存储失败')
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
  // 确保恢复页面滚动
  document.body.style.overflow = ''
})
</script>

<style scoped>
/* 原有样式保持不变... */
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

/* STRM扩展名显示样式 */
.strm-ext-display {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 200px;
}

.no-ext-text {
  font-size: 0.875rem;
  color: #6b7280;
  font-style: italic;
}

.ext-count {
  font-size: 0.875rem;
  color: #374151;
  font-weight: 500;
}

.ext-preview {
  font-size: 0.75rem;
  color: #6b7280;
  line-height: 1.4;
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
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
  border-radius: 6px;
}

.btn-xs {
  padding: 0.25rem 0.5rem;
  font-size: 0.625rem;
  border-radius: 4px;
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

/* 弹窗样式 - 修复滚动问题 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  overflow-y: auto;
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 600px;
  width: 100%;
  max-height: calc(100vh - 2rem);
  margin: auto;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #f3f4f6;
  flex-shrink: 0;
}

.modal-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #6b7280;
  cursor: pointer;
  padding: 0;
  line-height: 1;
  transition: color 0.2s;
}

.modal-close:hover {
  color: #374151;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.5rem;
  border-top: 1px solid #f3f4f6;
  flex-shrink: 0;
}

/* 扩展名编辑器样式 */
.ext-editor {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.ext-input-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.ext-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.ext-input-group {
  display: flex;
  gap: 0.5rem;
}

.ext-input {
  flex: 1;
  padding: 0.625rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: border-color 0.2s;
}

.ext-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.ext-list-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.ext-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.ext-actions {
  display: flex;
  gap: 0.5rem;
}

.ext-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  min-height: 100px;
  max-height: 200px;
  overflow-y: auto;
  padding: 0.75rem;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: #f9fafb;
}

.ext-tag {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: #3b82f6;
  color: white;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.ext-name {
  line-height: 1;
}

.ext-remove {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  font-size: 0.875rem;
  line-height: 1;
  padding: 0;
  margin-left: 0.125rem;
  transition: opacity 0.2s;
}

.ext-remove:hover {
  opacity: 0.7;
}

.no-ext-message {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  font-style: italic;
  font-size: 0.875rem;
  width: 100%;
  height: 100px;
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

  .strm-ext-display {
    min-width: auto;
    width: 100%;
  }

  .modal-overlay {
    padding: 0.5rem;
  }

  .modal-content {
    max-height: calc(100vh - 1rem);
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 1rem;
  }

  .ext-input-group {
    flex-direction: column;
  }

  .ext-list-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .ext-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
