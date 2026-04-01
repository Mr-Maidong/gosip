import axios from 'axios'
import { message } from 'ant-design-vue'
import { useUserStore } from '@/store/user'

// 创建 axios 实例
const request = axios.create({
  baseURL: 'http://localhost:8090/api',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    const res = response.data
    // 处理业务错误（code 不为 0 表示失败）
    if (res.code != 0) {
      message.error(res.msg || '请求失败')
      return Promise.reject(new Error(res.msg || '请求失败'))
    }
    return res
  },
  error => {
    const res = error.response?.data
    const errMsg = res ? `${res?.code}: ${res?.data}` : error.message || '请求失败'
    console.log(errMsg)
    message.error(errMsg ? errMsg : '请求失败')
    if (res?.code === 1004) {
      const userStore = useUserStore()
      userStore.logout()
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      window.location.href = '/login'
    }
    const err = new Error(errMsg)
    err.code = res?.code
    return Promise.reject(err)
  }
)

export default request
