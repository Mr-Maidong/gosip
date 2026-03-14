# GoSIP Web

GoSIP Web 管理平台 - 基于 Vue 3 + Pinia + Ant Design Vue 的前端项目

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **Pinia** - Vue 官方状态管理库
- **Vue Router** - 官方路由管理器
- **Ant Design Vue** - 企业级 UI 组件库
- **Axios** - HTTP 请求库
- **Less** - CSS 预处理器
- **Vite** - 下一代前端构建工具

## 目录结构

```
web/
├── src/
│   ├── api/              # API 请求封装
│   ├── assets/           # 静态资源
│   ├── components/       # 公共组件
│   ├── layouts/          # 布局组件
│   ├── router/           # 路由配置
│   ├── store/            # Pinia 状态管理
│   ├── styles/           # 全局样式
│   ├── utils/            # 工具函数
│   ├── views/            # 页面组件
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── index.html
├── package.json
├── vite.config.js
└── README.md
```

## 快速开始

### 安装依赖

```bash
yarn install
```

### 启动开发服务器

```bash
yarn dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
yarn build
```

### 代码检查和格式化

```bash
yarn lint
yarn format
```

## 配置说明

### 代理配置

在 `vite.config.js` 中配置了 API 代理，默认指向 `http://localhost:8090`

### 环境变量

创建 `.env` 文件配置环境变量：

```env
VITE_API_BASE_URL=/api
```

## 相关项目

- [GoSIP](https://github.com/panjjo/gosip) - GB28181 SIP 服务器
