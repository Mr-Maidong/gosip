/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      // 可以在这里自定义主题，覆盖默认颜色等
      colors: {
        primary: '#1890ff', // 与 Ant Design 主色保持一致
      }
    },
  },
  plugins: [],
  // 重要：禁用预飞行样式，避免与 Ant Design Vue 冲突
  corePlugins: {
    preflight: false,
  }
}
