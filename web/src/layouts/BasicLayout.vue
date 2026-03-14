<template>
  <a-layout class="layout">
    <a-layout-header class="header">
      <div class="logo">
        <span>GoSIP</span>
      </div>
      <a-menu v-model:selectedKeys="selectedKeys" theme="dark" mode="horizontal" :items="menuItems" />
      <div class="user-info">
        <a-dropdown>
          <span class="user-name">
            <a-avatar :size="24">U</a-avatar>
            管理员
          </span>
          <template #overlay>
            <a-menu>
              <a-menu-item key="settings">系统设置</a-menu-item>
              <a-menu-divider />
              <a-menu-item key="logout" @click="handleLogout">退出登录</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </a-layout-header>
    <a-layout-content class="content">
      <RouterView />
    </a-layout-content>
    <a-layout-footer class="footer">
      GoSIP Web ©2026 Created by GoSIP Team
    </a-layout-footer>
  </a-layout>
</template>

<script setup>
import { ref, computed, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  HomeOutlined,
  DesktopOutlined,
  VideoCameraOutlined,
  SettingOutlined
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()

const selectedKeys = computed(() => [route.path])

const menuItems = [
  {
    key: '/home',
    icon: h(HomeOutlined),
    label: '首页'
  },
  {
    key: '/devices',
    icon: h(DesktopOutlined),
    label: '设备管理'
  },
  {
    key: '/channels',
    icon: h(VideoCameraOutlined),
    label: '通道管理'
  },
  {
    key: '/streams',
    icon: h(VideoCameraOutlined),
    label: '流管理'
  },
  {
    key: '/settings',
    icon: h(SettingOutlined),
    label: '系统设置'
  }
]

const handleLogout = () => {
  // TODO: 实现登出逻辑
  router.push('/login')
}
</script>

<style lang="less" scoped>
.layout {
  min-height: 100vh;
}

.header {
  display: flex;
  align-items: center;
  padding: 0 24px;

  .logo {
    display: flex;
    align-items: center;
    margin-right: 24px;

    span {
      font-size: 20px;
      font-weight: bold;
      color: #fff;
    }
  }

  :deep(.ant-menu) {
    flex: 1;
    line-height: 64px;
  }

  .user-info {
    .user-name {
      color: #fff;
      cursor: pointer;
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }
}

.content {
  margin: 16px;
  padding: 24px;
  background: #fff;
  min-height: 280px;
}

.footer {
  text-align: center;
  color: #999;
}
</style>
