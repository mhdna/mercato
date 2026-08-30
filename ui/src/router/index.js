import { setupLayouts } from 'virtual:generated-layouts'
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/', component: () => import('@/pages/home.vue'), meta: { title: 'Home' } },
  { path: '/dashboard', component: () => import('@/pages/dashboard.vue'), meta: { title: 'Dashboard' } },
  { path: '/daily-income', component: () => import('@/pages/daily-income.vue'), meta: { title: 'Daily Income' } },
  { path: '/financials', component: () => import('@/pages/financials.vue'), meta: { title: 'Financials' } },
  { path: '/targets', component: () => import('@/pages/targets.vue'), meta: { title: 'Targets' } },
  { path: '/stock-health', component: () => import('@/pages/stock-health.vue'), meta: { title: 'Stock Health' } },
  { path: '/alerts', component: () => import('@/pages/alerts.vue'), meta: { title: 'Alerts' } },
  { path: '/sales-analytics', component: () => import('@/pages/sales-analytics.vue'), meta: { title: 'Sales Analytics' } },
  { path: '/sales-analysis', component: () => import('@/pages/sales-analysis.vue'), meta: { title: 'Sales Analysis' } },
  { path: '/sales-dashboard', component: () => import('@/pages/sales-dashboard.vue'), meta: { title: 'Sales Dashboard' } },
  { path: '/salesperson-performance', component: () => import('@/pages/salesperson-performance.vue'), meta: { title: 'Salesperson Performance' } },
  { path: '/about', component: () => import('@/pages/About.vue'), meta: { title: 'About' } },
  {
    path: '/login',
    component: () => import('@/pages/Login.vue'),
    meta: { title: 'Login', layout: 'auth' },
  },

  { path: '/products', component: () => import('@/pages/products.vue'), meta: { title: 'Products' } },
  { path: '/inventory', component: () => import('@/pages/inventory.vue'), meta: { title: 'Inventory' } },
  { path: '/inventory-analysis', component: () => import('@/pages/inventory-analysis.vue'), meta: { title: 'Inventory Analysis' } },
  { path: '/storage', component: () => import('@/pages/storage.vue'), meta: { title: 'Storage' } },
  { path: '/warehouses', component: () => import('@/pages/warehouses.vue'), meta: { title: 'Warehouses' } },
  { path: '/branches', component: () => import('@/pages/branches.vue'), meta: { title: 'Branches' } },
  { path: '/branch-invoices', redirect: '/invoices' },
  { path: '/transfers', component: () => import('@/pages/transfers.vue'), meta: { title: 'Transfers' } },
  { path: '/stock-movements', component: () => import('@/pages/stock-movements.vue'), meta: { title: 'Stock Movements' } },
  { path: '/barcode-printing', component: () => import('@/pages/barcode-printing.vue'), meta: { title: 'Barcode Printing' } },
  { path: '/colors-sizes', component: () => import('@/pages/Colors_Sizes.vue'), meta: { title: 'Colors & Sizes' } },
  { path: '/attributes', component: () => import('@/pages/attributes.vue'), meta: { title: 'Attributes' } },
  { path: '/barcodes', component: () => import('@/pages/barcodes.vue'), meta: { title: 'Barcodes' } },
  { path: '/assets', component: () => import('@/pages/assets.vue'), meta: { title: 'Assets' } },

  { path: '/invoices', component: () => import('@/pages/sales-invoices.vue'), meta: { title: 'Invoices' } },
  { path: '/sales-invoices', redirect: '/invoices' },
  { path: '/purchase-invoice', component: () => import('@/pages/purchase-invoice.vue'), meta: { title: 'Purchase Invoices' } },
  { path: '/expenses', component: () => import('@/pages/expenses.vue'), meta: { title: 'Expenses' } },
  { path: '/loans', component: () => import('@/pages/loans.vue'), meta: { title: 'Loans' } },
  { path: '/loan-payments', component: () => import('@/pages/loan-payments.vue'), meta: { title: 'Loan Payments' } },
  { path: '/discounts', redirect: '/discount-lists' },
  { path: '/pricing', component: () => import('@/pages/pricing.vue'), meta: { title: 'Coupons' } },
  { path: '/price-lists', component: () => import('@/pages/price-lists.vue'), meta: { title: 'Price Lists' } },
  { path: '/discount-lists', component: () => import('@/pages/discount-lists.vue'), meta: { title: 'Discount Lists' } },
  { path: '/invoice-types', redirect: '/invoices' },
  { path: '/currencies', component: () => import('@/pages/currencies.vue'), meta: { title: 'Currencies' } },
  { path: '/coupons', redirect: '/pricing' },
  { path: '/shifts', component: () => import('@/pages/shifts.vue'), meta: { title: 'Shifts' } },

  { path: '/clients', component: () => import('@/pages/clients.vue'), meta: { title: 'Clients' } },
  { path: '/clients/:id', component: () => import('@/pages/client-page.vue'), meta: { title: 'Client' } },
  { path: '/suppliers', component: () => import('@/pages/suppliers.vue'), meta: { title: 'Suppliers' } },

  { path: '/employees', component: () => import('@/pages/employees.vue'), meta: { title: 'Employees' } },
  { path: '/employees/:id', component: () => import('@/pages/employee_page.vue'), meta: { title: 'Employee' } },
  { path: '/attendance', component: () => import('@/pages/Attendance.vue'), meta: { title: 'Attendance' } },
  { path: '/calendar', component: () => import('@/pages/calendar.vue'), meta: { title: 'Calendar' } },

  { path: '/stats', component: () => import('@/pages/stats.vue'), meta: { title: 'Statistics' } },
  { path: '/users', component: () => import('@/pages/users.vue'), meta: { title: 'Users' } },
  { path: '/settings', component: () => import('@/pages/settings.vue'), meta: { title: 'Settings' } },
  { path: '/logs', component: () => import('@/pages/logs.vue'), meta: { title: 'Logs' } },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: setupLayouts(routes),
})

function updateDocumentTitle (route) {
  const pageTitle = route.meta.title
  document.title = pageTitle ? `GB Cloud - ${pageTitle}` : 'GB Cloud'
}

router.beforeEach(to => {
  const authStore = useAuthStore()
  if (to.path !== '/login' && !authStore.isAuthenticated) {
    return '/login'
  }
  if (to.path === '/login' && authStore.isAuthenticated) {
    return '/'
  }
})

router.afterEach(to => {
  updateDocumentTitle(to)
})

router.onError((err, to) => {
  if (err?.message?.includes?.('Failed to fetch dynamically imported module')) {
    if (localStorage.getItem('vuetify:dynamic-reload')) {
      console.error('Dynamic import error, reloading page did not fix it', err)
    } else {
      console.log('Reloading page to fix dynamic import error')
      localStorage.setItem('vuetify:dynamic-reload', 'true')
      location.assign(to.fullPath)
    }
  } else {
    console.error(err)
  }
})

router.isReady().then(() => {
  updateDocumentTitle(router.currentRoute.value)
  localStorage.removeItem('vuetify:dynamic-reload')
})

export default router
