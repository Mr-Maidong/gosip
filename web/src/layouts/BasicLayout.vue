<template>
  <a-layout class="layout">
    <!-- 垂直侧边栏 -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      class="sider"
      theme="light"
    >
      <!-- 顶部系统名称 -->
      <div class="logo">
        <div class="logo-icon">
          <svg viewBox="0 0 1147 1024" height="32"><path d="M577.980145 584.790361L955.058892 766.026024v38.640578l-377.078747 209.981687-377.06641-218.95094v-34.149783z" fill="#1890FF" opacity=".2" p-id="7886"></path><path d="M579.238554 978.15441L959.969157 766.13706 579.324916 531.863133l-2.553832-0.024675-380.69359 225.23065 380.718265 221.060627 2.430458 0.024675zM950.222651 765.902651L578.066506 973.157783 205.836337 757.019759 578.004819 536.847422 950.222651 765.902651z" fill="#FFFFFF" opacity=".2" p-id="7887"></path><path d="M577.77041 302.857253l502.771662 243.243181v43.846939L577.758072 871.732434 75.023422 577.90612v-37.826313z" fill="#1890FF" opacity=".4" p-id="7888"></path><path d="M579.028819 830.007518L1085.44 546.186795l-506.324819-313.615422-2.553832-0.024674L70.199518 534.046843l506.374169 295.936 2.455132 0.024675z m496.701687-284.030458L577.844434 824.998554 79.909012 533.997494l497.886072-296.454169 497.935422 308.433735z" fill="#FFFFFF" opacity=".4" p-id="7889"></path><path d="M571.515373 89.988627l565.605784 270.558072V412.06747L571.515373 725.497831 5.921928 398.656771v-44.809253z" fill="#1890FF" opacity=".6" p-id="7890"></path><path d="M572.761446 676.123759L1142.068434 360.645398 572.847807 12.028916l-2.541494-0.024675L1.061012 347.148337l569.269976 328.963085 2.430458 0.012337z m559.52347-315.700434L571.589398 671.127133 10.856867 347.098988 571.540048 17.025542l560.73253 343.410121z" fill="#FFFFFF" opacity=".6" p-id="7891"></path></svg>
        </div>
        <span v-if="!collapsed">YSIP</span>
      </div>

      <!-- 中间菜单 -->
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="light"
        mode="inline"
        :items="menuItems"
        @click="handleMenuClick"
        class="menu"
      />

      <!-- 底部用户信息 -->
      <div class="user-info">
        <a-dropdown placement="topRight">
          <div class="user-wrapper">
            <a-avatar :size="32" :src="userAvatar">
              <template #icon v-if="!userAvatar">{{ userName.charAt(0) }}</template>
            </a-avatar>
            <div class="user-details" v-if="!collapsed">
              <div class="user-name">{{ userName }}</div>
              <div class="user-role">{{ userRole }}</div>
            </div>
          </div>
          <template #overlay>
            <a-menu>
              <a-menu-item key="settings">系统设置</a-menu-item>
              <a-menu-divider />
              <a-menu-item key="logout" @click="handleLogout">退出登录</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </a-layout-sider>

    <!-- 主内容区 -->
    <a-layout class="main-layout">
      <div class="header-breadcrumb">
          <a-breadcrumb>
            <a-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
              {{ item.title }}
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>

      <!-- 内容区 -->
      <a-layout-content class="content">
        <RouterView />
      </a-layout-content>

      <!-- 底部 -->
      <a-layout-footer class="footer">
        YSIP Web ©2026 Created by YSIP Team
      </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { ref, computed, h, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  HomeOutlined,
  DesktopOutlined,
  VideoCameraOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined
} from '@ant-design/icons-vue'
import { useUserStore } from '@/store/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const collapsed = ref(false)

const selectedKeys = computed(() => [route.path])

const userName = computed(() => userStore.userInfo?.name || '管理员')

const userRole = computed(() => {
  const role = userStore.userInfo?.role
  const roleMap = {
    admin: '超级管理员',
    operator: '操作员',
    viewer: '观察者'
  }
  return roleMap[role] || '管理员'
})

// 头像处理：如果是 base64 编码，添加 data URI 前缀
const userAvatar = computed(() => {
  const avatar = userStore.userInfo?.avatar
  if (!avatar) return ''
  // 如果已经是 data URI 格式，直接返回
  if (avatar.startsWith('data:')) return avatar
  // 否则添加 base64 前缀
  return `data:image/png;base64,${avatar}`
})

// 面包屑导航
const breadcrumbs = computed(() => {
  const matched = route.matched.filter(item => item.meta && item.meta.title)
  return matched.map(item => ({
    path: item.path,
    title: item.meta.title
  }))
})

const menuItems = [
  {
    key: 'console_group',
    type: 'group',
    label: '控制台',
    children: [
      {
        key: '/home',
        icon: h('img', { src: new URL('@/assets/svgs/home.svg', import.meta.url).href }),
        label: '首页',
        title: '首页'
      },
    ]
  },
  {
    key: 'device-group',
    type: 'group',
    label: '设备管理',
    children: [
      {
        key: '/platform',
        icon: h('img', { src: new URL('@/assets/svgs/gb28181.svg', import.meta.url).href }),
        label: '平台管理',
        title: '平台管理'
      },
      {
        key: '/devices',
        icon: h('img', { src: new URL('@/assets/svgs/device.svg', import.meta.url).href }),
        label: '监控管理',
        title: '监控管理'
      }
    ]
  },
  {
    key: 'stream-group',
    type: 'group',
    label: '流媒体管理',
    children: [
      {
        key: '/streams',
        icon: h('img', { src: new URL('@/assets/svgs/stream.svg', import.meta.url).href }),
        label: '流管理',
        title: '流管理'
      }
    ]
  },
  {
    key: 'system-group',
    type: 'group',
    label: '系统管理',
    children: [
      {
        key: '/settings',
        icon: h('img', { src: new URL('@/assets/svgs/settings.svg', import.meta.url).href }),
        label: '系统设置',
        title: '系统设置'
      }
    ]
  }
]

const handleMenuClick = ({ key }) => {
  router.push(key)
}

const handleLogout = async () => {
  await userStore.logout()
  message.success('已退出登录')
  router.push('/login')
}

// 组件挂载时恢复用户状态
onMounted(() => {
  userStore.restoreFromStorage()
})
</script>

<style lang="less" scoped>
.layout {
  min-height: 100vh;
}

.sider {
  padding-inline: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.06);
  background: #fafafa;

  :deep(.ant-layout-sider-children) {
    display: flex;
    flex-direction: column;
  }

  .logo {
    height: 64px;
    padding: 0 20px;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 10px;
    transition: all 0.3s;
    border-bottom: 1px solid #f0f0f0;

    .logo-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
    }

    span {
      font-size: 16px;
      font-weight: 600;
      color: #333;
      white-space: nowrap;
      transition: opacity 0.3s;
    }
  }

  .menu {
    flex: 1;
    border-right: none !important;
    background: #fafafa;

    :deep(.ant-menu-item-group-title) {
      font-size: 12px;
    }

    :deep(.ant-menu-item-selected) {
      color: inherit;
      background-color: #00000010;
    }

  }

  .user-info {
    margin-block: 4px;
    height: 54px;
    display: flex;
    align-items: center;
    border-top: 1px solid #f0f0f0;

    .user-wrapper {
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;
      padding: 8px;
      border-radius: 8px;
      transition: background 0.3s;

      &:hover {
        background: #f5f5f5;
      }

      .user-details {
        display: flex;
        flex-direction: column;
        gap: 2px;
        overflow: hidden;
        flex: 1;

        .user-name {
          font-size: 14px;
          font-weight: 500;
          color: #333;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .user-role {
          font-size: 12px;
          color: #999;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
      }
    }
  }
}

.main-layout {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.content {
  flex: 1;
  margin: 16px;
  padding: 24px;
  background: #fff;
  border-radius: 8px;
  overflow-y: auto;
  min-height: 280px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.footer {
  text-align: center;
  color: #999;
  font-size: 13px;
}
</style>
