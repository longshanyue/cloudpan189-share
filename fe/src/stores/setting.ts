// stores/setting.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { settingApi, type Setting, type InitSystemRequest } from '@/api/setting'

export const useSettingStore = defineStore('setting', () => {
  // 状态
  const setting = ref<Setting | null>(null)
  const loading = ref(false)

  // 获取设置
  const fetchSetting = async () => {
    loading.value = true
    try {
      const data = await settingApi.getSetting()
      setting.value = data
      return data
    } catch (error) {
      console.error('获取设置失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 修改网站名称
  const modifyName = async (name: string) => {
    try {
      await settingApi.modifyName({ name })
      if (setting.value) {
        setting.value.title = name
      }
    } catch (error) {
      console.error('修改网站名称失败:', error)
      throw error
    }
  }

  // 修改基础URL
  const modifyBaseURL = async (baseURL: string) => {
    try {
      await settingApi.modifyBaseURL({ baseURL })
      if (setting.value) {
        setting.value.baseURL = baseURL
      }
    } catch (error) {
      console.error('修改基础URL失败:', error)
      throw error
    }
  }

  // 切换认证状态
  const toggleAuth = async (disable: boolean) => {
    try {
      await settingApi.toggleAuth({ disable })
      if (setting.value) {
        setting.value.enableAuth = !disable
      }
    } catch (error) {
      console.error('切换认证状态失败:', error)
      throw error
    }
  }

  // 刷新密钥
  const refreshKey = async () => {
    try {
      await settingApi.refreshKey()
    } catch (error) {
      console.error('刷新密钥失败:', error)
      throw error
    }
  }

  // 切换本地代理状态
  const toggleLocalProxy = async (disable: boolean) => {
    try {
      await settingApi.toggleLocalProxy({ disable })
      if (setting.value) {
        setting.value.localProxy = !disable
      }
    } catch (error) {
      console.error('切换本地代理失败:', error)
      throw error
    }
  }

  // 切换多线程流式下载状态
  const toggleMultipleStream = async (disable: boolean) => {
    try {
      await settingApi.toggleMultipleStream({ disable })
      if (setting.value) {
        setting.value.multipleStream = !disable
      }
    } catch (error) {
      console.error('切换多线程流式下载失败:', error)
      throw error
    }
  }

  // 切换挂载文件自动刷新状态
  const toggleEnableTopFileAutoRefresh = async (disable: boolean) => {
    try {
      await settingApi.toggleEnableTopFileAutoRefresh({ disable })
      if (setting.value) {
        setting.value.enableTopFileAutoRefresh = !disable
      }
    } catch (error) {
      console.error('切换挂载文件自动刷新状态失败:', error)
      throw error
    }
  }

  // 修改任务线程数
  const modifyJobThreadCount = async (threadCount: number) => {
    try {
      await settingApi.modifyJobThreadCount({ threadCount })
      if (setting.value) {
        setting.value.jobThreadCount = threadCount
      }
    } catch (error) {
      console.error('修改任务线程数失败:', error)
      throw error
    }
  }

  // 修改自动刷新间隔
  const modifyAutoRefreshMinutes = async (autoRefreshMinutes: number) => {
    try {
      await settingApi.modifyAutoRefreshMinutes({ autoRefreshMinutes })
      if (setting.value) {
        setting.value.autoRefreshMinutes = autoRefreshMinutes
      }
    } catch (error) {
      console.error('修改自动刷新间隔失败:', error)
      throw error
    }
  }

  // 修改多线程流线程数
  const modifyMultipleStreamThreadCount = async (threadCount: number) => {
    try {
      await settingApi.modifyMultipleStreamThreadCount({
        multipleStreamThreadCount: threadCount
      })
      if (setting.value) {
        setting.value.multipleStreamThreadCount = threadCount
      }
    } catch (error) {
      console.error('修改多线程流线程数失败:', error)
      throw error
    }
  }

  // 修改多线程流块大小
  const modifyMultipleStreamChunkSize = async (multipleStreamChunkSize: number) => {
    try {
      await settingApi.modifyMultipleStreamChunkSize({
        multipleStreamChunkSize: multipleStreamChunkSize
      })
      if (setting.value) {
        setting.value.multipleStreamChunkSize = multipleStreamChunkSize
      }
    } catch (error) {
      console.error('修改多线程流块大小失败:', error)
      throw error
    }
  }

  // 新增：切换STRM文件启用状态
  const toggleStrmFileEnable = async (enable: boolean) => {
    try {
      await settingApi.toggleStrmFileEnable({
        strmFileEnable: enable
      })
      if (setting.value) {
        setting.value.strmFileEnable = enable
      }
    } catch (error) {
      console.error('切换STRM文件启用状态失败:', error)
      throw error
    }
  }

  // 新增：修改STRM支持文件扩展名列表
  const modifyStrmSupportFileExtList = async (extList: string[]) => {
    try {
      await settingApi.modifyStrmSupportFileExtList({
        strmSupportFileExtList: extList
      })
      if (setting.value) {
        setting.value.strmSupportFileExtList = extList
      }
    } catch (error) {
      console.error('修改STRM支持文件扩展名列表失败:', error)
      throw error
    }
  }

  // 初始化系统
  const initSystem = async (data: InitSystemRequest) => {
    try {
      await settingApi.initSystem(data)
      // 初始化成功后重新获取设置信息
      await fetchSetting()
    } catch (error) {
      console.error('初始化系统失败:', error)
      throw error
    }
  }

  return {
    setting,
    loading,
    fetchSetting,
    modifyName,
    modifyBaseURL,
    toggleAuth,
    refreshKey,
    toggleLocalProxy,
    toggleMultipleStream,
    toggleEnableTopFileAutoRefresh,
    modifyJobThreadCount,
    modifyAutoRefreshMinutes,
    modifyMultipleStreamThreadCount,
    modifyMultipleStreamChunkSize,
    toggleStrmFileEnable,
    modifyStrmSupportFileExtList,
    initSystem
  }
})