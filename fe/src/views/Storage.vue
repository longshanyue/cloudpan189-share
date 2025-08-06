<template>
  <Layout>
    <!-- 存储管理主卡片 -->
    <PageCard title="存储管理" subtitle="管理存储挂载点和令牌绑定">
      <SectionDivider />

      <SubsectionTitle title="存储列表" />
      <!-- 操作栏 -->
      <div class="action-bar">
        <div class="search-section">
          <input v-model="searchName" type="text" placeholder="搜索存储名称..." class="search-input" @input="handleSearch">
        </div>
        <div class="action-buttons-group">
          <button @click="fetchStorages" class="btn btn-secondary" :disabled="loading">
            <Icons name="refresh" size="1rem" class="btn-icon" />
            {{ loading ? '刷新中...' : '刷新' }}
          </button>
          <button @click="showAddModal = true" class="btn btn-primary">
            <Icons name="add" size="1rem" class="btn-icon" />
            添加存储
          </button>
        </div>
      </div>

      <!-- 存储列表 -->
      <div class="storage-table-container">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <p>加载中...</p>
        </div>

        <div v-else-if="storages.length === 0" class="empty-state">
          <Icons name="storage" size="3rem" class="empty-icon" />
          <h3>暂无存储</h3>
          <p>{{ searchName ? '没有找到匹配的存储' : '还没有存储，点击上方按钮添加第一个存储' }}</p>
        </div>

        <table v-else class="storage-table">
          <thead>
            <tr>
              <th style="text-align: center">序号</th>
              <th style="text-align: center">挂载路径</th>
              <th style="text-align: center">协议类型</th>
              <th style="width: 260px; text-align: center">任务状态</th>
              <th style="text-align: center">令牌绑定</th>
              <th style="text-align: center">创建时间</th>
              <th style="text-align: center">修改时间</th>
              <th style="text-align: center">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(storage, index) in storages" :key="storage.id" class="storage-row">
              <td>{{ (currentPage - 1) * pageSize + index + 1 }}</td>
              <td>
                <div class="storage-info">
                  <div class="storage-icon">
                    <Icons name="storage" size="1.2rem" />
                  </div>
                  <div>
                    <div class="storage-name">{{ storage.localPath }}</div>
                  </div>
                </div>
              </td>
              <td>
                <span class="protocol-badge" :class="`protocol-${storage.osType}`">
                  {{ storage.osType === 'subscribe' ? '订阅' : '分享' }}
                </span>
              </td>
              <td>
                <div v-if="storage.jobStatus" class="job-status">
                  <div class="job-status-inline">
                    <span class="status-badge"
                      :class="storage.jobStatus.status === 0 ? 'status-waiting' : 'status-running'">
                      {{ JobStatusMap[storage.jobStatus.jobType] }}：
                      {{ storage.jobStatus.status === 0 ? '等待执行中' : '执行中' }}
                      <span
                        v-if="storage.jobStatus.status === 1 && (storage.jobStatus.jobType === 'refresh' || storage.jobStatus.jobType === 'deep_refresh')">
                        ，进度: {{ storage.jobStatus.scannedCount || 0 }} / {{ storage.jobStatus.waitCount || 0 }}
                      </span>
                    </span>
                  </div>
                </div>
                <span v-else class="no-job">-</span>
              </td>
              <td>
                <div v-if="storage.addition?.cloud_token" class="token-info">
                  <div class="token-name">令牌 #{{ storage.addition.cloud_token }}</div>
                  <div class="token-expire" v-if="getTokenById(storage.addition.cloud_token)">
                    {{ getTokenExpireText(getTokenById(storage.addition.cloud_token)) }}
                  </div>
                </div>
                <span v-else class="no-token">未绑定</span>
              </td>
              <td>{{ formatDate(storage.createdAt) }}</td>
              <td>{{ formatDate(storage.updatedAt) }}</td>
              <td>
                <div class="action-buttons">
                  <button @click="bindToken(storage)" class="btn btn-sm btn-secondary">
                    {{ storage.addition.cloud_token ? '重绑令牌' : '绑定令牌' }}
                  </button>
                  <button @click="deleteStorage(storage)" class="btn btn-sm btn-danger">
                    删除
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页组件 -->
        <Pagination v-if="total > 0" :current-page="currentPage" :page-size="pageSize" :total="total"
          @page-change="handlePageChange" @page-size-change="handlePageSizeChange" />
      </div>
    </PageCard>

    <!-- 添加存储弹窗 -->
    <div v-if="showAddModal" class="modal-overlay" @click="closeAddModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>添加存储</h3>
          <button @click="closeAddModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">
              <div>本地路径</div>
              <div style="font-size: 11px; color: #6b7280;">
                必须 / 开头，但不能只输入 / ，比如 /音视频文件夹/视频文件夹
              </div>
            </label>
            <input v-model="newStorage.localPath" type="text" class="form-input" placeholder="请输入本地路径" />
          </div>
          <div class="form-group">
            <label class="form-label">协议类型</label>
            <Select v-model="newStorage.protocol" :options="protocolOptions" placeholder="请选择协议类型" />
          </div>
          <div class="form-group">
            <label class="form-label">
              <div>云盘令牌（可选）</div>
              <div style="font-size: 11px; color: #6b7280;">
                如果没有绑定令牌，资源无法在线观看或下载
              </div>
            </label>
            <Select v-model="newStorage.cloudToken" :options="tokenOptions" placeholder="请选择令牌" />
          </div>
          <div v-if="newStorage.protocol === 'subscribe'" class="form-group">
            <label class="form-label">订阅用户ID</label>
            <input v-model="newStorage.subscribeUser" type="text" class="form-input" placeholder="请输入订阅用户ID" />
          </div>
          <div v-if="newStorage.protocol === 'share'">
            <div class="form-group">
              <label class="form-label">分享码</label>
              <input v-model="newStorage.shareCode" type="text" class="form-input" placeholder="请输入分享码" />
            </div>
            <div class="form-group">
              <label class="form-label">访问码（可选）</label>
              <input v-model="newStorage.shareAccessCode" type="text" class="form-input" placeholder="请输入访问码（可选）" />
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeAddModal" class="btn btn-secondary">取消</button>
          <button @click="confirmAddStorage" class="btn btn-primary" :disabled="addLoading">
            {{ addLoading ? '添加中...' : '确认添加' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 绑定令牌弹窗 -->
    <div v-if="showBindModal" class="modal-overlay" @click="closeBindModal">
      <div class="modal-content small" @click.stop>
        <div class="modal-header">
          <h3>{{ bindingStorage?.addition.cloud_token ? '重新绑定令牌' : '绑定令牌' }}</h3>
          <button @click="closeBindModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <p class="bind-info">为存储 <strong>{{ bindingStorage?.name }}</strong> {{ bindingStorage?.addition.cloud_token ?
            '重新绑定' : '绑定' }}令牌</p>
          <div class="form-group">
            <label class="form-label">选择令牌</label>
            <Select v-model="selectedTokenId" :options="tokenOptions" placeholder="请选择令牌" />
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeBindModal" class="btn btn-secondary">取消</button>
          <button @click="confirmBindToken" class="btn btn-primary" :disabled="bindLoading">
            {{ bindLoading ? '绑定中...' : '确认绑定' }}
          </button>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import Layout from '@/components/Layout.vue'
import Icons from '@/components/Icons.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'
import { ref, onMounted, computed, reactive, onUnmounted } from 'vue'
import { storageApi, type Storage, type AddStorageRequest } from '@/api/storage'
import { cloudTokenApi, type CloudToken } from '@/api/cloudtoken'
import { JobStatusMap } from '@/api/shared'
import { toast } from '@/utils/toast'
import { confirmDialog } from '@/utils/confirm'
import Select from '@/components/Select.vue'
import Pagination from '@/components/Pagination.vue'

// 响应式数据
const loading = ref(false)
const storages = ref<Storage[]>([])
const searchName = ref('')
const currentPage = ref(1)
const pageSize = ref(15)
const total = ref(0)
const showAddModal = ref(false)
const showBindModal = ref(false)
const addLoading = ref(false)
const bindLoading = ref(false)
const availableTokens = ref<CloudToken[]>([])
const bindingStorage = ref<Storage | null>(null)
const selectedTokenId = ref('')

// 新存储表单数据
const newStorage = reactive({
  localPath: '',
  protocol: 'subscribe',
  cloudToken: '',
  subscribeUser: '',
  shareCode: '',
  shareAccessCode: ''
})

// 协议选项
const protocolOptions = [
  { label: '订阅类型', value: 'subscribe' },
  { label: '分享类型', value: 'share' }
]

// 令牌选项
const tokenOptions = computed(() => {
  return availableTokens.value.map(token => ({
    label: token.name,
    value: token.id
  }))
})

// 获取存储列表
const fetchStorages = async () => {
  try {
    loading.value = true
    const response = await storageApi.list({
      currentPage: currentPage.value,
      pageSize: pageSize.value,
      name: searchName.value || undefined
    })
    storages.value = response.data
    total.value = response.total
    currentPage.value = response.currentPage
  } catch (error) {
    console.error('获取存储列表失败:', error)
    toast.error('获取存储列表失败')
  } finally {
    loading.value = false
  }
}

let refreshTimer: ReturnType<typeof setInterval>
onMounted(() => {
  refreshTimer = setInterval(() => {
    fetchStorages()
  }, 30 * 1000)
})

onUnmounted(() => {
  clearInterval(refreshTimer)
})

// 获取可用令牌列表
const fetchAvailableTokens = async () => {
  try {
    const response = await cloudTokenApi.list()
    availableTokens.value = response || []
  } catch (error) {
    console.error('获取令牌列表失败:', error)
    availableTokens.value = []
  }
}

// 分页处理
const handlePageChange = (page: number) => {
  if (page !== currentPage.value) {
    currentPage.value = page
    fetchStorages()
  }
}

const handlePageSizeChange = (size: number) => {
  if (size !== pageSize.value) {
    pageSize.value = size
    currentPage.value = 1
    localStorage.setItem('storageListPageSize', size.toString())
    fetchStorages()
  }
}

// 添加存储相关
const closeAddModal = () => {
  showAddModal.value = false
  Object.assign(newStorage, {
    localPath: '',
    protocol: 'subscribe',
    cloudToken: '',
    subscribeUser: '',
    shareCode: '',
    shareAccessCode: ''
  })
}

const confirmAddStorage = async () => {
  if (!newStorage.localPath.trim()) {
    toast.warning('请输入本地路径')
    return
  }

  if (newStorage.protocol === 'subscribe') {
    if (!newStorage.subscribeUser.trim()) {
      toast.warning('请输入订阅用户ID')
      return
    }
  } else if (newStorage.protocol === 'share') {
    if (!newStorage.shareCode.trim()) {
      toast.warning('请输入分享码')
      return
    }
  }

  try {
    addLoading.value = true
    const requestData: AddStorageRequest = {
      localPath: newStorage.localPath.trim(),
      protocol: newStorage.protocol
    }

    if (newStorage.cloudToken) {
      requestData.cloudToken = Number(newStorage.cloudToken)
    }

    if (newStorage.protocol === 'subscribe') {
      requestData.subscribeUser = newStorage.subscribeUser.trim()
    } else if (newStorage.protocol === 'share') {
      requestData.shareCode = newStorage.shareCode.trim()
      if (newStorage.shareAccessCode.trim()) {
        requestData.shareAccessCode = newStorage.shareAccessCode.trim()
      }
    }

    await storageApi.add(requestData)
    toast.success('添加存储成功，后台异步执行，如未显示，稍后刷新重新查看')
    closeAddModal()
    fetchStorages()
  } catch (error) {
    console.error('添加存储失败:', error)
    toast.error('添加存储失败')
  } finally {
    addLoading.value = false
  }
}

// 绑定令牌相关
const bindToken = (storage: Storage) => {
  bindingStorage.value = storage
  selectedTokenId.value = storage.addition.cloud_token?.toString() || ''
  showBindModal.value = true
}

const closeBindModal = () => {
  showBindModal.value = false
  bindingStorage.value = null
  selectedTokenId.value = ''
}

const confirmBindToken = async () => {
  if (!selectedTokenId.value) {
    toast.warning('请选择令牌')
    return
  }
  if (!bindingStorage.value) {
    return
  }

  try {
    bindLoading.value = true
    await storageApi.modifyToken({
      id: bindingStorage.value.id,
      cloudToken: parseInt(selectedTokenId.value)
    })
    toast.success('绑定令牌成功')
    closeBindModal()
    fetchStorages()
  } catch (error) {
    console.error('绑定令牌失败:', error)
    toast.error('绑定令牌失败')
  } finally {
    bindLoading.value = false
  }
}

// 删除存储
const deleteStorage = async (storage: Storage) => {
  const confirmed = await confirmDialog({
    title: '删除存储',
    message: `确定要删除存储 "${storage.name}" 吗？此操作不可恢复。`,
    confirmText: '删除',
    cancelText: '取消',
    isDanger: true
  })

  if (!confirmed) {
    return
  }

  try {
    await storageApi.delete({ id: storage.id })
    toast.success('删除存储成功')
    fetchStorages()
  } catch (error: any) {
    if (error?.message) {
      toast.error(error.message)
    } else {
      toast.error('删除存储失败')
    }
    console.error('删除存储失败:', error)
  }
}

// 格式化日期
const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 根据ID获取令牌信息
const getTokenById = (tokenId: number) => {
  return availableTokens.value.find(token => token.id === tokenId)
}

// 获取令牌到期时间文本
const getTokenExpireText = (token: CloudToken | undefined): string => {
  if (!token) return ''

  // expiresIn是13位时间戳，表示到期时间
  const expireTime = token.expiresIn
  const now = Date.now()

  if (expireTime <= now) {
    return '已过期'
  }

  const diffMs = expireTime - now
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  const diffHours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))

  if (diffDays > 0) {
    return `${diffDays}天后到期`
  } else if (diffHours > 0) {
    return `${diffHours}小时后到期`
  } else {
    return '即将到期'
  }
}

// 搜索处理
let searchTimer: NodeJS.Timeout
const handleSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1 // 搜索时重置到第一页
    fetchStorages()
  }, 500)
}

// 组件挂载时获取数据
onMounted(() => {
  // 从localStorage恢复页面大小设置
  const savedPageSize = localStorage.getItem('storageListPageSize')
  if (savedPageSize) {
    pageSize.value = parseInt(savedPageSize)
  }

  fetchStorages()
  fetchAvailableTokens()
})
</script>

<style scoped>
/* 操作栏样式 */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f9fafb;
  padding: 1.5rem;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  margin-bottom: 1.5rem;
}

.search-section {
  flex: 1;
  max-width: 300px;
}

.search-input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
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

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #4b5563;
}

.btn-warning {
  background: #f59e0b;
  color: white;
}

.btn-warning:hover:not(:disabled) {
  background: #d97706;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
}

.btn-sm {
  padding: 0.5rem 0.75rem;
  font-size: 0.75rem;
}

.btn-icon {
  font-size: 1rem;
}

/* 表格容器样式 */
.storage-table-container {
  background: white;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

/* 加载状态 */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  color: #6b7280;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e5e7eb;
  border-top: 3px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  text-align: center;
  color: #6b7280;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.6;
}

.empty-state h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #374151;
  margin: 0 0 0.5rem 0;
}

.empty-state p {
  margin: 0;
  font-size: 0.875rem;
}

/* 表格样式 */
.storage-table {
  width: 100%;
  border-collapse: collapse;
  text-align: center;
}

.storage-table th {
  background: #f9fafb;
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
}

.storage-table td {
  padding: 1rem;
  border-bottom: 1px solid #f3f4f6;
  font-size: 0.875rem;
}

.storage-row:hover {
  background: #f9fafb;
}

.storage-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.storage-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 0.875rem;
}

.storage-name {
  font-weight: 500;
  color: #1f2937;
}

.protocol-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.protocol-subscribe {
  background: #eff6ff;
  color: #1d4ed8;
}

.protocol-share {
  background: #fef3c7;
  color: #d97706;
}

.token-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.token-name {
  color: #059669;
  font-weight: 500;
}

.token-expire {
  color: #6b7280;
  font-size: 0.75rem;
}

.no-token {
  color: #9ca3af;
  font-style: italic;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.action-buttons-group {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.action-buttons-group .btn {
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
  font-weight: 600;
  letter-spacing: 0.025em;
}

.action-buttons-group .btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s;
}

.action-buttons-group .btn:hover::before {
  left: 100%;
}

.action-buttons-group .btn-secondary {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: 1px solid #10b981;
  box-shadow: 0 2px 4px 0 rgba(16, 185, 129, 0.2), 0 1px 2px 0 rgba(16, 185, 129, 0.1);
}

.action-buttons-group .btn-secondary:hover:not(:disabled) {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px 0 rgba(16, 185, 129, 0.3), 0 2px 4px 0 rgba(16, 185, 129, 0.2);
}

.action-buttons-group .btn-secondary:active {
  transform: translateY(0);
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1);
}

.action-buttons-group .btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: 1px solid #3b82f6;
  box-shadow: 0 2px 4px 0 rgba(59, 130, 246, 0.2), 0 1px 2px 0 rgba(59, 130, 246, 0.1);
}

.action-buttons-group .btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 8px 0 rgba(59, 130, 246, 0.3), 0 2px 4px 0 rgba(59, 130, 246, 0.2);
}

.action-buttons-group .btn-primary:active {
  transform: translateY(0);
  box-shadow: 0 1px 2px 0 rgba(59, 130, 246, 0.2);
}

.action-buttons-group .btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05) !important;
}

.action-buttons-group .btn-icon {
  transition: transform 0.3s ease;
}

.action-buttons-group .btn:hover:not(:disabled) .btn-icon {
  transform: scale(1.1);
}

/* 任务状态样式 */
.job-status {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 160px;
}

.job-status-inline {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.job-type {
  display: flex;
  align-items: center;
}

.job-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.375rem 0.875rem;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.025em;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  border: 1px solid transparent;
}

.job-del {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #dc2626;
  border-color: #fecaca;
}

.job-refresh {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #2563eb;
  border-color: #bfdbfe;
}

.job-deep_refresh {
  background: linear-gradient(135deg, #faf5ff 0%, #f3e8ff 100%);
  color: #7c3aed;
  border-color: #e9d5ff;
}

.job-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  border: 1px solid transparent;
  position: relative;
  overflow: hidden;
}

.status-waiting {
  background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  color: #d97706;
  border-color: #fed7aa;
}

.status-waiting::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #f59e0b;
}

.status-running {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  color: #16a34a;
  border-color: #bbf7d0;
  animation: pulse-green 2s infinite;
}

.status-running::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #22c55e;
}

@keyframes pulse-green {

  0%,
  100% {
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05), 0 0 0 0 rgba(34, 197, 94, 0.4);
  }

  50% {
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05), 0 0 0 4px rgba(34, 197, 94, 0.2);
  }
}

.progress-info {
  margin-top: 0.375rem;
  padding: 0.5rem;
  background: #f8fafc;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
}

.progress-text {
  font-size: 0.75rem;
  color: #475569;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.progress-text::before {
  content: '📊';
  font-size: 0.875rem;
}

.no-job {
  color: #94a3b8;
  font-style: italic;
  font-size: 0.875rem;
  text-align: center;
  padding: 1rem 0;
}

/* 弹窗样式 */
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
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-content.small {
  max-width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #6b7280;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-btn:hover {
  background: #f3f4f6;
}

.modal-body {
  padding: 1.5rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

/* 表单样式 */
.form-group {
  margin-bottom: 1.5rem;
}

.form-label {
  display: block;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.bind-info {
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
  font-size: 0.875rem;
  color: #92400e;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .action-bar {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }

  .search-section {
    max-width: none;
  }

  .action-buttons-group {
    flex-direction: column;
    gap: 0.5rem;
  }

  .action-buttons-group .btn {
    width: 100%;
    justify-content: center;
  }

  .storage-table {
    font-size: 0.75rem;
  }

  .storage-table th,
  .storage-table td {
    padding: 0.75rem 0.5rem;
  }

  .action-buttons {
    flex-direction: column;
  }

  .modal-content {
    margin: 0.5rem;
    max-width: none;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 1rem;
  }
}

@media (max-width: 640px) {
  .storage-table {
    display: block;
    overflow-x: auto;
    white-space: nowrap;
  }
}
</style>
