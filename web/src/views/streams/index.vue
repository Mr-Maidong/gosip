<template>
  <div class="streams-container">
    <div class="streams-header">
      <div class="header-left">
        <a-button-group>
          <a-button :type="batchMode ? 'primary' : 'default'" :disabled="streams.length === 0" @click="toggleBatchMode">
            <template #icon>
              <StopOutlined v-if="batchMode" />
              <BorderOutlined v-else />
            </template>
            {{ batchMode ? '退出批量' : '批量操作' }}
          </a-button>
          <a-dropdown :disabled="!batchMode || selectedRowKeys.length === 0" trigger="click">
            <a-button :type="batchMode ? 'primary' : 'default'" :disabled="!batchMode || selectedRowKeys.length === 0">
              <template #icon>
                <DownOutlined />
              </template>
            </a-button>
            <template #overlay>
              <a-menu @click="handleBatchAction">
                <a-menu-item key="stop"> <StopOutlined /> 批量停止 </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </a-button-group>
        <span v-if="batchMode && selectedRowKeys.length > 0" class="selection-tip"> 已选择 {{ selectedRowKeys.length }} 项 </span>
      </div>
      <div class="header-right">
        <a-button @click="fetchStreams">
          <template #icon>
            <ReloadOutlined />
          </template>
          刷新
        </a-button>
      </div>
    </div>

    <a-table
      :columns="columns"
      :data-source="streams"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ y: `calc(100vh - 334px)` }"
      :row-selection="{
        selectedRowKeys,
        onChange: handleSelectionChange,
        getCheckboxProps: () => ({ disabled: !batchMode })
      }"
      :row-key="record => record.id"
      class="streams-table"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'streamid'">
          <span class="stream-id-cell">
            <span class="stream-id-text">{{ record.streamid || '-' }}</span>
          </span>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-badge :status="getStatusBadge(record.status)" :text="getStatusText(record.status)" />
        </template>
        <template v-else-if="column.key === 'stream'">
          <a-tag :color="record.stream ? 'success' : 'default'" class="stream-tag">
            {{ record.stream ? '已接收' : '未接收' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'type'">
          <span class="type-cell">
            <PlayCircleOutlined v-if="record.t === 0" class="type-icon live" />
            <HistoryOutlined v-else class="type-icon replay" />
            {{ record.t === 0 ? '直播' : '回放' }}
          </span>
        </template>
        <template v-else-if="column.key === 'address'">
          <div class="address-cell">
            <div v-if="record.rtmp" class="address-item">
              <span class="address-text">{{ record.rtmp }}</span>
            </div>
          </div>
        </template>
        <template v-else-if="column.key === 'time'">
          <span class="time-cell">
            {{ formatTime(record.uptime) }}
          </span>
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
import { DownOutlined, ReloadOutlined, BorderOutlined, StopOutlined, PlayCircleOutlined, HistoryOutlined } from '@ant-design/icons-vue'

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
    title: '流ID',
    key: 'streamid',
    width: 140,
    ellipsis: true
  },
  {
    title: '设备 / 通道',
    key: 'device',
    width: 280,
    ellipsis: true,
    customRender: ({ record }) => `${record.deviceid} / ${record.channelid}`
  },
  {
    title: '类型',
    key: 'type',
    width: 90
  },
  {
    title: 'ZLM流',
    key: 'stream',
    width: 80
  },
  {
    title: '播放地址',
    key: 'address',
    width: 280,
    ellipsis: true
  },
  {
    title: '更新时间',
    key: 'time',
    width: 150
  },
  {
    title: '状态',
    key: 'status',
    width: 90,
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
    title: '操作',
    key: 'action',
    width: 80,
    fixed: 'right'
  }
])

const getStatusBadge = status => {
  const map = { 0: 'success', 1: 'error', '-1': 'warning' }
  return map[String(status)] || 'default'
}

const getStatusText = status => {
  const textMap = { 0: '正常', 1: '关闭', '-1': '未开始' }
  return textMap[String(status)] || '未知'
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
  margin: 16px 16px 0;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;

  .streams-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 0 16px 0px;

    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .selection-tip {
        font-size: 13px;
        color: #1890ff;
        font-weight: 500;
      }
    }

    .header-right {
      :deep(.ant-btn) {
        font-weight: 500;
      }
    }
  }

  .streams-table {
    :deep(.ant-table) {
      .ant-table-thead > tr > th {
        background: #fafafa;
        font-weight: 600;
        font-size: 13px;
        color: #262626;
      }

      .ant-table-tbody > tr > td {
        font-size: 13px;
      }

      .ant-table-tbody > tr:hover > td {
        background: #f5f5f5;
      }
    }
  }

  .stream-id-cell {
    .stream-id-text {
      font-family: 'Fira Code', 'Consolas', monospace;
      font-size: 12px;
      color: #595959;
      background: #f5f5f5;
      padding: 2px 6px;
      border-radius: 4px;
    }
  }

  .stream-tag {
    border: none;
    font-weight: 500;
  }

  .type-cell {
    display: flex;
    align-items: center;
    gap: 6px;

    .type-icon {
      font-size: 14px;

      &.live {
        color: #52c41a;
      }

      &.replay {
        color: #1890ff;
      }
    }
  }

  .address-cell {
    .address-item {
      .address-text {
        font-size: 13px;
        color: #8c8c8c;
      }
    }
  }

  .time-cell {
    font-size: 12px;
    color: #8c8c8c;
  }
}
</style>
