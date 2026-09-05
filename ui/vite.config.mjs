import { fileURLToPath, URL } from 'node:url'
import Vue from '@vitejs/plugin-vue'
// Plugins
import AutoImport from 'unplugin-auto-import/vite'
import Fonts from 'unplugin-fonts/vite'
import Components from 'unplugin-vue-components/vite'
import { VueRouterAutoImports } from 'unplugin-vue-router'
import VueRouter from 'unplugin-vue-router/vite'
// Utilities
import { defineConfig } from 'vite'

import Layouts from 'vite-plugin-vue-layouts-next'
import Vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    VueRouter(),
    Layouts(),
    Vue({
      template: { transformAssetUrls },
    }),
    Vuetify({
      autoImport: true,
      // `styles: true` ships Vuetify's precompiled CSS. Pointing `styles` at a
      // configFile makes the plugin recompile every component's SASS in dev,
      // and the emitted sourcemaps reference bare `VBtn.sass` / `VCode.sass`
      // paths the browser then 404s on. We have no SASS variable overrides
      // (settings.scss only held a plain `.card-title` rule, now in
      // global.scss), so the configFile bought us nothing. Re-add it here if
      // real `@use 'vuetify/settings' with (...)` overrides are needed.
      styles: true,
    }),
    Components(),
    Fonts({
      provider: 'none',
      fonts: {
        google: {
          families: [{
            name: 'Roboto',
            styles: 'wght@100;300;400;500;700;900',
          }],
        },
      },
    }),
    AutoImport({
      imports: [
        'vue',
        VueRouterAutoImports,
        {
          pinia: ['defineStore', 'storeToRefs'],
        },
      ],
      eslintrc: {
        enabled: true,
      },
      vueTemplate: true,
    }),
  ],
  build: {
    minify: 'esbuild',
    rollupOptions: {
      output: {
        manualChunks (id) {
          if (id.includes('node_modules')) {
            if (id.includes('echarts')) {
              return 'echarts'
            }
            if (id.includes('vuetify')) {
              return 'vuetify'
            }
            if (id.includes('@mdi')) {
              return 'mdi'
            }
          }
        },
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
      },
    },
    target: 'esnext',
    cssCodeSplit: true,
    assetsInlineLimit: 4096,
    chunkSizeWarningLimit: 600,
  },
  optimizeDeps: {
    exclude: [
      'vuetify',
      'vue-router',
      'unplugin-vue-router/runtime',
      'unplugin-vue-router/data-loaders',
      'unplugin-vue-router/data-loaders/basic',
    ],
  },
  define: { 'process.env': {} },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('src', import.meta.url)),
    },
    extensions: [
      '.js',
      '.json',
      '.jsx',
      '.mjs',
      '.ts',
      '.tsx',
      '.vue',
    ],
  },
  server: {
    port: 3000,
  },
})
