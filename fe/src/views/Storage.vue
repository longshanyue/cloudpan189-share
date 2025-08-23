<template>
  <div>
    <!-- 存储管理主卡片 -->
    <PageCard title="存储管理" subtitle="管理存储挂载点和令牌绑定">
      <template #extra>
        <div class="scan-info-panel">
          <div class="scan-info-item">
            <span class="info-label">自动扫描:</span>
            <span class="info-value" :class="settingStore.setting?.enableTopFileAutoRefresh ? 'status-enabled' : 'status-disabled'">
              {{ settingStore.setting?.enableTopFileAutoRefresh ? '已开启' : '已关闭' }}
            </span>
          </div>
          <div class="scan-info-item">
            <span class="info-label">扫描间隔:</span>
            <span class="info-value">{{ settingStore.setting?.autoRefreshMinutes || 10 }}分钟</span>
          </div>
          <div class="scan-info-item">
            <span class="info-label">任务线程:</span>
            <span class="info-value">{{ settingStore.setting?.jobThreadCount || 1 }}个</span>
          </div>
        </div>
      </template>
      <SectionDivider />

      <SubsectionTitle title="存储列表" />
      <!-- 操作栏 -->
      <div class="action-bar">
        <div class="search-section">
          <input v-model="searchName" type="text" placeholder="搜索存储名称..." class="search-input" @input="handleSearch">
        </div>
        <div class="action-buttons-group">
          <button 
            v-if="selectedCount > 0" 
            @click="openBatchBindModal" 
            class="btn btn-warning"
          >
            <Icons name="link" size="1rem" class="btn-icon" />
            批量绑定令牌 ({{ selectedCount }})
          </button>
          <button @click="scanTopFiles" class="btn btn-info" :disabled="scanTopLoading">
            <Icons name="search" size="1rem" class="btn-icon" />
            {{ scanTopLoading ? '扫描中...' : '扫描所有文件' }}
          </button>
          <button @click="fetchStorages" class="btn btn-secondary" :disabled="loading">
            <Icons name="refresh" size="1rem" class="btn-icon" />
            {{ loading ? '刷新中...' : '刷新' }}
          </button>
          <button @click="showSmartParseModal = true" class="btn btn-warning">
            <Icons name="link" size="1rem" class="btn-icon" />
            批量解析
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
              <th style="width: 50px; text-align: center">
                <input 
                  type="checkbox" 
                  :checked="isAllSelected" 
                  @change="toggleSelectAll"
                  class="checkbox"
                />
              </th>
              <th style="text-align: center">序号</th>
              <th style="text-align: center">挂载路径</th>
              <th style="width: 100px; text-align: center">协议类型</th>
              <th style="width: 260px; text-align: center">任务状态</th>
              <th style="text-align: center">令牌绑定</th>
              <th style="text-align: center">创建时间</th>
              <th style="text-align: center">修改时间</th>
              <th style="width: 350px; text-align: center">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(storage, index) in storages" :key="storage.id" class="storage-row">
              <td>
                <input 
                  type="checkbox" 
                  :checked="selectedStorageIds.has(storage.id)"
                  @change="toggleStorageSelection(storage.id)"
                  class="checkbox"
                />
              </td>
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
                  {{ getProtocolLabel(storage.osType) }}
                </span>
              </td>
              <td>
                <div v-if="storage.fileScanStat" class="scan-status">
                  <div class="scan-status-inline">
                    <span class="status-badge status-running">
                      刷新中：已完成 {{ storage.fileScanStat.scannedCount }} / 待扫描 {{ storage.fileScanStat.waitCount }}
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
                  <button @click="toggleAutoScan(storage)" class="btn btn-sm" 
                    :class="storage.addition.disable_auto_scan ? 'btn-success' : 'btn-info'"
                    :disabled="toggleAutoScanLoading.has(storage.id)">
                    <Icons :name="storage.addition.disable_auto_scan ? 'play' : 'pause'" size="0.875rem" class="btn-icon" />
                    {{ toggleAutoScanLoading.has(storage.id) ? '处理中...' : 
                       (storage.addition.disable_auto_scan ? '启用扫描' : '禁用扫描') }}
                  </button>
                  <button @click="refreshStorage(storage)" class="btn btn-sm btn-warning" :disabled="refreshingStorageIds.has(storage.id)">
                    <Icons name="refresh" size="0.875rem" class="btn-icon" />
                    {{ refreshingStorageIds.has(storage.id) ? '扫描中...' : '扫描文件' }}
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
              <div>云盘令牌{{ newStorage.protocol === 'person' || newStorage.protocol === 'family' ? '（必选）' : '（可选）' }}</div>
              <div style="font-size: 11px; color: #6b7280;">
                {{ newStorage.protocol === 'person' || newStorage.protocol === 'family' ?
                   '个人和家庭类型必须选择云盘令牌' :
                   '如果没有绑定令牌，资源无法在线观看或下载' }}
              </div>
            </label>
            <Select
              v-model="newStorage.cloudToken"
              :options="tokenOptions"
              placeholder="请选择令牌"
              @change="watchCloudToken"
            />
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
          <div v-if="newStorage.protocol === 'person'">
            <div class="form-group">
              <label class="form-label">选择文件夹</label>
              <div v-if="!newStorage.cloudToken" class="file-selector-placeholder">
                请先选择云盘令牌
              </div>
              <FileTreeSelector
                v-else
                v-model="newStorage.fileId"
                :cloud-token="newStorage.cloudToken"
                @select="handlePersonFileSelect"
              />
            </div>
            <div v-if="selectedPersonFileName" class="form-group">
              <label class="form-label">已选择文件夹</label>
              <div class="selected-file-info">
                <Icons name="folder" size="1rem" />
                <span>{{ selectedPersonFileName }}</span>
                <button type="button" @click="clearPersonFileSelection" class="clear-selection">
                  <Icons name="close" size="0.875rem" />
                </button>
              </div>
            </div>
          </div>
          <div v-if="newStorage.protocol === 'family'">
            <div class="form-group">
              <label class="form-label">选择家庭</label>
              <div v-if="!newStorage.cloudToken" class="file-selector-placeholder">
                请先选择云盘令牌
              </div>
              <div v-else-if="familyListLoading" class="file-selector-loading">
                <div class="loading-spinner-sm"></div>
                <span>加载家庭列表中...</span>
              </div>
              <Select
                v-else
                v-model="newStorage.familyId"
                :options="familyOptions"
                placeholder="请选择家庭"
                @change="handleFamilyChange"
              />
            </div>
            <div class="form-group">
              <label class="form-label">选择文件夹</label>
              <div v-if="!newStorage.cloudToken || !newStorage.familyId" class="file-selector-placeholder">
                请先选择云盘令牌和家庭
              </div>
              <FamilyFileTreeSelector
                v-else
                v-model="newStorage.fileId"
                :cloud-token="newStorage.cloudToken"
                :family-id="newStorage.familyId"
                @select="handleFamilyFileSelect"
              />
            </div>
            <div v-if="selectedFamilyFileName" class="form-group">
              <label class="form-label">已选择文件夹</label>
              <div class="selected-file-info">
                <Icons name="folder" size="1rem" />
                <span>{{ selectedFamilyFileName }}</span>
                <button type="button" @click="clearFamilyFileSelection" class="clear-selection">
                  <Icons name="close" size="0.875rem" />
                </button>
              </div>
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

    <!-- 批量绑定令牌弹窗 -->
    <div v-if="showBatchBindModal" class="modal-overlay" @click="closeBatchBindModal">
      <div class="modal-content small" @click.stop>
        <div class="modal-header">
          <h3>批量绑定令牌</h3>
          <button @click="closeBatchBindModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <p class="bind-info">
            将为 <strong>{{ selectedCount }}</strong> 个存储批量绑定令牌
          </p>
          <div class="form-group">
            <label class="form-label">选择令牌</label>
            <Select v-model="batchSelectedTokenId" :options="tokenOptions" placeholder="请选择令牌" />
          </div>
          <div class="selected-storages">
            <h4>已选择的存储：</h4>
            <ul>
              <li v-for="storage in storages.filter(s => selectedStorageIds.has(s.id))" :key="storage.id">
                {{ storage.localPath }}
              </li>
            </ul>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeBatchBindModal" class="btn btn-secondary">取消</button>
          <button @click="confirmBatchBindToken" class="btn btn-primary" :disabled="batchBindLoading">
            {{ batchBindLoading ? '绑定中...' : '确认批量绑定' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 智能解析弹窗 -->
    <div v-if="showSmartParseModal" class="modal-overlay" @click="closeSmartParseModal">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>智能解析批量导入</h3>
          <button @click="closeSmartParseModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">
              <div>粘贴天翼云盘链接</div>
              <div style="font-size: 11px; color: #6b7280;">
                支持天翼云盘分享链接和21cn订阅链接，可以一次性粘贴多个链接
              </div>
            </label>
            <textarea 
              v-model="parseInput" 
              class="form-textarea" 
              placeholder="请粘贴天翼云盘链接，支持多个链接同时解析...&#10;例如：&#10;https://cloud.189.cn/t/xxxxxx（访问码：1234）&#10;https://content.21cn.com/h5/subscrip/index.html#/pages/own-home/index?uuid=xxxxxxxx"
              rows="6"
            ></textarea>
          </div>
          <div class="form-group">
            <label class="form-label">默认本地路径前缀</label>
            <input 
              v-model="defaultPathPrefix" 
              type="text" 
              class="form-input" 
              placeholder="例如：/电影/2024年" 
            />
          </div>
          <div class="form-group">
            <label class="form-label">默认云盘令牌（可选）</label>
            <Select v-model="defaultCloudToken" :options="tokenOptions" placeholder="请选择令牌" />
          </div>
          <div class="parse-actions">
            <button @click="parseLinks" class="btn btn-info" :disabled="parseLoading || !parseInput.trim()">
              <Icons name="search" size="1rem" class="btn-icon" />
              {{ parseLoading ? '解析中...' : '解析链接' }}
            </button>
          </div>
          
          <!-- 解析结果列表 -->
          <div v-if="parsedResults.length > 0" class="parsed-results">
            <h4>解析结果 ({{ parsedResults.length }}个)</h4>
            <div class="results-table-container">
              <table class="results-table">
                <thead>
                  <tr>
                    <th width="50">
                      <input 
                        type="checkbox" 
                        :checked="allResultsSelected" 
                        @change="selectAllResults"
                        class="checkbox"
                      />
                    </th>
                    <th width="80">类型</th>
                    <th width="180">建议名称</th>
                    <th width="250">本地路径</th>
                    <th width="150">云盘令牌</th>
                    <th width="150">订阅用户ID</th>
                    <th width="120">分享码</th>
                    <th width="120">访问码</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(result, index) in parsedResults" :key="index" class="result-row">
                    <td>
                      <input 
                        type="checkbox" 
                        v-model="result.selected" 
                        class="checkbox"
                      />
                    </td>
                    <td>
                      <span class="result-type" :class="`type-${result.type}`">
                        {{ result.type === 'subscribe' ? '订阅' : '分享' }}
                      </span>
                    </td>
                    <td>
                      <span class="suggested-name" :title="result.suggestedName">
                        {{ result.suggestedName }}
                      </span>
                    </td>
                    <td>
                      <input 
                        v-model="result.localPath" 
                        type="text" 
                        class="table-input" 
                        placeholder="请输入本地路径"
                      />
                    </td>
                    <td>
                      <Select 
                        v-model="result.cloudToken" 
                        :options="tokenOptions" 
                        placeholder="选择令牌" 
                        size="sm"
                        class="table-select"
                      />
                    </td>
                    <td>
                      <input 
                        v-if="result.type === 'subscribe'"
                        v-model="result.subscribeUser" 
                        type="text" 
                        class="table-input" 
                        placeholder="订阅用户ID"
                      />
                      <span v-else class="text-muted">-</span>
                    </td>
                    <td>
                      <input 
                        v-if="result.type === 'share'"
                        v-model="result.shareCode" 
                        type="text" 
                        class="table-input" 
                        readonly
                      />
                      <span v-else class="text-muted">-</span>
                    </td>
                    <td>
                      <input 
                        v-if="result.type === 'share'"
                        v-model="result.accessCode" 
                        type="text" 
                        class="table-input" 
                        placeholder="访问码（可选）"
                      />
                      <span v-else class="text-muted">-</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeSmartParseModal" class="btn btn-secondary">取消</button>
          <button 
            v-if="parsedResults.length > 0" 
            @click="selectAllResults" 
            class="btn btn-info"
          >
            {{ allResultsSelected ? '取消全选' : '全选' }}
          </button>
          <button 
            v-if="selectedResultsCount > 0" 
            @click="batchImportStorages" 
            class="btn btn-primary" 
            :disabled="batchImportLoading"
          >
            {{ batchImportLoading ? '导入中...' : `批量导入 (${selectedResultsCount})` }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icons from '@/components/Icons.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'
import { ref, onMounted, computed, reactive, onUnmounted, watch } from 'vue'
import { storageApi, type Storage, type AddStorageRequest } from '@/api/storage'
import { cloudTokenApi, type CloudToken } from '@/api/cloudtoken'
import { toast } from '@/utils/toast'
import { confirmDialog } from '@/utils/confirm'
import Select from '@/components/Select.vue'
import Pagination from '@/components/Pagination.vue'
import FileTreeSelector from '@/components/FileTreeSelector.vue'
import FamilyFileTreeSelector from '@/components/FamilyFileTreeSelector.vue'
import { storageBridgeApi, type FamilyInfo } from '@/api/storage_bridge'
import { useSettingStore } from '@/stores/setting'
import { parseMultipleCloudPan189Links, type CloudPan189ParseResult } from '@/utils/cloudpan189Parser'

// Store 实例
const settingStore = useSettingStore()

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
const refreshingStorageIds = ref<Set<number>>(new Set())
const toggleAutoScanLoading = ref<Set<number>>(new Set())
const scanTopLoading = ref(false)

// 批量选择相关
const selectedStorageIds = ref<Set<number>>(new Set())
const showBatchBindModal = ref(false)
const batchBindLoading = ref(false)
const batchSelectedTokenId = ref('')

// 智能解析相关
const showSmartParseModal = ref(false)
const parseInput = ref('')
const defaultPathPrefix = ref('/电影')
const defaultCloudToken = ref('')
const parseLoading = ref(false)
const batchImportLoading = ref(false)
const parsedResults = ref<ParsedResult[]>([])

// 个人文件选择相关
const selectedPersonFileName = ref('')

// 家庭文件选择相关
const availableFamilies = ref<FamilyInfo[]>([])
const selectedFamilyFileName = ref('')
const familyListLoading = ref(false)

// 解析结果接口
interface ParsedResult extends CloudPan189ParseResult {
  selected: boolean
  suggestedName: string
  localPath: string
  cloudToken: string
  subscribeUser?: string
  shareCode?: string
}

// 新存储表单数据
const newStorage = reactive({
  localPath: '',
  protocol: 'subscribe',
  cloudToken: '',
  subscribeUser: '',
  shareCode: '',
  shareAccessCode: '',
  fileId: '',
  familyId: ''
})

// 协议选项
const protocolOptions = [
  { label: '订阅类型', value: 'subscribe' },
  { label: '分享类型', value: 'share' },
  { label: '个人类型', value: 'person' },
  { label: '家庭类型', value: 'family' }
]

// 令牌选项
const tokenOptions = computed(() => {
  return availableTokens.value.map(token => ({
    label: token.name,
    value: token.id
  }))
})

// 家庭选项
const familyOptions = computed(() => {
  return availableFamilies.value.map(family => ({
    label: family.remarkName || family.familyId,
    value: family.familyId
  }))
})

// 批量选择相关计算属性
const isAllSelected = computed(() => {
  return storages.value.length > 0 && storages.value.every(storage => selectedStorageIds.value.has(storage.id))
})

const selectedCount = computed(() => {
  return selectedStorageIds.value.size
})

// 智能解析相关计算属性
const allResultsSelected = computed(() => {
  return parsedResults.value.length > 0 && parsedResults.value.every(result => result.selected)
})

const selectedResultsCount = computed(() => {
  return parsedResults.value.filter(result => result.selected).length
})

// 检查是否有弹窗打开
const hasModalOpen = computed(() => {
  return showAddModal.value || showBindModal.value || showBatchBindModal.value || showSmartParseModal.value
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

// 获取家庭列表
const fetchFamilyList = async (cloudToken: string | number) => {
  if (!cloudToken) {
    availableFamilies.value = []
    return
  }
  
  try {
    familyListLoading.value = true
    const response = await storageBridgeApi.getFamilyList({
      cloudToken: Number(cloudToken)
    })
    availableFamilies.value = response || []
  } catch (error) {
    console.error('获取家庭列表失败:', error)
    availableFamilies.value = []
    toast.error('获取家庭列表失败')
  } finally {
    familyListLoading.value = false
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
    shareAccessCode: '',
    fileId: '',
    familyId: ''
  })
  // 清空个人文件选择相关数据
  selectedPersonFileName.value = ''
  // 清空家庭文件选择相关数据
  selectedFamilyFileName.value = ''
  availableFamilies.value = []
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
  } else if (newStorage.protocol === 'person') {
    if (!newStorage.cloudToken) {
      toast.warning('个人类型需要选择云盘令牌')
      return
    }
    if (!newStorage.fileId.trim()) {
      toast.warning('请选择文件')
      return
    }
  } else if (newStorage.protocol === 'family') {
    if (!newStorage.cloudToken) {
      toast.warning('家庭类型需要选择云盘令牌')
      return
    }
    if (!newStorage.familyId.trim()) {
      toast.warning('请选择家庭')
      return
    }
    if (!newStorage.fileId.trim()) {
      toast.warning('请选择文件夹')
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
    } else if (newStorage.protocol === 'person') {
      requestData.fileId = newStorage.fileId.trim()
    } else if (newStorage.protocol === 'family') {
      requestData.familyId = newStorage.familyId.trim()
      requestData.fileId = newStorage.fileId.trim()
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

// 刷新存储索引
const refreshStorage = async (storage: Storage) => {
  try {
    refreshingStorageIds.value.add(storage.id)
    await storageApi.deepRefreshFile({ id: storage.id })
    toast.success('刷新指令已发送，请查看任务状态')
    fetchStorages()
  } catch (error: any) {
    if (error?.message) {
      toast.error(error.message)
    } else {
      toast.error('刷新索引失败')
    }
    console.error('刷新索引失败:', error)
  } finally {
    refreshingStorageIds.value.delete(storage.id)
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

// 切换自动扫描
const toggleAutoScan = async (storage: Storage) => {
  const currentStatus = storage.addition.disable_auto_scan
  const action = currentStatus ? '启用' : '禁用'
  
  const confirmed = await confirmDialog({
    title: `${action}自动扫描`,
    message: `确定要${action}存储 "${storage.localPath}" 的自动扫描吗？`,
    confirmText: action,
    cancelText: '取消',
    isDanger: !currentStatus
  })

  if (!confirmed) {
    return
  }

  try {
    toggleAutoScanLoading.value.add(storage.id)
    await storageApi.toggleAutoScan({
      id: storage.id,
      disableAutoScan: !currentStatus
    })
    toast.success(`${action}自动扫描成功`)
    fetchStorages()
  } catch (error: any) {
    if (error?.message) {
      toast.error(error.message)
    } else {
      toast.error(`${action}自动扫描失败`)
    }
    console.error(`${action}自动扫描失败:`, error)
  } finally {
    toggleAutoScanLoading.value.delete(storage.id)
  }
}

// 扫描所有文件
const scanTopFiles = async () => {
  // 显示提示弹窗
  const confirmed = await confirmDialog({
    title: '扫描所有文件',
    message: '此操作会扫描所有启用了自动扫描的存储。对于禁用自动扫描的存储，需要手动点击"扫描文件"按钮进行扫描。是否开始扫描？',
    confirmText: '开始扫描',
    cancelText: '取消',
    isDanger: false
  })

  if (!confirmed) {
    return
  }

  try {
    scanTopLoading.value = true
    const response = await storageApi.scanTop()
    toast.success(response.message + '，如果存储禁用了自动扫描，需要手动点击"扫描文件"按钮进行扫描')
    // 刷新存储列表以查看任务状态
    fetchStorages()
  } catch (error: any) {
    if (error?.msg) {
      toast.error(error.msg)
    } else {
      toast.error('扫描所有文件失败')
    }
    console.error('扫描所有文件失败:', error)
  } finally {
    scanTopLoading.value = false
  }
}

// 批量选择相关方法
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedStorageIds.value.clear()
  } else {
    storages.value.forEach(storage => {
      selectedStorageIds.value.add(storage.id)
    })
  }
}

const toggleStorageSelection = (storageId: number) => {
  if (selectedStorageIds.value.has(storageId)) {
    selectedStorageIds.value.delete(storageId)
  } else {
    selectedStorageIds.value.add(storageId)
  }
}

const openBatchBindModal = () => {
  if (selectedStorageIds.value.size === 0) {
    toast.warning('请先选择要批量绑定的存储')
    return
  }
  showBatchBindModal.value = true
}

const closeBatchBindModal = () => {
  showBatchBindModal.value = false
  batchSelectedTokenId.value = ''
}

const confirmBatchBindToken = async () => {
  if (!batchSelectedTokenId.value) {
    toast.warning('请选择令牌')
    return
  }

  if (selectedStorageIds.value.size === 0) {
    toast.warning('请先选择要批量绑定的存储')
    return
  }

  try {
    batchBindLoading.value = true
    const result = await storageApi.batchBindToken({
      ids: Array.from(selectedStorageIds.value),
      cloudToken: parseInt(batchSelectedTokenId.value)
    })
    
    if (result.successCount > 0) {
      toast.success(`成功绑定 ${result.successCount} 个存储`)
    }
    
    if (result.failedCount > 0) {
      toast.warning(`${result.failedCount} 个存储绑定失败`)
    }
    
    closeBatchBindModal()
    selectedStorageIds.value.clear()
    fetchStorages()
  } catch (error) {
    console.error('批量绑定令牌失败:', error)
    toast.error('批量绑定令牌失败')
  } finally {
    batchBindLoading.value = false
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

// 智能解析相关方法
const closeSmartParseModal = () => {
  showSmartParseModal.value = false
  parseInput.value = ''
  parsedResults.value = []
  defaultPathPrefix.value = '/电影'
  defaultCloudToken.value = ''
}

const parseLinks = async () => {
  if (!parseInput.value.trim()) {
    toast.warning('请输入要解析的链接')
    return
  }

  try {
    parseLoading.value = true
    const results = parseMultipleCloudPan189Links(parseInput.value)
    
    if (results.length === 0) {
      toast.warning('未找到有效的天翼云盘链接')
      return
    }

    // 调用 preAdd 接口验证链接有效性并获取建议名称
    const validResults: ParsedResult[] = []
    const failedUrls: string[] = []
    
    for (const result of results) {
      try {
        const preAddData: any = {
          protocol: result.type === 'subscribe' ? 'subscribe' : 'share'
        }
        
        if (result.type === 'subscribe') {
          preAddData.subscribeUser = result.shareId
        } else if (result.type === 'share') {
          preAddData.shareCode = result.shareId
          if (result.accessCode) {
            preAddData.shareAccessCode = result.accessCode
          }
        }
        
        if (defaultCloudToken.value) {
          preAddData.cloudToken = Number(defaultCloudToken.value)
        }
        
        const preAddResponse = await storageApi.preAdd(preAddData)
        const suggestedName = preAddResponse.name || result.shareId
        
        validResults.push({
          ...result,
          selected: true,
          suggestedName,
          localPath: defaultPathPrefix.value ? `${defaultPathPrefix.value}/${suggestedName}` : `/${suggestedName}`,
          cloudToken: defaultCloudToken.value,
          subscribeUser: result.type === 'subscribe' ? result.shareId : '',
          shareCode: result.type === 'share' ? result.shareId : ''
        })
      } catch (error) {
        console.warn(`链接验证失败 (${result.originalUrl}):`, error)
        failedUrls.push(result.originalUrl)
      }
    }

    parsedResults.value = validResults
    
    if (validResults.length === 0) {
      toast.error('所有链接都无效或解析失败')
    } else if (failedUrls.length > 0) {
      toast.warning(`成功解析 ${validResults.length} 个链接，${failedUrls.length} 个链接无效`)
    } else {
      toast.success(`成功解析 ${validResults.length} 个链接`)
    }
  } catch (error: any) {
    console.error('解析链接失败:', error)
    toast.error(error.message || '解析链接失败')
  } finally {
    parseLoading.value = false
  }
}

const selectAllResults = () => {
  const newSelectedState = !allResultsSelected.value
  parsedResults.value.forEach(result => {
    result.selected = newSelectedState
  })
}

const batchImportStorages = async () => {
  const selectedResults = parsedResults.value.filter(result => result.selected)
  
  if (selectedResults.length === 0) {
    toast.warning('请选择要导入的存储')
    return
  }

  // 验证必填字段
  const invalidResults = selectedResults.filter(result => {
    if (!result.localPath.trim()) return true
    if (result.type === 'subscribe' && !result.subscribeUser?.trim()) return true
    if (result.type === 'share' && !result.shareCode?.trim()) return true
    return false
  })

  if (invalidResults.length > 0) {
    toast.warning('请完善所有选中项的必填信息')
    return
  }

  try {
    batchImportLoading.value = true
    let successCount = 0
    let failedCount = 0
    
    // 逐个添加存储
    for (const result of selectedResults) {
      try {
        const requestData: AddStorageRequest = {
          localPath: result.localPath.trim(),
          protocol: result.type === 'subscribe' ? 'subscribe' : 'share'
        }

        if (result.cloudToken) {
          requestData.cloudToken = Number(result.cloudToken)
        }

        if (result.type === 'subscribe') {
          requestData.subscribeUser = result.subscribeUser!.trim()
        } else if (result.type === 'share') {
          requestData.shareCode = result.shareCode!.trim()
          if (result.accessCode?.trim()) {
            requestData.shareAccessCode = result.accessCode.trim()
          }
        }

        await storageApi.add(requestData)
        successCount++
      } catch (error) {
        console.error(`添加存储失败 (${result.originalUrl}):`, error)
        failedCount++
      }
    }

    if (successCount > 0) {
      toast.success(`成功导入 ${successCount} 个存储，后台异步执行，如未显示，稍后刷新重新查看`)
    }
    
    if (failedCount > 0) {
      toast.warning(`${failedCount} 个存储导入失败`)
    }

    closeSmartParseModal()
    fetchStorages()
  } catch (error) {
    console.error('批量导入存储失败:', error)
    toast.error('批量导入存储失败')
  } finally {
    batchImportLoading.value = false
  }
}

// 监听云盘令牌变化
const watchCloudToken = () => {
  // 清空个人文件选择
  selectedPersonFileName.value = ''
  newStorage.fileId = ''
  
  // 清空家庭相关数据
  availableFamilies.value = []
  selectedFamilyFileName.value = ''
  newStorage.familyId = ''
  
  // 如果选择了令牌且协议是家庭类型，则获取家庭列表
  if (newStorage.cloudToken && newStorage.protocol === 'family') {
    fetchFamilyList(newStorage.cloudToken)
  }
}

// 处理个人文件选择
const handlePersonFileSelect = (file: any) => {
  newStorage.fileId = file.id
  selectedPersonFileName.value = file.name
}

// 清空个人文件选择
const clearPersonFileSelection = () => {
  newStorage.fileId = ''
  selectedPersonFileName.value = ''
}

// 处理家庭变化
const handleFamilyChange = () => {
  // 清空文件选择
  selectedFamilyFileName.value = ''
  newStorage.fileId = ''
}

// 处理家庭文件选择
const handleFamilyFileSelect = (file: any) => {
  newStorage.fileId = file.id
  selectedFamilyFileName.value = file.name
}

// 清空家庭文件选择
const clearFamilyFileSelection = () => {
  newStorage.fileId = ''
  selectedFamilyFileName.value = ''
}


// 获取协议标签
const getProtocolLabel = (protocol: string): string => {
  switch (protocol) {
    case 'subscribe':
      return '订阅'
    case 'share':
      return '分享'
    case 'cloud_folder':
      return '个人'
    case 'cloud_family_folder':
      return '家庭'
    default:
      return protocol
  }
}

// 监听弹窗状态，控制背景滚动
watch(hasModalOpen, (isOpen) => {
  if (isOpen) {
    // 禁用背景滚动
    document.body.style.overflow = 'hidden'
  } else {
    // 恢复背景滚动
    document.body.style.overflow = ''
  }
})

// 组件挂载时获取数据
onMounted(() => {
  // 从localStorage恢复页面大小设置
  const savedPageSize = localStorage.getItem('storageListPageSize')
  if (savedPageSize) {
    pageSize.value = parseInt(savedPageSize)
  }

  fetchStorages()
  fetchAvailableTokens()
  // 获取设置信息以显示扫描配置
  settingStore.fetchSetting()
  
  // 启动定时刷新
  refreshTimer = setInterval(() => {
    fetchStorages()
  }, 30 * 1000)
})

// 组件卸载时恢复滚动
onUnmounted(() => {
  // 确保在组件卸载时恢复滚动
  document.body.style.overflow = ''
  clearInterval(refreshTimer)
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

.btn-success {
  background: #10b981;
  color: white;
}

.btn-success:hover:not(:disabled) {
  background: #059669;
}

.btn-info {
  background: #0ea5e9;
  color: white;
}

.btn-info:hover:not(:disabled) {
  background: #0284c7;
}

.btn-warning {
  background: #f59e0b;
  color: white;
}

.btn-warning:hover:not(:disabled) {
  background: #d97706;
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
  text-align: center;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
}

.storage-table th:nth-child(2) {
  text-align: left;
}

.storage-table td {
  padding: 1rem;
  border-bottom: 1px solid #f3f4f6;
  font-size: 0.875rem;
  text-align: center;
}

.storage-table td:nth-child(2) {
  text-align: left;
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

.protocol-cloud_folder {
  background: #f0f9ff;
  color: #0369a1;
}

.protocol-cloud_family_folder {
  background: #fdf4ff;
  color: #a21caf;
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
  gap: 0.25rem;
  flex-wrap: nowrap;
  justify-content: center;
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

/* 扫描信息面板样式 */
.scan-info-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.75rem;
  min-width: 200px;
}

.scan-info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
}

.info-label {
  color: #64748b;
  font-weight: 500;
}

.info-value {
  font-weight: 600;
  color: #1e293b;
}

.status-enabled {
  color: #059669;
}

.status-disabled {
  color: #dc2626;
}

/* 扫描状态样式 */
.scan-status {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 160px;
}

.scan-status-inline {
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

.modal-content.large {
  max-width: 800px;
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

/* 批量绑定弹窗样式 */
.selected-storages {
  margin-top: 1rem;
  padding: 1rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.selected-storages h4 {
  margin: 0 0 0.75rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.selected-storages ul {
  margin: 0;
  padding: 0;
  list-style: none;
  max-height: 150px;
  overflow-y: auto;
}

.selected-storages li {
  padding: 0.5rem 0;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
  color: #6b7280;
}

.selected-storages li:last-child {
  border-bottom: none;
}

/* 复选框样式 */
.checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
}

/* 智能解析弹窗样式 */
.form-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  font-family: inherit;
  resize: vertical;
  min-height: 120px;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}

.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.parse-actions {
  margin: 1rem 0;
  text-align: center;
}

.parsed-results {
  margin-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
  padding-top: 1.5rem;
}

.parsed-results h4 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
}

/* 解析结果表格样式 */
.results-table-container {
  max-height: 400px;
  overflow-x: auto;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: white;
}

.results-table {
  width: 100%;
  min-width: 1100px;
  border-collapse: collapse;
  font-size: 0.875rem;
  table-layout: fixed;
}

.results-table th {
  background: #f9fafb;
  padding: 0.75rem 0.5rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 1;
}

.results-table td {
  padding: 0.75rem 0.5rem;
  border-bottom: 1px solid #f3f4f6;
  vertical-align: middle;
}

.result-row:hover {
  background: #f9fafb;
}

.result-type {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.type-subscribe {
  background: #eff6ff;
  color: #1d4ed8;
}

.type-share {
  background: #fef3c7;
  color: #d97706;
}



.result-url {
  font-size: 0.75rem;
  color: #6b7280;
  display: block;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.table-input {
  width: 100%;
  padding: 0.375rem 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.75rem;
  transition: border-color 0.2s;
  box-sizing: border-box;
  min-width: 0;
}

.table-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.table-input[readonly] {
  background: #f3f4f6;
  color: #6b7280;
}

.table-select {
  width: 100%;
  min-width: 0;
}

.suggested-name {
  font-weight: 500;
  color: #374151;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 0.25rem 0;
}

.text-muted {
  color: #9ca3af;
  font-size: 0.75rem;
  text-align: center;
}

.form-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.form-row:last-child {
  margin-bottom: 0;
}

.form-col {
  flex: 1;
}

.form-label-sm {
  display: block;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.25rem;
  font-size: 0.75rem;
}

.form-input-sm {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.75rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}
/* 文件选择器样式 */
.file-selector-placeholder {
  padding: 2rem;
  text-align: center;
  color: #9ca3af;
  background: #f9fafb;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
}

.file-selector-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 2rem;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  color: #6b7280;
  font-size: 0.875rem;
}

.loading-spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid #e5e7eb;
  border-top: 2px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.file-selector-empty {
  padding: 2rem;
  text-align: center;
  color: #9ca3af;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 0.875rem;
}


.selected-file-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #1d4ed8;
}

.clear-selection {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  margin-left: auto;
}

.clear-selection:hover {
  background: #f3f4f6;
  color: #374151;
}


.form-input-sm:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.form-input-sm[readonly] {
  background: #f3f4f6;
  color: #6b7280;
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
