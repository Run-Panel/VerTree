import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import router from './router'
import { createI18n } from 'vue-i18n'

// Import ECharts
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart, { THEME_KEY } from 'vue-echarts'

// Import global styles
import './styles/global.css'

import App from './App.vue'

// Register ECharts components
use([
  CanvasRenderer,
  LineChart,
  PieChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

// Import locale messages
import en from './locales/en.json'
import zh from './locales/zh.json'

// Create i18n instance
const i18n = createI18n({
  legacy: false,
  locale: 'zh',
  fallbackLocale: 'en',
  messages: {
    en,
    zh
  }
})

const app = createApp(App)
const pinia = createPinia()

// Register Element Plus Icons
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// Register VChart globally
app.component('VChart', VChart)
app.provide(THEME_KEY, 'light')

app.use(pinia)
app.use(router)
app.use(ElementPlus)
app.use(i18n)

// Initialize auth store after pinia is setup
import { useAuthStore } from './store/auth'
const authStore = useAuthStore()
authStore.initAuth()

app.mount('#app')
