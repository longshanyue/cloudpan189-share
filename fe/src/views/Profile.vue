<template>
  <div>
    <!-- 个人中心主卡片 -->
    <PageCard title="个人中心" subtitle="管理您的个人信息和账户设置">
      <SectionDivider />
      
      <SubsectionTitle title="基本信息" />
        <div class="profile-item">
          <div class="profile-label">
            <span class="label-text">用户名</span>
            <span class="label-desc">当前登录用户名</span>
          </div>
          <div class="profile-value">{{ authStore.user?.username || '-' }}</div>
        </div>
        
        <div class="profile-item">
          <div class="profile-label">
            <span class="label-text">状态</span>
            <span class="label-desc">账户状态</span>
          </div>
          <div class="profile-value">
            <span :class="['status-badge', authStore.user?.status === 1 ? 'status-active' : 'status-inactive']">
              {{ authStore.user?.status === 1 ? '正常' : '禁用' }}
            </span>
          </div>
        </div>
        
        <div class="profile-item">
          <div class="profile-label">
            <span class="label-text">创建时间</span>
            <span class="label-desc">账户创建时间</span>
          </div>
          <div class="profile-value">{{ formatDate(authStore.user?.createdAt) }}</div>
        </div>
        
        <div class="profile-item">
          <div class="profile-label">
            <span class="label-text">上次修改时间</span>
            <span class="label-desc">最后更新时间</span>
          </div>
          <div class="profile-value">{{ formatDate(authStore.user?.updatedAt) }}</div>
        </div>
        
        <SectionDivider />
        
        <SubsectionTitle title="权限信息" />
        <div class="permissions-grid">
          <div 
            v-for="permission in userPermissions" 
            :key="permission.value"
            class="permission-item"
          >
            <div class="permission-icon">
              <Icons 
                :name="getPermissionIcon(permission.value)" 
                size="1.2rem" 
              />
            </div>
            <div class="permission-details">
              <div class="permission-name">{{ permission.label }}</div>
              <div class="permission-desc">{{ permission.description }}</div>
            </div>
          </div>
        </div>
        <div v-if="userPermissions.length === 0" class="no-permissions">
          <Icons name="key" size="2rem" class="empty-icon" />
          <p>暂无权限信息</p>
        </div>
        
        <SectionDivider />
        
        <SubsectionTitle title="安全设置" />
        <div class="profile-item">
          <div class="profile-label">
            <span class="label-text">密码管理</span>
            <span class="label-desc">修改登录密码</span>
          </div>
          <div class="profile-control">
            <button @click="showPasswordModal = true" class="profile-btn profile-btn-primary">
              <Icons name="key" size="1rem" />
              修改密码
            </button>
          </div>
        </div>
    </PageCard>

    <!-- 修改密码弹窗 -->
    <div v-if="showPasswordModal" class="modal-overlay" @click="closePasswordModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <div class="modal-title">
            <Icons name="key" size="1.5rem" class="modal-icon" />
            <h3>修改密码</h3>
          </div>
          <button @click="closePasswordModal" class="close-btn">
            <span class="close-icon">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="handlePasswordChange">
            <div class="form-group">
              <label for="currentPassword">
                <Icons name="lock" size="1rem" />
                当前密码
              </label>
              <input
                id="currentPassword"
                v-model="passwordForm.currentPassword"
                type="password"
                required
                class="form-input"
                placeholder="请输入当前密码"
              />
            </div>
            <div class="form-group">
              <label for="newPassword">
                <Icons name="key" size="1rem" />
                新密码
              </label>
              <input
                id="newPassword"
                v-model="passwordForm.newPassword"
                type="password"
                required
                class="form-input"
                placeholder="请输入新密码（至少6位）"
              />
            </div>
            <div class="form-group">
              <label for="confirmPassword">
                <Icons name="check" size="1rem" />
                确认新密码
              </label>
              <input
                id="confirmPassword"
                v-model="passwordForm.confirmPassword"
                type="password"
                required
                class="form-input"
                placeholder="请再次输入新密码"
              />
            </div>
            <div class="form-actions">
              <button type="button" @click="closePasswordModal" class="cancel-btn">
                <Icons name="x" size="1rem" />
                取消
              </button>
              <button type="submit" class="submit-btn" :disabled="passwordLoading">
                <Icons name="check" size="1rem" />
                {{ passwordLoading ? '修改中...' : '确认修改' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Icons from '@/components/Icons.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'
import { useAuthStore } from '@/stores/auth'
import { getPermissionDetails } from '@/utils/permissions'
import { userApi } from '@/api/user'
import { toast } from '@/utils/toast'

const router = useRouter()
const authStore = useAuthStore()

// 密码修改相关
const showPasswordModal = ref(false)
const passwordLoading = ref(false)
const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 用户权限信息
const userPermissions = computed(() => {
  if (!authStore.user?.permissions) return []
  return getPermissionDetails(authStore.user.permissions)
})

// 格式化日期
const formatDate = (dateString: string | undefined): string => {
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

// 获取权限图标
const getPermissionIcon = (permissionValue: number): string => {
  switch (permissionValue) {
    case 1: return 'user'
    case 2: return 'storage'
    case 4: return 'settings'
    default: return 'key'
  }
}

// 关闭密码修改弹窗
const closePasswordModal = () => {
  showPasswordModal.value = false
  passwordForm.value = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}

// 处理密码修改
const handlePasswordChange = async () => {
  if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
    toast.error('新密码和确认密码不一致')
    return
  }

  if (passwordForm.value.newPassword.length < 6) {
    toast.error('新密码长度不能少于6位')
    return
  }

  try {
    passwordLoading.value = true
    await userApi.modifyOwnPassword({
      oldPassword: passwordForm.value.currentPassword,
      password: passwordForm.value.newPassword
    })
    toast.success('密码修改成功，即将跳转到登录页面')
    closePasswordModal()
    
    // 清除用户信息并跳转到登录页面
    setTimeout(() => {
      authStore.logout()
      router.push('/@login')
    }, 1500)
  } catch (error: any) {
    console.error('密码修改失败:', error)
    toast.error(error.response?.data?.msg || '密码修改失败')
  } finally {
    passwordLoading.value = false
  }
}

// 页面初始化
onMounted(() => {
  // 确保用户信息是最新的
  if (authStore.token) {
    authStore.fetchUserInfo()
  }
})
</script>

<style scoped>
/* 个人信息项样式 */
.profile-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid #f3f4f6;
}

.profile-item:last-child {
  border-bottom: none;
}

.profile-label {
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
}

.profile-value {
  font-size: 0.875rem;
  color: #1f2937;
  font-weight: 500;
}

.profile-control {
  display: flex;
  gap: 0.75rem;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}

.status-active {
  background: #dcfce7;
  color: #166534;
}

.status-inactive {
  background: #fef2f2;
  color: #dc2626;
}

/* 权限网格样式 */
.permissions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
  padding: 0 2rem 1rem 2rem;
}

.permission-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.permission-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border-radius: 8px;
}

.permission-details {
  flex: 1;
}

.permission-name {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 0.25rem;
}

.permission-desc {
  font-size: 0.875rem;
  color: #6b7280;
}

.no-permissions {
  text-align: center;
  padding: 2rem;
  color: #6b7280;
  margin: 0 2rem;
}

.empty-icon {
  opacity: 0.5;
  margin-bottom: 0.5rem;
}

/* 按钮样式 */
.profile-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  text-decoration: none;
}

.profile-btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border-color: #3b82f6;
}

.profile-btn-primary:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.profile-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border-radius: 12px 12px 0 0;
}

.modal-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.modal-icon {
  color: #fbbf24;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: white;
}

.close-btn {
  background: rgba(255, 255, 255, 0.1);
  border: none;
  cursor: pointer;
  color: white;
  padding: 0.5rem;
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: scale(1.1);
}

.close-icon {
  font-size: 1.5rem;
  font-weight: bold;
  line-height: 1;
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #374151;
  font-size: 0.95rem;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 1rem;
  transition: border-color 0.2s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 1px #3b82f6;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

.cancel-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.cancel-btn:hover {
  background: #f3f4f6;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.submit-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .permissions-grid {
    grid-template-columns: 1fr;
    padding: 0 1rem 1rem 1rem;
  }
  
  .profile-item {
    padding: 1rem;
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .modal-content {
    width: 95%;
    margin: 1rem;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .cancel-btn,
  .submit-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
