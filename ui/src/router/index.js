import { setupLayouts } from 'virtual:generated-layouts'
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  { path: '/', component: () => import('@/pages/dashboard.vue'), meta: { title: 'Dashboard' } },
  // { path: '/dashboard', component: () => import('@/pages/dashboard.vue'), meta: { title: 'Dashboard' } },
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
  { path: '/branch-invoices', component: () => import('@/pages/branch-invoices.vue'), meta: { title: 'Branch Invoices' } },
  { path: '/transfers', component: () => import('@/pages/transfers.vue'), meta: { title: 'Transfers' } },
  { path: '/barcode-printing', component: () => import('@/pages/barcode-printing.vue'), meta: { title: 'Barcode Printing' } },
  { path: '/colors-sizes', component: () => import('@/pages/Colors_Sizes.vue'), meta: { title: 'Colors & Sizes' } },
  { path: '/barcodes', component: () => import('@/pages/barcodes.vue'), meta: { title: 'Barcodes' } },
  { path: '/assets', component: () => import('@/pages/assets.vue'), meta: { title: 'Assets' } },

  { path: '/sales-invoices', component: () => import('@/pages/sales-invoices.vue'), meta: { title: 'Sales Invoices' } },
  { path: '/purchase-invoice', component: () => import('@/pages/purchase-invoice.vue'), meta: { title: 'Purchase Invoices' } },
  { path: '/expenses', component: () => import('@/pages/expenses.vue'), meta: { title: 'Expenses' } },
  { path: '/discounts', component: () => import('@/pages/discounts.vue'), meta: { title: 'Discounts' } },
  { path: '/invoice-types', component: () => import('@/pages/invoice-types.vue'), meta: { title: 'Invoice Types' } },
  { path: '/currencies', component: () => import('@/pages/currencies.vue'), meta: { title: 'Currencies' } },
  { path: '/coupons', component: () => import('@/pages/coupons.vue'), meta: { title: 'Coupons' } },
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
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: setupLayouts(routes),
})

router.beforeEach(to => {
  const authStore = useAuthStore()
  if (to.path !== '/login' && !authStore.isAuthenticated) {
    return '/login'
  }
  if (to.path === '/login' && authStore.isAuthenticated) {
    return '/'
  }
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
  localStorage.removeItem('vuetify:dynamic-reload')
})

export default router
