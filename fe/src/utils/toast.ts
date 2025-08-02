import { createApp, type App } from 'vue'
import Toast, { type ToastProps } from '@/components/Toast.vue'

interface ToastInstance {
  id: string
  app: App
  container: HTMLElement
}

class ToastManager {
  private instances: ToastInstance[] = []
  private idCounter = 0

  private createToast(props: ToastProps): ToastInstance {
    const id = `toast-${++this.idCounter}`
    const container = document.createElement('div')
    container.id = id
    document.body.appendChild(container)

    const app = createApp(Toast, {
      ...props,
      onClose: () => {
        this.removeToast(id)
        props.onClose?.()
      }
    })

    app.mount(container)

    const instance: ToastInstance = {
      id,
      app,
      container
    }

    this.instances.push(instance)
    return instance
  }

  private removeToast(id: string) {
    const index = this.instances.findIndex(instance => instance.id === id)
    if (index > -1) {
      const instance = this.instances[index]
      instance.app.unmount()
      document.body.removeChild(instance.container)
      this.instances.splice(index, 1)
    }
  }

  success(message: string, options?: Partial<ToastProps>) {
    return this.createToast({
      type: 'success',
      message,
      ...options
    })
  }

  warning(message: string, options?: Partial<ToastProps>) {
    return this.createToast({
      type: 'warning',
      message,
      ...options
    })
  }

  error(message: string, options?: Partial<ToastProps>) {
    return this.createToast({
      type: 'error',
      message,
      duration: 5000, // 错误消息显示更久
      ...options
    })
  }

  info(message: string, options?: Partial<ToastProps>) {
    return this.createToast({
      type: 'info',
      message,
      ...options
    })
  }

  // 清除所有Toast
  clear() {
    this.instances.forEach(instance => {
      instance.app.unmount()
      document.body.removeChild(instance.container)
    })
    this.instances = []
  }
}

export const toast = new ToastManager()

// 为了兼容原有的alert调用
export const showToast = {
  success: (message: string, title?: string) => toast.success(message, { title }),
  warning: (message: string, title?: string) => toast.warning(message, { title }),
  error: (message: string, title?: string) => toast.error(message, { title }),
  info: (message: string, title?: string) => toast.info(message, { title })
}

// 替换原生alert的函数
export const alert = (message: string) => {
  toast.info(message)
}

export default toast