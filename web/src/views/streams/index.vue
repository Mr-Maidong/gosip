<template>
  <div class="streams-container">
    <div class="page-header">
      <h2>流管理</h2>
      <a-space>
        <a-button type="primary" @click="fetchStreams"> 刷新 </a-button>
      </a-space>
    </div>

    <a-table :columns="columns" :data-source="streams" :loading="loading" :pagination="pagination" :scroll="{ y: 622 }" row-key="id" @change="handleTableChange">
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record.status)">
            {{ getStatusText(record.status) }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'stream'">
          <a-tag :color="record.stream ? 'success' : 'default'">
            {{ record.stream ? '已接收' : '未接收' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'type'">
          <a-space direction="vertical" :size="0">
            <span>{{ record.t === 0 ? '直播' : '回放' }}</span>
            <span class="type-sub">{{ record.streamtype === 'pull' ? '拉流' : '推流' }}</span>
          </a-space>
        </template>
        <template v-else-if="column.key === 'streamid'">
          <span class="stream-id">{{ record.streamid || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'address'">
          <a-space direction="vertical" :size="2" class="address-list">
            <span v-if="record.rtmp" class="address-item">
              <span class="address-label">RTMP:</span>
              <a :href="record.rtmp" target="_blank" class="address-link">{{ truncate(record.rtmp, 30) }}</a>
            </span>
            <span v-if="record.wsflv" class="address-item">
              <span class="address-label">WS-FLV:</span>
              <a :href="record.wsflv" target="_blank" class="address-link">{{ truncate(record.wsflv, 30) }}</a>
            </span>
          </a-space>
        </template>
        <template v-else-if="column.key === 'time'">
          <a-space direction="vertical" :size="2">
            <span>添加: {{ formatTime(record.addtime) }}</span>
            <span>更新: {{ formatTime(record.uptime) }}</span>
          </a-space>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-button type="link" danger size="small" :loading="record.stopping" @click="handleStop(record)"> 停止 </a-button>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getStreams, stopStream } from '@/api/stream'
import { message } from 'ant-design-vue'

const streams = ref([])
const loading = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`
})

const columns = [
  {
    title: '状态',
    key: 'status',
    width: 90,
    fixed: 'left'
  },
  {
    title: 'ZLM流',
    key: 'stream',
    width: 90
  },
  {
    title: '类型',
    key: 'type',
    width: 90
  },
  {
    title: '设备ID',
    dataIndex: 'deviceid',
    key: 'deviceid',
    width: 180,
    ellipsis: true
  },
  {
    title: '通道ID',
    dataIndex: 'channelid',
    key: 'channelid',
    width: 180,
    ellipsis: true
  },
  {
    title: '流ID',
    key: 'streamid',
    width: 150,
    ellipsis: true
  },
  {
    title: '播放地址',
    key: 'address',
    width: 220
  },
  {
    title: '时间信息',
    key: 'time',
    width: 160
  },
  {
    title: '操作',
    key: 'action',
    width: 80,
    fixed: 'right'
  }
]

const getStatusColor = status => {
  const colorMap = { 0: 'success', 1: 'error', '-1': 'warning' }
  return colorMap[String(status)] || 'default'
}

const getStatusText = status => {
  const textMap = { 0: '正常', 1: '关闭', '-1': '未开始' }
  return textMap[String(status)] || '未知'
}

const truncate = (str, len) => {
  if (!str) return '-'
  return str.length > len ? str.substring(0, len) + '...' : str
}

const formatTime = timestamp => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
}

const fetchStreams = async () => {
  loading.value = true
  try {
    const params = {
      limit: pagination.pageSize,
      skip: (pagination.current - 1) * pagination.pageSize
    }
    const res = await getStreams(params)
    streams.value = (res.data?.list || []).map(item => ({ ...item, stopping: false }))
    pagination.total = res.data?.total || 0
  } catch (error) {
    // error handled by request interceptor
  } finally {
    loading.value = false
  }
}

const handleTableChange = pag => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchStreams()
}

const handleStop = async record => {
  record.stopping = true
  try {
    await stopStream(record.streamid)
    message.success('流已停止')
    fetchStreams()
  } catch (error) {
    if (error.code === 1002) {
      const stream = streams.value.find(s => s.id === record.id)
      if (stream) {
        stream.status = 1
        stream.stop = true
        stream.stream = false
      }
      message.warning('流已关闭')
    }
  } finally {
    record.stopping = false
  }
}

onMounted(() => {
  fetchStreams()
})
</script>

<style lang="less" scoped>
.streams-container {
  margin: 18px 18px 0;
  padding: 16px 16px 0;
  background: #fff;
  border-radius: 8px;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;

    h2 {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
    }
  }

  .type-sub {
    font-size: 12px;
    color: #8c8c8c;
  }

  .stream-id {
    font-family: monospace;
    font-size: 12px;
  }

  .address-list {
    max-width: 260px;
  }

  .address-item {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    line-height: 1.4;
  }

  .address-label {
    color: #8c8c8c;
    flex-shrink: 0;
  }

  .address-link {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
