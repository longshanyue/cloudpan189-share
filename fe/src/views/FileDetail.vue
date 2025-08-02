<template>
  <div class="file-detail">
    <div class="file-detail-container">
      <!-- 现代化顶部导航栏 -->
      <div class="modern-header">
        <div class="header-left">
          <button @click="goBack" class="back-btn">
            <Icons name="arrow-left" size="1.2rem" />
          </button>
          <div class="breadcrumb-modern">
            <button 
              @click="navigateToPath('')" 
              class="breadcrumb-home"
            >
              <Icons name="home" size="1rem" />
            </button>
            <template v-for="(segment, index) in pathSegments" :key="index">
              <Icons name="chevron-right" size="0.8rem" class="breadcrumb-arrow" />
              <button 
                v-if="index < pathSegments.length - 1"
                @click="navigateToPath(getPathUpTo(index))" 
                class="breadcrumb-link"
                :title="segment"
              >
                {{ truncateText(segment, 8) }}
              </button>
              <span v-else class="breadcrumb-current" :title="segment">
                {{ truncateText(segment, 8) }}
              </span>
            </template>
          </div>
        </div>
        
        <div class="header-actions">
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>加载中...</p>
      </div>
      
      <!-- 错误状态 -->
      <div v-else-if="error" class="error-state">
        <Icons name="alert" size="2rem" class="error-icon" />
        <h3>访问出错</h3>
        <p>无法访问此文件，请检查路径是否正确</p>
        <button @click="navigateToPath('')" class="retry-btn">
          返回根目录
        </button>
      </div>

      <!-- 现代化文件详情内容 -->
      <div v-else-if="fileInfo" class="modern-content">
        <!-- 主要预览区域 -->
        <div class="preview-section">
          <!-- 视频预览 -->
          <div v-if="isVideoFile(fileInfo.name)" class="media-container">
            <div class="media-header">
              <div class="media-type-badge video">
                <Icons name="video" size="1rem" />
                Video
              </div>

            </div>
            <div class="media-wrapper">
              <MediaPlayer
                v-if="fileInfo.downloadURL"
                type="video"
                :src="fileInfo.downloadURL"
                :title="fileInfo.name"
                @loaded="onVideoLoaded"
                @error="onMediaError"
              />
              <div v-else class="loading-placeholder">
                <div class="loading-spinner"></div>
                <p>视频加载中...</p>
              </div>
            </div>
            <!-- 播放器下方的文件名和播放器选择 -->
            <div class="player-footer">
              <div class="file-title-section">
                <h2 class="current-file-name">{{ fileInfo.name }}</h2>
              </div>
              <div class="external-players">
                <span class="players-label">使用外部播放器：</span>
                <div class="player-buttons">
                  <button v-for="player in externalPlayers" :key="player.name" 
                          @click="openWithPlayer(player)" 
                          class="external-player-btn"
                          :title="`使用 ${player.name} 打开`">
                    <img :src="player.icon" :alt="player.name" class="player-icon" />
                    <span>{{ player.name }}</span>
                  </button>
                </div>
              </div>
            </div>
            
            <!-- 文件信息和操作区域 -->
            <div class="file-info-actions">
              <div class="file-meta-info">
                <div class="meta-item">
                  <span class="meta-label">类型：</span>
                  <span class="meta-value">{{ getFileTypeLabel(fileInfo.name) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">大小：</span>
                  <span class="meta-value">{{ formatFileSize(fileInfo.size || 0) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">修改时间：</span>
                  <span class="meta-value">{{ formatDate(fileInfo.modifyDate || '') }}</span>
                </div>
              </div>
              
              <div class="action-buttons">
                <button @click="copyFileLink" class="action-btn secondary">
                  <Icons name="link" size="1rem" />
                  <span>复制链接</span>
                </button>
                <button @click="downloadFile" class="action-btn primary">
                  <Icons name="download" size="1rem" />
                  <span>下载</span>
                </button>
                <div class="qr-code-container" @mouseenter="generateQRCode" @mouseleave="showQRCode = false">
                  <button class="action-btn secondary" title="二维码分享">
                    <Icons name="qr-code" size="1rem" />
                    <span>二维码</span>
                  </button>
                  <div v-if="qrCodeUrl && showQRCode" class="qr-popup">
                    <img :src="qrCodeUrl" alt="文件分享二维码" class="qr-popup-image" />
                    <p class="qr-popup-text">扫码访问页面</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 图片预览 -->
          <div v-else-if="isImageFile(fileInfo.name)" class="media-container">
            <div class="media-header">
              <div class="media-type-badge image">
                <Icons name="image" size="1rem" />
                Image
              </div>
            </div>
            <div class="media-wrapper">
              <MediaPlayer
                v-if="fileInfo.downloadURL"
                type="image"
                :src="fileInfo.downloadURL"
                :title="fileInfo.name"
                @loaded="onImageLoaded"
                @error="onMediaError"
              />
              <div v-else class="loading-placeholder">
                <div class="loading-spinner"></div>
                <p>图片加载中...</p>
              </div>
            </div>
            <div class="player-footer">
              <h2 class="current-file-name">{{ fileInfo.name }}</h2>
            </div>
            
            <!-- 文件信息和操作区域 -->
            <div class="file-info-actions">
              <div class="file-meta-info">
                <div class="meta-item">
                  <span class="meta-label">类型：</span>
                  <span class="meta-value">{{ getFileTypeLabel(fileInfo.name) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">大小：</span>
                  <span class="meta-value">{{ formatFileSize(fileInfo.size || 0) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">修改时间：</span>
                  <span class="meta-value">{{ formatDate(fileInfo.modifyDate || '') }}</span>
                </div>
              </div>
              
              <div class="action-buttons">
                <button @click="copyFileLink" class="action-btn secondary">
                  <Icons name="link" size="1rem" />
                  <span>复制链接</span>
                </button>
                <button @click="downloadFile" class="action-btn primary">
                  <Icons name="download" size="1rem" />
                  <span>下载</span>
                </button>
                <div class="qr-code-container" @mouseenter="generateQRCode" @mouseleave="showQRCode = false">
                  <button class="action-btn secondary" title="二维码分享">
                    <Icons name="qr-code" size="1rem" />
                    <span>二维码</span>
                  </button>
                  <div v-if="qrCodeUrl && showQRCode" class="qr-popup">
                    <img :src="qrCodeUrl" alt="文件分享二维码" class="qr-popup-image" />
                    <p class="qr-popup-text">扫码访问页面</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 其他文件类型 -->
          <div v-else class="file-preview-placeholder">
            <div class="placeholder-content">
              <Icons :name="getFileIcon(fileInfo.name)" size="5rem" class="file-type-icon" />
              <h2 class="file-name">{{ fileInfo.name }}</h2>
              <p class="file-type">{{ getFileTypeLabel(fileInfo.name) }}</p>
              <p class="preview-note">此文件类型暂不支持预览</p>
            </div>
            
            <!-- 文件信息和操作区域 -->
            <div class="file-info-actions">
              <div class="file-meta-info">
                <div class="meta-item">
                  <span class="meta-label">类型：</span>
                  <span class="meta-value">{{ getFileTypeLabel(fileInfo.name) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">大小：</span>
                  <span class="meta-value">{{ formatFileSize(fileInfo.size || 0) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">修改时间：</span>
                  <span class="meta-value">{{ formatDate(fileInfo.modifyDate || '') }}</span>
                </div>
              </div>
              
              <div class="action-buttons">
                <button @click="copyFileLink" class="action-btn secondary">
                  <Icons name="link" size="1rem" />
                  <span>复制链接</span>
                </button>
                <button @click="downloadFile" class="action-btn primary">
                  <Icons name="download" size="1rem" />
                  <span>下载</span>
                </button>
                <div class="qr-code-container" @mouseenter="generateQRCode" @mouseleave="showQRCode = false">
                  <button class="action-btn secondary" title="二维码分享">
                    <Icons name="qr-code" size="1rem" />
                    <span>二维码</span>
                  </button>
                  <div v-if="qrCodeUrl && showQRCode" class="qr-popup">
                    <img :src="qrCodeUrl" alt="文件分享二维码" class="qr-popup-image" />
                    <p class="qr-popup-text">扫码访问页面</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fileApi, type FileItem } from '@/api/file'
import Icons from '@/components/Icons.vue'
import MediaPlayer from '@/components/MediaPlayer.vue'
import { toast } from '@/utils/toast'
import vlcIcon from '@/assets/images/icons/vlc.webp'
import potplayerIcon from '@/assets/images/icons/potplayer.webp'
import mpcIcon from '@/assets/images/icons/mpc-hc.png'
import QRCode from 'qrcode'
const route = useRoute()
const router = useRouter()

// 响应式数据
const loading = ref(false)
const fileInfo = ref<FileItem | null>(null)
const error = ref<{ title: string; message: string } | null>(null)
const showQRCode = ref(false)
const qrCodeUrl = ref<string>('')


// 外部播放器配置
const externalPlayers = ref([
  {
    name: 'VLC',
    icon: vlcIcon,
    protocol: 'vlc://'
  },
  {
    name: 'PotPlayer',
    icon: potplayerIcon,
    protocol: 'potplayer://'
  },
  {
    name: 'MPC-HC',
    icon: mpcIcon,
    protocol: 'mpc-hc://'
  }
])

// 计算属性
const currentPath = computed(() => {
  const path = route.params.pathMatch
  if (path && typeof path === 'string') {
    return path.split('/').map(segment => decodeURIComponent(segment)).join('/')
  } else if (Array.isArray(path)) {
    return path.map(segment => decodeURIComponent(segment)).join('/')
  }
  return ''
})

const pathSegments = computed(() => {
  if (!currentPath.value || typeof currentPath.value !== 'string') {
    return []
  }
  return currentPath.value.split('/').filter(Boolean)
})

// 导航函数
const navigateToPath = (path: string) => {
  if (path === '') {
    router.push('/')
  } else {
    router.push('/' + path)
  }
}

const getPathUpTo = (index: number): string => {
  return pathSegments.value.slice(0, index + 1).join('/')
}

const goBack = () => {
  const parentPath = pathSegments.value.slice(0, -1).join('/')
  navigateToPath(parentPath)
}

// 文件类型判断
const isVideoFile = (filename: string): boolean => {
  if (!filename || typeof filename !== 'string') return false
  const ext = filename.split('.').pop()?.toLowerCase()
  return ['mp4', 'avi', 'mkv', 'mov', 'wmv', 'flv', 'webm'].includes(ext || '')
}

const isImageFile = (filename: string): boolean => {
  if (!filename || typeof filename !== 'string') return false
  const ext = filename.split('.').pop()?.toLowerCase()
  return ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'].includes(ext || '')
}

const isAudioFile = (filename: string): boolean => {
  if (!filename || typeof filename !== 'string') return false
  const ext = filename.split('.').pop()?.toLowerCase()
  return ['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a'].includes(ext || '')
}

// 获取文件图标
const getFileIcon = (filename: string): string => {
  if (!filename || typeof filename !== 'string') return 'file'
  
  const ext = filename.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'mp4':
    case 'avi':
    case 'mkv':
    case 'mov':
      return 'video'
    case 'mp3':
    case 'wav':
    case 'flac':
      return 'music'
    case 'jpg':
    case 'jpeg':
    case 'png':
    case 'gif':
      return 'image'
    case 'pdf':
      return 'file-text'
    case 'doc':
    case 'docx':
      return 'file-text'
    case 'zip':
    case 'rar':
    case '7z':
      return 'archive'
    default:
      return 'file'
  }
}

// 获取文件类型标签
const getFileTypeLabel = (filename: string): string => {
  if (!filename || typeof filename !== 'string') return '未知类型'
  
  const ext = filename.split('.').pop()?.toLowerCase()
  if (!ext) return '未知类型'
  
  const typeMap: Record<string, string> = {
    'pdf': 'PDF文档',
    'doc': 'Word文档',
    'docx': 'Word文档',
    'xls': 'Excel表格',
    'xlsx': 'Excel表格',
    'ppt': 'PowerPoint演示文稿',
    'pptx': 'PowerPoint演示文稿',
    'txt': '文本文件',
    'jpg': '图片文件',
    'jpeg': '图片文件',
    'png': '图片文件',
    'gif': '图片文件',
    'mp4': '视频文件',
    'avi': '视频文件',
    'mkv': '视频文件',
    'mp3': '音频文件',
    'wav': '音频文件',
    'zip': '压缩文件',
    'rar': '压缩文件',
    '7z': '压缩文件'
  }
  
  return typeMap[ext] || `${ext.toUpperCase()}文件`
}

// 格式化文件大小
const formatFileSize = (size: number): string => {
  if (size === 0) return '-'
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let index = 0
  let fileSize = size
  
  while (fileSize >= 1024 && index < units.length - 1) {
    fileSize /= 1024
    index++
  }
  
  return `${fileSize.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

// 格式化日期
const formatDate = (dateString: string): string => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 截断文字
const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) {
    return text
  }
  return text.substring(0, maxLength) + '...'
}



// 使用外部播放器打开
const openWithPlayer = (player: any) => {
  if (!fileInfo.value?.downloadURL) return
  
  try {
    const url = player.protocol + fileInfo.value.downloadURL
    window.open(url, '_blank')
    toast.success(`正在使用 ${player.name} 打开文件`)
  } catch (error) {
    console.error('打开播放器失败:', error)
    toast.error(`无法使用 ${player.name} 打开文件`)
  }
}

// 媒体加载完成
const onVideoLoaded = () => {
  console.log('视频加载完成')
}

const onImageLoaded = () => {
  console.log('图片加载完成')
}

const onAudioLoaded = () => {
  console.log('音频加载完成')
}

// 媒体加载错误
const onMediaError = (error: Error) => {
  console.error('媒体加载失败:', error)
  toast.error('媒体加载失败')
}

// 复制文件链接
const copyFileLink = async () => {
  if (!fileInfo.value?.downloadURL) return
  
  try {
    // 检查是否支持现代剪贴板API
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(fileInfo.value.downloadURL)
    } else {
      // 使用传统方法
      const textArea = document.createElement('textarea')
      textArea.value = fileInfo.value.downloadURL
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)
    }
    toast.success('链接已复制到剪贴板')
  } catch (error) {
    console.error('复制链接失败:', error)
    toast.error('复制链接失败')
  }
}

// 下载文件
const downloadFile = () => {
  if (!fileInfo.value?.downloadURL) return
  
  const link = document.createElement('a')
  link.href = fileInfo.value.downloadURL
  link.download = fileInfo.value.name
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  toast.success('文件下载已开始')
}

// 生成二维码
const generateQRCode = async () => {
  if (!qrCodeUrl.value) {
    try {
      // 使用本地二维码生成库，生成当前页面URL
      const currentUrl = window.location.href
      qrCodeUrl.value = await QRCode.toDataURL(currentUrl, {
        width: 120,
        margin: 1,
        color: {
          dark: '#000000',
          light: '#FFFFFF'
        }
      })
      showQRCode.value = true
    } catch (error) {
      console.error('生成二维码失败:', error)
      toast.error('生成二维码失败')
    }
  } else {
    showQRCode.value = true
  }
}

// 加载文件信息
const loadFileInfo = async () => {
  try {
    loading.value = true
    error.value = null
    
    const result = await fileApi.getFile(currentPath.value)
    fileInfo.value = result
  } catch (err: any) {
    console.error('加载文件失败:', err)
    error.value = {
      title: '加载失败',
      message: err.message || '无法加载文件信息'
    }
  } finally {
    loading.value = false
  }
}

// 组件挂载时加载文件信息
onMounted(() => {
  loadFileInfo()
})
</script>

<style scoped>
.file-detail {
  background: #f8fafc;
  padding: 1rem;
}

.file-detail-container {
  max-width: 980px;
  margin: 0 auto;
}

/* 现代化头部样式 */
.modern-header {
  background: white;
  border-radius: 1rem;
  padding: 1rem 1.5rem;
  margin-bottom: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid #e5e7eb;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
  min-width: 0;
}

.back-btn {
  background: #f3f4f6;
  border: none;
  border-radius: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.back-btn:hover {
  background: #e5e7eb;
  color: #374151;
}

.breadcrumb-modern {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
  -ms-overflow-style: none;
  position: relative;
}

.breadcrumb-modern::-webkit-scrollbar {
  display: none;
}

.breadcrumb-modern:hover::after {
  opacity: 1;
}

.breadcrumb-home {
  background: #eff6ff;
  border: 1px solid #dbeafe;
  color: #3b82f6;
  border-radius: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.breadcrumb-home:hover {
  background: #dbeafe;
}

.breadcrumb-arrow {
  color: #9ca3af;
  flex-shrink: 0;
}

.breadcrumb-link {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 0.375rem;
  transition: all 0.2s ease;
  font-size: 0.875rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.breadcrumb-link:hover {
  background: #f3f4f6;
  color: #374151;
}

.breadcrumb-current {
  color: #1f2937;
  font-weight: 500;
  font-size: 0.875rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 2rem;
  color: #475569;
  background: white;
  border-radius: 0.75rem;
  margin: 2rem 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid #e2e8f0;
}

.loading-spinner {
  width: 2rem;
  height: 2rem;
  border: 2px solid #e5e7eb;
  border-top: 2px solid #6b7280;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 2rem;
  text-align: center;
  color: #dc2626;
  background: #fef2f2;
  border-radius: 0.75rem;
  margin: 2rem 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid #fecaca;
}

.error-icon {
  margin-bottom: 1rem;
}

.error-state h3 {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
}

.error-state p {
  margin: 0 0 1.5rem 0;
  color: #7f1d1d;
}

.retry-btn {
  background: #dc2626;
  color: white;
  border: none;
  border-radius: 0.5rem;
  padding: 0.75rem 1.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.retry-btn:hover {
  background: #b91c1c;
}

/* 现代化内容区域 */
.modern-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.preview-section {
  background: white;
  border-radius: 1rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #e5e7eb;
}

/* 媒体容器样式 */
.media-container {
  display: flex;
  flex-direction: column;
}

.media-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #f3f4f6;
  background: #fafafa;
}

.media-type-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.75rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.media-type-badge.video {
  background: #fef3c7;
  color: #d97706;
}

.media-type-badge.image {
  background: #dcfce7;
  color: #16a34a;
}

.media-type-badge.audio {
  background: #e0e7ff;
  color: #4f46e5;
}

.media-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.control-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: none;
  border: none;
  color: #6b7280;
  font-size: 0.875rem;
  cursor: pointer;
  transition: color 0.2s ease;
}

.control-btn:hover {
  color: #374151;
}

.control-btn.active {
  color: #3b82f6;
}

.toggle-switch {
  width: 2rem;
  height: 1rem;
  background: #d1d5db;
  border-radius: 0.5rem;
  position: relative;
  transition: background 0.2s ease;
}

.toggle-switch::after {
  content: '';
  position: absolute;
  top: 0.125rem;
  left: 0.125rem;
  width: 0.75rem;
  height: 0.75rem;
  background: white;
  border-radius: 50%;
  transition: transform 0.2s ease;
}

.toggle-switch.active {
  background: #3b82f6;
}

.toggle-switch.active::after {
  transform: translateX(1rem);
}

/* 文件预览占位符样式 */
.file-preview-placeholder {
  background: white;
  border-radius: 1rem;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #e5e7eb;
  min-height: 400px;
  display: flex;
  flex-direction: column;
}

.placeholder-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  text-align: center;
  flex: 1;
  background: #fafafa;
}

.placeholder-content .file-type-icon {
  color: #9ca3af;
  margin-bottom: 1.5rem;
}

.placeholder-content .file-name {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.5rem 0;
  word-break: break-all;
}

.placeholder-content .file-type {
  font-size: 1rem;
  color: #6b7280;
  margin: 0 0 1rem 0;
}

.placeholder-content .preview-note {
  font-size: 0.875rem;
  color: #9ca3af;
  margin: 0;
}

/* 媒体预览区域 */
.media-preview {
  position: relative;
  background: #000;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.media-preview video {
  width: 100%;
  height: auto;
  max-height: 50vh;
  max-width: 100%;
}

.media-wrapper {
  width: 100%;
  border-radius: 0.5rem;
  overflow: hidden;
}

.loading-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: #6b7280;
  background: #f9fafb;
}

.media-preview img {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
}

.media-preview audio {
  width: 100%;
  background: #1f2937;
}

/* 外部播放器区域 */
.external-players {
  padding: 1rem 1.5rem;
  border-top: 1px solid #f3f4f6;
  background: #fafafa;
}

.external-players h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 0.75rem;
}

.player-list {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.player-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.75rem;
  color: #6b7280;
  transition: all 0.2s ease;
}

.player-btn:hover {
  background: #f9fafb;
  border-color: #d1d5db;
  color: #374151;
}

.player-icon {
  width: 1rem;
  height: 1rem;
}

/* 文件信息和操作区域样式 */
.file-info-actions {
  padding: 1.5rem;
  background: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
  border-top: 1px solid #f3f4f6;
}

.file-meta-info {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.meta-label {
  color: #6b7280;
  font-weight: 500;
}

.meta-value {
  color: #1f2937;
  font-weight: 600;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.action-btn.primary {
  background: #3b82f6;
  color: white;
}

.action-btn.primary:hover {
  background: #2563eb;
}

.action-btn.secondary {
  background: #f3f4f6;
  color: #6b7280;
  border-color: #e5e7eb;
}

.action-btn.secondary:hover {
  background: #e5e7eb;
  color: #374151;
}

/* 播放器底部区域样式 */
.player-footer {
  margin-top: 1rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.file-title-section {
  margin-bottom: 1rem;
}

.current-file-name {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  color: #333;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

.external-players {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.players-label {
  font-size: 0.9rem;
  color: #666;
  font-weight: 500;
}

.player-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.external-player-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border: none;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.1);
  cursor: pointer;
  transition: all 0.2s ease;
  backdrop-filter: blur(10px);
  font-size: 0.85rem;
  color: #333;
}

.external-player-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateY(-1px);
}

.player-icon {
  width: 20px;
  height: 20px;
  object-fit: contain;
}

.qr-code-container {
  position: relative;
}

.qr-popup {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-bottom: 0.5rem;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  padding: 1rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1001;
  text-align: center;
}

.qr-popup-image {
  width: 120px;
  height: 120px;
  border-radius: 0.375rem;
}

.qr-popup-text {
  font-size: 0.75rem;
  color: #6b7280;
  margin: 0.5rem 0 0 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .file-detail {
    padding: 0.5rem;
  }
  
  .modern-header {
    padding: 0.75rem 1rem;
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .header-left {
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .breadcrumb-modern {
    flex-wrap: wrap;
  }
  
  .header-actions {
    justify-content: center;
  }
  
  .media-preview {
    min-height: 300px;
  }
  
  .file-info-actions {
    flex-direction: column;
    align-items: stretch;
    gap: 1rem;
  }
  
  .file-meta-info {
    justify-content: center;
  }
  
  .action-buttons {
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .file-detail {
    padding: 0.25rem;
  }
  
  .modern-header {
    padding: 0.5rem 0.75rem;
  }
  
  .action-btn {
    padding: 0.5rem 0.75rem;
    font-size: 0.8rem;
  }
  
  .media-preview {
    min-height: 250px;
  }
  
  .player-list {
    justify-content: center;
  }
}
</style>
