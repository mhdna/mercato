export const navItems = [
  { title: 'Home', icon: 'mdi-home', to: '/' },
  {
    title: 'Analysis',
    icon: 'mdi-chart-bar',
    children: [
      { title: 'Dashboard', icon: 'mdi-view-dashboard', to: '/dashboard' },
      { title: 'Daily Income', icon: 'mdi-chart-bar', to: '/daily-income' },
      { title: 'Financials', icon: 'mdi-currency-usd', to: '/financials' },
      { title: 'Targets', icon: 'mdi-target', to: '/targets' },
      { title: 'Footfall', icon: 'mdi-walk', to: '/footfall' },
      { title: 'Conversion', icon: 'mdi-percent-outline', to: '/conversion' },
      { title: 'Sales Analytics', icon: 'mdi-chart-line', to: '/sales-analytics' },
      { title: 'Sales Analysis', icon: 'mdi-chart-box-outline', to: '/sales-analysis' },
      { title: 'Sales Dashboard', icon: 'mdi-view-dashboard', to: '/sales-dashboard' },
      { title: 'Salesperson Performance', icon: 'mdi-account-cash', to: '/salesperson-performance' },
    ],
  },
  { title: 'Assets Treemap', icon: 'mdi-view-gallery', to: '/assets-treemap' },
  {
    title: 'Alerts',
    icon: 'mdi-alert-circle-outline',
    to: '/alerts',
    append: 'alert-count',
  },
  {
    title: 'Stock Health',
    icon: 'mdi-heart-pulse',
    to: '/stock-health',
    append: 'status',
  },
  {
    title: 'Wholesale',
    icon: 'mdi-cart',
    children: [
      { title: 'Clients', icon: 'mdi-account', to: '/clients' },
      { title: 'Loyalty', icon: 'mdi-trophy', to: '/loyalty' },
      {
        title: 'Invoices',
        icon: 'mdi-format-list-bulleted-square',
        to: '/invoices',
      },
    ],
  },
  {
    title: 'Branches',
    icon: 'mdi-store-outline',
    children: [
      { title: 'Branches', icon: 'mdi-store-outline', to: '/branches' },
      { title: 'Branch Settings', icon: 'mdi-store-cog-outline', to: '/branch-settings' },
    ],
  },
  {
    title: 'Pricing',
    icon: 'mdi-sale',
    children: [
      { title: 'Price Lists', icon: 'mdi-format-list-bulleted', to: '/price-lists' },
      { title: 'Discount Lists', icon: 'mdi-sale', to: '/discount-lists' },
      { title: 'Coupons', icon: 'mdi-ticket-percent', to: '/pricing' },
    ],
  },
  {
    title: 'Products',
    icon: 'mdi-package-variant',
    children: [
      { title: 'Products', icon: 'mdi-package-variant', to: '/products' },
      { title: 'Colors & Sizes', icon: 'mdi-palette', to: '/colors-sizes' },
      { title: 'Attributes', icon: 'mdi-tag-multiple', to: '/attributes' },
    ],
  },
  {
    title: 'Accounting',
    icon: 'mdi-currency-usd',
    children: [
      { title: 'Expenses', icon: 'mdi-currency-usd-off', to: '/expenses' },
      { title: 'Loans', icon: 'mdi-hand-coin', to: '/loans' },
      { title: 'Currencies & Payments', icon: 'mdi-cash-multiple', to: '/currencies' },
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
      { title: 'Purchase Invoices', icon: 'mdi-invoice', to: '/purchase-invoice' },
      { title: 'Transfers', icon: 'mdi-transfer', to: '/transfers' },
      { title: 'Stock Counts', icon: 'mdi-clipboard-list-outline', to: '/stock-counts' },
      { title: 'Stock Movements', icon: 'mdi-swap-vertical', to: '/stock-movements' },
      { title: 'Suppliers', icon: 'mdi-truck-delivery', to: '/suppliers' },
      { title: 'Barcodes', icon: 'mdi-barcode', to: '/barcodes' },
    ],
  },
]

export const appendItems = [
  { title: 'Calendar', icon: 'mdi-calendar', to: '/calendar' },
  { title: 'Storage', icon: 'mdi-google-drive', to: '/storage' },
  { title: 'Logs', icon: 'mdi-text', to: '/logs' },
  { title: 'Audit Log', icon: 'mdi-history', to: '/audit-log' },
  { title: 'Settings', icon: 'mdi-cog-outline', to: '/settings' },
]
