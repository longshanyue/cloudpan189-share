<template>
  <div v-if="visible" class="confirm-overlay" @click="handleOverlayClick">
    <div class="confirm-dialog" @click.stop>
      <div class="confirm-header">
        <Icons name="alert-triangle" size="1.5rem" class="warning-icon" />
        <h3 class="confirm-title">{{ title }}</h3>
      </div>
      
      <div class="confirm-content">
        <p class="confirm-message">{{ message }}</p>
      </div>
      
      <div class="confirm-actions">
        <button @click="handleCancel" class="btn-cancel">
          {{ cancelText }}
        </button>
        <button @click="handleConfirm" class="btn-confirm" :class="{ 'btn-danger': isDanger }">
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import Icons from './Icons.vue'

interface Props {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  isDanger?: boolean
}

withDefaults(defineProps<Props>(), {
  title: '确认操作',
  confirmText: '确定',
  cancelText: '取消',
  isDanger: false
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const visible = ref(false)

const show = () => {
  visible.value = true
}

const hide = () => {
  visible.value = false
}

const handleConfirm = () => {
  emit('confirm')
  hide()
}

const handleCancel = () => {
  emit('cancel')
  hide()
}

const handleOverlayClick = () => {
  handleCancel()
}

// 监听弹窗显示状态，控制背景滚动
watch(visible, (isVisible) => {
  if (isVisible) {
    // 禁用背景滚动
    document.body.style.overflow = 'hidden'
  } else {
    // 恢复背景滚动
    document.body.style.overflow = ''
  }
})

defineExpose({
  show,
  hide
})
</script>

<style scoped>
.confirm-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(4px);
}

.confirm-dialog {
  background: white;
  border-radius: 1rem;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 400px;
  width: 90%;
  margin: 1rem;
  animation: slideIn 0.2s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.confirm-header {
  padding: 1.5rem 1.5rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.warning-icon {
  color: #f59e0b;
  flex-shrink: 0;
}

.confirm-title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #111827;
}

.confirm-content {
  padding: 0 1.5rem 1.5rem;
}

.confirm-message {
  margin: 0;
  color: #6b7280;
  line-height: 1.5;
}

.confirm-actions {
  padding: 1rem 1.5rem 1.5rem;
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.btn-cancel,
.btn-confirm {
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  border: none;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 80px;
}

.btn-cancel {
  background: #f3f4f6;
  color: #374151;
}

.btn-cancel:hover {
  background: #e5e7eb;
}

.btn-confirm {
  background: #3b82f6;
  color: white;
}

.btn-confirm:hover {
  background: #2563eb;
}

.btn-confirm.btn-danger {
  background: #ef4444;
}

.btn-confirm.btn-danger:hover {
  background: #dc2626;
}
</style>