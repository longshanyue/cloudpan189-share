import { createApp, ref } from 'vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

interface ConfirmOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  isDanger?: boolean
}

export const useConfirm = () => {
  const confirm = (options: ConfirmOptions): Promise<boolean> => {
    return new Promise((resolve) => {
      // 创建容器
      const container = document.createElement('div')
      document.body.appendChild(container)
      
      // 创建Vue应用实例
      const app = createApp(ConfirmDialog, {
        ...options,
        onConfirm: () => {
          cleanup()
          resolve(true)
        },
        onCancel: () => {
          cleanup()
          resolve(false)
        }
      })
      
      // 挂载组件
      const instance = app.mount(container)
      
      // 显示对话框
      ;(instance as any).show()
      
      // 清理函数
      const cleanup = () => {
        setTimeout(() => {
          app.unmount()
          document.body.removeChild(container)
        }, 200) // 等待动画完成
      }
    })
  }
  
  return { confirm }
}

// 全局确认函数
export const confirmDialog = (options: ConfirmOptions): Promise<boolean> => {
  const { confirm } = useConfirm()
  return confirm(options)
}