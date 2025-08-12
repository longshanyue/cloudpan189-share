<template>
  <div class="file-browser" ref="fileBrowserRef">
    <div class="browser-container" :class="{ 'header-fixed-padding': headerFixed }">
      <!-- 现代化头部样式 -->
      <div class="modern-header" :class="{ 'header-fixed': headerFixed }">
        <div class="header-left">
          <div class="breadcrumb-nav">
            <button
                @click="navigateToPath('')"
                class="breadcrumb-home"
                title="返回根目录"
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

        <div class="header-actions" style="display: flex; align-items: center;">
          <button @click="openSearchModal" class="action-btn secondary">
            <Icons name="search" size="1rem" />
            搜索
          </button>
          <button
              v-if="settingStore.setting?.strmFileEnable"
              @click="toggleShowStrmFiles"
              class="action-btn"
              :class="{ 'active': showStrmFiles }"
              :title="showStrmFiles ? '隐藏STRM文件' : '显示STRM文件'"
          >
            <Icons name="file" size="1rem" />
            {{ showStrmFiles ? '隐藏STRM' : '显示STRM' }}
          </button>
          <button @click="refreshCurrentPath" class="action-btn secondary" :disabled="loading">
            <Icons name="refresh" size="1rem" />
            刷新
          </button>
        </div>
      </div>

      <!-- 文件列表 -->
      <div class="file-list-container">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <p>加载中...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <Icons name="alert" size="2rem" class="error-icon" />
          <h3>访问出错</h3>
          <p>无法访问此路径，请检查路径是否正确</p>
          <button @click="navigateToPath('')" class="retry-btn">
            返回根目录
          </button>
        </div>

        <div v-else class="file-list">
          <!-- 文件列表头部 -->
          <div class="file-list-header">
            <div class="file-col-name">名称</div>
            <div class="file-col-size">大小</div>
            <div class="file-col-date">修改时间</div>
            <div class="file-col-actions">操作</div>
          </div>

          <!-- 文件和文件夹列表 -->
          <div
              v-for="item in fileList"
              :key="item.id"
              class="file-item"
              :class="{ folder: item.isFolder }"
              @click="handleItemClick(item)"
          >
            <div class="file-col-name">
              <Icons
                  :name="getFileIcon(item)"
                  size="1.2rem"
                  class="file-icon"
                  :class="{ 'folder-icon': item.isFolder }"
              />
              <span class="file-name">{{ item.name }}</span>
            </div>
            <div class="file-col-size">{{ formatFileSize(item.size) }}</div>
            <div class="file-col-date">{{ formatDate(item.modifyDate) }}</div>
            <div class="file-col-actions">
              <!-- 下载按钮 -->
              <button
                  v-if="!item.isFolder && item.downloadURL"
                  @click.stop="downloadFile(item)"
                  class="action-btn-small"
                  title="下载"
              >
                <Icons name="download" size="0.9rem" />
              </button>

              <!-- 刷新索引按钮 -->
              <button
                  v-if="item.isFolder"
                  @click.stop="refreshFolderIndex(item)"
                  class="action-btn-small refresh-btn"
                  title="如果文件夹内容未实际更新，点击刷新索引后会全量扫描文件夹内容及子文件夹内容"
              >
                <Icons name="refresh" size="0.9rem" />
                <span>刷新索引</span>
              </button>

              <!-- 删除按钮 -->
              <button
                  @click.stop="confirmDelete(item)"
                  class="action-btn-small delete-btn"
                  :disabled="item.isTop === 1"
                  :title="item.isTop === 1 ? '挂载点文件请在后台存储管理删除' : (item.isFolder ? '删除文件夹' : '删除文件')"
              >
                <Icons name="trash-2" size="0.9rem" />
                <span>删除</span>
              </button>
            </div>
          </div>

          <!-- 空文件夹提示 -->
          <div v-if="fileList.length === 0" class="empty-state">
            <Icons name="folder" size="3rem" class="empty-icon" />
            <p>此文件夹为空</p>
          </div>
        </div>
      </div>

    </div>

    <!-- 搜索弹窗 -->
    <div v-if="showSearchModal" class="search-modal-overlay" @click="closeSearchModal">
      <div class="search-modal" @click.stop>
        <div class="search-modal-header">
          <h3>文件搜索</h3>
          <button @click="closeSearchModal" class="close-btn">
            <Icons name="close" size="1.2rem" />
          </button>
        </div>

        <div class="search-modal-body">
          <div class="search-form">
            <div class="search-row">
              <div class="search-options-inline">
                <label class="inline-checkbox-label">
                  <input
                      v-model="globalSearch"
                      type="checkbox"
                      class="inline-checkbox"
                  />
                  <span class="inline-checkbox-text">全局搜索</span>
                </label>
              </div>

              <div class="search-input-group">
                <input
                    v-model="searchKeyword"
                    type="text"
                    placeholder="请输入搜索关键词"
                    class="search-input"
                    @keyup.enter="performSearch"
                />
                <button @click="performSearch" class="search-btn" :disabled="searchLoading || !searchKeyword.trim()">
                  <Icons name="search" size="1rem" />
                  搜索
                </button>
              </div>
            </div>
          </div>

          <!-- 搜索结果 -->
          <div v-if="searchResults.length > 0 || searchLoading" class="search-results">
            <div v-if="searchLoading" class="search-loading">
              <div class="loading-spinner"></div>
              <p>搜索中...</p>
            </div>

            <div v-else class="search-list">
              <div class="search-list-header">
                <div class="search-col-name">名称</div>
                <div class="search-col-type text-center">类型</div>
                <div class="search-col-size text-center">大小</div>
                <div class="search-col-path text-center">路径</div>
              </div>

              <div
                  v-for="item in searchResults"
                  :key="item.id"
                  class="search-item"
                  @click="navigateToSearchResult(item)"
              >
                <div class="search-col-name">
                  <Icons
                      :name="getFileIcon(item)"
                      size="1.2rem"
                      class="file-icon"
                      :class="{ 'folder-icon': item.isFolder }"
                  />
                  <span class="file-name">{{ item.name }}</span>
                </div>
                <div class="search-col-type text-center">{{ getFileType(item) }}</div>
                <div class="search-col-size text-center">{{ formatFileSize(item.size) }}</div>
                <div class="search-col-path text-center">{{ item.localPath }}</div>
              </div>
            </div>

            <!-- 分页 -->
            <div v-if="searchTotal > 0" class="search-pagination">
              <div class="pagination-info">
                共 {{ searchTotal }} 条结果，第 {{ searchCurrentPage }} / {{ totalPages }} 页
              </div>
              <div class="pagination-controls">
                <button
                    @click="searchCurrentPage > 1 && changePage(searchCurrentPage - 1)"
                    :disabled="searchCurrentPage <= 1 || searchLoading"
                    class="pagination-btn"
                >
                  上一页
                </button>
                <button
                    @click="searchCurrentPage < totalPages && changePage(searchCurrentPage + 1)"
                    :disabled="searchCurrentPage >= totalPages || searchLoading"
                    class="pagination-btn"
                >
                  下一页
                </button>
              </div>
            </div>
          </div>

          <div v-else-if="searchPerformed && !searchLoading" class="search-empty">
            <Icons name="search" size="2rem" class="empty-icon" />
            <p>未找到相关文件</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteModal" class="delete-modal-overlay" @click="closeDeleteModal">
      <div class="delete-modal" @click.stop>
        <div class="delete-modal-header">
          <h3>确认删除</h3>
          <button @click="closeDeleteModal" class="close-btn">
            <Icons name="x" size="1.2rem" />
          </button>
        </div>

        <div class="delete-modal-body">
          <div class="delete-warning">
            <Icons name="alert-triangle" size="2rem" class="warning-icon" />
            <p>您确定要删除以下{{ deleteTarget?.isFolder ? '文件夹' : '文件' }}吗？</p>
            <div class="delete-item-info">
              <Icons
                  :name="getFileIcon(deleteTarget)"
                  size="1.5rem"
                  class="file-icon"
                  :class="{ 'folder-icon': deleteTarget?.isFolder }"
              />
              <span class="item-name">{{ deleteTarget?.name }}</span>
            </div>
            <div class="delete-warnings">
              <p class="delete-note danger">
                {{ deleteTarget?.isFolder ? '此操作将删除文件夹及其所有内容，' : '' }}删除后无法恢复，请谨慎操作！
              </p>
              <p class="delete-note warning">
                <Icons name="info" size="1rem" class="info-icon" />
                注意：如果您开启了定时同步或手动同步刷新功能，删除的文件可能会在下次同步时重新出现。
              </p>
            </div>
          </div>

          <div class="delete-actions">
            <button @click="closeDeleteModal" class="cancel-btn">
              取消
            </button>
            <button
                @click="performDelete"
                class="confirm-delete-btn"
                :disabled="deleteLoading"
            >
              <Icons v-if="deleteLoading" name="loader" size="1rem" class="loading-icon" />
              {{ deleteLoading ? '删除中...' : '确认删除' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 回到顶部按钮 -->
    <transition name="fade">
      <button
          v-if="showBackToTop"
          @click="scrollToTop"
          class="back-to-top-btn"
          title="回到顶部"
      >
        <Icons name="arrow-up" size="1.2rem" />
      </button>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fileApi, type FileItem } from '@/api/file'
import {storageApi, type SearchRequest, type SearchItem as SearchFileItem} from '@/api/storage'
import { useSettingStore } from '@/stores/setting'
import { toast } from '@/utils/toast'
import Icons from '@/components/Icons.vue'

const route = useRoute()
const router = useRouter()
const settingStore = useSettingStore()

// 响应式数据
const loading = ref(false)
const currentFile = ref<FileItem | null>(null)
const error = ref<{ title: string; message: string } | null>(null)
const fileBrowserRef = ref<HTMLElement | null>(null)
const showBackToTop = ref(false)
const headerFixed = ref(false)
const showStrmFiles = ref(false)

// 搜索相关数据
const showSearchModal = ref(false)
const searchKeyword = ref('')
const globalSearch = ref(false)
const searchLoading = ref(false)
const searchResults = ref<SearchFileItem[]>([])
const searchTotal = ref(0)
const searchCurrentPage = ref(1)
const searchPageSize = ref(15)
const searchPerformed = ref(false)

// 删除相关数据
const showDeleteModal = ref(false)
const deleteTarget = ref<FileItem | null>(null)
const deleteLoading = ref(false)

// 计算属性
const currentPath = computed(() => {
  const path = route.params.pathMatch
  if (path && typeof path === 'string') {
    // 对路径的每个段分别解码
    return path.split('/').map(segment => decodeURIComponent(segment)).join('/')
  } else if (Array.isArray(path)) {
    // 如果是数组，合并后处理
    const pathStr = path.join('/')
    return pathStr.split('/').map(segment => decodeURIComponent(segment)).join('/')
  }
  return ''
})

const pathSegments = computed(() => {
  return currentPath.value ? currentPath.value.split('/').filter(Boolean) : []
})

const fileList = computed(() => {
  return currentFile.value?.children || []
})

// 搜索相关计算属性
const totalPages = computed(() => {
  return Math.ceil(searchTotal.value / searchPageSize.value)
})

// 获取指定索引之前的路径
const getPathUpTo = (index: number): string => {
  return pathSegments.value.slice(0, index + 1).join('/')
}

// 导航到指定路径
const navigateToPath = (path: string) => {
  if (path) {
    // 对路径的每个段分别编码，保留路径分隔符
    const encodedPath = path.split('/').map(segment => encodeURIComponent(segment)).join('/')
    router.push(`/${encodedPath}`)
  } else {
    router.push('/')
  }
}

// 切换显示STRM文件
const toggleShowStrmFiles = () => {
  showStrmFiles.value = !showStrmFiles.value
  // 重新加载当前路径以应用新的过滤设置
  loadFile(currentPath.value)
}

// 处理文件/文件夹点击
const handleItemClick = (item: FileItem) => {
  // 直接导航到对应路径，让路由守卫来处理文件和文件夹的区分
  const newPath = item.href.startsWith('/') ? item.href.substring(1) : item.href
  router.push('/' + newPath)
}

// 下载文件
const downloadFile = (item: FileItem) => {
  if (item.downloadURL) {
    const link = document.createElement('a')
    link.href = item.downloadURL
    link.download = item.name
    link.target = '_blank'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }
}

// 删除相关方法
const confirmDelete = (item: FileItem) => {
  // 检查是否为挂载点文件
  if (item.isTop === 1) {
    toast.warning('挂载点文件请在后台存储管理删除')
    return
  }

  deleteTarget.value = item
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  deleteTarget.value = null
  deleteLoading.value = false
}

const performDelete = async () => {
  if (!deleteTarget.value) return

  deleteLoading.value = true

  try {
    // 构建文件路径
    const itemPath = deleteTarget.value.path.startsWith('/')
        ? deleteTarget.value.path.substring(1)
        : deleteTarget.value.path

    await fileApi.deleteFile(itemPath)

    toast.success(`${deleteTarget.value.isFolder ? '文件夹' : '文件'}删除成功`)

    // 关闭弹窗
    closeDeleteModal()

    // 刷新当前路径
    await loadFile(currentPath.value)

  } catch (error: any) {
    console.error('删除失败:', error)

    // 根据错误类型显示不同的提示信息
    if (error.code === 400) {
      toast.error(error.message || '删除失败，请检查文件权限')
    } else if (error.code === 404) {
      toast.error('文件不存在或已被删除')
      // 文件不存在时也刷新列表
      await loadFile(currentPath.value)
    } else {
      toast.error(error.message || '删除失败，请稍后重试')
    }
  } finally {
    deleteLoading.value = false
  }
}

// 获取文件图标
const getFileIcon = (item: { isFolder: number; name: string} | null): string => {
  if (!item) return 'file'

  if (item.isFolder) {
    return 'folder'
  }

  const ext = item.name.split('.').pop()?.toLowerCase()
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
    case 'txt':
    case 'md':
      return 'file-text'
    case 'zip':
    case 'rar':
    case '7z':
      return 'archive'
    case 'strm':
      return 'video'
    default:
      return 'file'
  }
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

// 刷新当前路径
const refreshCurrentPath = () => {
  loadFile(currentPath.value)
}

// 刷新文件夹索引
const refreshFolderIndex = async (item: FileItem) => {
  try {
    await storageApi.deepRefreshFile({ id: item.id })
    toast.success('已下发刷新指令，请稍后查看结果~')
  } catch (error) {
    console.error('刷新文件夹索引失败:', error)
    toast.error('刷新文件夹索引失败')
  }
}

// 截断文字
const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) {
    return text
  }
  return text.substring(0, maxLength) + '...'
}

// 回到顶部相关方法
const handleScroll = () => {
  const scrollTop = window.pageYOffset || document.documentElement.scrollTop
  showBackToTop.value = scrollTop > 300
  headerFixed.value = scrollTop > 100
}

const scrollToTop = () => {
  window.scrollTo({
    top: 0,
    behavior: 'smooth'
  })
}

// 搜索相关方法
const openSearchModal = () => {
  showSearchModal.value = true
  searchKeyword.value = ''
  searchResults.value = []
  searchPerformed.value = false
  searchCurrentPage.value = 1
}

const closeSearchModal = () => {
  showSearchModal.value = false
  searchKeyword.value = ''
  searchResults.value = []
  searchPerformed.value = false
  globalSearch.value = false
}

const performSearch = async () => {
  if (!searchKeyword.value.trim()) {
    toast.warning('请输入搜索关键词')
    return
  }

  searchLoading.value = true
  searchCurrentPage.value = 1

  try {
    const searchParams: SearchRequest = {
      keyword: searchKeyword.value.trim(),
      pageSize: searchPageSize.value,
      currentPage: searchCurrentPage.value,
      global: globalSearch.value
    }

    if (!globalSearch.value && currentFile.value) {
      searchParams.pid = currentFile.value.id
    }

    const response = await storageApi.search(searchParams)
    searchResults.value = response.data
    searchTotal.value = response.total
    searchPerformed.value = true
  } catch (error: any) {
    console.error('搜索失败:', error)
    toast.error('搜索失败，请重试')
    searchResults.value = []
    searchTotal.value = 0
  } finally {
    searchLoading.value = false
  }
}

const changePage = async (page: number) => {
  if (page < 1 || page > totalPages.value || searchLoading.value) {
    return
  }

  searchCurrentPage.value = page
  searchLoading.value = true

  try {
    const searchParams: SearchRequest = {
      keyword: searchKeyword.value.trim(),
      pageSize: searchPageSize.value,
      currentPage: searchCurrentPage.value,
      global: globalSearch.value
    }

    if (!globalSearch.value && currentFile.value) {
      searchParams.pid = currentFile.value.id
    }

    const response = await storageApi.search(searchParams)
    searchResults.value = response.data
    searchTotal.value = response.total
  } catch (error: any) {
    console.error('搜索失败:', error)
    toast.error('搜索失败，请重试')
  } finally {
    searchLoading.value = false
  }
}

const navigateToSearchResult = (item: SearchFileItem) => {
  closeSearchModal()

  // 根据localPath导航到文件位置
  const path = item.localPath.startsWith('/') ? item.localPath.substring(1) : item.localPath

  if (item.isFolder) {
    // 如果是文件夹，直接导航到该路径
    navigateToPath(path)
  } else {
    // 如果是文件，导航到文件详情页
    router.push(`/file/${path}`)
  }
}

const getFileType = (item: SearchFileItem): string => {
  if (item.isFolder) {
    return '文件夹'
  }

  const ext = item.name.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'mp4':
    case 'avi':
    case 'mkv':
    case 'mov':
      return '视频'
    case 'mp3':
    case 'wav':
    case 'flac':
      return '音频'
    case 'jpg':
    case 'jpeg':
    case 'png':
    case 'gif':
      return '图片'
    case 'pdf':
      return 'PDF'
    case 'txt':
    case 'md':
      return '文本'
    case 'zip':
    case 'rar':
    case '7z':
      return '压缩包'
    case 'strm':
      return 'STRM'
    default:
      return '文件'
  }
}

// 加载文件/文件夹
const loadFile = async (path: string) => {
  try {
    loading.value = true
    error.value = null

    // 根据showStrmFiles状态决定是否包含STRM文件
    const options = {
      includeAutoGenerateStrmFile: showStrmFiles.value
    }

    const result = await fileApi.getFile(path, options)

    // 如果返回的是文件而不是文件夹，跳转到文件详情页面
    if (result && !result.isFolder && result.downloadURL) {
      router.push(`/file/${path}`)
      return
    }

    currentFile.value = result

  } catch (err: any) {
    console.error('加载文件失败:', err)

    if (err.code === 404) {
      error.value = {
        title: '文件不存在',
        message: err.message || '请求的文件或文件夹不存在'
      }
    } else {
      error.value = {
        title: '加载失败',
        message: err.message || '无法加载文件列表，请稍后重试'
      }
    }
  } finally {
    loading.value = false
  }
}

// 监听路由变化
watch(
    () => currentPath.value,
    (newPath) => {
      loadFile(newPath)
    },
    { immediate: true }
)

// 组件挂载时加载
onMounted(async () => {
  // 初始化设置store
  if (!settingStore.setting) {
    try {
      await settingStore.fetchSetting()
    } catch (error) {
      console.error('获取设置失败:', error)
    }
  }

  loadFile(currentPath.value)
  window.addEventListener('scroll', handleScroll)
})

// 组件卸载时清理
onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.file-browser {
  background: #f8fafc;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.browser-container {
  width: 100%;
  max-width: 980px;
}

.browser-container.header-fixed-padding {
  padding-top: 6rem;
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

.modern-header.header-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  width: 100%;
  max-width: none;
  z-index: 999;
  margin-bottom: 0;
  border-radius: 0;
  border-left: none;
  border-right: none;
  border-top: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 1rem;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
  min-width: 0;
}

.breadcrumb-nav {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  padding: 0.25rem 0;
  scrollbar-width: none;
  -ms-overflow-style: none;
  position: relative;
}

.breadcrumb-nav::-webkit-scrollbar {
  display: none;
}

.breadcrumb-nav:hover::after {
  opacity: 1;
}

.breadcrumb-home {
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

.breadcrumb-home:hover {
  background: #e5e7eb;
  color: #374151;
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
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
  white-space: nowrap;
  text-decoration: none;
}

.breadcrumb-link:hover {
  background: #f3f4f6;
  color: #374151;
}

.breadcrumb-current {
  color: #1f2937;
  font-size: 0.875rem;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  white-space: nowrap;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  color: #374151;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
  height: 2.5rem;
  flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.action-btn.secondary {
  background: #ffffff;
  border-color: #3b82f6;
  color: #3b82f6;
}

.action-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.action-btn:hover:not(:disabled) {
  background: #f9fafb;
  border-color: #9ca3af;
  color: #111827;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.action-btn.secondary:hover:not(:disabled) {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.action-btn.active:hover:not(:disabled) {
  background: #2563eb;
  border-color: #2563eb;
  color: white;
  box-shadow: 0 2px 4px rgba(37, 99, 235, 0.2);
}

.action-btn:active {
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  box-shadow: none;
}

.file-list-container {
  padding: 0;
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
  color: #ef4444;
  margin-bottom: 1rem;
}

.error-state h3 {
  color: #1f2937;
  margin-bottom: 0.5rem;
  font-size: 1.25rem;
}

.error-state p {
  color: #6b7280;
  margin-bottom: 1.5rem;
}

.retry-btn {
  background: #3b82f6;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background 0.2s;
}

.retry-btn:hover {
  background: #2563eb;
}

.file-list {
  background: white;
  border-radius: 0.75rem;
  border: 1px solid #e2e8f0;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.file-list-header {
  display: grid;
  grid-template-columns: 1fr 120px 180px 120px;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
  font-weight: 500;
  color: #374151;
  font-size: 0.875rem;
}

.file-item {
  display: grid;
  grid-template-columns: 1fr 120px 180px 120px;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid rgba(243, 244, 246, 0.5);
  cursor: pointer;
  align-items: center;
  position: relative;
}

.file-item:hover {
  background: #f8fafc;
}

.file-item:last-child {
  border-bottom: none;
}

.file-item.parent-dir {
  background: #fef3c7;
  border-bottom: 1px solid #fbbf24;
}

.file-item.parent-dir:hover {
  background: #fef3c7;
}

.file-col-name {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.file-icon {
  color: #6b7280;
  flex-shrink: 0;
  width: 1.5rem;
  height: 1.5rem;
}

.file-icon.folder-icon {
  color: #3b82f6;
}

.file-name {
  truncate: true;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
  color: #1f2937;
  font-size: 0.875rem;
}

.file-col-size,
.file-col-date,
.file-col-actions {
  font-size: 0.9rem;
  color: #6b7280;
  font-weight: 500;
  transition: color 0.3s ease;
}

.file-col-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.file-item:hover .file-col-size,
.file-item:hover .file-col-date {
  color: #374151;
}

.action-btn-small {
  background: none;
  border: 1px solid #d1d5db;
  color: #6b7280;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  transition: all 0.2s;
  font-size: 0.75rem;
  white-space: nowrap;
}

.action-btn-small:hover {
  background: #f3f4f6;
  color: #374151;
}

.refresh-btn {
  border-color: #3b82f6;
  color: #3b82f6;
}

.refresh-btn:hover {
  background: #eff6ff;
  border-color: #2563eb;
  color: #2563eb;
}

/* 删除按钮样式 */
.delete-btn {
  border-color: #ef4444;
  color: #ef4444;
}

.delete-btn:hover:not(:disabled) {
  background: #fef2f2;
  border-color: #dc2626;
  color: #dc2626;
}

.delete-btn:disabled {
  border-color: #d1d5db;
  color: #9ca3af;
  cursor: not-allowed;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 2rem;
  color: #6b7280;
  background: transparent;
  margin: 2rem 0;
}

.empty-icon {
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* 搜索弹窗样式 */
.search-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.search-modal {
  background: white;
  border-radius: 1rem;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  width: 100%;
  max-width: 800px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.search-modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 0.5rem;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

.search-modal-body {
  padding: 1.5rem;
  flex: 1;
}

.search-form {
  margin-bottom: 1.5rem;
}

.search-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.search-options-inline {
  flex-shrink: 0;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 0.5rem 0.75rem;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.search-input-group {
  display: flex;
  gap: 0.75rem;
  flex: 1;
}

.search-input {
  flex: 1;
  padding: 0.75rem 1rem;
  border: 1px solid #d1d5db;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.search-btn {
  background: #3b82f6;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background 0.2s;
}

.search-btn:hover:not(:disabled) {
  background: #2563eb;
}

.search-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.inline-checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: #374151;
  font-weight: 500;
  white-space: nowrap;
  transition: all 0.2s ease;
}

.inline-checkbox-label:hover {
  color: #1f2937;
}

.inline-checkbox {
  width: 1.125rem;
  height: 1.125rem;
  border: 2px solid #d1d5db;
  border-radius: 0.375rem;
  accent-color: #3b82f6;
  cursor: pointer;
  transition: all 0.2s ease;
}

.inline-checkbox:hover {
  border-color: #9ca3af;
}

.inline-checkbox:checked {
  background-color: #3b82f6;
  border-color: #3b82f6;
  transform: scale(1.05);
}

.inline-checkbox-text {
  user-select: none;
  color: #4b5563;
  font-weight: 500;
}

.text-left {
  text-align: left;
}

.text-center {
  text-align: center;
}

.search-results {
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  overflow: hidden;
}

.search-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: #6b7280;
}

.search-list {
  background: white;
}

.search-list-header {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 2fr;
  gap: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
  font-weight: 500;
  color: #374151;
  font-size: 0.875rem;
}

.search-item {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 2fr;
  gap: 1rem;
  padding: 1rem;
  border-bottom: 1px solid rgba(243, 244, 246, 0.5);
  cursor: pointer;
  align-items: center;
  transition: background 0.2s;
}

.search-item:hover {
  background: #f8fafc;
}

.search-item:last-child {
  border-bottom: none;
}

.search-col-name {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.search-col-type,
.search-col-size,
.search-col-path {
  font-size: 0.875rem;
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #f9fafb;
  border-top: 1px solid #e5e7eb;
}

.pagination-info {
  font-size: 0.875rem;
  color: #6b7280;
}

.pagination-controls {
  display: flex;
  gap: 0.5rem;
}

.pagination-btn {
  background: white;
  border: 1px solid #d1d5db;
  color: #374151;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.pagination-btn:hover:not(:disabled) {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.search-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: #6b7280;
  text-align: center;
}

.search-empty .empty-icon {
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* 删除确认弹窗样式 */
.delete-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.delete-modal {
  background: white;
  border-radius: 1rem;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  width: 100%;
  max-width: 520px;
  overflow: hidden;
}

.delete-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
  background: #fef2f2;
}

.delete-modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #dc2626;
}

.delete-modal-body {
  padding: 1.5rem;
}

.delete-warning {
  text-align: center;
  margin-bottom: 2rem;
}

.warning-icon {
  color: #f59e0b;
  margin-bottom: 1rem;
}

.delete-warning > p {
  color: #374151;
  font-size: 1rem;
  margin-bottom: 1rem;
  font-weight: 500;
}

.delete-item-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 0.5rem;
  margin: 1rem 0;
}

.item-name {
  font-weight: 500;
  color: #1f2937;
  font-size: 1rem;
}

.delete-warnings {
  text-align: left;
  margin-top: 1.5rem;
}

.delete-note {
  margin: 0.75rem 0;
  line-height: 1.5;
  font-size: 0.875rem;
}

.delete-note.danger {
  color: #dc2626;
  font-weight: 500;
  background: #fef2f2;
  padding: 0.75rem;
  border-radius: 0.5rem;
  border-left: 4px solid #ef4444;
}

.delete-note.warning {
  color: #92400e;
  background: #fef3c7;
  padding: 1rem;
  border-radius: 0.5rem;
  border-left: 4px solid #f59e0b;
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
}

.info-icon {
  color: #f59e0b;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.delete-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.cancel-btn {
  background: white;
  border: 1px solid #d1d5db;
  color: #374151;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
}

.cancel-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.confirm-delete-btn {
  background: #ef4444;
  border: 1px solid #ef4444;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.confirm-delete-btn:hover:not(:disabled) {
  background: #dc2626;
  border-color: #dc2626;
}

.confirm-delete-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading-icon {
  animation: spin 1s linear infinite;
}

/* 回到顶部按钮样式 */
.back-to-top-btn {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  width: 3rem;
  height: 3rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  transition: all 0.3s ease;
  z-index: 1000;
}

.back-to-top-btn:hover {
  background: #2563eb;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
}

.back-to-top-btn:active {
  transform: translateY(0);
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .file-browser {
    padding: 1rem;
  }

  .modern-header {
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
  }

  .breadcrumb-nav {
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .file-list-header,
  .file-item {
    grid-template-columns: 1fr 80px;
    gap: 0.5rem;
  }

  .file-col-size,
  .file-col-date {
    display: none;
  }

  .file-list-container {
    padding: 1rem;
  }

  .search-modal {
    margin: 1rem;
    max-height: 90vh;
  }

  .search-modal-header,
  .search-modal-body {
    padding: 1rem;
  }

  .search-input-group {
    flex-direction: column;
  }

  .search-list-header,
  .search-item {
    grid-template-columns: 1fr;
    gap: 0.5rem;
  }

  .search-col-type,
  .search-col-size {
    display: none;
  }

  .pagination-controls {
    flex-direction: column;
    gap: 0.25rem;
  }

  .delete-modal {
    margin: 1rem;
  }

  .delete-actions {
    flex-direction: column;
  }

  .cancel-btn,
  .confirm-delete-btn {
    width: 100%;
    justify-content: center;
  }

  .delete-note.warning {
    flex-direction: column;
    gap: 0.5rem;
  }
}

@media (max-width: 480px) {
  .file-browser {
    padding: 0.5rem;
  }

  .modern-header {
    padding: 0.75rem;
  }

  .breadcrumb-link {
    padding: 0.4rem 0.6rem;
    font-size: 0.75rem;
  }

  .action-btn {
    padding: 0.5rem 1rem;
    font-size: 0.8rem;
  }
}
</style>
