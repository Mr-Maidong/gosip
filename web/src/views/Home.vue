<template>
  <div class="home-container">
    <WelcomeCard />
  </div>
</template>

<script setup>
import WelcomeCard from '@/components/WelcomeCard.vue'
</script>

<style lang="less" scoped>
.home-container {
  margin: 16px;
  padding: 0;
  min-height: calc(100vh - 136px);
}
</style>
