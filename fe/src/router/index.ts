import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSettingStore } from '@/stores/setting'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'FileBrowser',
      component: () => import('@/views/FileBrowser.vue'),
      meta: {
        requiresAuth: true,
        title: '文件浏览'
      }
    },
    {
      path: '/@init',
      name: 'InitSystem',
      component: () => import('@/views/InitSystem.vue'),
      meta: {
        requiresAuth: false,
        requiresInit: false,
        title: '系统初始化'
      }
    },
    {
      path: '/@login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: {
        requiresAuth: false,
        title: '登录'
      }
    },
    {
      path: '/@admin',
      name: 'Admin',
      component: () => import('@/components/Layout.vue'),
      meta: {
        requiresAuth: true,
        title: '系统管理'
      },
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: {
            requiresAuth: true,
            title: '仪表盘'
          }
        },
        {
          path: 'users',
          name: 'Users',
          component: () => import('@/views/Users.vue'),
          meta: {
            requiresAuth: true,
            requiresAdmin: true,
            title: '用户管理'
          }
        },
        {
          path: 'user_groups',
          name: 'UserGroups',
          component: () => import('@/views/UserGroups.vue'),
          meta: {
            requiresAuth: true,
            requiresAdmin: true,
            title: '用户组管理'
          }
        },
        {
          path: 'storage',
          name: 'Storage',
          component: () => import('@/views/Storage.vue'),
          meta: {
            requiresAuth: true,
            requiresAdmin: true,
            title: '存储管理'
          }
        },
        {
          path: 'cloud_token',
          name: 'CloudToken',
          component: () => import('@/views/CloudToken.vue'),
          meta: {
            requiresAuth: true,
            title: '云盘登录'
          }
        },
        {
          path: 'setting',
          name: 'Settings',
          component: () => import('@/views/Settings.vue'),
          meta: {
            requiresAuth: true,
            requiresAdmin: true,
            title: '系统设置'
          }
        },
        {
          path: 'profile',
          name: 'Profile',
          component: () => import('@/views/Profile.vue'),
          meta: {
            requiresAuth: true,
            title: '个人中心'
          }
        },
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'FileBrowserPath',
      component: () => import('@/views/FileBrowser.vue'),
      meta: {
        requiresAuth: true,
        title: '文件浏览'
      },
      beforeEnter: (to, _, next) => {
        // 如果路径指向一个文件（包含文件扩展名），则跳转到文件详情页面
        const path = Array.isArray(to.params.pathMatch) ? to.params.pathMatch.join('/') : to.params.pathMatch || ''
        const lastSegment = path.split('/').pop() || ''
        const hasExtension = lastSegment.includes('.') && !lastSegment.startsWith('.')
        
        if (hasExtension) {
          // 跳转到文件详情页面
          next({ name: 'FileDetail', params: { pathMatch: to.params.pathMatch } })
        } else {
          next()
        }
      }
    },
    {
      path: '/file/:pathMatch(.*)*',
      name: 'FileDetail',
      component: () => import('@/views/FileDetail.vue'),
      meta: {
        requiresAuth: true,
        title: '文件详情'
      }
    }
  ]
})

// 路由守卫
router.beforeEach(async (to, _, next) => {
  const authStore = useAuthStore()
  const settingStore = useSettingStore()
  
  // 检查系统是否已初始化
  try {
    await settingStore.fetchSetting()
  } catch (error) {
    console.error('获取系统设置失败:', error)
  }
  
  // 如果系统未初始化且不是访问初始化页面，重定向到初始化页面
  if (settingStore.setting && !settingStore.setting.initialized && to.name !== 'InitSystem') {
    next('/@init')
    return
  }
  
  // 如果系统已初始化且访问初始化页面，重定向到登录页面
  if (settingStore.setting && settingStore.setting.initialized && to.name === 'InitSystem') {
    next('/@login')
    return
  }
  
  // 如果有token，先获取最新的用户信息
  if (authStore.token) {
    try {
      await authStore.initialize()
    } catch (error) {
      console.error('初始化认证状态失败:', error)
    }
  }
  
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - 云盘189分享管理系统`
  }
  
  // 检查是否需要认证
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/@login')
    return
  }
  
  // 检查是否需要管理员权限
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    next('/@dashboard')
    return
  }
  
  // 如果已登录且访问登录页，重定向到仪表盘
  if (to.name === 'Login' && authStore.isAuthenticated) {
    next('/@dashboard')
    return
  }
  
  next()
})

export default router