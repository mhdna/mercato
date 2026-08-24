export const navItems = [
  { title: 'Dashboard', icon: 'mdi-view-dashboard', to: '/' },
  {
    title: 'Inventory Analysis',
    icon: 'mdi-sine-wave',
    to: '/inventory-analysis',
  },
  { title: 'Clients', icon: 'mdi-account', to: '/clients' },
  { title: 'Suppliers', icon: 'mdi-truck-delivery', to: '/suppliers' },
  { title: 'Custom Reports', icon: 'mdi-chart-bar', to: '/stats' },
  { title: 'Discounts', icon: 'mdi-sale', to: '/discounts' },
  {
    title: 'Sales',
    icon: 'mdi-invoice',
    children: [
      {
        title: 'Sales Invoices',
        icon: 'mdi-format-list-bulleted-square',
        to: '/sales-invoices',
      },
      {
        title: 'Purchase Invoices',
        icon: 'mdi-invoice',
        to: '/purchase-invoice',
      },
      { title: 'Coupons', icon: 'mdi-ticket-percent', to: '/coupons' },
      { title: 'Shifts', icon: 'mdi-clock-outline', to: '/shifts' },
    ],
  },
  {
    title: 'Inventory',
    icon: 'mdi-cart',
    children: [
      { title: 'Products', icon: 'mdi-package-variant', to: '/products' },
      { title: 'Inventory', icon: 'mdi-warehouse', to: '/inventory' },
      { title: 'Barcodes', icon: 'mdi-barcode', to: '/barcodes' },
      { title: 'Transfers', icon: 'mdi-transfer', to: '/transfers' },
    ],
  },
  { title: 'Warehouses', icon: 'mdi-warehouse', to: '/warehouses' },
  { title: 'Assets', icon: 'mdi-hammer-wrench', to: '/assets' },
  { title: 'Colors & Sizes', icon: 'mdi-palette', to: '/colors-sizes' },
  { title: 'Currencies', icon: 'mdi-currency-usd', to: '/currencies' },
  { title: 'Expenses', icon: 'mdi-currency-usd-off', to: '/expenses' },
  { title: 'Employees', icon: 'mdi-account-group', to: '/employees' },
  {
    title: 'Attendance',
    icon: 'mdi-account-multiple-check-outline',
    to: '/attendance',
  },
]

export const appendItems = [
  { title: 'Calendar', icon: 'mdi-calendar', to: '/calendar' },
  { title: 'Storage', icon: 'mdi-google-drive', to: '/storage' },
]
