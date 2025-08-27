import axios from 'axios'
import { ElMessage } from 'element-plus'

// Create axios instance for authentication
const authApi = axios.create({
  baseURL: '/auth/api/v1',
  timeout: 10000
})

// Create axios instance for admin API
const request = axios.create({
  baseURL: '/admin/api/v1',
  timeout: 10000
})

// Add interceptors to both instances
function addInterceptors(instance) {
  // Request interceptor
  instance.interceptors.request.use(
    config => {
      // Add auth token if available
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

  // Response interceptor
  instance.interceptors.response.use(
    response => {
      const res = response.data
      
      // Check if response has the expected structure
      if (res && typeof res === 'object' && 'code' in res) {
        // If the response code is not 200, it is judged as an error
        if (res.code !== 200) {
          ElMessage.error(res.message || 'Request failed')
          return Promise.reject(new Error(res.message || 'Error'))
        } else {
          return res
        }
      } else {
        // If response doesn't have code field, return the raw data
        return res
      }
    },
    error => {
      console.error('Request error:', error)
      // Check if it's a network error
      if (error.code === 'ERR_NETWORK' || error.code === 'ERR_EMPTY_RESPONSE') {
        ElMessage.error('网络连接错误，请检查服务器状态')
      } else {
        ElMessage.error(error.message || 'Request failed')
      }
      return Promise.reject(error)
    }
  )
}

// Apply interceptors to both instances
addInterceptors(authApi)
addInterceptors(request)

export default request
export { authApi }
