<template>
  <a-layout class="layout">
    <!-- 垂直侧边栏 -->
    <a-layout-sider
      :trigger="null"
      class="sider"
      theme="light"
      width="220"
    >
      <!-- 顶部系统名称 -->
      <div class="logo">
        <div class="logo-icon">
          <svg
            viewBox="0 0 1147 1024"
            height="32"
          >
            <path
              d="M577.980145 584.790361L955.058892 766.026024v38.640578l-377.078747 209.981687-377.06641-218.95094v-34.149783z"
              fill="#1890FF"
              opacity=".2"
              p-id="7886"
            />
            <path
              d="M579.238554 978.15441L959.969157 766.13706 579.324916 531.863133l-2.553832-0.024675-380.69359 225.23065 380.718265 221.060627 2.430458 0.024675zM950.222651 765.902651L578.066506 973.157783 205.836337 757.019759 578.004819 536.847422 950.222651 765.902651z"
              fill="#FFFFFF"
              opacity=".2"
              p-id="7887"
            />
            <path
              d="M577.77041 302.857253l502.771662 243.243181v43.846939L577.758072 871.732434 75.023422 577.90612v-37.826313z"
              fill="#1890FF"
              opacity=".4"
              p-id="7888"
            />
            <path
              d="M579.028819 830.007518L1085.44 546.186795l-506.324819-313.615422-2.553832-0.024674L70.199518 534.046843l506.374169 295.936 2.455132 0.024675z m496.701687-284.030458L577.844434 824.998554 79.909012 533.997494l497.886072-296.454169 497.935422 308.433735z"
              fill="#FFFFFF"
              opacity=".4"
              p-id="7889"
            />
            <path
              d="M571.515373 89.988627l565.605784 270.558072V412.06747L571.515373 725.497831 5.921928 398.656771v-44.809253z"
              fill="#1890FF"
              opacity=".6"
              p-id="7890"
            />
            <path
              d="M572.761446 676.123759L1142.068434 360.645398 572.847807 12.028916l-2.541494-0.024675L1.061012 347.148337l569.269976 328.963085 2.430458 0.012337z m559.52347-315.700434L571.589398 671.127133 10.856867 347.098988 571.540048 17.025542l560.73253 343.410121z"
              fill="#FFFFFF"
              opacity=".6"
              p-id="7891"
            />
          </svg>
        </div>
        <span>YSIP</span>
      </div>

      <!-- 中间菜单 -->
      <a-menu
        v-model:selected-keys="selectedKeys"
        theme="light"
        mode="inline"
        :items="menuItems"
        class="menu"
        @click="handleMenuClick"
      />
    </a-layout-sider>

    <!-- 主内容区 -->
    <a-layout class="main-layout">
      <!-- 顶部 Header -->
      <a-layout-header class="header">
        <div class="header-left">
          <div
            class="tabs-wrapper"
            @dblclick="onTabsDblClick"
          >
            <a-tabs
              v-model:active-key="activeTab"
              type="card"
              :tab-bar-style="{ margin: '0', padding: '0' }"
              @change="onTabChange"
            >
              <a-tab-pane
                v-for="tab in openedTabs"
                :key="tab.path"
                :tab="tab.title"
              />
            </a-tabs>
          </div>
        </div>
        <div class="header-right">
          <a-dropdown placement="bottomRight">
            <div class="user-wrapper">
              <div class="user-details">
                <div class="user-name">
                  {{ userName }}
                </div>
                <div class="user-role">
                  {{ userRole }}
                </div>
              </div>
              <a-avatar
                class="user-avatar"
                :size="32"
                :src="userAvatar"
              >
                <template
                  v-if="!userAvatar"
                  #icon
                >
                  {{ userName.charAt(0) }}
                </template>
              </a-avatar>
            </div>
            <template #overlay>
              <a-menu>
                <a-menu-item key="settings">
                  系统设置
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item
                  key="logout"
                  @click="handleLogout"
                >
                  退出登录
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <!-- 内容区 -->
      <a-layout-content class="content">
        <RouterView v-slot="{ Component, route: currentRoute }">
          <keep-alive :include="cachedViews">
            <component
              :is="Component"
              :key="currentRoute.path"
            />
          </keep-alive>
        </RouterView>
        <!-- 底部 -->
        <a-layout-footer class="footer">
          YSIP Web ©2026 Created by YSIP Team
        </a-layout-footer>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup>
import { computed, h, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { useUserStore } from '@/store/user'
import { useTabsStore } from '@/store/tabs'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const tabsStore = useTabsStore()

const menuItems = computed(() => {
  const groupMap = {
    console: { label: '控制台', key: 'console_group' },
    device: { label: '设备管理', key: 'device_group' },
    stream: { label: '流媒体管理', key: 'stream_group' },
    system: { label: '系统管理', key: 'system_group' }
  }

  const groups = {}
  const layoutRoute = router.getRoutes().find(r => r.name === 'Layout')
  const children = layoutRoute?.children || []

  children.forEach(r => {
    if (!r.meta?.title || !r.meta?.group) return
    const { group, icon, title } = r.meta
    if (!groups[group]) {
      groups[group] = {
        key: groupMap[group]?.key || group,
        type: 'group',
        label: groupMap[group]?.label || group,
        children: []
      }
    }
    groups[group].children.push({
      key: r.path,
      icon: icon ? h('img', { src: new URL(icon, import.meta.url).href }) : null,
      label: title,
      title: title
    })
  })

  return Object.values(groups)
})

const selectedKeys = computed(() => [route.path])

// 标签页相关
const openedTabs = computed(() => tabsStore.openedTabs)
const activeTab = computed({
  get: () => tabsStore.activeTab,
  set: val => tabsStore.setActiveTab(val)
})
const cachedViews = computed(() => tabsStore.cachedViews)

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
  if (avatar.startsWith('data:')) return avatar
  return `data:image/png;base64,${avatar}`
})

// 点击菜单
const handleMenuClick = ({ key }) => {
  const route = router.getRoutes().find(r => r.path === key)
  if (route) {
    tabsStore.addTab({
      path: key,
      title: route.meta?.title || route.name,
      name: route.name,
      keepAlive: route.meta?.keepAlive !== false
    })
  }
  router.push(key)
}

// Tab 切换
const onTabChange = path => {
  router.push(path)
}

// tabs 双击事件处理
const onTabsDblClick = e => {
  const tab = e.target.closest('.ant-tabs-tab')
  if (!tab) return
  // 通过 tab 在 DOM 中的位置找到对应的 path
  const tabs = document.querySelectorAll('.ant-tabs-tab')
  const index = Array.from(tabs).indexOf(tab)
  if (index !== -1 && openedTabs.value[index]) {
    const tabPath = openedTabs.value[index].path
    if (tabPath !== '/home') {
      tabsStore.removeTab(tabPath)
      // removeTab 内部已处理激活标签切换逻辑，直接跳转
      router.push(activeTab.value)
    }
  }
}

const handleLogout = async () => {
  await userStore.logout()
  message.success('已退出登录')
  router.push('/login')
}

// 路由变化时自动添加 tab
watch(
  () => route.path,
  path => {
    if (path === '/login') return
    // 只在已恢复状态后才添加 tab，避免覆盖 activeTab
    if (!tabsStore.restored) return
    const routeInfo = router.getRoutes().find(r => r.path === path)
    if (routeInfo && routeInfo.meta?.title) {
      tabsStore.addTab({
        path: path,
        title: routeInfo.meta.title,
        name: routeInfo.name,
        keepAlive: routeInfo.meta.keepAlive !== false
      })
    }
  }
)

// 组件挂载时恢复状态
onMounted(() => {
  userStore.restoreFromStorage()
  tabsStore.restoreFromStorage()

  // 根据当前路由设置 activeTab
  const currentPath = route.path
  const exists = openedTabs.value.find(t => t.path === currentPath)
  tabsStore.setActiveTab(exists ? currentPath : '/home')
})
</script>

<style lang="less" scoped>
@import '@/styles/mixins.less';

.layout {
  height: 100vh;
  overflow: hidden;
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
}

.main-layout {
  width: calc(100vw - 220px);
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.header {
  background: #fff;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  height: 56px;

  .header-left {
    display: flex;
    align-items: center;
    flex: 1;

    .tabs-wrapper {
      flex: 1;
      overflow: hidden;

      :deep(.ant-tabs) {
        .ant-tabs-nav {
          margin: 0;
        }

        .ant-tabs-tab {
          position: relative;
          border: none !important;
          background: transparent;
          padding: 4px 12px;
          margin: 0 6px 0 0;
          font-size: 13px;
          color: #666;
          border-radius: 4px;
          transition: all 0.15s;
          user-select: none;

          &:hover {
            color: #1890ff;
            background: #f0f5ff;
          }
        }

        .ant-tabs-tab-active {
          color: #1890ff !important;
          background: #e6f7ff !important;
        }

        .ant-tabs-nav::before {
          display: none;
        }
      }
    }
  }

  .header-right {
    flex-shrink: 0;

    .user-wrapper {
      margin-right: -8px;
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;
      padding: 6px 10px;
      border-radius: 8px;
      transition: background 0.3s;

      &:hover {
        background: #f5f5f5;
      }

      .user-avatar {
        .user-avatar-style(32px);
      }

      .user-details {
        display: flex;
        flex-direction: column;
        line-height: 1;
        gap: 4px;

        .user-name {
          font-size: 14px;
          font-weight: 500;
          color: #333;
        }

        .user-role {
          font-size: 12px;
          color: #999;
        }
      }
    }
  }
}

.content {
  flex: 1;
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
