# 云盘189分享管理系统 - 前端

基于 Vue 3 + TypeScript + Vite 开发的现代化前端应用。

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全的 JavaScript 超集
- **Vite** - 快速的前端构建工具
- **Vue Router** - Vue.js 官方路由管理器
- **Pinia** - Vue 的状态管理库
- **Axios** - HTTP 客户端

## 项目结构

```
src/
├── api/           # API 接口封装
│   ├── index.ts   # Axios 配置和拦截器
│   └── user.ts    # 用户相关 API
├── router/        # 路由配置
│   └── index.ts   # 路由定义和守卫
├── stores/        # Pinia 状态管理
│   └── auth.ts    # 认证状态管理
├── styles/        # 样式文件
│   └── index.css  # 全局样式
├── views/         # 页面组件
│   ├── Login.vue      # 登录页面
│   ├── Dashboard.vue  # 仪表板
│   ├── Users.vue      # 用户管理
│   ├── Storage.vue    # 存储管理
│   └── Settings.vue   # 系统设置
├── App.vue        # 根组件
├── main.ts        # 应用入口
└── vite-env.d.ts  # TypeScript 声明文件
```

## 开发指南

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:5173

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

### 类型检查

```bash
npm run type-check
```

## 功能特性

### 已实现

- ✅ 项目基础架构搭建
- ✅ 用户认证系统（登录/登出）
- ✅ 路由配置和权限控制
- ✅ API 接口封装
- ✅ 响应式布局设计
- ✅ 登录页面
- ✅ 仪表板页面

### 待实现

- ⏳ 用户管理功能
- ⏳ 存储管理功能
- ⏳ 系统设置功能
- ⏳ 云盘Token管理
- ⏳ 文件分享管理

## API 接口

前端通过代理访问后端 API：

- 开发环境：`/api` -> `http://localhost:12395/api`
- 生产环境：需要配置 Nginx 反向代理

### 主要接口

- `POST /api/user/login` - 用户登录
- `POST /api/user/refresh_token` - 刷新Token
- `GET /api/user/info` - 获取用户信息
- `GET /api/user/list` - 获取用户列表（管理员）
- `POST /api/user/add` - 添加用户（管理员）

## 权限系统

系统采用基于位运算的权限控制：

- `PermissionBase = 1` - 基础权限
- `PermissionDavRead = 2` - DAV读取权限
- `PermissionAdmin = 4` - 管理员权限

## 样式设计

- 采用现代化的设计语言
- 响应式布局，支持移动端
- 自定义 CSS 类，无第三方 UI 库依赖
- 渐变背景和卡片式设计
- 统一的颜色主题和间距规范

## 开发注意事项

1. 所有 API 调用都通过 Axios 拦截器自动添加认证头
2. 路由守卫会自动检查用户认证状态和权限
3. 使用 Pinia 进行状态管理，支持持久化存储
4. TypeScript 严格模式，确保类型安全
5. 遵循 Vue 3 Composition API 最佳实践