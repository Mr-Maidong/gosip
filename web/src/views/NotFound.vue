<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <h1>404</h1>
      <p>抱歉，您访问的页面不存在</p>
      <a-button
        type="primary"
        @click="goHome"
      >
        返回首页
      </a-button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<style lang="less" scoped>
.not-found-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: #f0f2f5;
}

.not-found-content {
  text-align: center;

  h1 {
    font-size: 72px;
    color: #1890ff;
    margin-bottom: 16px;
  }

  p {
    margin-bottom: 24px;
    color: #666;
  }
}
</style>
