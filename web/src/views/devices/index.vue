<template>
  <div class="devices-container">
    <div class="devices-header">
      <div class="header-left">
        <template v-if="mode === 'channels'">
          <a-button @click="mode = 'devices'">
            <template #icon>
              <ArrowLeftOutlined />
            </template>
            返回
          </a-button>
          <span class="mode-title">{{ currentDevice?.name || currentDevice?.deviceid }} - 通道列表</span>
        </template>
        <template v-else>
          <a-button-group>
            <a-button
              :type="batchMode ? 'primary' : 'default'"
              @click="toggleBatchMode"
            >
              <template #icon>
                <CheckOutlined v-if="batchMode" />
                <BorderOutlined v-else />
              </template>
              {{ batchMode ? '退出批量' : '批量操作' }}
            </a-button>
            <a-dropdown
              :disabled="!batchMode || selectedRowKeys.length === 0"
              trigger="click"
            >
              <a-button
                :type="batchMode ? 'primary' : 'default'"
                :disabled="!batchMode || selectedRowKeys.length === 0"
              >
                <template #icon>
                  <DownOutlined />
                </template>
              </a-button>
              <template #overlay>
                <a-menu @click="handleBatchAction">
                  <a-menu-item key="delete">
                    <DeleteOutlined /> 批量删除
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </a-button-group>
          <span
            v-if="batchMode && selectedRowKeys.length > 0"
            class="selection-tip"
          > 已选择 {{ selectedRowKeys.length }} 项 </span>
        </template>
      </div>
      <div class="header-right">
        <a-button
          v-if="mode === 'devices'"
          type="primary"
          @click="openAddModal"
        >
          <template #icon>
            <PlusOutlined />
          </template>
          添加设备
        </a-button>
        <a-button @click="refresh">
          <template #icon>
            <ReloadOutlined />
          </template>
          刷新
        </a-button>
      </div>
    </div>

    <!-- 设备列表 -->
    <a-table
      v-if="mode === 'devices'"
      :columns="deviceColumns"
      :data-source="devices"
      :loading="loading"
      :pagination="pagination"
      :row-selection="batchMode ? { selectedRowKeys, onChange: handleSelectionChange } : null"
      :row-key="record => record.id"
      class="devices-table"
      @change="handleTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <span class="device-name-cell">
            <PlaySquareOutlined class="device-icon" />
            <span class="device-name">{{ record.name || record.deviceid }}</span>
          </span>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-badge
            :status="getDeviceStatus(record).badge"
            :text="getDeviceStatus(record).text"
            class="status-badge"
            @click="openDetailModal(record)"
          />
        </template>
        <template v-else-if="column.key === 'info'">
          <span class="info-cell">
            <span
              v-if="record.manufacturer"
              class="info-item"
            >{{ record.manufacturer }}</span>
            <span
              v-if="record.model"
              class="info-item"
            >{{ record.model }}</span>
            <span
              v-if="record.devicetype"
              class="info-item"
            >{{ record.devicetype }}</span>
          </span>
        </template>
        <template v-else-if="column.key === 'sipip'">
          <span class="host-cell">{{ record.sipip || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'streamip'">
          <span class="host-cell">{{ record.streamip || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'source'">
          <span class="host-cell">{{ record.source || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'active'">
          <span class="time-cell">{{ formatTime(record.active) }}</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space :size="0">
            <a-button
              type="link"
              size="small"
              :loading="record.syncing"
              @click="handleSync(record)"
            >
              同步
            </a-button>
            <a-button
              type="link"
              size="small"
              @click="openChannelsModal(record)"
            >
              通道
            </a-button>
            <a-button
              type="link"
              size="small"
              @click="openEditModal(record)"
            >
              编辑
            </a-button>
            <a-button
              type="link"
              danger
              size="small"
              :loading="record.deleting"
              @click="handleDelete(record)"
            >
              删除
            </a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- 通道列表 -->
    <a-table
      v-else
      :columns="channelColumns"
      :data-source="channels"
      :loading="loading"
      :pagination="pagination"
      :row-key="record => record.id"
      class="devices-table"
      @change="handleChannelTableChange"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'name'">
          <span class="device-name-cell">
            <VideoCameraOutlined class="device-icon" />
            <span class="device-name">{{ record.name || record.channelid }}</span>
          </span>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-badge
            :status="record.status === 'ON' ? 'success' : 'default'"
            :text="record.status === 'ON' ? '在线' : '离线'"
          />
        </template>
        <template v-else-if="column.key === 'info'">
          <span class="info-cell">
            <span
              v-if="record.manufacturer"
              class="info-item"
            >{{ record.manufacturer }}</span>
            <span
              v-if="record.model"
              class="info-item"
            >{{ record.model }}</span>
          </span>
        </template>
        <template v-else-if="column.key === 'streamtype'">
          <a-tag :color="record.streamtype === 'pull' ? 'blue' : 'green'">
            {{ record.streamtype === 'pull' ? '拉流' : '推流' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'address'">
          <span class="address-cell">{{ record.address || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'active'">
          <span class="time-cell">{{ formatTime(record.active) }}</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space :size="0">
            <a-button
              type="link"
              size="small"
              :loading="record.starting"
              @click="handleStartLive(record)"
            >
              直播
            </a-button>
            <a-button
              type="link"
              size="small"
              danger
              :loading="record.stopping"
              @click="handleStopLive(record)"
            >
              停止
            </a-button>
            <a-button
              type="link"
              size="small"
              disabled
            >
              回放
            </a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <DeviceForm
      v-model="editModalVisible"
      type="edit"
      :device="currentDevice"
      @success="handleEditSuccess"
    />
    <DeviceForm
      v-model="addModalVisible"
      type="add"
      @success="handleAddSuccess"
    />

    <!-- 设备详情弹窗 -->
    <DeviceDetailModal
      v-model:open="detailModalVisible"
      :device="detailDevice"
    />

    <!-- 直播弹窗 -->
    <a-modal
      v-model:open="liveModalVisible"
      title="直播"
      :footer="null"
      width="800px"
      :destroy-on-close="true"
    >
      <div class="live-player-wrapper">
        <LivePlayer
          v-if="liveUrl"
          :url="liveUrl"
        />
        <div
          v-else
          class="live-loading"
        >
          <a-spin size="large" />
          <span>正在开启直播...</span>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { getDevices, createDevice, deleteDevice, syncChannels } from '@/api/device'
import { getChannels } from '@/api/channel'
import { startStream, stopStream } from '@/api/stream'
import { message } from 'ant-design-vue'
import { DownOutlined, ReloadOutlined, CheckOutlined, BorderOutlined, PlusOutlined, DeleteOutlined, PlaySquareOutlined, ArrowLeftOutlined, VideoCameraOutlined } from '@ant-design/icons-vue'
import { LivePlayer } from '@/components'
import DeviceForm from './DeviceForm.vue'
import DeviceDetailModal from './DeviceDetailModal.vue'
import request from '@/api/request'

const mode = ref('devices')
const devices = ref([])
const channels = ref([])
const loading = ref(false)
const batchMode = ref(false)
const selectedRowKeys = ref([])

const currentDevice = ref(null)

const editModalVisible = ref(false)
const addModalVisible = ref(false)
const detailModalVisible = ref(false)
const detailDevice = ref(null)
const liveModalVisible = ref(false)
const liveUrl = ref('')
const currentChannel = ref(null)

const statusFilter = ref(null)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`
})

const deviceColumns = computed(() => [
  {
    title: '设备名称',
    key: 'name',
    width: 180,
    ellipsis: true
  },
  {
    title: '设备ID',
    dataIndex: 'deviceid',
    key: 'deviceid',
    width: 180,
    ellipsis: true
  },
  {
    title: '状态',
    key: 'status',
    width: 90
  },
  {
    title: '设备信息',
    key: 'info',
    width: 180,
    ellipsis: true
  },
  {
    title: '源地址',
    key: 'source',
    width: 160,
    ellipsis: true
  },
  {
    title: '最后活跃',
    key: 'active',
    width: 150
  },
  {
    title: '操作',
    key: 'action',
    width: 160,
    fixed: 'right'
  }
])

const channelColumns = computed(() => [
  {
    title: '通道名称',
    key: 'name',
    width: 200,
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
    title: '状态',
    key: 'status',
    width: 80
  },
  {
    title: '设备信息',
    key: 'info',
    width: 150,
    ellipsis: true
  },
  {
    title: '播放类型',
    key: 'streamtype',
    width: 88
  },
  {
    title: '最后活跃',
    key: 'active',
    width: 150
  },
  {
    title: '操作',
    key: 'action',
    width: 140,
    fixed: 'right'
  }
])

const getDeviceStatus = record => {
  if (!record.regist) {
    return { badge: 'default', text: '未注册' }
  }
  if (record.online) {
    return { badge: 'success', text: '在线' }
  }
  return { badge: 'error', text: '离线' }
}

const formatTime = timestamp => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
}

const refresh = () => {
  if (mode.value === 'devices') {
    fetchDevices()
  } else {
    fetchChannels()
  }
}

const fetchDevices = async () => {
  loading.value = true
  try {
    const params = {
      limit: pagination.pageSize,
      skip: (pagination.current - 1) * pagination.pageSize
    }
    if (statusFilter.value !== null) {
      params.filters = JSON.stringify([{ field_name: 'regist', opertator: '=', value: statusFilter.value }])
    }
    const res = await getDevices(params)
    devices.value = (res.data?.list || []).map(item => ({ ...item, deleting: false, syncing: false }))
    pagination.total = res.data?.total || 0
  } catch (error) {
    // error handled by request interceptor
  } finally {
    loading.value = false
  }
}

const fetchChannels = async () => {
  if (!currentDevice.value) return
  loading.value = true
  try {
    const params = {
      limit: pagination.pageSize,
      skip: (pagination.current - 1) * pagination.pageSize,
      filters: JSON.stringify([{ field_name: 'deviceid', opertator: '=', value: currentDevice.value.deviceid }])
    }
    const res = await getChannels(params)
    channels.value = (res.data?.list || []).map(item => ({ ...item, starting: false, stopping: false, hasStream: false }))
    pagination.total = res.data?.total || 0
  } catch (error) {
    // error handled by request interceptor
  } finally {
    loading.value = false
  }
}

watch(statusFilter, () => {
  pagination.current = 1
  fetchDevices()
})

const handleTableChange = (pag, filters) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  if (filters.regist && filters.regist.length > 0) {
    statusFilter.value = filters.regist[0]
  }
  fetchDevices()
}

const handleChannelTableChange = pag => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchChannels()
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

const openChannelsModal = record => {
  currentDevice.value = record
  pagination.current = 1
  mode.value = 'channels'
  fetchChannels()
}

const openEditModal = record => {
  currentDevice.value = record
  editModalVisible.value = true
}

const handleEditSuccess = async formData => {
  try {
    const data = { name: formData.name }
    if (formData.pwd) {
      data.pwd = formData.pwd
    }
    if (formData.host !== undefined) {
      data.host = formData.host
    }
    if (formData.sipip !== undefined) {
      data.sipip = formData.sipip
    }
    if (formData.streamip !== undefined) {
      data.streamip = formData.streamip
    }
    if (formData.manufacturer !== undefined) {
      data.manufacturer = formData.manufacturer
    }
    if (formData.model !== undefined) {
      data.model = formData.model
    }
    // 订阅设置
    if (formData.subscribe) {
      data.subscribe = formData.subscribe
    }
    await request({
      url: `/v1/devices/${formData.deviceid}`,
      method: 'post',
      data
    })
    message.success('设备更新成功')
    editModalVisible.value = false
    fetchDevices()
  } catch (error) {
    // error handled by request interceptor
  }
}

const openAddModal = () => {
  addModalVisible.value = true
}

const openDetailModal = record => {
  detailDevice.value = record
  detailModalVisible.value = true
}

const handleAddSuccess = async formData => {
  try {
    const data = { name: formData.name, pwd: formData.pwd }
    if (formData.deviceId) {
      data.deviceId = formData.deviceId
    }
    await createDevice(data)
    message.success('设备创建成功')
    addModalVisible.value = false
    fetchDevices()
  } catch (error) {
    // error handled by request interceptor
  }
}

const handleDelete = async record => {
  record.deleting = true
  try {
    await deleteDevice(record.deviceid)
    message.success('设备删除成功')
    fetchDevices()
  } catch (error) {
    // error handled by request interceptor
  } finally {
    record.deleting = false
  }
}

const handleSync = async record => {
  record.syncing = true
  try {
    await syncChannels(record.deviceid)
    message.success('通道同步请求已发送')
  } catch (error) {
    // error handled by request interceptor
  } finally {
    record.syncing = false
  }
}

const handleStartLive = async record => {
  if (record.status !== 'ON') {
    message.warning('设备离线，无法开启直播')
    return
  }
  record.starting = true
  liveUrl.value = ''
  currentChannel.value = record
  liveModalVisible.value = true
  try {
    const res = await startStream(record.channelid, {})
    if (res.data) {
      liveUrl.value = res.data.wsflv || res.data.rtmp || ''
      // 保存 streamid 用于停止直播
      record.streamid = res.data.streamid
      record.hasStream = true
      if (!liveUrl.value) {
        message.warning('未获取到播放地址')
      }
    }
  } catch (error) {
    liveModalVisible.value = false
  } finally {
    record.starting = false
  }
}

const handleStopLive = async record => {
  if (record.status !== 'ON') {
    message.warning('设备离线')
    return
  }
  record.stopping = true
  try {
    // 使用 streamid 而不是 channelid
    const streamId = record.streamid || record.channelid
    await stopStream(streamId)
    message.success('直播已停止')
    record.hasStream = false
    record.streamid = null
    liveModalVisible.value = false
    liveUrl.value = ''
    fetchChannels()
  } catch (error) {
    // 响应拦截器已显示错误消息，这里不需要重复显示
  } finally {
    record.stopping = false
  }
}

const handleBatchAction = async ({ key }) => {
  if (key === 'delete') {
    handleBatchDelete()
  }
}

const handleBatchDelete = async () => {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择设备')
    return
  }
  message.info('批量删除功能开发中')
  selectedRowKeys.value = []
  batchMode.value = false
}

onMounted(() => {
  fetchDevices()
})
</script>

<style lang="less" scoped>
.devices-container {
  margin: 16px 16px 0;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;

  .devices-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 0 16px 0px;

    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .mode-title {
        font-weight: 500;
        color: #262626;
      }

      .selection-tip {
        font-size: 13px;
        color: #1890ff;
        font-weight: 500;
      }
    }

    .header-right {
      display: flex;
      gap: 8px;

      :deep(.ant-btn) {
        font-weight: 500;
      }
    }
  }

  .devices-table {
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

  .device-name-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .device-icon {
      font-size: 16px;
      color: #1890ff;
    }

    .device-name {
      font-weight: 500;
      color: #262626;
    }
  }

  .info-cell {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;

    .info-item {
      font-size: 12px;
      color: #8c8c8c;
      background: #f5f5f5;
      padding: 2px 6px;
      border-radius: 4px;
    }
  }

  .host-cell {
    font-family: 'Fira Code', 'Consolas', monospace;
    font-size: 12px;
    color: #595959;
  }

  .address-cell {
    font-family: 'Fira Code', 'Consolas', monospace;
    font-size: 12px;
    color: #595959;
  }

  .time-cell {
    font-size: 12px;
    color: #8c8c8c;
  }
}

.status-badge {
  cursor: pointer;
  &:hover {
    background-color: #f5f5f5;
  }
}

.live-player-wrapper {
  width: 100%;
  height: 450px;
  background: #000;
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;

  .live-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    color: #8c8c8c;
  }
}
</style>
