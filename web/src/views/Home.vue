<template>
  <div class="home-container">
    <!-- 顶部欢迎区 -->
    <div class="welcome-section">
      <div class="welcome-left">
        <h1>欢迎回来，{{ userName }}</h1>
        <p>YSIP 视频监控管理平台 · 实时运行状态</p>
      </div>
      <div class="welcome-right">
        <div class="status-badge" :class="systemStatus.type">
          <span class="status-dot"></span>
          {{ systemStatus.text }}
        </div>
        <div class="current-time">{{ currentTime }}</div>
      </div>
    </div>

    <!-- 统计卡片区 -->
    <a-row :gutter="[16, 16]" class="stats-row">zui
      <a-col :span="6" v-for="(stat, index) in stats" :key="index">
        <div class="stat-card" :style="{ borderLeftColor: stat.color }">
          <div class="stat-icon" :style="{ background: stat.bgColor }">
            <component :is="stat.icon" :style="{ color: stat.color }" />
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stat.value }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </div>
      </a-col>
    </a-row>

    <!-- 负载与流量区 -->
    <a-row :gutter="[16, 16]" class="charts-row">
      <a-col :xs="24" :lg="12">
        <div class="chart-card">
          <div class="chart-title">系统负载</div>
          <div class="load-items">
            <div class="load-item">
              <div class="load-header">
                <span class="load-label">CPU 使用率</span>
                <span class="load-value" :style="{ color: getLoadColor(systemLoad.cpu) }">
                  {{ systemLoad.cpu }}%
                </span>
              </div>
              <a-progress
                :percent="systemLoad.cpu"
                :show-info="false"
                :stroke-color="getLoadColor(systemLoad.cpu)"
                :trail-color="'#f0f0f0'"
                size="small"
              />
            </div>
            <div class="load-item">
              <div class="load-header">
                <span class="load-label">内存使用率</span>
                <span class="load-value" :style="{ color: getLoadColor(systemLoad.memory) }">
                  {{ systemLoad.memory }}%
                </span>
              </div>
              <a-progress
                :percent="systemLoad.memory"
                :show-info="false"
                :stroke-color="getLoadColor(systemLoad.memory)"
                :trail-color="'#f0f0f0'"
                size="small"
              />
            </div>
            <div class="load-item">
              <div class="load-header">
                <span class="load-label">网络 I/O</span>
                <span class="load-value">
                  <span class="traffic-up">{{ systemLoad.networkIn }}</span>
                  <span class="traffic-sep">/</span>
                  <span class="traffic-down">{{ systemLoad.networkOut }}</span>
                </span>
              </div>
              <div class="network-bars">
                <div class="net-bar-item">
                  <span class="net-bar-label">入</span>
                  <div class="net-bar-track">
                    <div class="net-bar-fill" :style="{ width: systemLoad.networkInPercent + '%' }"></div>
                  </div>
                </div>
                <div class="net-bar-item">
                  <span class="net-bar-label">出</span>
                  <div class="net-bar-track">
                    <div class="net-bar-fill out" :style="{ width: systemLoad.networkOutPercent + '%' }"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </a-col>
      <a-col :xs="24" :lg="12">
        <div class="chart-card">
          <div class="chart-title">流量统计</div>
          <div class="traffic-stats">
            <div class="traffic-item">
              <div class="traffic-icon in">
                <ArrowUpOutlined />
              </div>
              <div class="traffic-info">
                <div class="traffic-label">入口流量</div>
                <div class="traffic-value">{{ trafficStats.input.speed }} <span class="traffic-unit">{{ trafficStats.input.unit }}</span></div>
                <div class="traffic-total">总流量: {{ trafficStats.input.total }}</div>
              </div>
            </div>
            <div class="traffic-item">
              <div class="traffic-icon out">
                <ArrowDownOutlined />
              </div>
              <div class="traffic-info">
                <div class="traffic-label">出口流量</div>
                <div class="traffic-value">{{ trafficStats.output.speed }} <span class="traffic-unit">{{ trafficStats.output.unit }}</span></div>
                <div class="traffic-total">总流量: {{ trafficStats.output.total }}</div>
              </div>
            </div>
          </div>
          <div class="traffic-trend">
            <div class="trend-title">实时流量趋势</div>
            <div class="trend-bars">
              <div
                v-for="(bar, index) in trafficTrend"
                :key="index"
                class="trend-bar-item"
              >
                <div class="trend-bar-wrap">
                  <div class="trend-bar in" :style="{ height: bar.in + '%' }"></div>
                  <div class="trend-bar out" :style="{ height: bar.out + '%' }"></div>
                </div>
                <div class="trend-time">{{ bar.time }}</div>
              </div>
            </div>
          </div>
        </div>
      </a-col>
    </a-row>

    <!-- 最近活动流 + 推流统计 -->
    <a-row :gutter="[16, 16]" class="charts-row">
      <a-col :xs="24" :lg="14">
        <div class="chart-card">
          <div class="chart-title">最近活动流</div>
          <a-table
            :columns="streamColumns"
            :data-source="recentStreams"
            :pagination="false"
            size="small"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'type'">
                <a-tag :color="getStreamTypeColor(record.type)">{{ record.type }}</a-tag>
              </template>
              <template v-if="column.key === 'status'">
                <span class="stream-status" :class="record.status">
                  <span class="stream-dot"></span>
                  {{ getStreamStatusText(record.status) }}
                </span>
              </template>
              <template v-if="column.key === 'action'">
                <a-button type="link" size="small" danger @click="handleStopStream(record)">停止</a-button>
              </template>
            </template>
          </a-table>
        </div>
      </a-col>
      <a-col :xs="24" :lg="10">
        <div class="chart-card">
          <div class="chart-title">推流统计</div>
          <div class="stream-stats">
            <div class="stream-stats-type">
              <div class="stats-type-title">按推流类型</div>
              <div class="stats-bar-item" v-for="item in streamStatsByType" :key="item.type">
                <span class="stats-bar-label">{{ item.type }}</span>
                <div class="stats-bar-track">
                  <div class="stats-bar-fill" :style="{ width: item.percent + '%', background: item.color }"></div>
                </div>
                <span class="stats-bar-value">{{ item.count }}</span>
              </div>
            </div>
            <a-divider style="margin: 12px 0" />
            <div class="stream-stats-device">
              <div class="stats-type-title">按设备</div>
              <div class="stats-bar-item" v-for="item in streamStatsByDevice" :key="item.name">
                <span class="stats-bar-label">{{ item.name }}</span>
                <div class="stats-bar-track">
                  <div class="stats-bar-fill" :style="{ width: item.percent + '%', background: item.color }"></div>
                </div>
                <span class="stats-bar-value">{{ item.count }}</span>
              </div>
            </div>
          </div>
        </div>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  ApartmentOutlined,
  DesktopOutlined,
  VideoCameraOutlined,
  FieldTimeOutlined,
  UserOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined
} from '@ant-design/icons-vue'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const userName = computed(() => userStore.userInfo?.name || '管理员')

// 当前时间
const currentTime = ref('')
let timeTimer = null

const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

// 系统状态
const systemStatus = computed(() => {
  const load = systemLoad.value
  if (load.cpu > 90 || load.memory > 90) {
    return { type: 'danger', text: '负载过高' }
  } else if (load.cpu > 70 || load.memory > 70) {
    return { type: 'warning', text: '负载较高' }
  }
  return { type: 'normal', text: '运行正常' }
})

// 统计数据
const stats = ref([
  { label: '平台总数', value: 3, icon: ApartmentOutlined, color: '#1890ff', bgColor: '#e6f7ff' },
  { label: '设备总数', value: 48, icon: DesktopOutlined, color: '#52c41a', bgColor: '#f6ffed' },
  { label: '通道总数', value: 256, icon: VideoCameraOutlined, color: '#722ed1', bgColor: '#f9f0ff' },
  { label: '在线流', value: 12, icon: FieldTimeOutlined, color: '#fa8c16', bgColor: '#fff7e6' },
])

// 系统负载
const systemLoad = ref({
  cpu: 45,
  memory: 62,
  networkIn: '1.2Gbps',
  networkOut: '800Mbps',
  networkInPercent: 60,
  networkOutPercent: 40
})

const getLoadColor = (value) => {
  if (value >= 90) return '#ff4d4f'
  if (value >= 70) return '#faad14'
  return '#52c41a'
}

// 流量统计
const trafficStats = ref({
  input: { speed: '1.2', unit: 'Gbps', total: '12.5 TB' },
  output: { speed: '800', unit: 'Mbps', total: '8.3 TB' }
})

const trafficTrend = ref([
  { time: '14:00', in: 80, out: 60 },
  { time: '14:05', in: 65, out: 45 },
  { time: '14:10', in: 90, out: 70 },
  { time: '14:15', in: 75, out: 55 },
  { time: '14:20', in: 60, out: 40 },
  { time: '14:25', in: 85, out: 65 },
  { time: '14:30', in: 70, out: 50 },
  { time: '14:35', in: 95, out: 75 },
  { time: '14:40', in: 60, out: 40 },
  { time: '14:45', in: 80, out: 60 }
])

// 流列表
const streamColumns = [
  { title: '通道名称', dataIndex: 'name', key: 'name' },
  { title: '类型', dataIndex: 'type', key: 'type', width: 100 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '开始时间', dataIndex: 'startTime', key: 'startTime', width: 100 },
  { title: '操作', key: 'action', width: 80 }
]

const recentStreams = ref([
  { id: 1, name: 'Camera-01 - 门口监控', type: 'GB28181', status: 'active', startTime: '14:30:15' },
  { id: 2, name: 'Camera-02 - 大厅监控', type: 'GB28181', status: 'active', startTime: '14:28:42' },
  { id: 3, name: 'Camera-03 - 停车场', type: 'RTMP', status: 'active', startTime: '14:25:33' },
  { id: 4, name: 'Camera-04 - 电梯监控', type: 'GB28181', status: 'active', startTime: '14:20:08' },
  { id: 5, name: 'RTSP-Camera-01', type: 'RTSP', status: 'active', startTime: '14:15:22' }
])

const getStreamTypeColor = (type) => {
  const colors = { GB28181: 'blue', RTMP: 'red', RTSP: 'orange' }
  return colors[type] || 'default'
}

const getStreamStatusText = (status) => {
  return status === 'active' ? '直播中' : '已停止'
}

const handleStopStream = (record) => {
  console.log('停止流:', record.name)
}

// 推流统计 - 按类型
const streamStatsByType = ref([
  { type: 'GB28181', count: 12, percent: 60, color: '#1890ff' },
  { type: 'RTMP', count: 5, percent: 25, color: '#ff4d4f' },
  { type: 'RTSP', count: 3, percent: 15, color: '#fa8c16' }
])

// 推流统计 - 按设备
const streamStatsByDevice = ref([
  { name: 'Device-A', count: 6, percent: 50, color: '#52c41a' },
  { name: 'Device-B', count: 4, percent: 33, color: '#1890ff' },
  { name: 'Device-C', count: 2, percent: 17, color: '#722ed1' }
])

onMounted(() => {
  updateTime()
  timeTimer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (timeTimer) clearInterval(timeTimer)
})
</script>

<style lang="less" scoped>
.home-container {
  margin: 16px 16px 0;
  padding: 16px;
  min-height: calc(100vh - 180px);
  background: #fdfdfd;
  border-radius: 12px;
}

// 顶部欢迎区
.welcome-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px 24px;
  background: linear-gradient(135deg, #1890ff 0%, #42b4ff 100%);
  border-radius: 8px;
  color: #fff;

  .welcome-left {
    h1 {
      margin: 0 0 4px 0;
      font-size: 24px;
      font-weight: 600;
    }
    p {
      margin: 0;
      font-size: 14px;
      opacity: 0.9;
    }
  }

  .welcome-right {
    text-align: right;

    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 4px 12px;
      border-radius: 12px;
      font-size: 13px;
      font-weight: 500;
      margin-bottom: 8px;

      &.normal {
        background: rgba(255, 255, 255, 0.2);
      }
      &.warning {
        background: rgba(250, 173, 20, 0.6);
      }
      &.danger {
        background: rgba(255, 77, 79, 0.6);
      }

      .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: currentColor;
        animation: pulse 1.5s infinite;
      }
    }

    .current-time {
      font-size: 13px;
      opacity: 0.9;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

// 统计卡片
.stats-row {
  margin-bottom: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  border-left: 3px solid;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);

  .stat-icon {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    flex-shrink: 0;
  }

  .stat-info {
    .stat-value {
      font-size: 24px;
      font-weight: 600;
      color: #333;
      line-height: 1;
    }
    .stat-label {
      font-size: 13px;
      color: #999;
      margin-top: 4px;
    }
  }
}

// 图表卡片
.charts-row {
  margin-bottom: 16px;
}

.chart-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  height: 100%;

  .chart-title {
    font-size: 15px;
    font-weight: 600;
    color: #333;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid #f0f0f0;
  }
}

// 系统负载
.load-items {
  .load-item {
    &:not(:last-child) {
      margin-bottom: 16px;
    }

    .load-header {
      display: flex;
      justify-content: space-between;
      margin-bottom: 6px;

      .load-label {
        font-size: 13px;
        color: #666;
      }
      .load-value {
        font-size: 13px;
        font-weight: 500;
      }
    }
  }
}

.traffic-up { color: #52c41a; }
.traffic-sep { color: #999; margin: 0 2px; }
.traffic-down { color: #1890ff; }

.network-bars {
  margin-top: 8px;
  .net-bar-item {
    display: flex;
    align-items: center;
    gap: 8px;
    &:not(:last-child) { margin-bottom: 4px; }

    .net-bar-label {
      width: 16px;
      font-size: 11px;
      color: #999;
    }
    .net-bar-track {
      flex: 1;
      height: 6px;
      background: #f0f0f0;
      border-radius: 3px;
      overflow: hidden;
    }
    .net-bar-fill {
      height: 100%;
      background: #52c41a;
      border-radius: 3px;
      transition: width 0.3s;
      &.out { background: #1890ff; }
    }
  }
}

// 流量统计
.traffic-stats {
  display: flex;
  gap: 24px;
  margin-bottom: 16px;

  .traffic-item {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: #fafafa;
    border-radius: 6px;

    .traffic-icon {
      width: 40px;
      height: 40px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 18px;
      &.in { background: #e6f7ff; color: #1890ff; }
      &.out { background: #f6ffed; color: #52c41a; }
    }

    .traffic-info {
      .traffic-label {
        font-size: 12px;
        color: #999;
      }
      .traffic-value {
        font-size: 18px;
        font-weight: 600;
        color: #333;
        .traffic-unit {
          font-size: 12px;
          font-weight: normal;
          color: #999;
        }
      }
      .traffic-total {
        font-size: 11px;
        color: #999;
      }
    }
  }
}

.traffic-trend {
  .trend-title {
    font-size: 13px;
    color: #666;
    margin-bottom: 12px;
  }
  .trend-bars {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    height: 60px;
    gap: 4px;

    .trend-bar-item {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      height: 100%;

      .trend-bar-wrap {
        flex: 1;
        width: 100%;
        display: flex;
        align-items: flex-end;
        justify-content: center;
        gap: 2px;
      }
      .trend-bar {
        width: 6px;
        background: #52c41a;
        border-radius: 2px 2px 0 0;
        min-height: 4px;
        &.out { background: #1890ff; }
      }
      .trend-time {
        font-size: 10px;
        color: #999;
        margin-top: 4px;
      }
    }
  }
}

// 流状态
.stream-status {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  &.active { color: #52c41a; }
  &.stopped { color: #999; }

  .stream-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: currentColor;
    &.active { animation: pulse 1.5s infinite; }
  }
}

// 推流统计
.stream-stats {
  .stream-stats-type,
  .stream-stats-device {
    .stats-type-title {
      font-size: 13px;
      color: #666;
      margin-bottom: 10px;
    }
  }

  .stats-bar-item {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 8px;

    .stats-bar-label {
      width: 70px;
      font-size: 12px;
      color: #666;
      flex-shrink: 0;
    }
    .stats-bar-track {
      flex: 1;
      height: 8px;
      background: #f0f0f0;
      border-radius: 4px;
      overflow: hidden;
    }
    .stats-bar-fill {
      height: 100%;
      border-radius: 4px;
      transition: width 0.3s;
    }
    .stats-bar-value {
      width: 30px;
      font-size: 12px;
      font-weight: 500;
      color: #333;
      text-align: right;
    }
  }
}
</style>
