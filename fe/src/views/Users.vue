<template>
  <div>
    <!-- 用户管理主卡片 -->
    <PageCard title="用户管理" subtitle="管理系统用户和权限设置">
      <SectionDivider />

      <SubsectionTitle title="用户列表" />
      <!-- 操作栏 -->
      <div class="action-bar">
        <div class="search-section">
          <label for="searchUsername" class="sr-only">搜索用户名</label>
          <input
              v-model="searchUsername"
              type="text"
              placeholder="搜索用户名..."
              class="search-input"
              name="searchUsername"
              id="searchUsername"
              @input="handleSearch"
          >
        </div>
        <button @click="showAddModal = true" class="btn btn-primary">
          <Icons name="add" size="1rem" class="btn-icon" />
          添加用户
        </button>
      </div>

      <!-- 用户列表 -->
      <div class="users-table-container">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <p>加载中...</p>
        </div>

        <div v-else-if="users.length === 0" class="empty-state">
          <Icons name="users" size="3rem" class="empty-icon" />
          <h3>暂无用户</h3>
          <p>{{ searchUsername ? '没有找到匹配的用户' : '还没有用户，点击上方按钮添加第一个用户' }}</p>
        </div>

        <table v-else class="users-table">
          <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>状态</th>
            <th>用户组</th>
            <th>权限</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="user in users" :key="user.id" class="user-row">
            <td>{{ user.id }}</td>
            <td>
              <div class="user-info">
                <div class="user-avatar">
                  <span>{{ user.username.charAt(0).toUpperCase() }}</span>
                </div>
                <span class="username">{{ user.username }}</span>
              </div>
            </td>
            <td>
                  <span :class="['status-badge', user.status === 1 ? 'status-active' : 'status-inactive']">
                    {{ user.status === 1 ? '正常' : '禁用' }}
                  </span>
            </td>
            <td>
              <span class="group-name">{{ user.groupName || '默认用户组' }}</span>
            </td>
            <td>
              <div class="permissions">
                    <span v-for="permission in getPermissionLabels(user.permissions)"
                          :key="permission"
                          class="permission-tag">
                      {{ permission }}
                    </span>
              </div>
            </td>
            <td>{{ formatDate(user.createdAt) }}</td>
            <td>
              <div class="action-buttons">
                <button @click="editUser(user)" class="btn btn-sm btn-secondary">
                  编辑
                </button>
                <button @click="bindUserGroup(user)" class="btn btn-sm btn-primary">
                  绑定用户组
                </button>
                <button @click="resetPassword(user)" class="btn btn-sm btn-warning">
                  重置密码
                </button>
                <button @click="deleteUser(user)" class="btn btn-sm btn-danger">
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

    <!-- 添加用户弹窗 -->
    <div v-if="showAddModal" class="modal-overlay" @click="closeAddModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>添加用户</h3>
          <button @click="closeAddModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label" for="newUsername">用户名</label>
            <input
                v-model="newUser.username"
                type="text"
                class="form-input"
                placeholder="请输入用户名（3-20字符）"
                maxlength="20"
                name="newUsername"
                id="newUsername"
            >
          </div>
          <div class="form-group">
            <label class="form-label" for="newPassword">密码</label>
            <input
                v-model="newUser.password"
                type="password"
                class="form-input"
                placeholder="请输入密码（6-20字符）"
                maxlength="20"
                name="newPassword"
                id="newPassword"
            >
          </div>
          <div class="form-group">
            <fieldset>
              <legend class="form-label">权限设置</legend>
              <div class="permission-checkboxes">
                <label class="checkbox-item" for="newBasePermission">
                  <input
                      type="checkbox"
                      v-model="newUser.permissions.base"
                      name="newBasePermission"
                      id="newBasePermission"
                  >
                  <span>基础权限（登录和查看信息）</span>
                </label>
                <label class="checkbox-item" for="newDavReadPermission">
                  <input
                      type="checkbox"
                      v-model="newUser.permissions.davRead"
                      name="newDavReadPermission"
                      id="newDavReadPermission"
                  >
                  <span>WebDAV访问权限</span>
                </label>
                <label class="checkbox-item" for="newAdminPermission">
                  <input
                      type="checkbox"
                      v-model="newUser.permissions.admin"
                      name="newAdminPermission"
                      id="newAdminPermission"
                  >
                  <span>管理员权限</span>
                </label>
              </div>
            </fieldset>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeAddModal" class="btn btn-secondary">取消</button>
          <button @click="confirmAddUser" class="btn btn-primary" :disabled="addLoading">
            {{ addLoading ? '添加中...' : '确认添加' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑用户弹窗 -->
    <div v-if="showEditModal" class="modal-overlay" @click="closeEditModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>编辑用户</h3>
          <button @click="closeEditModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label" for="editUsername">用户名</label>
            <input
                v-model="editingUser.username"
                type="text"
                class="form-input"
                disabled
                name="editUsername"
                id="editUsername"
            >
          </div>
          <div class="form-group">
            <fieldset>
              <legend class="form-label">权限设置</legend>
              <div class="permission-checkboxes">
                <label class="checkbox-item" for="editBasePermission">
                  <input
                      type="checkbox"
                      v-model="editingUser.permissions.base"
                      name="editBasePermission"
                      id="editBasePermission"
                  >
                  <span>基础权限（登录和查看信息）</span>
                </label>
                <label class="checkbox-item" for="editDavReadPermission">
                  <input
                      type="checkbox"
                      v-model="editingUser.permissions.davRead"
                      name="editDavReadPermission"
                      id="editDavReadPermission"
                  >
                  <span>WebDAV访问权限</span>
                </label>
                <label class="checkbox-item" for="editAdminPermission">
                  <input
                      type="checkbox"
                      v-model="editingUser.permissions.admin"
                      name="editAdminPermission"
                      id="editAdminPermission"
                  >
                  <span>管理员权限（可以管理用户和系统设置）</span>
                </label>
              </div>
            </fieldset>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeEditModal" class="btn btn-secondary">取消</button>
          <button @click="confirmEditUser" class="btn btn-primary" :disabled="editLoading">
            {{ editLoading ? '保存中...' : '保存更改' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 重置密码弹窗 -->
    <div v-if="showPasswordModal" class="modal-overlay" @click="closePasswordModal">
      <div class="modal-content small" @click.stop>
        <div class="modal-header">
          <h3>重置密码</h3>
          <button @click="closePasswordModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <p class="reset-info">为用户 <strong>{{ resetPasswordUser?.username }}</strong> 重置密码</p>
          <div class="form-group">
            <label class="form-label" for="resetPasswordInput">新密码</label>
            <input
                v-model="newPassword"
                type="password"
                class="form-input"
                placeholder="请输入新密码（6-20字符）"
                maxlength="20"
                name="resetPasswordInput"
                id="resetPasswordInput"
            >
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closePasswordModal" class="btn btn-secondary">取消</button>
          <button @click="confirmResetPassword" class="btn btn-warning" :disabled="passwordLoading">
            {{ passwordLoading ? '重置中...' : '确认重置' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 绑定用户组弹窗 -->
    <div v-if="showBindGroupModal" class="modal-overlay" @click="closeBindGroupModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>绑定用户组</h3>
          <button @click="closeBindGroupModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <p class="bind-info">为用户 <strong>{{ bindGroupUser?.username }}</strong> 绑定用户组</p>

          <!-- 用户组选择 -->
          <div class="form-group">
            <label for="userGroupSelect">
              <div class="flex" style="justify-content: space-between">
                <span>选择用户组：</span>
                <span style="font-size: 11px; color: #666;">用户组用于控制用户可以看到哪些文件，默认用户组可看到所有文件。</span>
              </div>
            </label>
            <Select
                id="userGroupSelect"
                v-model="selectedGroupId"
                :options="userGroupOptions"
                :disabled="groupLoading"
                placeholder="请选择用户组"
                empty-text="暂无用户组"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button @click="closeBindGroupModal" class="btn btn-secondary">取消</button>
          <button @click="confirmBindUserGroup" class="btn btn-primary" :disabled="bindGroupLoading">
            {{ bindGroupLoading ? '绑定中...' : '确认绑定' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import Layout from '@/components/Layout.vue'
import Icons from '@/components/Icons.vue'
import Pagination from '@/components/Pagination.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'

import Select from '@/components/Select.vue'
import { userApi, type User, type AddUserRequest, type UpdateUserRequest, type ModifyPasswordRequest } from '@/api/user'
import { userGroupApi, type UserGroup } from '@/api/usergroup'
import { toast } from '@/utils/toast'
import { confirmDialog } from '@/utils/confirm'
import { getPermissionLabels, calculatePermissions, parsePermissions, PERMISSIONS } from '@/utils/permissions'

// 响应式数据
const users = ref<User[]>([])
const loading = ref(false)
const searchUsername = ref('')

// 分页相关
const currentPage = ref(1)
const pageSize = ref(parseInt(localStorage.getItem('userListPageSize') || '10'))
const total = ref(0)

// 添加用户相关
const showAddModal = ref(false)
const addLoading = ref(false)
const newUser = reactive({
  username: '',
  password: '',
  permissions: {
    base: true,
    davRead: false,
    admin: false
  }
})

// 编辑用户相关
const showEditModal = ref(false)
const editLoading = ref(false)
const editingUser = ref<any>({})

// 重置密码相关
const showPasswordModal = ref(false)
const passwordLoading = ref(false)
const resetPasswordUser = ref<User | null>(null)
const newPassword = ref('')

// 绑定用户组相关
const showBindGroupModal = ref(false)
const bindGroupLoading = ref(false)
const groupLoading = ref(false)
const bindGroupUser = ref<User | null>(null)
const allUserGroups = ref<UserGroup[]>([])
const selectedGroupId = ref<number>(0)

const userGroupOptions = computed(() => {
  const options = allUserGroups.value.map(group => ({
    label: group.name,
    value: group.id
  }))
  // 添加默认用户组选项
  options.unshift({ label: '默认用户组', value: 0 })
  return options
})

// 获取用户列表
const fetchUsers = async () => {
  try {
    loading.value = true
    const params: any = {
      currentPage: currentPage.value,
      pageSize: pageSize.value
    }
    if (searchUsername.value.trim()) {
      params.username = searchUsername.value.trim()
    }
    const response = await userApi.getUserList(params)
    // 处理分页响应格式
    users.value = response.data || []
    total.value = response.total || 0
  } catch (error) {
    console.error('获取用户列表失败:', error)
    toast.error('获取用户列表失败')
    users.value = []
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
    fetchUsers()
  }, 500)
}

// 分页处理
const handlePageChange = (page: number) => {
  if (page !== currentPage.value) {
    currentPage.value = page
    fetchUsers()
  }
}

const handlePageSizeChange = (size: number) => {
  if (size !== pageSize.value) {
    pageSize.value = size
    currentPage.value = 1
    localStorage.setItem('userListPageSize', size.toString())
    fetchUsers()
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

// 添加用户相关函数
const closeAddModal = () => {
  showAddModal.value = false
  newUser.username = ''
  newUser.password = ''
  newUser.permissions = {
    base: true,
    davRead: false,
    admin: false
  }
}

const confirmAddUser = async () => {
  if (!newUser.username.trim() || newUser.username.length < 3) {
    toast.error('用户名长度必须在3-20字符之间')
    return
  }
  if (!newUser.password.trim() || newUser.password.length < 6) {
    toast.error('密码长度必须在6-20字符之间')
    return
  }

  try {
    addLoading.value = true
    const addData: AddUserRequest = {
      username: newUser.username.trim(),
      password: newUser.password,
      is_super: newUser.permissions.admin ? 1 : 0
    }

    // 先添加用户
    const result = await userApi.addUser(addData)

    // 如果需要设置完整权限，再调用更新接口
    const permissions = calculatePermissions(newUser.permissions)
    if (permissions !== (newUser.permissions.admin ? PERMISSIONS.ADMIN | PERMISSIONS.BASE : PERMISSIONS.BASE)) {
      const updateData: UpdateUserRequest = {
        id: result.id,
        permissions
      }
      await userApi.updateUser(updateData)
    }

    toast.success('用户添加成功')
    closeAddModal()
    fetchUsers()
  } catch (error: any) {
    console.error('添加用户失败:', error)
    toast.error(error.msg || '添加用户失败')
  } finally {
    addLoading.value = false
  }
}

// 编辑用户相关函数
const editUser = (user: User) => {
  editingUser.value = {
    id: user.id,
    username: user.username,
    permissions: parsePermissions(user.permissions)
  }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  editingUser.value = {}
}

const confirmEditUser = async () => {
  try {
    editLoading.value = true
    const permissions = calculatePermissions(editingUser.value.permissions)
    const updateData: UpdateUserRequest = {
      id: editingUser.value.id,
      permissions
    }

    await userApi.updateUser(updateData)
    toast.success('用户权限更新成功')
    closeEditModal()
    fetchUsers()
  } catch (error: any) {
    console.error('更新用户失败:', error)
    toast.error(error.msg || '更新用户失败')
  } finally {
    editLoading.value = false
  }
}

// 重置密码相关函数
const resetPassword = (user: User) => {
  resetPasswordUser.value = user
  newPassword.value = ''
  showPasswordModal.value = true
}

const closePasswordModal = () => {
  showPasswordModal.value = false
  resetPasswordUser.value = null
  newPassword.value = ''
}

const confirmResetPassword = async () => {
  if (!newPassword.value.trim() || newPassword.value.length < 6) {
    toast.error('密码长度必须在6-20字符之间')
    return
  }

  try {
    passwordLoading.value = true
    const modifyData: ModifyPasswordRequest = {
      id: resetPasswordUser.value!.id,
      password: newPassword.value
    }

    await userApi.modifyPassword(modifyData)
    toast.success('密码重置成功')
    closePasswordModal()
  } catch (error: any) {
    console.error('重置密码失败:', error)
    toast.error(error.msg || '重置密码失败')
  } finally {
    passwordLoading.value = false
  }
}

// 绑定用户组相关函数
const bindUserGroup = async (user: User) => {
  bindGroupUser.value = user
  showBindGroupModal.value = true

  // 先设置当前用户的用户组ID，确保在获取用户组列表之前就有默认值
  selectedGroupId.value = user.groupId || 0

  try {
    groupLoading.value = true
    // 获取所有用户组
    const groupsResponse = await userGroupApi.getUserGroupList({ currentPage: 1, pageSize: 1000 })
    allUserGroups.value = groupsResponse.data || []

    // 再次确认设置用户当前绑定的用户组ID（防止异步问题）
    selectedGroupId.value = user.groupId || 0
  } catch (error: any) {
    console.error('获取用户组列表失败:', error)
    toast.error('获取用户组列表失败')
  } finally {
    groupLoading.value = false
  }
}

const closeBindGroupModal = () => {
  showBindGroupModal.value = false
  bindGroupUser.value = null
  allUserGroups.value = []
  selectedGroupId.value = 0
}

const confirmBindUserGroup = async () => {
  try {
    bindGroupLoading.value = true

    // 调用绑定用户组接口
    await userApi.bindGroup({
      userId: bindGroupUser.value!.id,
      groupId: selectedGroupId.value
    })

    toast.success('用户组绑定成功')
    closeBindGroupModal()
    fetchUsers() // 刷新用户列表
  } catch (error: any) {
    console.error('绑定用户组失败:', error)
    toast.error(error.msg || '绑定用户组失败')
  } finally {
    bindGroupLoading.value = false
  }
}

// 删除用户
const deleteUser = async (user: User) => {
  const confirmed = await confirmDialog({
    title: '删除用户',
    message: `确定要删除用户 "${user.username}" 吗？此操作不可恢复。`,
    confirmText: '删除',
    cancelText: '取消',
    isDanger: true
  })

  if (!confirmed) {
    return
  }

  try {
    await userApi.deleteUser(user.id)
    toast.success('用户删除成功')
    fetchUsers()
  } catch (error: any) {
    console.error('删除用户失败:', error)
    toast.error(error.msg || '删除用户失败')
  }
}

// 页面初始化
onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
/* 屏幕阅读器专用的隐藏类 */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* fieldset 和 legend 样式 */
fieldset {
  border: none;
  padding: 0;
  margin: 0;
}

legend {
  padding: 0;
  margin-bottom: 0.5rem;
}

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
.users-table-container {
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
.users-table {
  width: 100%;
  border-collapse: collapse;
}

.users-table th {
  background: #f9fafb;
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
}

.users-table td {
  padding: 1rem;
  border-bottom: 1px solid #f3f4f6;
  font-size: 0.875rem;
}

.user-row:hover {
  background: #f9fafb;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
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

.username {
  font-weight: 500;
  color: #1f2937;
}

/* 状态徽章 */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-active {
  background: #dcfce7;
  color: #166534;
}

.status-inactive {
  background: #fee2e2;
  color: #991b1b;
}

/* 用户组名称 */
.group-name {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  background: #f0f9ff;
  color: #0369a1;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  border: 1px solid #e0f2fe;
}

/* 权限标签 */
.permissions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.permission-tag {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  background: #eff6ff;
  color: #1d4ed8;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
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

.form-input:disabled {
  background: #f9fafb;
  color: #6b7280;
  cursor: not-allowed;
}

/* 权限复选框 */
.permission-checkboxes {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.checkbox-item:hover {
  background: #f9fafb;
}

.checkbox-item input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: #3b82f6;
  margin: 0;
}

.checkbox-item span {
  font-size: 0.875rem;
  color: #374151;
}

/* 重置密码信息 */
.reset-info {
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
  font-size: 0.875rem;
  color: #92400e;
}

/* 绑定用户组信息 */
.bind-info {
  background: #eff6ff;
  border: 1px solid #3b82f6;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
  font-size: 0.875rem;
  color: #1e40af;
}

/* 表单选择框样式 */
.form-select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  background-color: white;
  transition: border-color 0.2s, box-shadow 0.2s;
  box-sizing: border-box;
}

.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-select:disabled {
  background-color: #f9fafb;
  color: #6b7280;
  cursor: not-allowed;
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

  .users-table {
    font-size: 0.75rem;
  }

  .users-table th,
  .users-table td {
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
  .users-table {
    display: block;
    overflow-x: auto;
    white-space: nowrap;
  }
}
</style>
