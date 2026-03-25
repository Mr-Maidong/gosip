# YSIP Web 产品需求文档

## 项目概述

YSIP Web 是 YSIP（GB28181 SIP 服务器）的 Web 管理平台，提供设备管理、通道管理、流媒体管理、用户管理等功能。

## 技术栈

- **前端框架**: Vue 3 (Composition API)
- **构建工具**: Vite
- **UI 组件库**: Ant Design Vue
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP 请求**: Axios (基于封装)

## 功能清单

### 1. 用户认证

| 功能 | 描述 | 状态 |
|------|------|------|
| 用户登录 | 用户名密码登录 | ✅ 已实现 |
| 用户登出 | 退出登录，清除认证状态 | ✅ 已实现 |
| 登录状态持久化 | localStorage 保存 token | ✅ 已实现 |
| 路由鉴权 | 未登录自动跳转登录页 | ✅ 已实现 |
| 角色映射 | admin/operator/viewer 角色显示 | ✅ 已实现 |

### 2. 首页

| 功能 | 描述 | 状态 |
|------|------|------|
| 欢迎页 | 展示平台基本信息 | ✅ 已实现 |

### 3. 平台管理 (/platform)

| 功能 | 描述 | 状态 |
|------|------|------|
| 平台列表 | 显示 GB28181 平台列表 | 🔲 待开发 |

### 4. 监控管理 (/devices)

| 功能 | 描述 | 状态 |
|------|------|------|
| 设备列表 | 获取设备列表 (API: getDevices) | 🔲 待开发 |
| 创建设备 | 新增设备 (API: createDevice) | 🔲 待开发 |
| 删除设备 | 删除设备 (API: deleteDevice) | 🔲 待开发 |
| PTZ 控制 | 云台控制 (API: ptzControl) | 🔲 待开发 |

### 5. 流管理 (/streams)

| 功能 | 描述 | 状态 |
|------|------|------|
| 流列表 | 获取流列表 (API: getStreams) | 🔲 待开发 |
| 开始播放 | 启动推流 (API: startStream) | 🔲 待开发 |
| 停止播放 | 停止推流 (API: stopStream) | 🔲 待开发 |

### 6. 通道管理 (集成在设备管理中)

| 功能 | 描述 | 状态 |
|------|------|------|
| 通道列表 | 获取通道列表 (API: getChannels) | 🔲 待开发 |
| 创建通道 | 新增通道 (API: createChannel) | 🔲 待开发 |
| 录像回放 | 获取录像列表 (API: getRecords) | 🔲 待开发 |

### 7. 系统设置 (/settings)

| 功能 | 描述 | 状态 |
|------|------|------|
| 系统配置 | 系统参数配置 | 🔲 待开发 |

### 8. 用户管理 (集成在系统设置中)

| 功能 | 描述 | 状态 |
|------|------|------|
| 用户列表 | 获取用户列表 (API: getUserList) | 🔲 待开发 |
| 创建用户 | 新增用户 (API: createUser) | 🔲 待开发 |
| 更新用户 | 编辑用户信息 (API: updateUser) | 🔲 待开发 |
| 删除用户 | 删除用户 (API: deleteUser) | 🔲 待开发 |
| 启用用户 | 启用用户 (API: enableUser) | 🔲 待开发 |
| 禁用用户 | 禁用用户 (API: disableUser) | 🔲 待开发 |
| 修改密码 | 修改用户密码 (API: changePassword) | 🔲 待开发 |

## 界面布局

### 整体结构

```
┌─────────────────────────────────────────────────┐
│  Logo  │           Tabs (Header)       │  User  │
├────────┼─────────────────────────────────────────┤
│        │                                         │
│  Menu  │              Content                    │
│        │                                         │
│        │                                         │
├────────┴─────────────────────────────────────────┤
│                   Footer                         │
└─────────────────────────────────────────────────┘
```

### 组件

| 组件 | 描述 |
|------|------|
| BasicLayout | 基础布局容器 |
| Header | 顶部导航栏（含标签页、用户信息） |
| Sider | 左侧菜单栏 |
| Content | 主内容区 |
| Footer | 底部版权信息 |

### 界面特性

| 特性 | 描述 | 状态 |
|------|------|------|
| 标签页管理 | 多标签页切换 | ✅ 已实现 |
| 标签页双击关闭 | 双击标签页关闭 | ✅ 已实现 |
| 首页固定 | 首页始终在第一个位置 | ✅ 已实现 |
| Tab 高亮同步 | 路由变化时高亮对应 Tab | ✅ 已实现 |
| 用户信息展示 | 头像、用户名、角色 | ✅ 已实现 |
| 侧边栏折叠 | 可折叠的侧边菜单 | ❌ 已移除 |

## API 路由

| 路径 | 方法 | 功能 |
|------|------|------|
| /api/v1/login | POST | 用户登录 |
| /api/v1/logout | POST | 用户登出 |
| /api/v1/users/current | GET | 获取当前用户 |
| /api/v1/users | GET | 获取用户列表 |
| /api/v1/users/create | POST | 创建用户 |
| /api/v1/users/:id | POST | 更新用户 |
| /api/v1/users/:id | DELETE | 删除用户 |
| /api/v1/users/:id/enable | POST | 启用用户 |
| /api/v1/users/:id/disable | POST | 禁用用户 |
| /api/v1/users/:id/password | POST | 修改密码 |
| /api/v1/devices | GET | 获取设备列表 |
| /api/v1/devices/create | POST | 创建设备 |
| /api/v1/devices/:id | DELETE | 删除设备 |
| /api/v1/devices/ptz | POST | PTZ 控制 |
| /api/v1/channels | GET | 获取通道列表 |
| /api/v1/devices/:deviceId/channels | POST | 创建通道 |
| /api/v1/channels/:channelId/streams | POST | 开始播放 |
| /api/v1/channels/:channelId/records | GET | 获取录像列表 |
| /api/v1/streams | GET | 获取流列表 |
| /api/v1/streams/:id | DELETE | 停止播放 |

## 数据模型

### 用户 (User)

| 字段 | 类型 | 描述 |
|------|------|------|
| id | number | 用户 ID |
| username | string | 用户名 |
| name | string | 姓名 |
| role | string | 角色 (admin/operator/viewer) |
| avatar | string | 头像 (base64) |
| status | string | 状态 (enabled/disabled) |

### 设备 (Device)

| 字段 | 类型 | 描述 |
|------|------|------|
| id | number | 设备 ID |
| name | string | 设备名称 |
| deviceId | string | GB28181 设备 ID |
| host | string | 设备 IP |
| port | number | 端口 |
| status | string | 在线状态 |

### 通道 (Channel)

| 字段 | 类型 | 描述 |
|------|------|------|
| id | number | 通道 ID |
| deviceId | string | 所属设备 ID |
| channelId | string | GB28181 通道 ID |
| name | string | 通道名称 |
| status | string | 在线状态 |

### 流 (Stream)

| 字段 | 类型 | 描述 |
|------|------|------|
| id | string | 流 ID |
| channelId | string | 通道 ID |
| name | string | 流名称 |
| status | string | 播放状态 |
| url | string | 流地址 |

## 版本信息

- 创建日期: 2026-03-25
- 版本: 1.0.0
