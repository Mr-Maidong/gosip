import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
  state: () => ({
    collapsed: false,
    loading: false
  }),

  actions: {
    toggleCollapsed() {
      this.collapsed = !this.collapsed
    },

    setLoading(loading) {
      this.loading = loading
    }
  }
})
