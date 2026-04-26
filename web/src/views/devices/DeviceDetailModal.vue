<template>
  <a-modal v-model:open="visible" :title="device?.name || device?.deviceid || '设备详情'" :footer="null" width="720px" :destroy-on-close="true">
    <a-tabs v-model:activeKey="activeTab">
      <a-tab-pane key="events" tab="注册记录">
        <a-spin :spinning="loading.events" class="timeline-spin">
          <a-timeline v-if="events.length > 0" class="custom-timeline">
            <a-timeline-item v-for="(event, index) in events" :key="event.id" :color="event.eventtype === 'ONLINE' ? 'green' : 'red'" :class="{ 'last-item': index === events.length - 1 }">
              <div class="event-time">
                {{ formatTime(event.eventtime) }}
              </div>
              <div class="event-source">{{ event.source || '-' }}</div>
            </a-timeline-item>
          </a-timeline>
          <a-empty v-else description="暂无记录" />
        </a-spin>
      </a-tab-pane>

      <a-tab-pane key="positions" tab="GPS轨迹">
        <a-spin :spinning="loading.positions" class="timeline-spin">
          <a-timeline v-if="positions.length > 0" class="custom-timeline">
            <a-timeline-item v-for="(pos, index) in positions" :key="pos.id" color="blue" :class="{ 'last-item': index === positions.length - 1 }">
              <div class="event-time">{{ formatTime(pos.gpstime) }}</div>
              <div>经度: {{ pos.longitude }}, 纬度: {{ pos.latitude }}</div>
              <div v-if="pos.speed" class="event-source">速度: {{ pos.speed }}</div>
            </a-timeline-item>
          </a-timeline>
          <a-empty v-else description="近15分钟无上报轨迹" />
        </a-spin>
      </a-tab-pane>

      <a-tab-pane key="alarms" tab="告警信息">
        <a-table :data-source="alarms" :columns="alarmColumns" :pagination="false" size="small">
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'alarmLevel'">
              <a-tag :color="getAlarmColor(record.alarmlevel)">{{ record.alarmlevel }}</a-tag>
            </template>
            <template v-else-if="column.key === 'alarmTime'">
              {{ formatTime(record.alarmtime) }}
            </template>
          </template>
        </a-table>
      </a-tab-pane>
    </a-tabs>
  </a-modal>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { getDeviceEvents, getDevicePositions, getDeviceAlarms } from '@/api/device'
import dayjs from 'dayjs'

const props = defineProps({ open: Boolean, device: Object })
const emit = defineEmits(['update:open'])

const activeTab = ref('events')
const events = ref([])
const positions = ref([])
const alarms = ref([])
const loading = ref({ events: false, positions: false, alarms: false })

const visible = computed({ get: () => props.open, set: val => emit('update:open', val) })

watch(
  () => [props.open, props.device],
  ([open, device]) => {
    if (open && device) loadTabData(activeTab.value)
  },
  { immediate: true }
)

watch(activeTab, val => {
  if (props.open && props.device) loadTabData(val)
})

const loadTabData = async tab => {
  if (tab === 'events' && !loading.value.events) {
    loading.value.events = true
    try {
      const res = await getDeviceEvents(props.device.deviceid)
      events.value = res.data?.list || []
    } catch (e) {
      console.error('加载事件失败:', e)
    }
    loading.value.events = false
  } else if (tab === 'positions' && !loading.value.positions) {
    loading.value.positions = true
    try {
      const res = await getDevicePositions(props.device.deviceid)
      positions.value = res.data?.list || []
    } catch (e) {
      console.error('加载位置失败:', e)
    }
    loading.value.positions = false
  } else if (tab === 'alarms' && !loading.value.alarms) {
    loading.value.alarms = true
    try {
      const res = await getDeviceAlarms(props.device.deviceid)
      alarms.value = res.data?.list || []
    } catch (e) {
      console.error('加载告警失败:', e)
    }
    loading.value.alarms = false
  }
}

const formatTime = t => dayjs.unix(t).format('YYYY-MM-DD HH:mm:ss')
const getAlarmColor = l => ({ NORMAL: 'green', WARNING: 'orange', CRITICAL: 'red' })[l] || 'default'

const alarmColumns = [
  { title: '时间', key: 'alarmTime', width: 160 },
  { title: '类型', dataIndex: 'alarmtype', key: 'alarmType', width: 120 },
  { title: '级别', key: 'alarmLevel', width: 100 },
  { title: '消息', dataIndex: 'alarmmsg', ellipsis: true }
]
</script>

<style lang="less" scoped>
.custom-timeline {
  padding-top: 8px;
  max-height: 420px;
  overflow-y: auto;
}

.custom-timeline :deep(.ant-timeline-item-last) {
  padding-bottom: 0;
}

.event-time {
  font-size: 12px;
  color: #999;
}

#map-container {
  width: 100%;
  border-radius: 4px;
}
</style>
