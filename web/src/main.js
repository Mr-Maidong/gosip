import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import App from './App.vue'
import router from './router'

// 引入样式
import 'ant-design-vue/dist/reset.css'
import './styles/tailwind.css'  // Tailwind CSS
import './styles/index.less'     // 自定义 Less 样式

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(Antd)

app.mount('#app')
