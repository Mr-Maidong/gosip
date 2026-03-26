<template>
  <div class="welcome-card">
    <div class="card-header">
      <div class="user-info">
        <a-avatar :size="48" :src="userAvatar" class="user-avatar">
          <template #icon>
            <UserOutlined />
          </template>
        </a-avatar>
        <div class="user-details">
          <h2 class="greeting">{{ greeting }}，{{ userStore.userInfo?.name || userStore.userInfo?.username }}</h2>
          <span class="meta">
            <a-tag :color="roleColor">{{ roleText }}</a-tag>
            <span class="date">{{ currentDate }}</span>
          </span>
        </div>
      </div>
    </div>

    <div class="card-body">
      <div class="stats-row">
        <div class="stat-item" @click="$router.push('/platform')">
          <span class="stat-value">{{ stats.platform }}</span>
          <span class="stat-label">平台</span>
        </div>
        <div class="stat-divider" />
        <div class="stat-item" @click="$router.push('/devices')">
          <span class="stat-value">{{ stats.devices }}</span>
          <span class="stat-label">设备</span>
        </div>
        <div class="stat-divider" />
        <div class="stat-item" @click="$router.push('/streams')">
          <span class="stat-value">{{ stats.streams }}</span>
          <span class="stat-label">在线流</span>
        </div>
      </div>

      <div class="quick-actions">
        <a-button type="primary" @click="$router.push('/devices')"> 添加设备 </a-button>
        <a-button @click="$router.push('/streams')"> 查看流 </a-button>
        <a-button @click="$router.push('/settings')"> 系统设置 </a-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/store/user'
import { getDevices } from '@/api/device'
import { getChannels } from '@/api/channel'
import { getStreams } from '@/api/stream'
import { UserOutlined } from '@ant-design/icons-vue'

const userStore = useUserStore()

const userAvatar = computed(() => {
  const avatar = userStore.userInfo?.avatar
  if (!avatar) return ''
  if (avatar.startsWith('data:')) return avatar
  return `data:image/png;base64,${avatar}`
})

const stats = ref({
  platform: 0,
  devices: 0,
  streams: 0
})

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '凌晨好'
  if (hour < 9) return '早上好'
  if (hour < 12) return '上午好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  if (hour < 22) return '晚上好'
  return '夜里好'
})

const roleText = computed(() => {
  const roleMap = { admin: '管理员', operator: '操作员', viewer: '观察者' }
  return roleMap[userStore.userInfo?.role] || '用户'
})

const roleColor = computed(() => {
  const colorMap = { admin: 'red', operator: 'blue', viewer: 'green' }
  return colorMap[userStore.userInfo?.role] || 'default'
})

const currentDate = computed(() => {
  const now = new Date()
  const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日 ${weekDays[now.getDay()]}`
})

const fetchStats = async () => {
  try {
    const [devicesRes, channelsRes, streamsRes] = await Promise.all([getDevices({ limit: 1 }), getChannels({ limit: 1 }), getStreams({ limit: 1 })])
    stats.value = {
      platform: channelsRes.total || 0,
      devices: devicesRes.total || 0,
      streams: streamsRes.total || 0
    }
  } catch (error) {
    // error handled by request interceptor
  }
}

onMounted(() => {
  fetchStats()
})
</script>

<style lang="less" scoped>
@import '@/styles/mixins.less';

.welcome-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px 32px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);

  .card-header {
    margin-bottom: 24px;

    .user-info {
      display: flex;
      align-items: center;
      gap: 14px;

      .user-avatar {
        .user-avatar-style(48px);
      }

      .user-details {
        .greeting {
          font-size: 18px;
          font-weight: 600;
          color: #262626;
          margin: 0 0 4px 0;
          line-height: 1.4;
        }

        .meta {
          display: flex;
          align-items: center;
          gap: 10px;

          .ant-tag {
            margin: 0;
            border: none;
          }

          .date {
            font-size: 13px;
            color: #8c8c8c;
          }
        }
      }
    }
  }

  .card-body {
    .stats-row {
      display: flex;
      align-items: center;
      padding: 20px 0;
      margin-bottom: 20px;
      border-top: 1px solid #f0f0f0;
      border-bottom: 1px solid #f0f0f0;

      .stat-item {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        cursor: pointer;
        padding: 8px 16px;
        border-radius: 6px;
        transition: background 0.2s;

        &:hover {
          background: #f5f5f5;
        }

        .stat-value {
          font-size: 28px;
          font-weight: 600;
          color: #1890ff;
          line-height: 1.2;
        }

        .stat-label {
          font-size: 13px;
          color: #8c8c8c;
          margin-top: 4px;
        }
      }

      .stat-divider {
        width: 1px;
        height: 40px;
        background: #e8e8e8;
      }
    }

    .quick-actions {
      display: flex;
      gap: 12px;

      :deep(.ant-btn) {
        border-radius: 4px;
        font-weight: 500;
      }
    }
  }
}

@media (max-width: 640px) {
  .welcome-card {
    padding: 20px;

    .card-body {
      .stats-row {
        flex-wrap: wrap;
        gap: 12px;

        .stat-item {
          flex: 1 1 45%;
        }

        .stat-divider {
          display: none;
        }
      }

      .quick-actions {
        flex-wrap: wrap;

        :deep(.ant-btn) {
          flex: 1;
          min-width: 100px;
        }
      }
    }
  }
}
</style>
