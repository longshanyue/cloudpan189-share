<template>
  <div>
    <!-- 仪表盘主卡片 -->
    <PageCard :title="'仪表盘'" :subtitle="getGreeting() + '，' + authStore.user?.username + '！'">
      <SectionDivider />
      
      <SubsectionTitle title="用户信息" />
        <div class="info-section">
          <div class="info-item">
            <div class="info-label">
              <span class="label-text">当前用户</span>
              <span class="label-desc">登录用户身份信息</span>
            </div>
            <div class="info-value">
              <div class="user-badge">
                <div class="user-avatar">
                  <span class="avatar-text">{{ authStore.user?.username?.charAt(0).toUpperCase() }}</span>
                </div>
                <div class="user-details">
                  <span class="user-name">{{ authStore.user?.username }}</span>
                  <span class="user-role" :class="{ 'admin-role': authStore.isAdmin }">
                    {{ authStore.isAdmin ? '管理员' : '普通用户' }}
                  </span>
                </div>
                <div class="user-status">
                  <div class="status-dot"></div>
                  <span class="status-text">在线</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <SectionDivider />
        
        <SubsectionTitle title="用户权限" />
        <div class="info-section">
          <div class="permissions-grid">
            <div 
              v-for="permission in getPermissionDetails(authStore.user?.permissions || 0)" 
              :key="permission.value"
              class="permission-card"
            >
              <div class="permission-header">
                <Icons 
                  :name="permission.value === 1 ? 'user' : permission.value === 2 ? 'storage' : 'settings'" 
                  class="permission-icon" 
                />
                <span class="permission-name">{{ permission.label }}</span>
              </div>
              <p class="permission-desc">{{ permission.description }}</p>
            </div>
          </div>
        </div>
        
        <SectionDivider />
        
        <SubsectionTitle title="系统状态" />
        <div class="info-section">
          <div class="info-item">
            <div class="info-label">
              <span class="label-text">运行时间</span>
              <span class="label-desc">系统持续运行时长</span>
            </div>
            <div class="info-value">
               <span class="runtime-display">{{ formatRunTime(settingStore.setting?.runTimes || 0) }}</span>
             </div>
          </div>
        </div>
        
        <SectionDivider />
        
        <SubsectionTitle title="快速使用" />
        <div class="feature-section">
          <div class="dav-grid">
            <div class="dav-item">
              <div class="dav-header">
                <div class="dav-icon">
                  <Icons name="folder" size="1.2rem" />
                </div>
                <div class="dav-info">
                  <h4 class="dav-title">WebDAV</h4>
                  <p class="dav-desc">通过 WebDAV 协议访问文件系统</p>
                  <p class="dav-hint">适用于各种支持 WebDAV 的客户端</p>
                </div>
              </div>
              <div class="dav-url">
                <input
                  type="text"
                  :value="getDavUrl('/dav')"
                  readonly
                  class="url-input"
                  @click="selectText"
                />
                <button @click="copyToClipboard(getDavUrl('/dav'))" class="copy-btn">
                  <Icons name="copy" size="0.9rem" />
                </button>
              </div>
            </div>
          </div>
        </div>
    </PageCard>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useSettingStore } from '@/stores/setting'
import Icons from '@/components/Icons.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'
import { getPermissionDetails } from '@/utils/permissions'
import { toast } from '@/utils/toast'

const authStore = useAuthStore()
const settingStore = useSettingStore()

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

// 定时器引用
const timer = ref<NodeJS.Timeout | null>(null)

// 获取设置数据
const fetchSettingData = async () => {
  try {
    await settingStore.fetchSetting()
  } catch (error) {
    console.error('获取网站设置失败:', error)
  }
}

// 获取问候语
const getGreeting = (): string => {
  const hour = new Date().getHours()
  if (hour >= 5 && hour < 12) {
    return '上午好'
  } else if (hour >= 12 && hour < 14) {
    return '中午好'
  } else if (hour >= 14 && hour < 18) {
    return '下午好'
  } else {
    return '晚上好'
  }
}

// 获取DAV URL
const getDavUrl = (path: string): string => {
  const baseURL = settingStore.setting?.baseURL || window.location.origin
  return baseURL.endsWith('/') ? baseURL + path.substring(1) : baseURL + path
}

// 选中文本
const selectText = (event: Event) => {
  const input = event.target as HTMLInputElement
  input.select()
}

// 复制到剪贴板
const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    toast.success('DAV 地址已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    // 降级方案：使用传统方法
    const textArea = document.createElement('textarea')
    textArea.value = text
    document.body.appendChild(textArea)
    textArea.select()
    try {
      document.execCommand('copy')
      toast.success('DAV 地址已复制到剪贴板')
    } catch (fallbackError) {
      console.error('降级复制也失败:', fallbackError)
      toast.error('复制失败，请手动复制')
    }
    document.body.removeChild(textArea)
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
/* 信息区域样式 */
.info-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 2rem;
  border-bottom: 1px solid #f3f4f6;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
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

.info-value {
  display: flex;
  align-items: center;
}

/* 美化的用户徽章样式 */
.user-badge {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
  min-width: 280px;
}

.user-badge:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  transform: translateY(-1px);
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 1rem;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  flex-shrink: 0;
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.user-name {
  font-weight: 600;
  color: #1f2937;
  font-size: 0.9rem;
}

.user-role {
  font-size: 0.75rem;
  color: #6b7280;
  padding: 0.125rem 0.5rem;
  background: #f3f4f6;
  border-radius: 12px;
  display: inline-block;
  width: fit-content;
}

.user-role.admin-role {
  background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  color: #92400e;
  border: 1px solid #f59e0b;
}

.user-status {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  flex-shrink: 0;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.status-text {
  font-size: 0.75rem;
  color: #10b981;
  font-weight: 500;
}

/* 运行时间显示 */
.runtime-display {
  font-weight: 600;
  color: #059669;
  font-size: 0.875rem;
  padding: 0.5rem 0.75rem;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdf4 100%);
  border-radius: 8px;
  border: 1px solid #bbf7d0;
}

/* 功能区域样式 */
.feature-section {
  margin-top: 0.5rem;
}

/* DAV 网格布局 */
.dav-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 1rem;
}

/* DAV 项目样式 */
.dav-item {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 1.25rem;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.dav-item:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  transform: translateY(-1px);
}

/* DAV 头部 */
.dav-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.dav-icon {
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  border: 1px solid #3b82f6;
  border-radius: 8px;
  color: #1d4ed8;
  flex-shrink: 0;
}


.dav-info {
  flex: 1;
}

.dav-title {
  font-size: 1rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.25rem 0;
}

.dav-desc {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0;
  line-height: 1.4;
}

.dav-hint {
  font-size: 0.75rem;
  color: #f59e0b;
  margin: 0.375rem 0 0 0;
  line-height: 1.4;
  font-style: italic;
}

/* DAV URL 输入框和复制按钮 */
.dav-url {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.url-input {
  flex: 1;
  padding: 0.75rem 1rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  background: #f9fafb;
  font-size: 0.875rem;
  color: #374151;
  font-family: 'Courier New', monospace;
  cursor: pointer;
  transition: all 0.2s ease;
}

.url-input:hover {
  border-color: #3b82f6;
  background: #f0f9ff;
}

.url-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  background: white;
}

.copy-btn {
  padding: 0.75rem;
  background: #3b82f6;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2.5rem;
  height: 2.5rem;
}

.copy-btn:hover {
  background: #1d4ed8;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
}

.copy-btn:active {
  transform: translateY(0);
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: white;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  transition: all 0.2s ease;
  cursor: pointer;
}

.feature-item:hover {
  background: #f9fafb;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.feature-icon {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f1f5f9;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  flex-shrink: 0;
  transition: all 0.2s ease;
  color: #6b7280;
}

.feature-item:hover .feature-icon {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

/* 权限卡片样式 */
.permissions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 0.75rem;
  padding: 0 2rem 1rem 2rem;
}

.permission-card {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 0.875rem;
  transition: all 0.2s ease;
}

.permission-card:hover {
  background: #f3f4f6;
  border-color: #d1d5db;
}

.permission-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.375rem;
}

.permission-icon {
  width: 1rem;
  height: 1rem;
  color: #3b82f6;
}

.permission-name {
  font-weight: 600;
  color: #1f2937;
  font-size: 0.8rem;
}

.permission-desc {
  color: #6b7280;
  font-size: 0.75rem;
  line-height: 1.4;
  margin: 0;
}

.feature-content {
  flex: 1;
}

.feature-title {
  font-size: 0.8rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.25rem 0;
}

.feature-desc {
  font-size: 0.7rem;
  color: #6b7280;
  margin: 0;
  line-height: 1.4;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .info-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
  }
  
  .info-value {
    width: 100%;
    justify-content: flex-start;
  }
  
  .user-badge {
    width: 100%;
    min-width: auto;
  }
  
  .feature-grid {
    grid-template-columns: 1fr;
    gap: 0.5rem;
  }
  
  .feature-item {
    padding: 0.75rem;
  }
  
  .permissions-grid {
    grid-template-columns: 1fr;
    padding: 0 1rem 1rem 1rem;
  }
  
  /* DAV 响应式 */
  .dav-grid {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }
  
  .dav-item {
    padding: 1rem;
  }
  
  .dav-header {
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }
  
  .dav-icon {
    width: 2rem;
    height: 2rem;
  }
  
  .dav-title {
    font-size: 0.9rem;
  }
  
  .dav-desc {
    font-size: 0.8rem;
  }
}

@media (max-width: 480px) {
  .feature-item {
    flex-direction: column;
    text-align: center;
    gap: 0.5rem;
  }
  
  .feature-icon {
    align-self: center;
  }
  
  .user-badge {
    flex-direction: column;
    text-align: center;
    gap: 0.75rem;
  }
  
  .user-status {
    justify-content: center;
  }
  
  /* DAV 小屏幕响应式 */
  .dav-item {
    padding: 0.75rem;
  }
  
  .dav-header {
    flex-direction: column;
    text-align: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }
  
  .dav-icon {
    align-self: center;
    width: 1.75rem;
    height: 1.75rem;
  }
  
  .dav-info {
    text-align: center;
  }
  
  .dav-title {
    font-size: 0.85rem;
  }
  
  .dav-desc {
    font-size: 0.75rem;
  }
  
  .dav-url {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .url-input {
    font-size: 0.8rem;
    padding: 0.6rem 0.8rem;
  }
  
  .copy-btn {
    width: 100%;
    min-width: auto;
    height: 2.25rem;
  }
}
</style>
