/**
 * plugins/vuetify.js
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Composables
import { createVuetify } from 'vuetify'
import {
  VFileUpload,
  VFileUploadDropzone,
  VFileUploadItem,
  VFileUploadList,
  VIconBtn,
} from 'vuetify/components'
import { VCommandPalette } from 'vuetify/labs/VCommandPalette'

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
          error: '#B71C1C',
          primary: '#5C6BC0', // Indigo 400
        },
      },
    },
  },
  typography: {
    fontFamily: '\'Roboto\', sans-serif',
  },
  // Kill native browser autofill/autocomplete app-wide. Vuetify's "suppress"
  // emits autocomplete="off" *and* randomises the input name on focus, which
  // is what actually stops Chrome/Brave from re-filling fields.
  defaults: {
    VDataTable: { disableSort: true, itemsPerPage: 14 },
    VDataTableServer: { disableSort: true, itemsPerPage: 14 },
    VDataTableVirtual: { disableSort: true },
    VTextField: { autocomplete: 'suppress' },
    VTextarea: { autocomplete: 'suppress' },
    VAutocomplete: { autocomplete: 'suppress' },
    VCombobox: { autocomplete: 'suppress' },
    VSelect: { autocomplete: 'suppress' },
    VNumberInput: { autocomplete: 'suppress' },
    VOtpInput: { autocomplete: 'suppress' },
    VFileInput: { autocomplete: 'suppress' },
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
