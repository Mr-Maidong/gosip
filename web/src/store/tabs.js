import { defineStore } from 'pinia'

export const useTabsStore = defineStore('tabs', {
  state: () => ({
    // 已打开的标签页
    openedTabs: [],
    // 当前激活的标签页
    activeTab: '',
    // 是否已从 localStorage 恢复
    restored: false
  }),

  getters: {
    // 获取需要缓存的组件名列表
    cachedViews: state => {
      return state.openedTabs.filter(tab => tab.keepAlive).map(tab => tab.name)
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
        const newTab = {
          path: tab.path,
          title: tab.title,
          name: tab.name,
          keepAlive: tab.keepAlive !== false
        }
        // 如果是首页，插入到第一个位置
        if (tab.path === '/home') {
          this.openedTabs.unshift(newTab)
        } else {
          // 其他标签添加到末尾
          this.openedTabs.push(newTab)
        }
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
     * 从 localStorage 恢复状态
     */
    restoreFromStorage() {
      const savedTabs = localStorage.getItem('openedTabs')

      if (savedTabs) {
        try {
          this.openedTabs = JSON.parse(savedTabs)
        } catch {
          this.openedTabs = []
        }
      }

      // 确保至少有一个首页标签，且在第一个位置
      const homeIndex = this.openedTabs.findIndex(t => t.path === '/home')
      if (homeIndex === -1) {
        this.openedTabs.unshift({
          path: '/home',
          title: '首页',
          name: 'Home',
          keepAlive: true
        })
      } else if (homeIndex !== 0) {
        // 将首页移到第一个位置
        const [homeTab] = this.openedTabs.splice(homeIndex, 1)
        this.openedTabs.unshift(homeTab)
      }
      this.restored = true
    },

    /**
     * 保存到 localStorage
     */
    saveToStorage() {
      localStorage.setItem('openedTabs', JSON.stringify(this.openedTabs))
    }
  }
})
