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
      <nav class="sidebar">
        <div class="sidebar-content">
          <div class="nav-section">
            <h3 class="nav-section-title">主要功能</h3>
            <ul class="nav-menu">
              <li class="nav-item">
                <router-link to="/@dashboard" class="nav-link" :class="{ active: $route.name === 'Dashboard' }">
                  <Icons name="dashboard" class="nav-icon" />
                  <span class="nav-text">仪表盘</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@profile" class="nav-link" :class="{ active: $route.name === 'Profile' }">
                  <Icons name="user" class="nav-icon" />
                  <span class="nav-text">个人中心</span>
                </router-link>
              </li>
            </ul>
          </div>

          <div class="nav-section">
            <h3 class="nav-section-title">文件浏览</h3>
            <ul class="nav-menu">
              <li class="nav-item">
                <router-link to="/" class="nav-link" :class="{ active: $route.name === 'FileBrowser' || $route.name === 'FileBrowserPath' }">
                  <Icons name="folder" class="nav-icon" />
                  <span class="nav-text">文件浏览</span>
                </router-link>
              </li>
            </ul>
          </div>
          
          <div v-if="authStore.isAdmin" class="nav-section">
            <h3 class="nav-section-title">管理功能</h3>
            <ul class="nav-menu">
              <li class="nav-item">
                <router-link to="/@users" class="nav-link" :class="{ active: $route.name === 'Users' }">
                  <Icons name="users" class="nav-icon" />
                  <span class="nav-text">用户管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@usergroups" class="nav-link" :class="{ active: $route.name === 'UserGroups' }">
                  <Icons name="group" class="nav-icon" />
                  <span class="nav-text">用户组管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@storage" class="nav-link" :class="{ active: $route.name === 'Storage' }">
                  <Icons name="storage" class="nav-icon" />
                  <span class="nav-text">存储管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@tokens" class="nav-link" :class="{ active: $route.name === 'CloudToken' }">
                  <Icons name="tokens" class="nav-icon" />
                  <span class="nav-text">令牌管理</span>
                </router-link>
              </li>
              <li class="nav-item">
                <router-link to="/@setting" class="nav-link" :class="{ active: $route.name === 'Settings' }">
                  <Icons name="settings" class="nav-icon" />
                  <span class="nav-text">系统设置</span>
                </router-link>
              </li>
            </ul>
          </div>
        </div>
      </nav>
      
      <!-- 主内容区 -->
      <main class="main-content">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettingStore } from '@/stores/setting'
import Icons from './Icons.vue'

const router = useRouter()
const authStore = useAuthStore()
const settingStore = useSettingStore()

// 退出登录
const handleLogout = async () => {
  await authStore.logout()
  router.push('/@login')
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
  width: 250px;
  background: white;
  border-right: 1px solid #e5e7eb;
  flex-shrink: 0;
  position: fixed;
  left: 0;
  height: calc(100vh - 73px);
  overflow-y: auto;
  z-index: 10;
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
  margin-left: 250px;
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
  
  .main-content {
    margin-left: 0;
  }
  
  .main-content {
    padding: 1rem;
    order: 1;
  }
}
</style>