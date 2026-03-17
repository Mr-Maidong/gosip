# GoSIP 项目上下文文档

## 项目概述

**GoSIP** 是一个基于 Go 语言开发的 GB28181 SIP 服务器，主要用于视频监控领域的信令控制。该项目与 [ZLMediaKit](https://github.com/xia-chu/ZLMediaKit) 媒体服务器配合使用，提供完整的 GB28181 国标协议支持。

### 核心技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.19+ |
| Web 框架 | Gin |
| ORM | GORM (MySQL) |
| 配置管理 | Viper |
| 日志 | Logrus |
| API 文档 | Swagger/OpenAPI |
| 认证授权 | JWT + RBAC |
| 媒体服务器 | ZLMediaKit |

### 主要功能

- ✅ 设备注册管理（NVR/DVR/IPC 摄像头）
- ✅ 通道管理
- ✅ 实时预览（直播流）
- ✅ 远程回放（录像回放）
- ✅ 录像历史文件获取
- ✅ 流管理（MySQL 存储，支持服务重启后恢复）
- ✅ 异步通知（Webhook）
- ✅ 语音对讲
- ✅ PTZ 控制
- ✅ 用户管理（新增/编辑/删除/启用/禁用/密码修改）
- ✅ **RBAC 权限管理**（用户 - 角色 - 权限三层模型）
- ✅ **角色管理**（默认角色：admin、operator、viewer）
- ✅ **权限管理**（menu/button/api 类型）
- ✅ **JWT 认证**（Token 有效期 7 天）
- ✅ **登录/登出**接口

### 项目结构

```
gosip/
├── main.go              # 程序入口，启动 SIP 服务和 API 服务
├── config.yml           # 主配置文件
├── Makefile             # 构建和 Docker 打包脚本
├── Dockerfile           # Docker 镜像构建文件
├── api/                 # RESTful API 层
│   ├── main.go          # API 路由初始化
│   ├── c/               # API 控制器
│   │   ├── channels.go  # 通道管理接口
│   │   ├── devices.go   # 设备管理接口
│   │   ├── files.go     # 文件管理接口
│   │   ├── records.go   # 录像接口
│   │   ├── streams.go   # 流管理接口
│   │   ├── talk.go      # 语音对讲接口
│   │   ├── users.go     # 用户管理接口
│   │   ├── roles.go     # 角色管理接口 ⭐新增
│   │   ├── permissions.go # 权限管理接口 ⭐新增
│   │   └── zlm.go       # ZLMediaKit Webhook 接口
│   ├── model/           # 数据模型 ⭐新增
│   │   └── model.go     # RBAC 模型定义
│   ├── service/         # 服务层 ⭐新增
│   │   └── permission.go # 权限服务
│   └── middleware/      # 中间件
│       ├── auth.go      # JWT 认证中间件 ⭐新增
│       ├── permission.go # 权限验证中间件 ⭐新增
│       ├── cors.go      # CORS 中间件
│       └── recovery.go  # 异常恢复中间件
├── sip/                 # SIP 协议处理层
│   ├── handler.go       # SIP 消息处理器
│   ├── devices.go       # 设备注册/注销处理
│   ├── play.go          # 播放信令处理
│   ├── record.go        # 录像信令处理
│   ├── stream.go        # 流管理
│   ├── talk.go          # 语音对讲处理
│   ├── zlm.go           # ZLMediaKit 集成
│   ├── user.go          # 用户模型
│   ├── user_auth.go     # 用户认证（密码加密）
│   ├── keepalive.go     # 心跳处理
│   ├── notify.go        # 通知处理
│   ├── files.go         # 文件处理
│   ├── sys.go           # 系统信息
│   └── s/               # SIP 协议底层实现
├── db/                  # 数据库层
│   ├── gorm.go          # GORM 配置和通用方法
│   ├── model.go         # 数据模型定义
│   ├── filter.go        # 查询过滤器
│   └── tx.go            # 事务处理
├── m/                   # 配置和全局模块
│   ├── config.go        # 配置结构体定义和加载
│   ├── logger.go        # 日志配置
│   └── m.go             # 全局变量
├── utils/               # 工具函数
│   ├── utils.go         # 通用工具函数（含 JWT、密码加密）
│   └── logger.go        # 日志工具
├── docs/                # Swagger API 文档
├── demo/                # 演示/示例配置
└── web/                 # Web 相关资源
```

## 构建和运行

### 环境要求

- Go 1.19+
- MySQL 数据库
- ZLMediaKit 媒体服务器

### 本地开发

```bash
# 安装依赖
go mod download

# 直接运行（需要 config.yml 配置）
go run main.go

# 或使用 Makefile 构建 Linux 版本
make build
```

### Docker 部署

```bash
# 构建并推送 Docker 镜像
make all

# 或单独构建 Docker 镜像
make docker
```

### 配置说明

编辑 `config.yml` 配置以下关键参数：

```yaml
mod: release                    # 运行模式：debug/release
database:
  dialect: mysql
  url: mysip:password@tcp(host:port)/mysip?charset=utf8&parseTime=True
udp: 0.0.0.0:55060             # SIP 服务器 UDP 端口
tcp: 0.0.0.0:55060             # SIP 服务器 TCP 端口
api: 0.0.0.0:8090              # RESTful API 端口
secret: your-secret-key        # API 验证密钥
logger: trace                  # 日志级别：trace/debug/info/warn/error
media:
  restful: http://host:18080   # ZLMediaKit RESTful API 地址
  http: http://host:18080      # ZLMediaKit HTTP 地址
  ws: ws://host:18080          # ZLMediaKit WebSocket 地址
  rtmp: rtmp://host:1935       # RTMP 流地址
  rtsp: rtsp://host:8554       # RTSP 流地址
  rtp: http://host:10000       # RTP 推流地址
  secret: zlm-secret           # ZLMediaKit 密钥
stream:
  hls: 1                       # 是否开启 HLS 转码
  rtmp: 1                      # 是否开启 RTMP 转码
gb28181:
  lid: 34020000002000000001    # 系统 ID
  region: 3402000000           # 系统域
  did: 34020000001118          # 设备 ID 前缀
  cid: 34020000001318          # 通道 ID 前缀
  dnum: 0                      # 设备数量
  cnum: 0                      # 通道数量
```

### API 文档

启动服务后访问：`http://localhost:8090/swagger/index.html`

### 主要 API 接口

| 类别 | 端点 | 方法 | 描述 |
|------|------|------|------|
| 设备 | `/api/v1/devices` | GET | 设备列表 |
| 设备 | `/api/v1/devices/create` | POST | 创建设备 |
| 设备 | `/api/v1/devices/:id` | POST/DELETE | 更新/删除设备 |
| 设备 | `/api/v1/devices/ptz` | POST | PTZ 控制 |
| 通道 | `/api/v1/channels` | GET | 通道列表 |
| 通道 | `/api/v1/devices/:id/channels` | POST | 创建通道 |
| 通道 | `/api/v1/devices/:id/channels_sync` | POST | 通道同步 |
| 通道 | `/api/v1/channels/:id` | POST/DELETE | 更新/删除通道 |
| 流 | `/api/v1/streams` | GET | 流列表 |
| 流 | `/api/v1/channels/:id/streams` | POST | 开始播放 |
| 流 | `/api/v1/streams/:id` | DELETE | 停止播放 |
| 语音 | `/api/v1/channels/:id/start_talk` | POST | 开始语音对讲 |
| 录像 | `/api/v1/channels/:id/records` | GET | 录像列表 |
| ZLM | `/zlm/webhook/:method` | POST | ZLMediaKit Webhook |
| ZLM | `/api/stream/keepalive` | POST | ZLM 心跳接口 |
| ZLM | `/api/v1/media/status` | GET | 获取媒体服务器状态 |
| 用户 | `/api/v1/users` | GET | 用户列表 |
| 用户 | `/api/v1/users/create` | POST | 创建用户 |
| 用户 | `/api/v1/users/:id` | POST/DELETE | 更新/删除用户 |
| 用户 | `/api/v1/users/:id/enable` | POST | 启用用户 |
| 用户 | `/api/v1/users/:id/disable` | POST | 禁用用户 |
| 用户 | `/api/v1/users/:id/password` | POST | 修改密码 |

## 开发规范

### 代码风格

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 错误处理使用 `errors.Is` 进行错误类型判断

### 数据库操作

- 所有数据库操作通过 `db/` 包封装的通用方法进行
- 使用 `DBModel` 作为基础模型，包含 `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` 字段
- 时间戳使用 Unix 时间戳（int64）存储
- 使用事务时通过 `db.NewTx` 创建事务对象，完成后调用 `tx.Commit` 提交

### 日志级别

配置文件中支持以下日志级别：
- `trace` - 最详细日志
- `debug` - 调试日志
- `info` - 信息日志
- `warn` - 警告日志
- `error` - 错误日志

### 定时任务

项目使用 `cron` 库执行以下定时任务：
- 每 5 分钟检查并关闭无效流（`CheckStreams`）
- 每 5 分钟清理过期录制文件（`ClearFiles`）

## 工作流程

### 设备注册流程

1. 调用 `POST /devices/create` 创建设备
2. 系统返回设备 SIP ID 和服务器配置
3. 在设备端配置 SIP 服务器信息
4. 设备向服务器发送 REGISTER 请求
5. 服务器处理设备注册并存储

### 直播播放流程

1. 调用 `POST /channels/:id/streams` 请求播放
2. 服务器发送 INVITE 请求到设备
3. 设备推送 RTP 流到 ZLMediaKit
4. ZLMediaKit 触发 Webhook 通知 GoSIP
5. 返回播放地址（RTMP/HLS/FLV 等）

### 录像回放流程

1. 调用 `GET /channels/:id/records` 获取录像列表
2. 选择时间段调用 `POST /channels/:id/streams`（带时间参数）
3. 服务器发送 PLAY 请求到设备
4. 设备推送历史录像流
5. 回放结束后调用 `DELETE /streams/:id` 关闭流

### 语音对讲流程

1. 调用 `POST /channels/:id/start_talk` 开始对讲
2. 服务器建立语音通道
3. 双向音频传输
4. 结束后关闭通道

## RBAC 权限管理

### 数据模型

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│    User     │────▶│  UserRole    │◀────│    Role     │
│   (用户)    │     │ (用户角色关联) │     │   (角色)    │
└─────────────┘     └──────────────┘     └─────────────┘
                           │                    │
                           │                    │
                           ▼                    ▼
                    ┌──────────────┐     ┌──────────────┐
                    │RolePermission│◀────│  Permission  │
                    │(角色权限关联) │     │   (权限)     │
                    └──────────────┘     └──────────────┘
```

### 默认角色

| 角色编码 | 角色名称 | 描述 | 权限 |
|----------|----------|------|------|
| admin | 超级管理员 | 系统超级管理员，拥有所有权限 | 全部权限（自动分配） |
| operator | 操作员 | 系统操作员，拥有操作权限 | 需手动分配 |
| viewer | 观察者 | 只读权限用户 | 需手动分配 |

### 默认用户

| 用户名 | 密码 | 角色 | 说明 |
|--------|------|------|------|
| admin | admin123 | admin | 系统启动时自动创建 |

### 认证流程

```
1. 用户登录 → POST /api/v1/login
2. 验证成功 → 返回 JWT Token（有效期 7 天）
3. 后续请求 → Header 携带 Authorization: Bearer <token>
4. 服务端验证 → Auth 中间件解析 Token，获取用户信息
5. 权限检查 → （可选）PermissionAuth 中间件检查权限
```

### 权限类型

| 类型 | 说明 | 示例 |
|------|------|------|
| menu | 菜单权限 | 设备管理、通道管理 |
| button | 按钮权限 | 创建、编辑、删除 |
| api | 接口权限 | GET /api/v1/devices |

### 使用示例

```bash
# 1. 登录获取 Token
curl -X POST http://localhost:8090/api/v1/login \
  -d "username=admin" -d "password=admin123"

# 2. 携带 Token 访问接口
curl -X GET http://localhost:8090/api/v1/devices \
  -H "Authorization: Bearer <your_token>"

# 3. 无 Token 访问（返回认证错误）
curl -X GET http://localhost:8090/api/v1/devices
# 返回：{"code":1004,"msg":"未提供认证信息"}
```

### 白名单接口（无需认证）

- `/api/v1/health` - 健康检查
- `/api/v1/login` - 登录
- `/api/v1/logout` - 登出
- `/zlm/webhook/*` - ZLMediaKit Webhook
- `/swagger/*` - Swagger 文档

## 注意事项

1. **ZLMediaKit 集成**：必须正确配置 ZLMediaKit 的 Webhook 指向本项目的 API 地址
2. **数据库初始化**：首次启动会自动创建数据库表（包括 RBAC 相关表）
3. **流管理**：直播流 5 分钟无人观看自动关闭，回放流需手动关闭
4. **接口安全**：所有 API 请求需要通过 JWT Token 进行认证
5. **端口配置**：确保 SIP 端口（UDP/TCP）和 API 端口未被占用
6. **pprof 调试**：服务启动后会开启 pprof 调试端口（6060）
7. **ZLM 心跳**：ZLM 需要定期调用 `/api/stream/keepalive` 接口保持心跳，否则媒体服务器状态会显示为离线
8. **JWT 密钥**：使用 config.yml 中的 `secret` 配置作为 JWT 密钥，生产环境请修改默认值
9. **默认密码**：首次启动会创建 admin 用户（密码 admin123），建议及时修改
