<template>
  <Layout>
    <!-- 令牌管理主卡片 -->
    <PageCard title="令牌管理" subtitle="管理云盘189的访问令牌，用于获取下载链接">
      <SectionDivider />
        
        <!-- 操作区域 -->
        <div class="action-section">
          <button class="action-btn primary" @click="handleAddToken" :disabled="loading">
            <Icons name="add" size="1rem" class="btn-icon" />
            添加新令牌
          </button>
        </div>
        
        <SectionDivider />
        
        <!-- 令牌列表 -->
        <SubsectionTitle title="令牌列表" />
        <div class="token-list" v-if="tokens.length > 0">
          <div class="token-item" v-for="token in tokens" :key="token.id">
            <div class="token-info">
              <div class="token-name">{{ token.name }}</div>
              <div class="token-meta">
                <span class="token-status" :class="getStatusClass(token.status)">{{ getStatusText(token.status) }}</span>
                <span class="token-login-type" :class="getLoginTypeClass(token.loginType)">{{ getLoginTypeText(token.loginType) }}</span>
                <span v-if="token.status === 1" class="token-expiry" :class="getExpiryClass(token.expiresIn)">{{ formatExpiry(token.expiresIn) }}</span>
                <span class="token-date">创建于 {{ formatDate(token.createdAt) }}</span>
                <span v-if="token.loginType ===2" style="color: #999; font-size: 12px">密码登录将在令牌有效期还剩7天时尝试更新</span>
              </div>
            </div>
            <div class="token-actions">
              <button class="action-btn secondary" @click="handleEditTokenName(token)" :disabled="loading">
                修改名称
              </button>
              <button class="action-btn primary" @click="handleUpdateToken(token)" :disabled="loading">
                更新令牌
              </button>
              <button class="action-btn danger" @click="handleDeleteToken(token)" :disabled="loading">
                删除
              </button>
            </div>
          </div>
        </div>
        
        <!-- 空状态 -->
        <div class="empty-state" v-else>
          <Icons name="tokens" size="3rem" class="empty-icon" />
          <div class="empty-title">暂无令牌</div>
          <div class="empty-desc">点击上方按钮添加您的第一个云盘令牌</div>
        </div>
    </PageCard>
    
    <!-- 添加/编辑令牌弹窗 -->
    <div class="modal-overlay" v-if="showModal" @click="closeModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">{{ isEditing ? '编辑令牌' : '添加令牌' }}</h3>
          <button class="modal-close" @click="closeModal">✕</button>
        </div>
        
        <div class="modal-body">
          <!-- 登录方式选择 -->
          <div class="login-type-selection" v-if="!selectedLoginType">
            <div class="login-type-title">选择登录方式</div>
            <div class="login-type-options">
              <div class="login-type-option" @click="selectLoginType(1)">
                 <div class="login-type-icon qr-icon">
                   <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                     <rect x="3" y="3" width="8" height="8" rx="1" stroke="currentColor" stroke-width="2" fill="none"/>
                     <rect x="13" y="3" width="8" height="8" rx="1" stroke="currentColor" stroke-width="2" fill="none"/>
                     <rect x="3" y="13" width="8" height="8" rx="1" stroke="currentColor" stroke-width="2" fill="none"/>
                     <rect x="5" y="5" width="4" height="4" fill="currentColor"/>
                     <rect x="15" y="5" width="4" height="4" fill="currentColor"/>
                     <rect x="5" y="15" width="4" height="4" fill="currentColor"/>
                     <rect x="13" y="13" width="2" height="2" fill="currentColor"/>
                     <rect x="17" y="13" width="2" height="2" fill="currentColor"/>
                     <rect x="19" y="15" width="2" height="2" fill="currentColor"/>
                     <rect x="15" y="17" width="2" height="2" fill="currentColor"/>
                     <rect x="13" y="19" width="2" height="2" fill="currentColor"/>
                     <rect x="17" y="19" width="2" height="2" fill="currentColor"/>
                     <rect x="19" y="19" width="2" height="2" fill="currentColor"/>
                   </svg>
                 </div>
                 <div class="login-type-name">扫码登录</div>
                 <div class="login-type-desc">使用天翼云盘APP扫描二维码</div>
               </div>
               <div class="login-type-option" @click="selectLoginType(2)">
                 <div class="login-type-icon pwd-icon">
                   <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                     <rect x="3" y="11" width="18" height="10" rx="2" ry="2" stroke="currentColor" stroke-width="2" fill="none"/>
                     <circle cx="12" cy="16" r="1" fill="currentColor"/>
                     <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
                   </svg>
                 </div>
                 <div class="login-type-name">密码登录</div>
                 <div class="login-type-desc">使用用户名和密码登录</div>
               </div>
            </div>
          </div>
          
          <!-- 二维码显示区域 -->
          <div class="qrcode-section" v-if="selectedLoginType === 1 && qrCodeData">
            <div class="qrcode-container">
              <div class="qrcode-display" v-if="qrCodeData.qrCodeUrl">
                 <img :src="qrCodeData.qrCodeUrl" alt="二维码" class="qrcode-image" />
                 <div class="qrcode-text">请使用天翼云盘APP扫描二维码</div>
                 <div v-if="scanTimeLeft > 0 && scanTimeLeft < 120" class="qrcode-countdown">
                   请在 {{ Math.floor(scanTimeLeft / 60) }}分{{ scanTimeLeft % 60 }}秒 内完成扫码
                 </div>
                 <div class="qrcode-uuid">{{ qrCodeData.uuid }}</div>
               </div>
              <div class="qrcode-placeholder" v-else>
                <div class="qrcode-icon">
                  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <rect x="2" y="3" width="20" height="14" rx="2" stroke="currentColor" stroke-width="2" fill="none"/>
                    <circle cx="12" cy="10" r="3" stroke="currentColor" stroke-width="2" fill="none"/>
                    <path d="M12 21l-3-3h6l-3 3z" fill="currentColor"/>
                  </svg>
                </div>
                <div class="qrcode-text">正在生成二维码...</div>
                <div class="qrcode-uuid">{{ qrCodeData.uuid }}</div>
              </div>
            </div>
            <div class="qrcode-tips">
              <p>扫描步骤：</p>
              <ol>
                <li>打开天翼云盘APP</li>
                <li>点击右上角扫一扫</li>
                <li>扫描上方二维码</li>
                <li>确认授权登录</li>
              </ol>
            </div>
            <div class="qrcode-status" v-if="isScanning">
              <div class="status-indicator">
                <div class="spinner"></div>
                <span>正在检测扫码状态... ({{ scanTimeLeft }}秒)</span>
              </div>
            </div>
          </div>
          
          <!-- 扫码登录未生成二维码时的提示 -->
          <div v-if="selectedLoginType === 1 && !qrCodeData" class="qrcode-placeholder-empty">
            <div class="empty-qr-icon">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <rect x="3" y="3" width="8" height="8" rx="1" stroke="currentColor" stroke-width="2" fill="none"/>
                <rect x="13" y="3" width="8" height="8" rx="1" stroke="currentColor" stroke-width="2" fill="none"/>
                <rect x="3" y="13" width="8" height="8" rx="1" stroke="currentColor" stroke-width="2" fill="none"/>
                <rect x="5" y="5" width="4" height="4" fill="currentColor"/>
                <rect x="15" y="5" width="4" height="4" fill="currentColor"/>
                <rect x="5" y="15" width="4" height="4" fill="currentColor"/>
                <rect x="13" y="13" width="2" height="2" fill="currentColor"/>
                <rect x="17" y="13" width="2" height="2" fill="currentColor"/>
                <rect x="19" y="15" width="2" height="2" fill="currentColor"/>
                <rect x="15" y="17" width="2" height="2" fill="currentColor"/>
                <rect x="13" y="19" width="2" height="2" fill="currentColor"/>
                <rect x="17" y="19" width="2" height="2" fill="currentColor"/>
                <rect x="19" y="19" width="2" height="2" fill="currentColor"/>
              </svg>
            </div>
            <div class="empty-qr-text">{{ isEditing ? '重新生成二维码进行授权' : '点击下方按钮生成二维码开始授权' }}</div>
          </div>
          
          <!-- 密码登录表单 -->
          <div class="password-login-section" v-if="selectedLoginType === 2 || (isEditing && showPasswordForm)">
            <div v-if="formReady" class="password-form">
              <div class="form-group">
                <label class="form-label">用户名</label>
                <input 
                  v-model="username" 
                  type="text" 
                  placeholder="请输入天翼云盘用户名"
                  class="form-input"
                  autocomplete="off"
                  :name="'username_' + randomId"
                  readonly
                  @focus="removeReadonly"
                />
              </div>
              <div class="form-group">
                <label class="form-label">密码</label>
                <input 
                  v-model="password" 
                  type="password" 
                  placeholder="请输入密码"
                  class="form-input"
                  autocomplete="new-password"
                  :name="'password_' + randomId"
                  readonly
                  @focus="removeReadonly"
                />
              </div>
            </div>
            <div v-else class="form-loading">
              <div class="form-loading-text">正在准备表单...</div>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button class="modal-btn secondary" @click="closeModal" :disabled="loading">
            取消
          </button>
          
          <!-- 返回按钮（在选择了登录方式后显示） -->
          <button 
            v-if="selectedLoginType && !isEditing" 
            class="modal-btn secondary" 
            @click="goBack" 
            :disabled="loading"
          >
            返回
          </button>
          
          <!-- 扫码登录按钮 -->
          <button 
            v-if="selectedLoginType === 1 || (isEditing && !showPasswordForm)" 
            class="modal-btn primary" 
            @click="qrCodeData ? (isScanning ? stopQrcodeCheck() : startQrcodeCheck()) : handleGenerateQrcode()" 
            :disabled="loading"
          >
            {{ loading ? '处理中...' : (qrCodeData ? (isScanning ? '停止检测' : '我已扫码') : '开始扫码') }}
          </button>
          
          <!-- 密码登录按钮 -->
          <button 
            v-if="(selectedLoginType === 2 || (isEditing && showPasswordForm)) && formReady" 
            class="modal-btn primary" 
            @click="handlePasswordLogin" 
            :disabled="loading || !username.trim() || !password.trim()"
          >
            {{ loading ? '登录中...' : (isEditing ? '更新令牌' : '添加令牌') }}
          </button>
        </div>
      </div>
    </div>
    
    <!-- 修改名称弹窗 -->
    <div class="modal-overlay" v-if="showNameModal" @click="closeNameModal">
      <div class="modal-content small" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">修改令牌名称</h3>
          <button class="modal-close" @click="closeNameModal">✕</button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">令牌名称</label>
            <input 
              v-model="newTokenName" 
              type="text" 
              placeholder="请输入新的令牌名称"
              class="form-input"
              maxlength="50"
            />
          </div>
        </div>
        
        <div class="modal-footer">
          <button class="modal-btn secondary" @click="closeNameModal" :disabled="loading">
            取消
          </button>
          <button class="modal-btn primary" @click="confirmNameEdit" :disabled="loading || !newTokenName.trim()">
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 确认删除弹窗 -->
    <div class="modal-overlay" v-if="showDeleteModal" @click="closeDeleteModal">
      <div class="modal-content small" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">确认删除</h3>
          <button class="modal-close" @click="closeDeleteModal">✕</button>
        </div>
        
        <div class="modal-body">
          <p>确定要删除令牌 <strong>{{ deleteTarget?.name }}</strong> 吗？</p>
          <p class="warning-text">此操作不可撤销，删除后相关的存储配置可能会失效。</p>
        </div>
        
        <div class="modal-footer">
          <button class="modal-btn secondary" @click="closeDeleteModal" :disabled="loading">
            取消
          </button>
          <button class="modal-btn danger" @click="confirmDelete" :disabled="loading">
            {{ loading ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { cloudTokenApi, type CloudToken } from '@/api/cloudtoken'
import Layout from '@/components/Layout.vue'
import Icons from '@/components/Icons.vue'
import PageCard from '@/components/PageCard.vue'
import SectionDivider from '@/components/SectionDivider.vue'
import SubsectionTitle from '@/components/SubsectionTitle.vue'
import QRCode from 'qrcode'
import { toast } from '@/utils/toast'

// 响应式数据
const tokens = ref<CloudToken[]>([])
const loading = ref(false)
const showModal = ref(false)
const showDeleteModal = ref(false)
const showNameModal = ref(false)
const isEditing = ref(false)
const qrCodeData = ref<{ uuid: string; qrCodeUrl?: string } | null>(null)
const deleteTarget = ref<CloudToken | null>(null)
const editingToken = ref<CloudToken | null>(null)
const newTokenName = ref('')
const checkInterval = ref<NodeJS.Timeout | null>(null)
const isScanning = ref(false)
const scanTimeLeft = ref(120)
const scanTimer = ref<NodeJS.Timeout | null>(null)
const editingTokenId = ref<number | null>(null)

// 表单相关数据
const selectedLoginType = ref<number | null>(null) // 1: 扫码登录, 2: 密码登录
const username = ref('')
const password = ref('')
const showPasswordForm = ref(false)

// 防止自动填充的相关数据
const formReady = ref(false)
const randomId = ref(Math.random().toString(36).substr(2, 9))

// 移除readonly属性的方法
const removeReadonly = (event: Event) => {
  const target = event.target as HTMLInputElement
  target.removeAttribute('readonly')
}

// 获取令牌列表
const fetchTokens = async () => {
  try {
    loading.value = true
    tokens.value = await cloudTokenApi.list()
  } catch (error) {
    console.error('获取令牌列表失败:', error)
    toast.error('获取令牌列表失败')
  } finally {
    loading.value = false
  }
}

// 添加令牌
const handleAddToken = () => {
  isEditing.value = false
  editingTokenId.value = null
  selectedLoginType.value = null
  username.value = ''
  password.value = ''
  qrCodeData.value = null
  isScanning.value = false
  scanTimeLeft.value = 120
  showPasswordForm.value = false
  formReady.value = false
  randomId.value = Math.random().toString(36).substr(2, 9)
  showModal.value = true
}

// 修改令牌名称
const handleEditTokenName = (token: CloudToken) => {
  editingToken.value = token
  newTokenName.value = token.name
  showNameModal.value = true
}

// 更新令牌
const handleUpdateToken = (token: CloudToken) => {
  isEditing.value = true
  editingTokenId.value = token.id
  editingToken.value = token
  password.value = ''
  qrCodeData.value = null
  isScanning.value = false
  scanTimeLeft.value = 120
  formReady.value = false
  randomId.value = Math.random().toString(36).substr(2, 9)
  
  // 根据令牌的登录类型直接设置对应的显示方式
  if (token.loginType === 1) {
    // 扫码登录
    selectedLoginType.value = 1
    showPasswordForm.value = false
    username.value = ''
  } else {
    // 密码登录 - 自动填充用户名
    selectedLoginType.value = 2
    showPasswordForm.value = true
    username.value = token.username || '' // 从令牌数据中获取用户名
    
    // 延迟显示表单，避免自动填充
    nextTick(() => {
      setTimeout(() => {
        formReady.value = true
      }, 300)
    })
  }
  
  showModal.value = true
}

// 选择登录类型
const selectLoginType = (type: number) => {
  selectedLoginType.value = type
  
  if (type === 2) {
    // 选择密码登录时，延迟显示表单
    formReady.value = false
    randomId.value = Math.random().toString(36).substr(2, 9)
    
    nextTick(() => {
      setTimeout(() => {
        formReady.value = true
      }, 300)
    })
  }
}

// 返回登录方式选择
const goBack = () => {
  selectedLoginType.value = null
  qrCodeData.value = null
  stopQrcodeCheck()
  username.value = ''
  password.value = ''
  formReady.value = false
}

// 密码登录
const handlePasswordLogin = async () => {
  try {
    loading.value = true
    
    const loginData: any = {
      username: username.value.trim(),
      password: password.value.trim()
    }
    
    // 如果是编辑模式，传递令牌ID
    if (isEditing.value && editingTokenId.value) {
      loginData.id = editingTokenId.value
    }
    
    await cloudTokenApi.usernameLogin(loginData)
    
    toast.success(isEditing.value ? '令牌更新成功' : '令牌添加成功')
    closeModal()
    await fetchTokens()
  } catch (error: any) {
    console.error('密码登录失败:', error)
    toast.error(error?.msg || '用户名或密码错误')
  } finally {
    loading.value = false
  }
}

// 生成二维码
const handleGenerateQrcode = async () => {
  try {
    loading.value = true
    const response = await cloudTokenApi.initQrcode()
    qrCodeData.value = { uuid: response.uuid }
    
    // 生成二维码图片
    await generateQrCodeImage(response.uuid)
    
    // 二维码生成成功后立即开始倒计时
    startCountdown()
  } catch (error) {
    console.error('生成二维码失败:', error)
    toast.error('生成二维码失败')
  } finally {
    loading.value = false
  }
}

// 生成二维码图片
const generateQrCodeImage = async (uuid: string) => {
  try {
    const qrCodeUrl = await QRCode.toDataURL(uuid, {
      width: 200,
      margin: 2,
      color: {
        dark: '#000000',
        light: '#FFFFFF'
      }
    })
    
    if (qrCodeData.value) {
      qrCodeData.value.qrCodeUrl = qrCodeUrl
    }
  } catch (error) {
    console.error('生成二维码图片失败:', error)
    toast.error('生成二维码图片失败')
  }
}

// 开始倒计时
const startCountdown = () => {
  if (scanTimer.value) {
    clearInterval(scanTimer.value)
  }
  
  scanTimeLeft.value = 120
  
  // 倒计时定时器
  scanTimer.value = setInterval(() => {
    scanTimeLeft.value--
    if (scanTimeLeft.value <= 0) {
      stopQrcodeCheck()
      toast.error('扫码超时，请重新生成二维码')
      qrCodeData.value = null
    }
  }, 1000)
}

// 开始二维码检查
const startQrcodeCheck = () => {
  if (!qrCodeData.value) {
    toast.error('请先生成二维码')
    return
  }
  
  if (checkInterval.value) {
    clearInterval(checkInterval.value)
  }
  
  isScanning.value = true
  
  // 检查二维码状态
  checkInterval.value = setInterval(async () => {
    if (!qrCodeData.value) return
    
    try {
      const checkData: any = {
        uuid: qrCodeData.value.uuid
      }
      
      // 如果是编辑模式，传递令牌ID
      if (isEditing.value && editingTokenId.value) {
        checkData.id = editingTokenId.value
      }
      
      const res = await cloudTokenApi.checkQrcode(checkData)
      if (res.code !== 200) {
        return
      }
      
      // 扫码成功
      stopQrcodeCheck()
      toast.success(isEditing.value ? '令牌更新成功' : '令牌添加成功')
      closeModal()
      await fetchTokens()
    } catch (error: any) {
      // 如果是扫码未完成，继续等待
      if (error?.message && (error.message.includes('登录查询失败') || error.message.includes('二维码未扫描'))) {
        // 继续等待扫码
        console.log('等待用户扫码...')
      } else {
        console.error('检查二维码状态失败:', error)
        stopQrcodeCheck()
        toast.error('扫码验证失败: ' + (error?.message || '未知错误'))
      }
    }
  }, 3000)
}

// 停止二维码检查
const stopQrcodeCheck = () => {
  if (checkInterval.value) {
    clearInterval(checkInterval.value)
    checkInterval.value = null
  }
  
  if (scanTimer.value) {
    clearInterval(scanTimer.value)
    scanTimer.value = null
  }
  
  isScanning.value = false
  loading.value = false
}

// 删除令牌
const handleDeleteToken = (token: CloudToken) => {
  deleteTarget.value = token
  showDeleteModal.value = true
}

// 确认删除
const confirmDelete = async () => {
  if (!deleteTarget.value) return
  
  try {
    loading.value = true
    await cloudTokenApi.delete({ id: deleteTarget.value.id })
    
    toast.success('令牌删除成功')
    closeDeleteModal()
    await fetchTokens()
  } catch (error) {
    console.error('删除令牌失败:', error)
    toast.error('删除令牌失败')
  } finally {
    loading.value = false
  }
}

// 关闭弹窗
const closeModal = () => {
  showModal.value = false
  stopQrcodeCheck()
  qrCodeData.value = null
  isEditing.value = false
  editingTokenId.value = null
  editingToken.value = null
  selectedLoginType.value = null
  username.value = ''
  password.value = ''
  showPasswordForm.value = false
  formReady.value = false
  scanTimeLeft.value = 120
  randomId.value = Math.random().toString(36).substr(2, 9)
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  deleteTarget.value = null
}

const closeNameModal = () => {
  showNameModal.value = false
  editingToken.value = null
  newTokenName.value = ''
}

// 确认修改名称
const confirmNameEdit = async () => {
  if (!editingToken.value || !newTokenName.value.trim()) return
  
  try {
    loading.value = true
    await cloudTokenApi.modifyName({
       id: editingToken.value.id,
       name: newTokenName.value.trim()
     })
    
    toast.success('令牌名称修改成功')
    closeNameModal()
    await fetchTokens()
  } catch (error) {
    console.error('修改令牌名称失败:', error)
    toast.error('修改令牌名称失败')
  } finally {
    loading.value = false
  }
}

// 工具函数
const getStatusClass = (status: number) => {
  return status === 1 ? 'active' : 'inactive'
}

const getStatusText = (status: number) => {
  return status === 1 ? '正常' : '异常'
}

const getLoginTypeText = (loginType: number) => {
  return loginType === 1 ? '扫码登录' : '密码登录'
}

const getLoginTypeClass = (loginType: number) => {
  return loginType === 1 ? 'login-type-qrcode' : 'login-type-password'
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatExpiry = (expiresIn: number) => {
  if (!expiresIn || expiresIn <= 0) return '永久有效'
  
  // 计算过期时间：更新时间 + 有效期毫秒数
  const expiryDate = new Date(expiresIn)
  const now = new Date()
  const diffTime = expiryDate.getTime() - now.getTime()
  
  if (diffTime <= 0) {
    return '已过期'
  }
  
  // 计算剩余时间
  const diffSeconds = Math.floor(diffTime / 1000)
  const days = Math.floor(diffSeconds / (24 * 60 * 60))
  const hours = Math.floor((diffSeconds % (24 * 60 * 60)) / (60 * 60))
  const minutes = Math.floor((diffSeconds % (60 * 60)) / 60)
  
  let result = '剩余'
  if (days > 0) {
    result += `${days}天`
  }
  if (hours > 0) {
    result += `${hours}小时`
  }
  if (minutes > 0 && days === 0) { // 只有在天数为0时才显示分钟
    result += `${minutes}分`
  }
  
  return result
}

const getExpiryClass = (expiresIn: number) => {
  if (!expiresIn || expiresIn <= 0) return 'expiry-permanent'
  
  const expiryDate = new Date(expiresIn)
  const now = new Date()
  const diffTime = expiryDate.getTime() - now.getTime()
  
  if (diffTime <= 0) {
    return 'expiry-expired'
  }
  
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return 'expiry-today'
  } else if (diffDays <= 7) {
    return 'expiry-warning'
  } else {
    return 'expiry-normal'
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchTokens()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  stopQrcodeCheck()
})
</script>

<style scoped>
/* 操作区域 */
.action-section {
  display: flex;
  justify-content: flex-start;
  margin: 1rem 0;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
  text-decoration: none;
}

.action-btn.primary {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}

.action-btn.primary:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.action-btn.secondary {
  background: #f8fafc;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.action-btn.secondary:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.action-btn.danger {
  background: #fee2e2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.action-btn.danger:hover {
  background: #fecaca;
  border-color: #f87171;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.btn-icon {
  font-size: 1rem;
}

/* 令牌列表 */
.token-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.token-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  background: linear-gradient(135deg, #fafbfc 0%, #f8fafc 100%);
  border-radius: 12px;
  border: 1px solid #f1f5f9;
  transition: all 0.2s ease;
}

.token-item:hover {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.token-info {
  flex: 1;
}

.token-name {
  font-size: 1rem;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 0.5rem;
}

.token-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.875rem;
}

.token-status {
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.75rem;
}

.token-status.active {
  background: #dcfce7;
  color: #166534;
}

.token-status.inactive {
  background: #fee2e2;
  color: #dc2626;
}

.token-login-type {
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
  font-weight: 500;
  font-size: 0.75rem;
}

.token-login-type.login-type-qrcode {
  background: #e0f2fe;
  color: #0369a1;
}

.token-login-type.login-type-password {
  background: #fef3c7;
  color: #d97706;
}

.token-date {
  color: #6b7280;
}

.token-expiry {
  font-weight: 500;
  font-size: 0.875rem;
}

.token-expiry.expiry-normal {
  color: #059669;
}

.token-expiry.expiry-warning {
  color: #f59e0b;
}

.token-expiry.expiry-today {
  color: #ef4444;
  font-weight: 600;
}

.token-expiry.expiry-expired {
  color: #dc2626;
  font-weight: 600;
}

.token-expiry.expiry-permanent {
  color: #6366f1;
}

.token-actions {
  display: flex;
  gap: 0.5rem;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #6b7280;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.empty-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 0.5rem;
}

.empty-desc {
  font-size: 0.875rem;
}

/* 登录方式选择样式 */
.login-type-selection {
  text-align: center;
  padding: 1rem 0;
}

.login-type-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 1.5rem;
}

.login-type-options {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.login-type-option {
  flex: 1;
  max-width: 200px;
  padding: 1.5rem 1rem;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fafbfc;
}

.login-type-option:hover {
  border-color: #3b82f6;
  background: #f8fafc;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
}

.login-type-icon {
  font-size: 2rem;
  margin-bottom: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  border-radius: 8px;
  background: #f5f5f5;
  margin: 0 auto 0.75rem auto;
}

.qr-icon {
  background: #3b82f6;
  color: white;
}

.pwd-icon {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.empty-qr-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
  color: #9ca3af;
}

.qrcode-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.5rem;
  color: #6b7280;
}

.login-type-name {
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 0.5rem;
}

.login-type-desc {
  font-size: 0.875rem;
  color: #6b7280;
  line-height: 1.4;
}

.password-login-section {
  padding: 1rem 0;
}

.password-form {
  animation: fadeIn 0.3s ease-in-out;
}

.form-loading {
  text-align: center;
  padding: 2rem 0;
}

.form-loading-text {
  color: #6b7280;
  font-size: 0.875rem;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.modal-content.small {
  max-width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #f1f5f9;
}

.modal-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #6b7280;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.modal-close:hover {
  background: #f3f4f6;
  color: #374151;
}

.modal-body {
  padding: 1.5rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.5rem;
  border-top: 1px solid #f1f5f9;
}

.form-group {
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #374151;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s ease;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input::placeholder {
  color: #9ca3af;
}

.form-input[readonly] {
  background-color: #f9fafb;
  cursor: pointer;
}

.form-input[readonly]:focus {
  background-color: white;
}

.modal-btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.modal-btn.primary {
  background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
  color: white;
}

.modal-btn.primary:hover {
  background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
}

.modal-btn.secondary {
  background: #f8fafc;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.modal-btn.secondary:hover {
  background: #f1f5f9;
}

.modal-btn.danger {
  background: #dc2626;
  color: white;
}

.modal-btn.danger:hover {
  background: #b91c1c;
}

.modal-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 二维码区域 */
.qrcode-section {
  text-align: center;
}

.qrcode-container {
  margin-bottom: 1.5rem;
}

.qrcode-placeholder {
  width: 200px;
  height: 200px;
  margin: 0 auto;
  border: 2px dashed #d1d5db;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f9fafb;
}

.qrcode-display {
  text-align: center;
  margin: 0 auto;
}

.qrcode-image {
  width: 200px;
  height: 200px;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  margin-bottom: 1rem;
  background: white;
}

.qrcode-icon {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.qrcode-text {
  font-size: 0.875rem;
  color: #6b7280;
  margin-bottom: 0.5rem;
}

.qrcode-countdown {
  color: #ef4444;
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  padding: 0.5rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.qrcode-uuid {
  font-size: 0.75rem;
  color: #9ca3af;
  font-family: monospace;
  word-break: break-all;
  padding: 0 1rem;
}

.qrcode-tips {
  text-align: left;
  background: #f0f9ff;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #e0f2fe;
}

.qrcode-tips p {
  margin: 0 0 0.5rem 0;
  font-weight: 600;
  color: #0369a1;
}

.qrcode-tips ol {
  margin: 0;
  padding-left: 1.5rem;
  color: #0369a1;
}

.qrcode-tips li {
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
}

/* 扫码状态指示器 */
.qrcode-status {
  margin-top: 1rem;
  padding: 1rem;
  background: #f0f9ff;
  border-radius: 8px;
  border: 1px solid #e0f2fe;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: #0369a1;
  font-weight: 500;
}

.spinner {
  width: 1rem;
  height: 1rem;
  border: 2px solid #e0f2fe;
  border-top: 2px solid #0369a1;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 空状态二维码占位符 */
.qrcode-placeholder-empty {
  text-align: center;
  padding: 3rem 1rem;
  color: #6b7280;
}

.empty-qr-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.empty-qr-text {
  font-size: 1rem;
  color: #374151;
}

.warning-text {
  color: #dc2626;
  font-size: 0.875rem;
  margin-top: 0.5rem;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .token-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .token-actions {
    width: 100%;
    justify-content: flex-end;
  }
  
  .modal-content {
    width: 95%;
    margin: 1rem;
  }
}

@media (max-width: 480px) {
  .token-actions {
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .action-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
