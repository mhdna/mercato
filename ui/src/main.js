/**
 * main.js
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Composables
import { createApp } from 'vue'

import { installClientLogger } from '@/composables/useClientLogs'
// Plugins
import { registerPlugins } from '@/plugins'

// Components
import App from './App.vue'

// ECharts - register only needed modules (tree-shakes the rest)
import '@/plugins/echarts'

// Fonts
import '@fontsource/roboto'
import '@/styles/global.scss'

// Styles
import 'unfonts.css'

installClientLogger()

const app = createApp(App)

app.config.errorHandler = (error, instance, info) => {
  console.error(`Vue error (${info})`, error)
}

registerPlugins(app)

app.mount('#app')
