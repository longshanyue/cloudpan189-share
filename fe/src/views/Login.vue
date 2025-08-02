<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="bg-circle bg-circle-1"></div>
      <div class="bg-circle bg-circle-2"></div>
      <div class="bg-circle bg-circle-3"></div>
      <div class="bg-wave"></div>
    </div>
    
    <!-- 左侧信息面板 -->
    <div class="login-info">
      <div class="info-content">
        <div class="logo">
          <div class="logo-icon">☁️</div>
          <h1 class="logo-text">{{ settingStore.setting?.title || 'CloudPan189' }}</h1>
        </div>
        <p class="info-description">
          安全、高效的云盘分享管理平台
        </p>
        <div class="feature-list">
          <div class="feature-item">
            <div class="feature-icon-wrapper">
              <span class="feature-icon">🔒</span>
            </div>
            <div class="feature-content">
              <span class="feature-title">安全可靠</span>
              <span class="feature-desc">鉴权机制防护</span>
            </div>
          </div>
          <div class="feature-item">
            <div class="feature-icon-wrapper">
              <span class="feature-icon">⚡</span>
            </div>
            <div class="feature-content">
              <span class="feature-title">高效管理</span>
              <span class="feature-desc">智能文件管理</span>
            </div>
          </div>
          <div class="feature-item">
            <div class="feature-icon-wrapper">
              <span class="feature-icon">🌐</span>
            </div>
            <div class="feature-content">
              <span class="feature-title">多线程加载</span>
              <span class="feature-desc">极速访问体验</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 右侧登录表单 -->
    <div class="login-panel">
      <div class="login-form">
        <div class="form-header">
          <div class="welcome-badge">
            <span class="badge-icon">👋</span>
            <span class="badge-text">欢迎回来</span>
          </div>
          <h2 class="form-title">登录您的账户</h2>
          <p class="form-subtitle">请输入您的登录凭据以继续</p>
        </div>
        
        <form @submit.prevent="handleLogin" class="form-content">
          <div class="form-group">
            <label class="form-label">用户名</label>
            <div class="input-wrapper">
              <div class="input-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                  <circle cx="12" cy="7" r="4"></circle>
                </svg>
              </div>
              <input
                v-model="loginForm.username"
                type="text"
                placeholder="请输入用户名"
                class="form-input"
                :class="{ error: errors.username }"
              />
            </div>
            <span v-if="errors.username" class="error-message">{{ errors.username }}</span>
          </div>
          
          <div class="form-group">
            <label class="form-label">密码</label>
            <div class="input-wrapper">
              <div class="input-icon">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <circle cx="12" cy="16" r="1"></circle>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
              </div>
              <input
                v-model="loginForm.password"
                type="password"
                placeholder="请输入密码"
                class="form-input"
                :class="{ error: errors.password }"
              />
            </div>
            <span v-if="errors.password" class="error-message">{{ errors.password }}</span>
          </div>
          
          <div v-if="errorMessage" class="error-alert">
            <div class="error-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="8" x2="12" y2="12"></line>
                <line x1="12" y1="16" x2="12.01" y2="16"></line>
              </svg>
            </div>
            <span class="error-text">{{ errorMessage }}</span>
          </div>
          
          <button
            type="submit"
            class="login-button"
            :disabled="loading"
          >
            <span v-if="loading" class="loading-spinner"></span>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M15 3h6v6"></path>
              <path d="M10 14 21 3"></path>
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
            </svg>
            {{ loading ? '登录中...' : '立即登录' }}
          </button>
        </form>
        
        <div class="login-footer">
          <div class="footer-divider"></div>
          <p class="footer-text">© 2025 {{ settingStore.setting?.title || 'CloudPan189' }}. 保留所有权利</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettingStore } from '@/stores/setting'

const router = useRouter()
const authStore = useAuthStore()
const settingStore = useSettingStore()

// 表单数据
const loginForm = reactive({
  username: '',
  password: ''
})

// 错误信息
const errors = reactive({
  username: '',
  password: ''
})

const errorMessage = ref('')
const loading = ref(false)

// 表单验证
const validateForm = () => {
  errors.username = ''
  errors.password = ''
  
  if (!loginForm.username) {
    errors.username = '请输入用户名'
    return false
  }
  
  if (loginForm.username.length < 3 || loginForm.username.length > 20) {
    errors.username = '用户名长度应在3-20个字符之间'
    return false
  }
  
  if (!loginForm.password) {
    errors.password = '请输入密码'
    return false
  }
  
  if (loginForm.password.length < 6 || loginForm.password.length > 20) {
    errors.password = '密码长度应在6-20个字符之间'
    return false
  }
  
  return true
}

// 处理登录
const handleLogin = async () => {
  if (!validateForm()) {
    return
  }
  
  loading.value = true
  errorMessage.value = ''
  
  try {
    await authStore.login(loginForm)
    router.push('/@dashboard')
  } catch (error: any) {
    errorMessage.value = error.msg || '登录失败，请检查用户名和密码'
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取设置
onMounted(async () => {
  try {
    await settingStore.fetchSetting()
  } catch (error) {
    console.error('获取网站设置失败:', error)
  }
})
</script>

<style scoped>
.login-container {
  display: flex;
  position: relative;
  overflow: hidden;
  flex: 1;
}

/* 背景装饰 */
.bg-decoration {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 50%, #1e40af 100%);
}

.bg-wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 200px;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1200 120'%3E%3Cpath d='M321.39,56.44c58-10.79,114.16-30.13,172-41.86,82.39-16.72,168.19-17.73,250.45-.39C823.78,31,906.67,72,985.66,92.83c70.05,18.48,146.53,26.09,214.34,3V0H0V27.35A600.21,600.21,0,0,0,321.39,56.44Z' fill='rgba(255,255,255,0.1)'%3E%3C/path%3E%3C/svg%3E") no-repeat;
  background-size: cover;
}

.bg-circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
  animation: float 8s ease-in-out infinite;
  backdrop-filter: blur(2px);
}

.bg-circle-1 {
  width: 300px;
  height: 300px;
  top: 5%;
  left: 5%;
  animation-delay: 0s;
}

.bg-circle-2 {
  width: 200px;
  height: 200px;
  top: 50%;
  right: 10%;
  animation-delay: 3s;
}

.bg-circle-3 {
  width: 150px;
  height: 150px;
  bottom: 15%;
  left: 15%;
  animation-delay: 6s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px) scale(1);
    opacity: 0.8;
  }
  50% {
    transform: translateY(-30px) scale(1.05);
    opacity: 1;
  }
}

/* 左侧信息面板 */
.login-info {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
  padding: 3rem;
}

.info-content {
  max-width: 500px;
  color: white;
  text-align: left;
}

.logo {
  margin-bottom: 3rem;
  text-align: center;
}

.logo-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  display: block;
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.2));
}

.logo-text {
  font-size: 2.8rem;
  font-weight: 800;
  margin: 0;
  text-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
  background: linear-gradient(45deg, #ffffff, #e0f2fe);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.info-description {
  font-size: 1.2rem;
  line-height: 1.6;
  margin-bottom: 3rem;
  opacity: 0.95;
  text-align: center;
  font-weight: 300;
}

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: all 0.3s ease;
}

.feature-item:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateX(5px);
}

.feature-icon-wrapper {
  width: 50px;
  height: 50px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.feature-icon {
  font-size: 1.5rem;
}

.feature-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.feature-title {
  font-size: 1.1rem;
  font-weight: 600;
}

.feature-desc {
  font-size: 0.9rem;
  opacity: 0.8;
}

/* 右侧登录面板 */
.login-panel {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  position: relative;
  z-index: 1;
  padding: 3rem;
}

.login-form {
  width: 100%;
  max-width: 420px;
}

.form-header {
  text-align: center;
  margin-bottom: 2.5rem;
}

.welcome-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: #eff6ff;
  color: #3b82f6;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  margin-bottom: 1.5rem;
}

.badge-icon {
  font-size: 1rem;
}

.form-title {
  font-size: 2.2rem;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 0.5rem;
  background: linear-gradient(135deg, #1f2937, #3b82f6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.form-subtitle {
  color: #6b7280;
  font-size: 1rem;
  margin: 0;
  font-weight: 400;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 0.5rem;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 1rem;
  color: #6b7280;
  z-index: 1;
  display: flex;
  align-items: center;
}

.form-input {
  width: 100%;
  padding: 1rem 1rem 1rem 3rem;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 1rem;
  transition: all 0.3s ease;
  background: #f9fafb;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  background: white;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1);
  transform: translateY(-1px);
}

.form-input.error {
  border-color: #ef4444;
  background: #fef2f2;
}

.error-message {
  color: #ef4444;
  font-size: 0.875rem;
  margin-left: 0.5rem;
  font-weight: 500;
}

.error-alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  color: #dc2626;
  font-size: 0.875rem;
  font-weight: 500;
}

.error-icon {
  color: #ef4444;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.login-button {
  width: 100%;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.login-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(59, 130, 246, 0.4);
  background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
}

.login-button:active:not(:disabled) {
  transform: translateY(0);
}

.login-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.login-footer {
  margin-top: 2.5rem;
  text-align: center;
}

.footer-divider {
  height: 1px;
  background: linear-gradient(to right, transparent, #e5e7eb, transparent);
  margin-bottom: 1.5rem;
}

.footer-text {
  color: #6b7280;
  font-size: 0.875rem;
  margin: 0;
  font-weight: 400;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .login-info {
    padding: 2rem;
  }
  
  .login-panel {
    padding: 2rem;
  }
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
  }
  
  .login-info {
    min-height: 45vh;
    padding: 1.5rem;
  }
  
  .info-content {
    max-width: 100%;
    text-align: center;
  }
  
  .logo-text {
    font-size: 2.2rem;
  }
  
  .info-description {
    font-size: 1rem;
    margin-bottom: 2rem;
  }
  
  .feature-list {
    gap: 1rem;
  }
  
  .feature-item {
    padding: 0.75rem;
  }
  
  .login-panel {
    padding: 1.5rem;
    min-height: 55vh;
  }
  
  .form-title {
    font-size: 1.8rem;
  }
  
  .bg-circle-1 {
    width: 200px;
    height: 200px;
  }
  
  .bg-circle-2 {
    width: 150px;
    height: 150px;
  }
  
  .bg-circle-3 {
    width: 100px;
    height: 100px;
  }
}

@media (max-width: 480px) {
  .login-info {
    padding: 1rem;
  }
  
  .login-panel {
    padding: 1rem;
  }
  
  .logo-text {
    font-size: 1.8rem;
  }
  
  .form-title {
    font-size: 1.6rem;
  }
  
  .form-input {
    padding: 0.875rem 0.875rem 0.875rem 2.5rem;
  }
  
  .input-icon {
    left: 0.75rem;
  }
}
</style>
