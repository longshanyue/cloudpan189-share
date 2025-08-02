<template>
  <div id="app-container">
    <router-view />
    <footer class="app-footer">
      <div class="footer-content">
        <a href="https://github.com/xxcheng123/cloudpan189-share" target="_blank" class="footer-link">
          由 cloudpan189-share 驱动
        </a>
        <span class="footer-separator">|</span>
        <router-link to="/@dashboard" class="footer-link">
          管理
        </router-link>
      </div>
      <div class="footer-content" style="margin-top: 4px">
        <span style="font-size: 12px; color: #999;">本系统旨在为您提供便捷的文件访问体验。为确保您的数据和账号安全，非必要勿将您的网站或链接分享到网络，防止因多人异地访问导致被官方限速或封禁账号。</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// 应用初始化时检查认证状态
onMounted(async () => {
  // 如果有token，则获取最新的用户信息
  if (authStore.token) {
    try {
      await authStore.initialize()
    } catch (error) {
      console.error('初始化用户信息失败:', error)
    }
  }
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f5f5f5;
  color: #333;
}

#app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f8fafc;
}

.app-footer {
  margin-top: auto;
  padding: 20px 0;
  background-color: #fff;
  border-top: 1px solid #e5e5e5;
  position: sticky;
  bottom: 0;
  left: 50%;
  width: 100%;
  z-index: 999;
}

.footer-content {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  font-size: 14px;
  color: #666;
}

.footer-link {
  color: #666;
  text-decoration: none;
  transition: color 0.2s ease;
}

.footer-link:hover {
  color: #007bff;
  text-decoration: underline;
}

.footer-separator {
  color: #ccc;
  font-weight: normal;
}
</style>