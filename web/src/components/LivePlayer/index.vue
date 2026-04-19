<template>
  <div
    ref="containerRef"
    class="live-player-container"
  >
    <a-spin
      :spinning="loading"
      class="player-spin"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { message, Spin as ASpin } from 'ant-design-vue'

const props = defineProps({
  url: {
    type: String,
    required: true
  },
  isLive: {
    type: Boolean,
    default: true
  },
  muted: {
    type: Boolean,
    default: false
  },
  stretch: {
    type: Boolean,
    default: true
  }
})

const containerRef = ref(null)
const loading = ref(true)
let player = null

const loadAndInitPlayer = () => {
  if (!containerRef.value || !props.url) return

  loading.value = true

  if (typeof window.EasyPlayerPro === 'undefined') {
    message.error('播放器加载失败，请刷新页面重试')
    loading.value = false
    return
  }

  player = new window.EasyPlayerPro(containerRef.value, {
    isLive: props.isLive,
    isMute: props.muted,
    stretch: props.stretch,
    hasAudio: true,
    bufferTime: 0.2,
    loadTimeOut: 10,
    debug: false,
    btns: {
      fullscreen: true,
      screenshot: true,
      play: true,
      audio: true,
      stretch: false
    }
  })

  player.on('play', () => {
    console.log('播放器开始播放')
    loading.value = false
  })

  player.on('error', (err) => {
    console.error('播放器错误:', err)
    message.error('视频播放失败')
    loading.value = false
  })

  player.on('timeout', () => {
    message.error('视频加载超时')
    loading.value = false
  })

  player.on('liveEnd', () => {
    message.warning('直播已结束')
  })

  player.play(props.url)
}

onMounted(() => {
  if (window.EasyPlayerPro) {
    loadAndInitPlayer()
  } else {
    const script = document.createElement('script')
    script.src = '/js/EasyPlayer-pro.js'
    script.onload = () => {
      loadAndInitPlayer()
    }
    script.onerror = () => {
      message.error('播放器脚本加载失败')
    }
    document.head.appendChild(script)
  }
})

onBeforeUnmount(() => {
  if (player) {
    player.destroy()
    player = null
  }
})

watch(() => props.url, (newUrl) => {
  if (player && newUrl) {
    loading.value = true
    player.play(newUrl)
  }
})
</script>

<style scoped>
.live-player-container {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 200px;
  background: #000;
  border-radius: 4px;
  overflow: hidden;
}

.player-spin {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 10;
}
</style>
