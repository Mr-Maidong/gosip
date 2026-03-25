import { defineStore } from 'pinia'

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    // 已打开的标签页
    openedTabs: [],
    // 当前激活的标签页
    activeTab: ''
  }),

  getters: {
    // 获取需要缓存的组件名列表
    cachedViews: (state) => {
      return state.openedTabs
        .filter(tab => tab.keepAlive)
        .map(tab => tab.name)
    }
  },

  actions: {
    /**
     * 添加标签页
     * @param {Object} tab - { path, title, name, keepAlive }
     */
    addTab(tab) {
      // 检查是否已存在
      const exists = this.openedTabs.find(t => t.path === tab.path)
      if (!exists) {
        this.openedTabs.push({
          path: tab.path,
          title: tab.title,
          name: tab.name,
          keepAlive: tab.keepAlive !== false
        })
      }
      this.activeTab = tab.path
      this.saveToStorage()
    },

    /**
     * 移除标签页
     * @param {string} path - 要关闭的标签页路径
     */
    removeTab(path) {
      const index = this.openedTabs.findIndex(t => t.path === path)
      if (index === -1) return

      const removed = this.openedTabs[index]
      // 不允许关闭首页
      if (removed.path === '/home') return

      this.openedTabs.splice(index, 1)

      // 如果关闭的是当前激活的标签，切换到前一个
      if (this.activeTab === path) {
        const newIndex = Math.min(index, this.openedTabs.length - 1)
        this.activeTab = this.openedTabs[newIndex]?.path || '/home'
      }
      this.saveToStorage()
    },

    /**
     * 切换标签页
     * @param {string} path - 要切换到的标签页路径
     */
    setActiveTab(path) {
      this.activeTab = path
      // 同时更新 localStorage
      localStorage.setItem('activeTab', path)
    },

    /**
     * 关闭其他标签页
     * @param {string} path - 保留的标签页路径
     */
    closeOtherTabs(path) {
      this.openedTabs = this.openedTabs.filter(
        t => t.path === path || t.path === '/home'
      )
      this.activeTab = path
      this.saveToStorage()
    },

    /**
     * 关闭所有标签页，回到首页
     */
    closeAllTabs() {
      this.openedTabs = this.openedTabs.filter(t => t.path === '/home')
      this.activeTab = '/home'
      this.saveToStorage()
    },

    /**
     * 从 localStorage 恢复状态
     */
    restoreFromStorage() {
      const savedTabs = localStorage.getItem('openedTabs')
      const savedActive = localStorage.getItem('activeTab')

      if (savedTabs) {
        try {
          this.openedTabs = JSON.parse(savedTabs)
        } catch {
          this.openedTabs = []
        }
      }

      if (savedActive) {
        this.activeTab = savedActive
      }

      // 确保至少有一个首页标签
      if (this.openedTabs.length === 0) {
        this.openedTabs.push({
          path: '/home',
          title: '首页',
          name: 'Home',
          keepAlive: true
        })
        this.activeTab = '/home'
      }
    },

    /**
     * 保存到 localStorage
     */
    saveToStorage() {
      localStorage.setItem('openedTabs', JSON.stringify(this.openedTabs))
    }
  }
})
