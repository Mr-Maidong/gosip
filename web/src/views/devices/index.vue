<template>
  <div class="devices-container">
    <div class="devices-header">
      <div class="header-left">
        <a-button-group>
          <a-button :type="batchMode ? 'primary' : 'default'" @click="toggleBatchMode">
            <template #icon>
              <CheckOutlined v-if="batchMode" />
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
                <a-menu-item key="delete"> <DeleteOutlined /> 批量删除 </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </a-button-group>
        <span v-if="batchMode && selectedRowKeys.length > 0" class="selection-tip"> 已选择 {{ selectedRowKeys.length }} 项 </span>
      </div>
      <div class="header-right">
        <a-button type="primary" @click="openAddModal">
          <template #icon>
            <PlusOutlined />
          </template>
          添加设备
        </a-button>
        <a-button @click="fetchDevices">
          <template #icon>
            <ReloadOutlined />
          </template>
          刷新
        </a-button>
      </div>
    </div>

    <a-table
      :columns="columns"
      :data-source="devices"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ y: 540 }"
      :row-selection="
        batchMode
          ? {
              selectedRowKeys,
              onChange: handleSelectionChange
            }
          : null
      "
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
        <template v-else-if="column.key === 'regist'">
          <a-badge :status="record.regist ? 'success' : 'default'" :text="record.regist ? '已注册' : '未注册'" />
        </template>
        <template v-else-if="column.key === 'info'">
          <span class="info-cell">
            <span v-if="record.manufacturer" class="info-item">{{ record.manufacturer }}</span>
            <span v-if="record.model" class="info-item">{{ record.model }}</span>
            <span v-if="record.devicetype" class="info-item">{{ record.devicetype }}</span>
          </span>
        </template>
        <template v-else-if="column.key === 'host'">
          <span class="host-cell">{{ record.host || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'address'">
          <span class="address-cell">{{ record.host }}:{{ record.port }}</span>
        </template>
        <template v-else-if="column.key === 'active'">
          <span class="time-cell">{{ formatTime(record.active) }}</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="openEditModal(record)"> 编辑 </a-button>
            <a-button type="link" danger size="small" :loading="record.deleting" @click="handleDelete(record)"> 删除 </a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <DeviceForm v-model="editModalVisible" type="edit" :device="currentDevice" @success="handleEditSuccess" />
    <DeviceForm v-model="addModalVisible" type="add" @success="handleAddSuccess" />
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { getDevices, createDevice, deleteDevice } from '@/api/device'
import { message } from 'ant-design-vue'
import { DownOutlined, ReloadOutlined, CheckOutlined, BorderOutlined, PlusOutlined, DeleteOutlined, PlaySquareOutlined } from '@ant-design/icons-vue'
import DeviceForm from './DeviceForm.vue'
import request from '@/api/request'

const devices = ref([])
const loading = ref(false)
const batchMode = ref(false)
const selectedRowKeys = ref([])

const editModalVisible = ref(false)
const addModalVisible = ref(false)
const currentDevice = ref(null)

const statusFilter = ref(null)

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
    title: '设备名称',
    key: 'name',
    width: 200,
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
    title: '注册状态',
    key: 'regist',
    width: 90
  },
  {
    title: '设备信息',
    key: 'info',
    width: 180,
    ellipsis: true
  },
  {
    title: '收流地址',
    key: 'host',
    width: 140,
    ellipsis: true
  },
  {
    title: '地址',
    key: 'address',
    width: 180,
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
    width: 120,
    fixed: 'right'
  }
])

const formatTime = timestamp => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-')
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
    devices.value = (res.data?.list || []).map(item => ({ ...item, deleting: false }))
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

const handleSelectionChange = keys => {
  selectedRowKeys.value = keys
}

const toggleBatchMode = () => {
  batchMode.value = !batchMode.value
  if (!batchMode.value) {
    selectedRowKeys.value = []
  }
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
  margin: 16px;
  background: #fff;
  border-radius: 8px;
  overflow: hidden;

  .devices-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;
    background: #fafafa;

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
</style>
