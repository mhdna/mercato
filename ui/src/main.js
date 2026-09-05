import { createApp } from 'vue'

import { installClientLogger } from '@/composables/useClientLogs'
import { registerPlugins } from '@/plugins'

import App from './App.vue'

import '@/plugins/echarts'
import '@/styles/global.scss'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'

installClientLogger()

const app = createApp(App)
app.config.errorHandler = (err, _i, info) => console.error(`Vue error (${info})`, err)
registerPlugins(app)
app.mount('#app')
