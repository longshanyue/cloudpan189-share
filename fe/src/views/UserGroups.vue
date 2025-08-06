<template>
  <Layout>
    <!-- 用户组管理主卡片 -->
    <PageCard title="用户组管理" subtitle="用户组用于控制用户可以看到哪些文件，默认用户组默认存在且无法查看删除，默认用户组可看到所有文件。">
      <SectionDivider />
      
      <SubsectionTitle title="用户组列表" />
        <!-- 操作栏 -->
        <div class="action-bar">
          <div class="search-section">
            <input 
              v-model="searchName" 
              type="text" 
              placeholder="搜索用户组名称..." 
              class="search-input"
              @input="handleSearch"
            >
          </div>
          <button @click="showAddModal = true" class="btn btn-primary">
            <Icons name="add" size="1rem" class="btn-icon" />
            添加用户组
          </button>
        </div>

        <!-- 用户组列表 -->
        <div class="usergroups-table-container">
          <div v-if="loading" class="loading-state">
            <div class="loading-spinner"></div>
            <p>加载中...</p>
          </div>
          
          <div v-else-if="userGroups.length === 0" class="empty-state">
            <Icons name="users" size="3rem" class="empty-icon" />
            <h3>暂无用户组</h3>
            <p>{{ searchName ? '没有找到匹配的用户组' : '还没有用户组，点击上方按钮添加第一个用户组' }}</p>
          </div>
          
          <table v-else class="usergroups-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>用户组名称</th>
                <th>创建时间</th>
                <th>更新时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="group in userGroups" :key="group.id" class="group-row">
                <td>{{ group.id }}</td>
                <td>
                  <span class="group-name">{{ group.name }}</span>
                </td>
                <td>{{ formatDate(group.createdAt) }}</td>
                <td>{{ formatDate(group.updatedAt) }}</td>
                 <td>
                   <div class="action-buttons">
                     <button @click="editGroup(group)" class="btn btn-sm btn-secondary">
                       修改名称
                     </button>
                     <button @click="manageStoragePermissions(group)" class="btn btn-sm btn-info">
                       绑定存储权限
                     </button>
                     <button @click="deleteGroup(group)" class="btn btn-sm btn-danger">
                       删除
                     </button>
                   </div>
                 </td>
               </tr>
             </tbody>
           </table>
            
            <!-- 分页组件 -->
            <Pagination 
              v-if="total > 0"
              :current-page="currentPage"
              :page-size="pageSize"
              :total="total"
              @page-change="handlePageChange"
              @page-size-change="handlePageSizeChange"
            />
        </div>
    </PageCard>

    <!-- 添加用户组弹窗 -->
    <div v-if="showAddModal" class="modal-overlay" @click="closeAddModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>添加用户组</h3>
          <button @click="closeAddModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">用户组名称</label>
            <input 
              v-model="newGroup.name" 
              type="text" 
              class="form-input"
              placeholder="请输入用户组名称（1-255字符）"
              maxlength="255"
            >
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeAddModal" class="btn btn-secondary">取消</button>
          <button @click="confirmAddGroup" class="btn btn-primary" :disabled="addLoading">
            {{ addLoading ? '添加中...' : '确认添加' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 修改用户组名称弹窗 -->
    <div v-if="showEditModal" class="modal-overlay" @click="closeEditModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>修改用户组名称</h3>
          <button @click="closeEditModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
           <div class="form-group">
             <label class="form-label">用户组名称</label>
             <input 
               v-model="editingGroup.name" 
               type="text" 
               class="form-input"
               placeholder="请输入用户组名称（1-255字符）"
               maxlength="255"
             >
           </div>
         </div>
        <div class="modal-footer">
          <button @click="closeEditModal" class="btn btn-secondary">取消</button>
          <button @click="confirmEditGroup" class="btn btn-primary" :disabled="editLoading">
            {{ editLoading ? '保存中...' : '保存更改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 绑定存储权限弹窗 -->
    <div v-if="showStorageModal" class="modal-overlay" @click="closeStorageModal">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>绑定存储权限 - {{ currentGroup?.name }}</h3>
          <button @click="closeStorageModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <p class="storage-info">为用户组 <strong>{{ currentGroup?.name }}</strong> 配置可访问的存储</p>
          
          <!-- 穿梭框 -->
          <TransferBox
             :all-items="allStorages"
             :bound-item-ids="boundStorageIds"
             item-key="id"
             name-field="name"
             desc-field="localPath"
             left-title="可选存储"
             right-title="已绑定存储"
             left-empty-text="暂无可选存储"
             right-empty-text="暂无已绑定存储"
             :loading="storageLoading"
             @change="handleStorageSelectionChange"
           />
        </div>
        <div class="modal-footer">
          <button @click="closeStorageModal" class="btn btn-secondary">取消</button>
          <button @click="confirmManageStoragePermissions" class="btn btn-primary" :disabled="storagePermissionLoading">
            {{ storagePermissionLoading ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import Layout from '@/components/Layout.vue'
import Icons from '@/components/Icons.vue'
import Pagination from '@/components/Pagination.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'
import TransferBox from '@/components/TransferBox.vue'
import { userGroupApi, type UserGroup, type AddUserGroupRequest, type ModifyUserGroupNameRequest, type BatchBindFilesRequest } from '@/api/usergroup'
import { storageApi, type Storage } from '@/api/storage'
import { toast } from '@/utils/toast'
import { confirmDialog } from '@/utils/confirm'

// 响应式数据
const userGroups = ref<UserGroup[]>([])
const loading = ref(false)
const searchName = ref('')

// 分页相关
const currentPage = ref(1)
const pageSize = ref(parseInt(localStorage.getItem('userGroupListPageSize') || '10'))
const total = ref(0)

// 添加用户组相关
const showAddModal = ref(false)
const addLoading = ref(false)
const newGroup = reactive({
  name: ''
})

// 编辑用户组相关
const showEditModal = ref(false)
const editLoading = ref(false)
const editingGroup = ref<any>({})

// 绑定存储权限相关
const showStorageModal = ref(false)
const storagePermissionLoading = ref(false)
const storageLoading = ref(false)
const currentGroup = ref<UserGroup | null>(null)
const allStorages = ref<Storage[]>([])
const boundStorageIds = ref<number[]>([])
const availableStorages = ref<Storage[]>([])
const boundStorages = ref<Storage[]>([])

// 获取用户组列表
const fetchUserGroups = async () => {
  try {
    loading.value = true
    const params: any = {
      currentPage: currentPage.value,
      pageSize: pageSize.value
    }
    if (searchName.value.trim()) {
      params.name = searchName.value.trim()
    }
    const response = await userGroupApi.getUserGroupList(params)
    // 处理分页响应格式
    userGroups.value = response.data || []
    total.value = response.total || 0
  } catch (error) {
    console.error('获取用户组列表失败:', error)
    toast.error('获取用户组列表失败')
    userGroups.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 搜索处理
let searchTimer: NodeJS.Timeout
const handleSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1 // 搜索时重置到第一页
    fetchUserGroups()
  }, 500)
}

// 分页处理
const handlePageChange = (page: number) => {
  if (page !== currentPage.value) {
    currentPage.value = page
    fetchUserGroups()
  }
}

const handlePageSizeChange = (size: number) => {
  if (size !== pageSize.value) {
    pageSize.value = size
    currentPage.value = 1
    localStorage.setItem('userGroupListPageSize', size.toString())
    fetchUserGroups()
  }
}

// 格式化日期
const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 添加用户组相关函数
const closeAddModal = () => {
  showAddModal.value = false
  newGroup.name = ''
}

const confirmAddGroup = async () => {
  if (!newGroup.name.trim() || newGroup.name.length < 1) {
    toast.error('用户组名称长度必须在1-255字符之间')
    return
  }

  try {
    addLoading.value = true
    const addData: AddUserGroupRequest = {
      name: newGroup.name.trim()
    }
    
    await userGroupApi.addUserGroup(addData)
    toast.success('用户组添加成功')
    closeAddModal()
    fetchUserGroups()
  } catch (error: any) {
    console.error('添加用户组失败:', error)
    toast.error(error.msg || '添加用户组失败')
  } finally {
    addLoading.value = false
  }
}

// 编辑用户组相关函数
const editGroup = (group: UserGroup) => {
  editingGroup.value = {
    id: group.id,
    name: group.name
  }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  editingGroup.value = {}
}

const confirmEditGroup = async () => {
  if (!editingGroup.value.name.trim() || editingGroup.value.name.length < 1) {
    toast.error('用户组名称长度必须在1-255字符之间')
    return
  }

  try {
    editLoading.value = true
    const updateData: ModifyUserGroupNameRequest = {
      id: editingGroup.value.id,
      name: editingGroup.value.name.trim()
    }
    
    await userGroupApi.modifyUserGroupName(updateData)
    toast.success('用户组名称更新成功')
    closeEditModal()
    fetchUserGroups()
  } catch (error: any) {
    console.error('更新用户组失败:', error)
    toast.error(error.msg || '更新用户组失败')
  } finally {
    editLoading.value = false
  }
}

// 获取所有存储列表
const fetchAllStorages = async () => {
  try {
    storageLoading.value = true
    const response = await storageApi.list({ noPaginate: true })
    allStorages.value = response.data || []
  } catch (error) {
    console.error('获取存储列表失败:', error)
    toast.error('获取存储列表失败')
    allStorages.value = []
  } finally {
    storageLoading.value = false
  }
}

// 获取用户组已绑定的文件
const fetchBoundFiles = async (groupId: number) => {
  try {
    const response = await userGroupApi.getBindFiles({ groupId })
    boundStorageIds.value = response.files.map(file => file.fileId)
  } catch (error) {
    console.error('获取绑定文件失败:', error)
    boundStorageIds.value = []
  }
}

// 更新可选和已绑定的存储列表
const updateStorageLists = () => {
  availableStorages.value = allStorages.value.filter(storage => !boundStorageIds.value.includes(storage.id))
  boundStorages.value = allStorages.value.filter(storage => boundStorageIds.value.includes(storage.id))
}

// 绑定存储权限相关函数
const manageStoragePermissions = async (group: UserGroup) => {
  currentGroup.value = group
  
  // 获取所有存储和已绑定的文件
  await Promise.all([
    fetchAllStorages(),
    fetchBoundFiles(group.id)
  ])
  
  updateStorageLists()
  showStorageModal.value = true
}

const closeStorageModal = () => {
  showStorageModal.value = false
  currentGroup.value = null
  allStorages.value = []
  boundStorageIds.value = []
  availableStorages.value = []
  boundStorages.value = []
}

// 处理存储选择变化
 const handleStorageSelectionChange = (boundIds: number[]) => {
   boundStorageIds.value = boundIds
   updateStorageLists()
 }

const confirmManageStoragePermissions = async () => {
  try {
    storagePermissionLoading.value = true
    
    const bindData: BatchBindFilesRequest = {
      groupId: currentGroup.value!.id,
      fileIds: boundStorageIds.value
    }
    
    const result = await userGroupApi.batchBindFiles(bindData)
    toast.success(`存储权限配置成功：绑定 ${result.bindCount} 个存储`)
    closeStorageModal()
  } catch (error: any) {
    console.error('配置存储权限失败:', error)
    toast.error(error.msg || '配置存储权限失败')
  } finally {
    storagePermissionLoading.value = false
  }
}

// 删除用户组
const deleteGroup = async (group: UserGroup) => {
  const confirmed = await confirmDialog({
    title: '删除用户组',
    message: `确定要删除用户组 "${group.name}" 吗？此操作不可恢复。`,
    confirmText: '删除',
    cancelText: '取消',
    isDanger: true
  })
  
  if (!confirmed) {
    return
  }

  try {
    await userGroupApi.deleteUserGroup({ id: group.id })
    toast.success('用户组删除成功')
    fetchUserGroups()
  } catch (error: any) {
    console.error('删除用户组失败:', error)
    toast.error(error.msg || '删除用户组失败')
  }
}

// 页面初始化
onMounted(() => {
  fetchUserGroups()
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

.btn-info {
  background: #06b6d4;
  color: white;
}

.btn-info:hover:not(:disabled) {
  background: #0891b2;
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
.usergroups-table-container {
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
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
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
.usergroups-table {
  width: 100%;
  border-collapse: collapse;
}

.usergroups-table th {
  background: #f9fafb;
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
}

.usergroups-table td {
  padding: 1rem;
  border-bottom: 1px solid #f3f4f6;
  font-size: 0.875rem;
}

.group-row:hover {
  background: #f9fafb;
}

.group-name {
  font-weight: 500;
  color: #1f2937;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
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

.modal-content.large {
  max-width: 600px;
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
  width: 2rem;
  height: 2rem;
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
  gap: 0.75rem;
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
  border-radius: 6px;
  font-size: 0.875rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
  resize: vertical;
  min-height: 100px;
}

.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-hint {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: #6b7280;
}

.storage-info {
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: #f0f9ff;
  border: 1px solid #e0f2fe;
  border-radius: 6px;
  font-size: 0.875rem;
  color: #0369a1;
}

/* 存储项内容样式 */
.storage-item-content {
  display: flex;
  flex-direction: column;
}

.storage-name {
  font-weight: 500;
  color: #1f2937;
  font-size: 0.875rem;
  margin-bottom: 0.25rem;
}

.storage-path {
  font-size: 0.75rem;
  color: #6b7280;
  word-break: break-all;
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
  
  .usergroups-table th,
  .usergroups-table td {
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
  .usergroups-table {
    display: block;
    overflow-x: auto;
    white-space: nowrap;
  }
}
</style>