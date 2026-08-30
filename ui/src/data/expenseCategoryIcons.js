// The icon + colour palette the "add / edit category" dialog offers. Every
// icon here must be registered in src/plugins/icons.js, and every colour is
// a Vuetify colour token so `:color` works directly on v-avatar / v-icon.

export const EXPENSE_CATEGORY_ICONS = [
  'mdi-flash',
  'mdi-water',
  'mdi-lightbulb',
  'mdi-wifi',
  'mdi-phone',
  'mdi-home-city',
  'mdi-cash-multiple',
  'mdi-cash',
  'mdi-credit-card-outline',
  'mdi-bank',
  'mdi-scale-balance',
  'mdi-shield-check',
  'mdi-wrench',
  'mdi-broom',
  'mdi-truck',
  'mdi-car',
  'mdi-gas-station',
  'mdi-package-variant-closed',
  'mdi-cart',
  'mdi-bullhorn',
  'mdi-food',
  'mdi-coffee',
  'mdi-printer',
  'mdi-laptop',
  'mdi-medical-bag',
  'mdi-account-group',
  'mdi-gift-outline',
  'mdi-tag-outline',
]

export const EXPENSE_CATEGORY_COLORS = [
  'red',
  'pink',
  'purple',
  'deep-purple',
  'indigo',
  'blue',
  'cyan',
  'teal',
  'green',
  'light-green',
  'amber',
  'orange',
  'deep-orange',
  'brown',
  'blue-grey',
  'grey',
]

export const DEFAULT_EXPENSE_CATEGORY_ICON = 'mdi-tag-outline'
export const DEFAULT_EXPENSE_CATEGORY_COLOR = 'blue-grey'

export function categoryIcon (category) {
  return category?.icon || DEFAULT_EXPENSE_CATEGORY_ICON
}

export function categoryColor (category) {
  return category?.color || DEFAULT_EXPENSE_CATEGORY_COLOR
}
