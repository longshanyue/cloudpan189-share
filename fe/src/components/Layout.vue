<template>
  <div class="dashboard">
    <!-- 顶部导航栏 -->
    <header class="dashboard-header">
      <div class="header-content">
        <div class="header-left">
          <div class="logo">
            <Icons name="cloud" size="1.5rem" class="logo-icon" />
            <!-- 头部Logo文字 -->
            <span class="logo-text">{{ settingStore.setting?.title || 'CloudPan189' }}</span>
          </div>
        </div>
        
        <div class="header-right">
          <!-- 总线状态显示 (仅管理员可见) -->
          <BusStatus v-if="authStore.isAdmin" />
          
          <div class="user-menu">
            <div class="user-avatar">
              <span class="avatar-text">{{ authStore.user?.username?.charAt(0).toUpperCase() }}</span>
            </div>
            <div class="user-info">
              <router-link to="/@profile" class="profile-link">
                <span class="user-name">{{ authStore.user?.username }}</span>
                <span class="user-role">{{ authStore.isAdmin ? '管理员' : '普通用户' }}</span>
              </router-link>
            </div>
            <button @click="handleLogout" class="logout-btn">
              <Icons name="logout" size="1rem" class="logout-icon" />
              退出
            </button>
          </div>
        </div>
      </div>
    </header>
    
    <div class="dashboard-layout">
      <!-- 侧边栏 -->
      <nav class="sidebar" :class="{ collapsed: sidebarCollapsed }">
        <div class="sidebar-content">
          <div class="nav-section">
            <h3 class="nav-section-title" v-show="!sidebarCollapsed">主要功能</h3>
            <ul class="nav-menu">
              <li class="nav-item">
                <router-link to="/@admin/dashboard" class="nav-link" :class="{ active: $route.name === 'Dashboard' }" :title="sidebarCollapsed ? '仪表盘' : ''">
                  <Icons name="dashboard" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">仪表盘</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@admin/profile" class="nav-link" :class="{ active: $route.name === 'Profile' }" :title="sidebarCollapsed ? '个人中心' : ''">
                  <Icons name="user" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">个人中心</span>
                </router-link>
              </li>
            </ul>
          </div>

          <div class="nav-section">
            <h3 class="nav-section-title" v-show="!sidebarCollapsed">文件浏览</h3>
            <ul class="nav-menu">
              <li class="nav-item">
                <router-link to="/" class="nav-link" :class="{ active: $route.name === 'FileBrowser' || $route.name === 'FileBrowserPath' }" :title="sidebarCollapsed ? '文件浏览' : ''">
                  <Icons name="folder" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">文件浏览</span>
                </router-link>
              </li>
            </ul>
          </div>
          
          <div v-if="authStore.isAdmin" class="nav-section">
            <h3 class="nav-section-title" v-show="!sidebarCollapsed">管理功能</h3>
            <ul class="nav-menu">
              <li class="nav-item">
                <router-link to="/@admin/users" class="nav-link" :class="{ active: $route.name === 'Users' }" :title="sidebarCollapsed ? '用户管理' : ''">
                  <Icons name="users" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">用户管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@admin/user_groups" class="nav-link" :class="{ active: $route.name === 'UserGroups' }" :title="sidebarCollapsed ? '用户组管理' : ''">
                  <Icons name="group" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">用户组管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@admin/storage" class="nav-link" :class="{ active: $route.name === 'Storage' }" :title="sidebarCollapsed ? '存储管理' : ''">
                  <Icons name="storage" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">存储管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@admin/cloud_token" class="nav-link" :class="{ active: $route.name === 'CloudToken' }" :title="sidebarCollapsed ? '令牌管理' : ''">
                  <Icons name="tokens" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">令牌管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@admin/setting" class="nav-link" :class="{ active: $route.name === 'Settings' }" :title="sidebarCollapsed ? '系统设置' : ''">
                  <Icons name="settings" class="nav-icon" />
                  <span class="nav-text" v-show="!sidebarCollapsed">系统设置</span>
                </router-link>
              </li>
            </ul>
          </div>
        </div>
        
        <!-- 侧边栏切换按钮 -->
        <div class="sidebar-footer">
          <button @click="toggleSidebar" class="sidebar-toggle">
            <Icons :name="sidebarCollapsed ? 'chevron-right' : 'arrow-left'" size="1.25rem" />
          </button>
        </div>
      </nav>
      
      <!-- 主内容区 -->
      <main class="main-content" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettingStore } from '@/stores/setting'
import Icons from './Icons.vue'
import BusStatus from './BusStatus.vue'

const router = useRouter()
const authStore = useAuthStore()
const settingStore = useSettingStore()

// 侧边栏收缩状态
const sidebarCollapsed = ref(false)

// 切换侧边栏状态
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
  // 保存状态到localStorage
  localStorage.setItem('sidebarCollapsed', String(sidebarCollapsed.value))
}

// 组件挂载时恢复侧边栏状态
onMounted(() => {
  const saved = localStorage.getItem('sidebarCollapsed')
  if (saved !== null) {
    sidebarCollapsed.value = saved === 'true'
  }
})

// 退出登录
const handleLogout = async () => {
  authStore.logout()

  await router.push('/@login')
}
</script>

<style scoped>
.dashboard {
  background: #f8fafc;
}

.dashboard-header {
  background: white;
  border-bottom: 1px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  width: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.sidebar-footer {
  position: absolute;
  bottom: 1rem;
  left: 0;
  right: 0;
  padding: 0 1rem;
  z-index: 1001;
  display: flex;
}

.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 40px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  color: #64748b;
  transition: all 0.2s;
}

.sidebar-toggle:hover {
  background: #e2e8f0;
  color: #475569;
}

.sidebar.collapsed .sidebar-toggle {
  width: 40px;
  margin: 0 auto;
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo-icon {
  font-size: 1.5rem;
}

.logo-text {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
}

.avatar-text {
  font-size: 1rem;
}

.user-info {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.profile-link {
  text-decoration: none;
  color: inherit;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  transition: opacity 0.2s ease;
}

.profile-link:hover {
  opacity: 0.8;
}

.user-name {
  font-weight: 600;
  color: #1f2937;
}

.user-role {
  font-size: 0.875rem;
  color: #6b7280;
}

.logout-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.logout-btn:hover {
  background: #dc2626;
}

.logout-icon {
  font-size: 1rem;
}

.dashboard-layout {
  display: flex;
}

.sidebar {
  width: 12rem;
  background: white;
  border-right: 1px solid #e5e7eb;
  flex-shrink: 0;
  position: fixed;
  left: 0;
  height: calc(100vh - 73px);
  overflow-y: auto;
  z-index: 1000;
  padding-bottom: 100px;
  transition: width 0.3s ease;
}

.sidebar.collapsed {
  width: 4rem;
}

.sidebar.collapsed .nav-section-title {
  display: none;
}

.sidebar.collapsed .nav-text {
  display: none;
}

.sidebar.collapsed .nav-link {
  justify-content: center;
  padding: 0.75rem;
}

.sidebar-content {
  padding: 1.5rem 0;
}

.nav-section {
  margin-bottom: 2rem;
}

.nav-section-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 1rem 1.5rem;
}

.nav-menu {
  list-style: none;
  margin: 0;
  padding: 0;
}

.nav-item {
  margin: 0;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1.5rem;
  color: #6b7280;
  text-decoration: none;
  transition: all 0.2s;
  border-right: 3px solid transparent;
}

.nav-link:hover {
  background: #f3f4f6;
  color: #1f2937;
}

.nav-link.active {
  background: #eff6ff;
  color: #3b82f6;
  border-right-color: #3b82f6;
}

.nav-icon {
  font-size: 1.125rem;
  width: 1.25rem;
  text-align: center;
}

.nav-text {
  font-weight: 500;
  font-size: 0.875rem;
}

.main-content {
  flex: 1;
  padding: 2rem;
  margin-left: 12rem;
  transition: margin-left 0.3s ease;
}

.main-content.sidebar-collapsed {
  margin-left: 4rem;
}

@media (max-width: 768px) {
  .dashboard-layout {
    flex-direction: column;
  }
  
  .sidebar {
    width: 100%;
    order: 2;
    position: relative;
    top: auto;
    left: auto;
    height: auto;
  }
  
  .sidebar.collapsed {
    width: 100%;
  }
  
  .main-content {
    margin-left: 0;
  }
  
  .main-content.sidebar-collapsed {
    margin-left: 0;
  }
  
  .main-content {
    padding: 1rem;
    order: 1;
  }
}
</style>