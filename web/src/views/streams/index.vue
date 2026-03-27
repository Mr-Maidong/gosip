<template>
  <div class="streams-container">
    <div class="page-header">
      <a-space :size="16">
        <a-button-group>
          <a-button :type="batchMode ? 'primary' : 'default'" @click="toggleBatchMode">
            {{ batchMode ? '取消批量' : '批量操作' }}
          </a-button>
          <a-dropdown trigger="click" :disabled="!batchMode || selectedRowKeys.length === 0">
            <a-button :type="batchMode ? 'primary' : 'default'" :disabled="!batchMode || selectedRowKeys.length === 0">
              <DownOutlined />
            </a-button>
            <template #overlay>
              <a-menu @click="handleBatchAction">
                <a-menu-item key="stop"> 批量停止 </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </a-button-group>
        <a-button @click="fetchStreams"> 刷新 </a-button>
      </a-space>
    </div>

    <a-table
      :columns="columns"
      :data-source="streams"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ y: 622 }"
      :row-selection="{
        selectedRowKeys,
        onChange: handleSelectionChange,
        getCheckboxProps: () => ({
          disabled: !batchMode
        })
      }"
      :row-key="record => record.id"
      @change="handleTableChange"
    >
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
          </a-space>
        </template>
        <template v-else-if="column.key === 'streamid'">
          <span class="stream-id">{{ record.streamid || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'address'">
          <a-space direction="vertical" :size="0" class="address-list">
            <span v-if="record.rtmp" class="address-item">
              <span class="address-label">RTMP:</span>
              <a-button type="link" class="address-link">{{ truncate(record.rtmp, 40) }}</a-button>
            </span>
          </a-space>
        </template>
        <template v-else-if="column.key === 'time'">
          <a-space direction="vertical" :size="2">
            <span>更新: {{ formatTime(record.uptime) }}</span>
          </a-space>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-button type="link" danger size="small" :loading="record.stopping" :disabled="record.status === 1" @click="handleStop(record)"> 停止 </a-button>
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { getStreams, stopStream } from '@/api/stream'
import { message } from 'ant-design-vue'
import { DownOutlined } from '@ant-design/icons-vue'

const streams = ref([])
const loading = ref(false)
const batchMode = ref(false)
const selectedRowKeys = ref([])
const batchStopping = ref(false)

const statusFilter = ref(0)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`
})

const columns = computed(() => [
  {
    title: '设备ID',
    dataIndex: 'deviceid',
    key: 'deviceid',
    width: 160,
    ellipsis: true
  },
  {
    title: '通道ID',
    dataIndex: 'channelid',
    key: 'channelid',
    width: 160,
    ellipsis: true
  },
  {
    title: '流ID',
    key: 'streamid',
    width: 120,
    ellipsis: true
  },
  {
    title: 'ZLM流',
    key: 'stream',
    width: 68
  },
  {
    title: '类型',
    key: 'type',
    width: 60
  },
  {
    title: '播放地址',
    key: 'address',
    width: 300,
    ellipsis: true
  },
  {
    title: '状态',
    key: 'status',
    width: 72,
    filters: [
      { text: '正常', value: 0 },
      { text: '关闭', value: 1 },
      { text: '未开始', value: -1 }
    ],
    filterMultiple: false,
    filteredValue: statusFilter.value !== null ? [statusFilter.value] : null,
    onFilter: (value, record) => record.status === value
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
])

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
    if (statusFilter.value !== null) {
      const filters = [{ field_name: 'status', opertator: '=', value: statusFilter.value }]
      params.filters = JSON.stringify(filters)
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

watch(statusFilter, () => {
  pagination.current = 1
  fetchStreams()
})

const handleTableChange = (pag, filters) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  if (filters.status && filters.status.length > 0) {
    statusFilter.value = filters.status[0]
  }
  fetchStreams()
}

const handleSelectionChange = keys => {
  selectedRowKeys.value = keys
}

const clearSelection = () => {
  selectedRowKeys.value = []
}

const toggleBatchMode = () => {
  batchMode.value = !batchMode.value
  if (!batchMode.value) {
    selectedRowKeys.value = []
  }
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

const handleBatchAction = async ({ key }) => {
  if (key === 'stop') {
    handleBatchStop()
  }
}

const handleBatchStop = async () => {
  const selectedStreams = streams.value.filter(s => selectedRowKeys.value.includes(s.id))
  if (selectedStreams.length === 0) {
    message.warning('请先选择流')
    return
  }
  batchStopping.value = true
  let successCount = 0
  let failCount = 0
  for (const stream of selectedStreams) {
    try {
      await stopStream(stream.streamid)
      successCount++
    } catch (error) {
      if (error.code === 1002) {
        successCount++
      } else {
        failCount++
      }
    }
  }
  batchStopping.value = false
  selectedRowKeys.value = []
  if (failCount === 0) {
    message.success(`已停止 ${successCount} 个流`)
  } else {
    message.warning(`停止 ${successCount} 个流，失败 ${failCount} 个`)
  }
  fetchStreams()
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
    justify-content: start;
    align-items: center;
    margin-bottom: 24px;
  }

  .selection-info {
    margin-bottom: 12px;
    padding: 8px 12px;
    background: #f5f5f5;
    border-radius: 4px;
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 13px;

    a {
      color: #1890ff;
      cursor: pointer;
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

  .address-label {
    color: #8c8c8c;
  }
}
</style>
