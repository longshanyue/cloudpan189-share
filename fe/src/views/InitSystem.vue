<template>
  <div class="init-system-container">
    <!-- 背景装饰 -->
    <div class="bg-decoration">
      <div class="bg-circle bg-circle-1"></div>
      <div class="bg-circle bg-circle-2"></div>
      <div class="bg-circle bg-circle-3"></div>
      <div class="bg-wave"></div>
    </div>
    
    <div class="init-card">
      <div class="init-header">
        <h1>初始化系统</h1>
        <p>欢迎使用云盘189分享管理系统，请完成系统初始化设置</p>
      </div>
      
      <form @submit.prevent="handleSubmit" class="init-form">
        <div class="form-group">
          <label for="title">网站名称</label>
          <input
            id="title"
            v-model="formData.title"
            type="text"
            placeholder="请输入网站名称"
            required
            :disabled="loading"
          />
        </div>
        
        <div class="form-group">
          <label for="baseURL">基础URL</label>
          <div class="input-group">
            <input
              id="baseURL"
              v-model="formData.baseURL"
              type="url"
              placeholder="例如：https://your-domain.com"
              required
              :disabled="loading"
            />
            <button
              type="button"
              @click="autoFillBaseURL"
              class="btn-auto-fill"
              :disabled="loading"
            >
              自动获取
            </button>
          </div>
          <small class="form-hint">请输入完整的基础URL，例如：https://your-domain.com（不要以/结尾）</small>
        </div>
        
        <div class="form-group">
          <label>WebDAV访问认证</label>
          <div class="toggle-buttons">
            <button
              type="button"
              @click="formData.enableAuth = true"
              :class="['toggle-btn', { active: formData.enableAuth }]"
              :disabled="loading"
            >
              启用
            </button>
            <button
              type="button"
              @click="formData.enableAuth = false"
              :class="['toggle-btn', { active: !formData.enableAuth }]"
              :disabled="loading"
            >
              禁用
            </button>
          </div>
          <small class="form-hint">启用后，WebDAV访问需要用户认证</small>
        </div>
        
        <div class="form-group">
          <label for="superUsername">管理员账号</label>
          <input
            id="superUsername"
            v-model="formData.superUsername"
            type="text"
            placeholder="请输入管理员账号（3-20个字符）"
            required
            minlength="3"
            maxlength="20"
            :disabled="loading"
          />
        </div>
        
        <div class="form-group">
          <label for="superPassword">管理员密码</label>
          <input
            id="superPassword"
            v-model="formData.superPassword"
            type="password"
            placeholder="请输入管理员密码（6-20个字符）"
            required
            minlength="6"
            maxlength="20"
            :disabled="loading"
          />
        </div>
        
        <div class="form-group">
          <label for="confirmPassword">确认密码</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            placeholder="请再次输入管理员密码"
            required
            :disabled="loading"
          />
          <small v-if="passwordMismatch" class="error-hint">两次输入的密码不一致</small>
        </div>
        
        <button
          type="submit"
          class="btn-submit"
          :disabled="loading || !isFormValid"
        >
          <span v-if="loading">初始化中...</span>
          <span v-else>初始化系统</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingStore } from '@/stores/setting'
import type { InitSystemRequest } from '@/api/setting'
import { toast } from '@/utils/toast'

const router = useRouter()
const settingStore = useSettingStore()

const loading = ref(false)
const confirmPassword = ref('')

const formData = ref<InitSystemRequest>({
  title: '天翼订阅小站',
  enableAuth: true,
  baseURL: '',
  superUsername: '',
  superPassword: ''
})

// 计算属性
const passwordMismatch = computed(() => {
  return confirmPassword.value && formData.value.superPassword && 
         confirmPassword.value !== formData.value.superPassword
})

const isFormValid = computed(() => {
  return formData.value.title &&
         formData.value.baseURL &&
         formData.value.superUsername &&
         formData.value.superPassword &&
         confirmPassword.value &&
         !passwordMismatch.value &&
         formData.value.superUsername.length >= 3 &&
         formData.value.superUsername.length <= 20 &&
         formData.value.superPassword.length >= 6 &&
         formData.value.superPassword.length <= 20
})

// 自动填充基础URL
const autoFillBaseURL = () => {
  let url = window.location.origin
  // 确保URL不以/结尾
  if (url.endsWith('/')) {
    url = url.slice(0, -1)
  }
  formData.value.baseURL = url
}

// 提交表单
const handleSubmit = async () => {
  if (!isFormValid.value) {
    return
  }
  
  // 确保baseURL不以/结尾
  if (formData.value.baseURL.endsWith('/')) {
    formData.value.baseURL = formData.value.baseURL.slice(0, -1)
  }
  
  loading.value = true
  
  try {
    await settingStore.initSystem(formData.value)
    
    // 显示初始化成功提示
    toast.success('系统初始化成功！跳转登录页面')
    
    // 初始化成功，直接跳转到登录页面
    // 由于已经更新了系统状态，路由守卫会允许访问登录页面
    router.push('/@login')
  } catch (error: any) {
    console.error('初始化失败:', error)
    toast.error(error.response?.data?.msg || '初始化失败，请重试')
  } finally {
    loading.value = false
  }
}

// 组件挂载时自动填充基础URL
onMounted(() => {
  autoFillBaseURL()
})
</script>

<style scoped>
.init-system-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  position: relative;
  box-sizing: border-box;
  flex: 1;
}

/* 背景装饰 */
.bg-decoration {
  position: fixed;
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

.init-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.15);
  padding: 32px;
  width: 100%;
  max-width: 480px;
  position: relative;
  z-index: 1;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.init-header {
  text-align: center;
  margin-bottom: 24px;
}

.init-header h1 {
  color: #1f2937;
  margin: 0 0 8px 0;
  font-size: 26px;
  font-weight: 700;
}

.init-header p {
  color: #6b7280;
  margin: 0;
  font-size: 14px;
  line-height: 1.4;
  font-weight: 400;
}

.init-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-weight: 500;
  color: #333;
  font-size: 14px;
}

.form-group input[type="text"],
.form-group input[type="url"],
.form-group input[type="password"] {
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  border-radius: 8px;
  font-size: 14px;
  transition: border-color 0.2s ease;
}

.form-group input[type="text"]:focus,
.form-group input[type="url"]:focus,
.form-group input[type="password"]:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1);
  transform: translateY(-1px);
}

.form-group input:disabled {
  background-color: #f5f5f5;
  cursor: not-allowed;
}

.input-group {
  display: flex;
  gap: 8px;
}

.input-group input {
  flex: 1;
}

.btn-auto-fill {
  padding: 12px 16px;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
  white-space: nowrap;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-auto-fill:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(59, 130, 246, 0.4);
  background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
}

.btn-auto-fill:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.toggle-buttons {
  display: flex;
  gap: 8px;
}

.toggle-btn {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #e1e5e9;
  background: white;
  color: #666;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.toggle-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s;
}

.toggle-btn:hover:not(:disabled) {
  border-color: #3b82f6;
  color: #3b82f6;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.toggle-btn:hover:not(:disabled)::before {
  left: 100%;
}

.toggle-btn.active {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  border-color: #3b82f6;
  color: white;
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
  transform: translateY(-2px);
}

.toggle-btn.active::before {
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
}

.toggle-btn:active:not(:disabled) {
  transform: translateY(0);
  transition: transform 0.1s;
}

.toggle-btn:disabled {
  background: #f5f5f5;
  border-color: #e1e5e9;
  color: #ccc;
  cursor: not-allowed;
  transform: none;
}

.form-hint {
  color: #666;
  font-size: 12px;
  margin-top: 4px;
}

.error-hint {
  color: #e74c3c;
  font-size: 12px;
  margin-top: 4px;
}

.btn-submit {
  padding: 14px 24px;
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-top: 10px;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-submit:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(59, 130, 246, 0.4);
  background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
}

.btn-submit:active:not(:disabled) {
  transform: translateY(0);
}

.btn-submit:disabled {
  background: #ccc;
  cursor: not-allowed;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .init-system-container {
    padding: 1rem;
  }
  
  .init-card {
    padding: 24px 20px;
    max-width: 100%;
  }
  
  .input-group {
    flex-direction: column;
  }
  
  .btn-auto-fill {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .init-system-container {
    padding: 0.5rem;
  }
  
  .init-card {
    padding: 20px 16px;
  }
  
  .init-header h1 {
    font-size: 22px;
  }
  
  .init-header p {
    font-size: 13px;
  }
}

/* 确保在小屏幕上也能正常滚动 */
@media (max-height: 700px) {
  .init-system-container {
    align-items: flex-start;
    padding-top: 2rem;
    padding-bottom: 2rem;
  }
}
</style>
