<template>
  <Teleport to="body">
    <Transition name="toast" appear>
      <div v-if="visible" class="toast-container" :class="`toast-${type}`">
        <div class="toast-content">
          <div class="toast-icon">
            <Icons :name="iconName" size="1.2rem" />
          </div>
          <div class="toast-message">
            <div v-if="title" class="toast-title">{{ title }}</div>
            <div class="toast-text">{{ message }}</div>
          </div>
          <button v-if="closable" @click="close" class="toast-close">
            <Icons name="close" size="1rem" />
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icons from './Icons.vue'

export interface ToastProps {
  type?: 'success' | 'warning' | 'error' | 'info'
  title?: string
  message: string
  duration?: number
  closable?: boolean
  onClose?: () => void
}

const props = withDefaults(defineProps<ToastProps>(), {
  type: 'info',
  duration: 3000,
  closable: true
})

const visible = ref(true)

const iconName = computed(() => {
  switch (props.type) {
    case 'success':
      return 'check-circle'
    case 'warning':
      return 'warning'
    case 'error':
      return 'alert'
    case 'info':
    default:
      return 'info'
  }
})

const close = () => {
  visible.value = false
  setTimeout(() => {
    props.onClose?.()
  }, 300)
}

onMounted(() => {
  if (props.duration > 0) {
    setTimeout(() => {
      close()
    }, props.duration)
  }
})
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  max-width: 400px;
  min-width: 300px;
}

.toast-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  background: white;
  border-radius: 0.75rem;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  border-left: 4px solid;
}

.toast-success .toast-content {
  border-left-color: #10b981;
}

.toast-warning .toast-content {
  border-left-color: #f59e0b;
}

.toast-error .toast-content {
  border-left-color: #ef4444;
}

.toast-info .toast-content {
  border-left-color: #3b82f6;
}

.toast-icon {
  flex-shrink: 0;
}

.toast-success .toast-icon {
  color: #10b981;
}

.toast-warning .toast-icon {
  color: #f59e0b;
}

.toast-error .toast-icon {
  color: #ef4444;
}

.toast-info .toast-icon {
  color: #3b82f6;
}

.toast-message {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
}

.toast-text {
  color: #6b7280;
  font-size: 0.875rem;
  line-height: 1.4;
  word-break: break-word;
}

.toast-close {
  flex-shrink: 0;
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  transition: all 0.2s ease;
}

.toast-close:hover {
  background: #f3f4f6;
  color: #6b7280;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100%) scale(0.95);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100%) scale(0.95);
}

@media (max-width: 480px) {
  .toast-container {
    top: 10px;
    right: 10px;
    left: 10px;
    max-width: none;
    min-width: auto;
  }
  
  .toast-content {
    padding: 0.875rem 1rem;
  }
  
  .toast-title,
  .toast-text {
    font-size: 0.8rem;
  }
}
</style>