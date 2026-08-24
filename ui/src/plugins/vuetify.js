/**
 * plugins/vuetify.js
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Composables
import { createVuetify } from 'vuetify'
import { VCommandPalette } from 'vuetify/labs/VCommandPalette'
import {
  VFileUpload,
  VFileUploadDropzone,
  VFileUploadItem,
  VFileUploadList,
} from 'vuetify/labs/VFileUpload'
import { VIconBtn } from 'vuetify/labs/VIconBtn'

import { aliases, mdi } from './icons'
// Styles
import 'vuetify/styles'

function getSystemTheme () {
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark'
  }
  return 'light'
}

function getInitialTheme () {
  const saved = localStorage.getItem('theme')
  if (saved) {
    return saved
  }
  return getSystemTheme()
}

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  theme: {
    defaultTheme: getInitialTheme(),
    themes: {
      light: {
        colors: {
          primary: '#3F51B5', // Indigo 500
        },
      },
      dark: {
        dark: true,
        colors: {
          background: '#000000',
          primary: '#5C6BC0', // Indigo 400
          surface: '#000000',
        },
      },
    },
  },
  typography: {
    fontFamily: '\'Roboto\', sans-serif',
  },
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: {
      mdi,
    },
  },
  components: {
    VIconBtn,
    VCommandPalette,
    VFileUpload,
    VFileUploadDropzone,
    VFileUploadItem,
    VFileUploadList,
  },
})
