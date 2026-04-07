# YSIP 前端产品需求文档 (PRD)

## 1. 文档概述

### 1.1 文档目的

本文档描述 YSIP (原 GoSIP) 视频监控系统前端的产品需求，包括已完成功能和待开发功能。用于指导前端开发、测试和产品迭代。

### 1.2 适用范围

- **项目名称**: YSIP 前端 (web/)
- **技术栈**: Vue 3 + Vite + Ant Design Vue 4 + Pinia + Vue Router 4
- **后端服务**: Go + Gin + GB28181 SIP + ZLMediaKit
- **目标用户**: 视频监控系统管理员、操作员

### 1.3 术语说明

| 术语 | 说明 |
|------|------|
| YSIP / GoSIP | 基于 GB28181 协议的视频监控系统 |
| ZLM | ZLMediaKit 流媒体服务器 |
| 设备 | NVR/DVR 或支持 GB28181 的摄像头 |
| 通道 | 设备下的视频通道（摄像头） |
| 流 | 视频流（直播/回放） |
| GB28181 | 中国公共安全视频监控联网标准 |

---

## 2. 产品概述

### 2.1 产品背景

YSIP 是一个基于 GB28181 协议的视频监控系统平台，与 ZLMediaKit 配合提供完整的视频监控解决方案。前端基于 Vue 3 构建，提供设备管理、通道管理、流管理等功能。

### 2.2 产品目标

- ✅ 提供直观的设备管理和通道管理界面
- ✅ 支持实时视频监控和录像回放
- ✅ 支持流管理和状态监控
- ✅ 提供 RBAC 权限管理（后端支持）

### 2.3 用户角色

| 角色 | 说明 | 权限 |
|------|------|------|
| 管理员 (admin) | 系统管理员 | 所有权限 |
| 操作员 (operator) | 日常操作人员 | 设备管理、流管理 |
| 观察者 (observer) | 只读用户 | 查看权限 |

---

## 3. 产品架构

### 3.1 技术栈

| 类别 | 技术 | 版本 |
|------|------|------|
| **框架** | Vue 3 (Composition API) | ^3.4.0 |
| **构建工具** | Vite | ^5.0.10 |
| **UI 库** | Ant Design Vue | ^4.1.0 |
| **状态管理** | Pinia | ^2.1.7 |
| **路由** | Vue Router | ^4.2.5 |
| **HTTP 客户端** | Axios | ^1.6.5 |
| **时间处理** | dayjs | ^1.11.10 |
| **视频播放器** | EasyPlayer Pro | (内嵌 JS 库) |
| **CSS 预处理器** | Less | ^4.2.0 |
| **代码规范** | ESLint + Prettier | - |

### 3.2 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                      浏览器 (前端)                        │
│                                                         │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │  Vue 3      │  │  Pinia Store │  │  Vue Router   │  │
│  │  Components │  │  (状态管理)   │  │  (路由管理)    │  │
│  └──────┬──────┘  └──────┬───────┘  └───────┬───────┘  │
│         │                │                   │          │
│         └────────────────┼───────────────────┘          │
│                          │                              │
│                   ┌──────▼───────┐                       │
│                   │  Axios API   │                       │
│                   │  (HTTP 客户端)│                       │
│                   └──────┬───────┘                       │
└──────────────────────────┼──────────────────────────────┘
                           │ HTTP + JWT Token
                           ▼
┌─────────────────────────────────────────────────────────┐
│                    GoSIP 后端服务                         │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │  Gin     │  │  SIP     │  │  MySQL   │              │
│  │  (API)   │  │  Protocol│  │  (DB)    │              │
│  └──────────┘  └──────────┘  └──────────┘              │
│                                                         │
│  ┌──────────┐  ┌──────────┐                            │
│  │  Redis   │  │  ZLMediaKit (流媒体)                   │
│  │  (缓存)  │  │          │                            │
│  └──────────┘  └──────────┘                            │
└─────────────────────────────────────────────────────────┘
```

### 3.3 目录结构

```
web/
├── src/
│   ├── api/                    # API 请求模块
│   │   ├── request.js          # Axios 实例 + 拦截器
│   │   ├── user.js             # 用户相关 API
│   │   ├── device.js           # 设备相关 API
│   │   ├── channel.js          # 通道相关 API
│   │   └── stream.js           # 流相关 API
│   ├── components/             # 全局共享组件
│   │   ├── index.js            # 组件统一导出
│   │   └── LivePlayer/         # 视频播放器组件
│   │       └── index.vue
│   ├── layouts/                # 布局组件
│   │   └── BasicLayout.vue     # 主布局 (侧边栏 + 顶栏 + 内容区 + 标签页)
│   ├── router/                 # 路由配置
│   │   └── index.js            # 路由定义 + 导航守卫
│   ├── store/                  # Pinia 状态管理
│   │   ├── index.js            # Store 统一导出
│   │   ├── user.js             # 用户状态 (登录/登出/Token)
│   │   ├── app.js              # 应用状态 (侧边栏折叠/加载)
│   │   └── tabs.js             # 标签页状态 (打开/关闭/缓存)
│   ├── styles/                 # Less 样式
│   │   ├── variables.less      # 主题变量
│   │   ├── mixins.less         # Less Mixin
│   │   └── index.less          # 全局样式
│   ├── utils/                  # 工具函数
│   │   └── index.js            # 格式化/防抖/节流/深拷贝
│   ├── views/                  # 页面组件
│   │   ├── home/               # 首页模块
│   │   ├── devices/            # 设备管理模块
│   │   ├── platform/           # 平台管理模块 (占位)
│   │   ├── streams/            # 流管理模块
│   │   ├── settings/           # 系统设置模块 (占位)
│   │   ├── Login.vue           # 登录页
│   │   └── NotFound.vue        # 404 页
│   ├── App.vue                 # 根组件
│   └── main.js                 # 应用入口
├── public/
│   └── js/                     # 第三方 JS 库
│       └── EasyPlayer-*.js     # 视频播放器
├── package.json
├── vite.config.js
└── index.html
```

---

## 4. 功能需求

### 4.1 功能矩阵

| 模块 | 功能 | 状态 | 优先级 |
|------|------|------|--------|
| **用户认证** | 登录/登出 | ✅ 完成 | P0 |
| | JWT Token 认证 | ✅ 完成 | P0 |
| | 路由守卫 | ✅ 完成 | P0 |
| | localStorage 持久化 | ✅ 完成 | P1 |
| **首页** | 欢迎卡片 | ✅ 完成 | P1 |
| | 统计概览 | ✅ 完成 | P1 |
| | 快捷操作 | ✅ 完成 | P2 |
| **设备管理** | 设备列表 | ✅ 完成 | P0 |
| | 设备增删改 | ✅ 完成 | P0 |
| | 通道列表 | ✅ 完成 | P0 |
| | 通道同步 | ✅ 完成 | P1 |
| | 直播播放 | ✅ 完成 | P0 |
| | 批量删除 | ⏳ 待开发 | P2 |
| **流管理** | 流列表 | ✅ 完成 | P0 |
| | 状态筛选 | ✅ 完成 | P1 |
| | 停止流 | ✅ 完成 | P0 |
| | 批量停止 | ✅ 完成 | P1 |
| **标签页** | 多标签切换 | ✅ 完成 | P1 |
| | KeepAlive 缓存 | ✅ 完成 | P1 |
| | 双击关闭 | ✅ 完成 | P2 |
| **平台管理** | 级联平台管理 | ⏳ 占位 | P2 |
| **系统设置** | 系统配置 | ⏳ 占位 | P2 |
| **录像回放** | 回放播放 | ⏳ 待开发 | P1 |
| **云台控制** | PTZ 控制 | ⏳ 待开发 | P2 |
| **用户管理** | 用户 CRUD | ⏳ 待开发 | P2 |

---

## 5. 页面详细说明

### 5.1 登录页 (/login)

**功能描述**: 用户登录认证入口

**页面元素**:
- 用户名输入框（必填）
- 密码输入框（必填）
- 登录按钮
- 加载状态指示器

**交互逻辑**:
1. 用户输入用户名和密码
2. 点击登录按钮触发验证
3. 调用 `POST /api/v1/login` 接口
4. 登录成功：保存 Token 到 localStorage，跳转首页
5. 登录失败：显示错误 message

**样式要求**:
- 渐变紫色背景
- 浮动动画效果
- Logo 脉冲动画
- 居中卡片布局

---

### 5.2 首页 (/home)

**功能描述**: 系统概览和快捷入口

**页面元素**:

| 区域 | 内容 |
|------|------|
| 欢迎卡片 | 用户头像、问候语、角色标签、日期星期 |
| 快捷操作 | 添加设备、查看流、系统设置 |
| 统计概览 | 设备总数、通道总数、在线流总数 |

**问候语逻辑**:
| 时间段 | 问候语 |
|--------|--------|
| 00:00-06:00 | 凌晨好 |
| 06:00-09:00 | 早上好 |
| 09:00-12:00 | 上午好 |
| 12:00-14:00 | 中午好 |
| 14:00-18:00 | 下午好 |
| 18:00-22:00 | 晚上好 |
| 22:00-24:00 | 夜里好 |

**统计数据来源**:
- 设备总数：`GET /api/v1/devices?limit=1` → `total`
- 通道总数：`GET /api/v1/channels?limit=1` → `total`
- 在线流：`GET /api/v1/streams?limit=1&filters=[{"field_name":"status","opertator":"=","value":0}]` → `total`

**快捷操作跳转**:
- 添加设备 → 打开设备添加弹窗
- 查看流 → `/streams`
- 系统设置 → `/settings`

---

### 5.3 监控管理 (/devices)

**功能描述**: 设备管理和通道管理

#### 5.3.1 设备列表模式（默认）

**表格字段**:

| 字段 | 说明 | 样式 |
|------|------|------|
| 设备名称 | 显示名称或 deviceid | 带设备图标 📹 |
| 设备ID | 唯一标识 | 等宽字体 |
| 注册状态 | 已注册/未注册 | 徽章：绿色/灰色 |
| 厂商/型号/设备类型 | 设备信息 | 标签形式 |
| 收流地址 | 设备 IP | 文本 |
| 源地址 | 来源标识 | 文本 |
| 最后活跃 | 时间戳 | 本地时间格式化 |
| 操作 | 同步/通道/编辑/删除 | 按钮组 |

**操作功能**:

| 操作 | 说明 | API |
|------|------|-----|
| 同步 | 向设备发送通道同步请求 | `POST /api/v1/devices/:id/channels_sync` |
| 通道 | 切换到通道列表模式，显示该设备的通道 | 前端状态切换 |
| 编辑 | 打开编辑弹窗 | `POST /api/v1/devices/:id` |
| 删除 | 确认后删除设备 | `DELETE /api/v1/devices/:id` |

#### 5.3.2 通道列表模式

**表格字段**:

| 字段 | 说明 | 样式 |
|------|------|------|
| 通道名称 | 显示名称或 channelid | 带摄像头图标 🎥 |
| 通道ID | 唯一标识 | 等宽字体 |
| 状态 | 在线/离线 | 徽章：绿色/灰色 |
| 厂商/型号 | 设备信息 | 标签形式 |
| 播放类型 | 拉流/推流 | 标签：蓝色/绿色 |
| 最后活跃 | 时间戳 | 本地时间格式化 |
| 操作 | 直播/回放 | 按钮组 |

**直播功能**:
1. 点击"直播"按钮
2. 调用 `POST /api/v1/channels/:id/streams` 获取播放地址
3. 打开直播弹窗，使用 LivePlayer 组件播放
4. 支持 WS-FLV / RTMP 协议

**回放功能**: 当前未实现（按钮存在但无逻辑）

#### 5.3.3 设备表单

**添加设备模式**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 设备ID | 文本 | 否 | 不填自动生成 |
| 设备名称 | 文本 | 是 | 显示名称 |
| 设备密码 | 密码 | 是 | 设备接入密码 |

**编辑设备模式**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| 设备ID | 文本 | 是 | 禁用状态 |
| 设备名称 | 文本 | 是 | 显示名称 |
| 收流地址 | 文本 | 否 | 设备 IP |
| 厂商 | 文本 | 否 | 设备厂商 |
| 型号 | 文本 | 否 | 设备型号 |
| 设备密码 | 密码 | 否 | 留空不修改 |

#### 5.3.4 批量操作（待开发）

- 批量模式切换：进入/退出批量模式
- 批量选择：多选设备
- 批量删除：确认后批量删除（当前提示"功能开发中"）

#### 5.3.5 分页和筛选

| 功能 | 说明 |
|------|------|
| 分页 | 每页 20 条，支持切换 pageSize |
| 总数显示 | 显示总数、快速跳转 |
| 筛选 | 按注册状态筛选（已注册/未注册） |

---

### 5.4 流管理 (/streams)

**功能描述**: 视频流管理和状态监控

**表格字段**:

| 字段 | 说明 | 样式 |
|------|------|------|
| 流ID | 视频流唯一标识 | 等宽字体，灰色背景 |
| 设备/通道 | 显示 "deviceid / channelid" | 文本 |
| 类型 | 直播/回放 | 图标：绿色播放/蓝色历史 |
| ZLM流 | 已接收/未接收 | 标签：绿色/灰色 |
| 播放地址 | RTMP 地址 | 文本，可点击复制 |
| 更新时间 | 时间戳 | 本地时间格式化 |
| 状态 | 正常/关闭/未开始 | 徽章：绿色/红色/橙色 |
| 操作 | 停止 | 按钮 |

**操作功能**:

| 操作 | 说明 | API |
|------|------|-----|
| 停止单流 | 停止指定流 | `DELETE /api/v1/streams/:id` |
| 批量停止 | 选择多个流后批量停止 | 循环调用 `DELETE /api/v1/streams/:id` |

**状态筛选**:
- 按状态筛选：正常/关闭/未开始
- 单选筛选器

**分页**:
- 每页 20 条
- 显示总数、快速跳转

**批量停止统计**:
- 显示成功/失败数量
- 失败时显示错误信息

---

### 5.5 平台管理 (/platform)

**当前状态**: 占位页面

**计划功能**: 级联平台管理（GB28181 级联）

---

### 5.6 系统设置 (/settings)

**当前状态**: 占位页面

**计划功能**: 系统参数配置

---

### 5.7 404 页面 (/:pathMatch(*)*)

**功能**:
- 显示 404 大标题
- 提示文案 "抱歉，您访问的页面不存在"
- "返回首页" 按钮

---

## 6. 数据模型

### 6.1 核心数据结构

#### 设备 (Device)

```javascript
{
  id: number,              // 数据库 ID
  deviceid: string,        // 设备唯一标识 (SIP ID)
  name: string,            // 设备名称
  host: string,            // 收流地址 (IP)
  port: number,            // SIP 端口
  manufacturer: string,    // 厂商
  model: string,           // 型号
  firmware: string,        // 固件版本
  regist: boolean,         // 注册状态
  active: number,          // 最后活跃时间戳
  source: string,          // 源地址
  created_at: string,      // 创建时间
  updated_at: string       // 更新时间
}
```

#### 通道 (Channel)

```javascript
{
  id: number,              // 数据库 ID
  channelid: string,       // 通道唯一标识
  deviceid: string,        // 所属设备 ID
  name: string,            // 通道名称
  status: number,          // 状态 (1=在线, 0=离线)
  streamtype: string,      // 播放类型 (push/pull)
  url: string,             // 拉流地址 (streamtype=pull 时有效)
  manufacturer: string,    // 厂商
  model: string,           // 型号
  active: number,          // 最后活跃时间戳
  created_at: string,      // 创建时间
  updated_at: string       // 更新时间
}
```

#### 流 (Stream)

```javascript
{
  id: number,              // 数据库 ID
  streamid: string,        // 流 ID (SSRC 16进制)
  deviceid: string,        // 设备 ID
  channelid: string,       // 通道 ID
  t: number,               // 类型 (0=直播, 1=回放)
  streamtype: string,      // 播放类型 (push/pull)
  status: number,          // 状态 (0=正常, 1=关闭, -1=未开始)
  stream: boolean,         // ZLM 是否收到流
  rtmp: string,            // RTMP 播放地址
  http: string,            // HLS 播放地址
  wsflv: string,           // WS-FLV 播放地址
  rtsp: string,            // RTSP 播放地址
  callid: string,          // SIP Call-ID
  stop: boolean,           // 是否停止
  created_at: string,      // 创建时间
  updated_at: string       // 更新时间
}
```

#### 用户 (User)

```javascript
{
  id: number,              // 数据库 ID
  username: string,        // 用户名 (登录用)
  name: string,            // 显示名称
  role: string,            // 角色 (admin/operator/observer)
  avatar: string,          // 头像 URL
  email: string,           // 邮箱
  phone: string,           // 手机
  status: number,          // 状态 (1=启用, 0=禁用)
  created_at: string,      // 创建时间
  updated_at: string       // 更新时间
}
```

### 6.2 状态管理 (Pinia Stores)

#### user store

```javascript
{
  token: string,           // JWT Token
  userInfo: {              // 用户信息
    username: string,
    name: string,
    role: string,
    user_id: number,
    avatar: string
  }
}
```

#### tabs store

```javascript
{
  openedTabs: [            // 已打开的标签页
    {
      path: string,        // 路由路径
      title: string,       // 标签页标题
      name: string,        // 组件名
      keepAlive: boolean   // 是否缓存
    }
  ],
  activeTab: string,       // 当前激活的标签页路径
  restored: boolean        // 是否已从 localStorage 恢复
}
```

#### app store

```javascript
{
  collapsed: boolean,      // 侧边栏折叠状态
  loading: boolean         // 全局加载状态
}
```

---

## 7. 非功能需求

### 7.1 性能要求

| 指标 | 要求 |
|------|------|
| 首屏加载 | < 3 秒 |
| 页面切换 | < 500ms |
| API 超时 | 30 秒 |
| 列表分页 | 默认 20 条/页 |
| 视频播放延迟 | < 2 秒 |

### 7.2 安全要求

| 项目 | 要求 |
|------|------|
| 认证方式 | JWT Token (Bearer) |
| Token 存储 | localStorage |
| Token 有效期 | 7 天 (后端配置) |
| 路由守卫 | 未登录跳转登录页 |
| 响应拦截 | code=1004 自动登出 |

### 7.3 兼容性

| 浏览器 | 版本 | 支持程度 |
|--------|------|---------|
| Chrome | >= 90 | ✅ 完全支持 |
| Firefox | >= 88 | ✅ 完全支持 |
| Edge | >= 90 | ✅ 完全支持 |
| Safari | >= 14 | ✅ 完全支持 |

### 7.4 样式规范

| 类别 | 值 |
|------|-----|
| 主色 | #1890ff |
| 成功色 | #52c41a |
| 警告色 | #faad14 |
| 错误色 | #f5222d |
| 基础字号 | 14px |
| 背景色 | #ffffff / #f0f2f5 |
| 圆角 | 4px / 8px |
| 间距 | 4px / 8px / 16px / 24px / 32px |

---

## 8. 待开发功能清单

### 8.1 高优先级 (P1)

| 功能 | 说明 | 涉及文件 |
|------|------|---------|
| **录像回放** | 设备通道列表中的回放功能 | `devices/index.vue` |
| **用户管理页面** | 用户 CRUD 管理界面 | 新建 `views/users/index.vue` |

### 8.2 中优先级 (P2)

| 功能 | 说明 | 涉及文件 |
|------|------|---------|
| **平台管理** | 级联平台管理功能 | `platform/index.vue` |
| **系统设置** | 系统参数配置 | `settings/index.vue` |
| **设备批量删除** | 批量选择并删除设备 | `devices/index.vue` |
| **PTZ 云台控制** | 云台方向控制和变焦 | 新建组件或集成到直播弹窗 |

### 8.3 低优先级 (P3)

| 功能 | 说明 | 备注 |
|------|------|------|
| 响应式布局优化 | 支持移动端和平板 | 当前仅支持桌面端 |
| 主题切换 | 支持深色/浅色主题 | Ant Design Vue 内置支持 |
| 国际化 | 支持多语言 | 当前仅中文 |

---

## 9. 附录

### 9.1 API 端点列表

#### 用户模块

| 方法 | 端点 | 函数名 | 状态 |
|------|------|--------|------|
| POST | `/api/v1/login` | `login(data)` | ✅ 已使用 |
| POST | `/api/v1/logout` | `logout()` | ✅ 已使用 |
| GET | `/api/v1/users` | `getUserList(params)` | ⏳ 待使用 |
| POST | `/api/v1/users/create` | `createUser(data)` | ⏳ 待使用 |
| POST | `/api/v1/users/:id` | `updateUser(id, data)` | ⏳ 待使用 |
| DELETE | `/api/v1/users/:id` | `deleteUser(id)` | ⏳ 待使用 |
| POST | `/api/v1/users/:id/enable` | `enableUser(id)` | ⏳ 待使用 |
| POST | `/api/v1/users/:id/disable` | `disableUser(id)` | ⏳ 待使用 |
| POST | `/api/v1/users/:id/password` | `changePassword(id, data)` | ⏳ 待使用 |

#### 设备模块

| 方法 | 端点 | 函数名 | 状态 |
|------|------|--------|------|
| GET | `/api/v1/devices` | `getDevices(params)` | ✅ 已使用 |
| POST | `/api/v1/devices/create` | `createDevice(data)` | ✅ 已使用 |
| DELETE | `/api/v1/devices/:id` | `deleteDevice(id)` | ✅ 已使用 |
| POST | `/api/v1/devices/:id` | (匿名) | ✅ 已使用 (编辑) |
| POST | `/api/v1/devices/ptz` | `ptzControl(data)` | ⏳ 待使用 |
| POST | `/api/v1/devices/:id/channels_sync` | `syncChannels(id)` | ✅ 已使用 |

#### 通道模块

| 方法 | 端点 | 函数名 | 状态 |
|------|------|--------|------|
| GET | `/api/v1/channels` | `getChannels(params)` | ✅ 已使用 |
| POST | `/api/v1/devices/:deviceId/channels` | `createChannel(deviceId, data)` | ⏳ 待使用 |
| GET | `/api/v1/channels/:channelId/records` | `getRecords(channelId, params)` | ⏳ 待使用 |

#### 流模块

| 方法 | 端点 | 函数名 | 状态 |
|------|------|--------|------|
| GET | `/api/v1/streams` | `getStreams(params)` | ✅ 已使用 |
| POST | `/api/v1/channels/:channelId/streams` | `startStream(channelId, data)` | ✅ 已使用 |
| DELETE | `/api/v1/streams/:id` | `stopStream(id)` | ✅ 已使用 |

### 9.2 关键文件路径

| 用途 | 文件路径 |
|------|---------|
| **应用入口** | `web/src/main.js` |
| **根组件** | `web/src/App.vue` |
| **路由配置** | `web/src/router/index.js` |
| **主布局** | `web/src/layouts/BasicLayout.vue` |
| **登录页** | `web/src/views/Login.vue` |
| **404 页** | `web/src/views/NotFound.vue` |
| **首页** | `web/src/views/home/index.vue` |
| **欢迎卡片** | `web/src/views/home/WelcomeCard.vue` |
| **设备列表** | `web/src/views/devices/index.vue` |
| **设备表单** | `web/src/views/devices/DeviceForm.vue` |
| **平台管理** | `web/src/views/platform/index.vue` |
| **流管理** | `web/src/views/streams/index.vue` |
| **系统设置** | `web/src/views/settings/index.vue` |
| **视频播放器** | `web/src/components/LivePlayer/index.vue` |
| **API 请求封装** | `web/src/api/request.js` |
| **用户 API** | `web/src/api/user.js` |
| **设备 API** | `web/src/api/device.js` |
| **通道 API** | `web/src/api/channel.js` |
| **流 API** | `web/src/api/stream.js` |
| **用户 Store** | `web/src/store/user.js` |
| **标签页 Store** | `web/src/store/tabs.js` |
| **应用 Store** | `web/src/store/app.js` |
| **主题变量** | `web/src/styles/variables.less` |
| **Mixins** | `web/src/styles/mixins.less` |
| **全局样式** | `web/src/styles/index.less` |
| **工具函数** | `web/src/utils/index.js` |
| **Vite 配置** | `web/vite.config.js` |
| **依赖清单** | `web/package.json` |

### 9.3 路由配置

| 路由路径 | 页面名称 | 标签页标题 | 菜单分组 | KeepAlive | 认证 |
|----------|---------|-----------|---------|-----------|------|
| `/` | Layout | - | - | - | ✅ |
| `/home` | Home | 首页 | 控制台 | ✅ | ✅ |
| `/platform` | Platform | 平台管理 | 设备管理 | ✅ | ✅ |
| `/devices` | Devices | 监控管理 | 设备管理 | ✅ | ✅ |
| `/streams` | Streams | 流管理 | 流媒体管理 | ✅ | ✅ |
| `/settings` | Settings | 系统设置 | 系统管理 | ✅ | ✅ |
| `/login` | Login | 登录 | - | ❌ | ❌ |
| `/:pathMatch(.*)*` | NotFound | 404 | - | ❌ | ❌ |

### 9.4 响应格式

**成功响应**:
```json
{
  "code": "0",
  "msgid": "...",
  "data": { ... }
}
```

**错误响应**:
```json
{
  "code": "1000",
  "msgid": "...",
  "data": "错误信息"
}
```

**状态码说明**:

| 状态码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1000 | 认证错误 |
| 1001 | 数据库错误 |
| 1002 | 参数错误 |
| 1003 | 系统错误 |
| 1004 | Token 过期/无效 (前端自动登出) |

---

## 10. 更新日志

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.0.0 | 2026-04-06 | 初始版本，包含设备管理、流管理、首页、登录认证 |

---

**文档维护**: 请在每次功能更新后同步更新此文档。
