export const navItems = [
  { title: 'Dashboard', icon: 'mdi-view-dashboard', to: '/' },
  {
    title: 'Wholesale',
    icon: 'mdi-cart',
    children: [
      { title: 'Clients', icon: 'mdi-account', to: '/clients' },
      { title: 'Suppliers', icon: 'mdi-truck-delivery', to: '/suppliers' },
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
    ],
  },
  {
    title: 'Branches',
    icon: 'mdi-store-outline',
    children: [
      { title: 'Branches', icon: 'mdi-store-outline', to: '/branches' },
      {
        title: 'Branch Invoices',
        icon: 'mdi-invoice',
        to: '/branch-invoices',
      },
    ],
  },
  {
    title: 'Coupons & Discounts',
    icon: 'mdi-sale',
    children: [
      { title: 'Discounts', icon: 'mdi-sale', to: '/discounts' },
      { title: 'Coupons', icon: 'mdi-ticket-percent', to: '/coupons' },
    ],
  },
  {
    title: 'Products',
    icon: 'mdi-package-variant',
    children: [
      { title: 'Products', icon: 'mdi-package-variant', to: '/products' },
      { title: 'Colors & Sizes', icon: 'mdi-palette', to: '/colors-sizes' },
      {
        title: 'Invoice Types',
        icon: 'mdi-tag-multiple',
        to: '/invoice-types',
      },
    ],
  },
  {
    title: 'Accounting',
    icon: 'mdi-currency-usd',
    children: [
      { title: 'Expenses', icon: 'mdi-currency-usd-off', to: '/expenses' },
      { title: 'Loans', icon: 'mdi-hand-coin', to: '/loans' },
      { title: 'Loan Payments', icon: 'mdi-cash-refund', to: '/loan-payments' },
      { title: 'Currencies', icon: 'mdi-currency-usd', to: '/currencies' },
      { title: 'Assets', icon: 'mdi-hammer-wrench', to: '/assets' },
    ],
  },
  {
    title: 'Human Resources',
    icon: 'mdi-account-group',
    children: [
      { title: 'Employees', icon: 'mdi-account-group', to: '/employees' },
      {
        title: 'Attendance',
        icon: 'mdi-account-multiple-check-outline',
        to: '/attendance',
      },
      { title: 'Shifts', icon: 'mdi-clock-outline', to: '/shifts' },
    ],
  },
  {
    title: 'Inventory',
    icon: 'mdi-warehouse',
    children: [
      { title: 'Inventory', icon: 'mdi-warehouse', to: '/inventory' },
      { title: 'Warehouses', icon: 'mdi-warehouse', to: '/warehouses' },
      { title: 'Barcodes', icon: 'mdi-barcode', to: '/barcodes' },
      { title: 'Transfers', icon: 'mdi-transfer', to: '/transfers' },
    ],
  },
  {
    title: 'Stats',
    icon: 'mdi-chart-bar',
    children: [
      {
        title: 'Inventory Analysis',
        icon: 'mdi-sine-wave',
        to: '/inventory-analysis',
      },
      { title: 'Custom Reports', icon: 'mdi-chart-bar', to: '/stats' },
    ],
  },
]

export const appendItems = [
  { title: 'Calendar', icon: 'mdi-calendar', to: '/calendar' },
  { title: 'Storage', icon: 'mdi-google-drive', to: '/storage' },
  { title: 'Settings', icon: 'mdi-cog-outline', to: '/settings' },
]
