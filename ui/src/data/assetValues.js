// Fake asset-register data for the assets treemap.
// Later this will be replaced by real data coming from the assets API.
// Each location holds a list of coloured category groups; every group holds
// assets with a book value (USD), a unit count and an asset-tag code.

const GROUP_COLORS = {
  'IT Equipment': '#1565C0',
  'Furniture': '#6D4C41',
  'Vehicles': '#2E7D32',
  'Machinery': '#C62828',
  'Fixtures & Fittings': '#F57F17',
  'Office Equipment': '#00838F',
  'Security': '#4A148C',
  'Kitchen': '#AD1457',
}

function group (name, items) {
  return { name, color: GROUP_COLORS[name] || '#546E7A', items }
}

export const locations = [
  {
    id: 'head-office',
    name: 'Head Office',
    groups: [
      group('IT Equipment', [
        { name: 'Laptops', value: 42_000, qty: 35, code: 'IT-LT-001' },
        { name: 'Desktops', value: 18_500, qty: 22, code: 'IT-DT-014' },
        { name: 'Servers', value: 55_000, qty: 4, code: 'IT-SRV-002' },
        { name: 'Network Switches', value: 9800, qty: 12, code: 'IT-NET-041' },
        { name: 'Monitors', value: 7600, qty: 48, code: 'IT-MON-077' },
      ]),
      group('Furniture', [
        { name: 'Desks', value: 12_400, qty: 40, code: 'FN-DSK-003' },
        { name: 'Office Chairs', value: 9600, qty: 55, code: 'FN-CHR-018' },
        { name: 'Meeting Tables', value: 6800, qty: 8, code: 'FN-TBL-006' },
        { name: 'Storage Cabinets', value: 4200, qty: 24, code: 'FN-CAB-030' },
      ]),
      group('Office Equipment', [
        { name: 'Printers', value: 8300, qty: 9, code: 'OE-PRN-011' },
        { name: 'Projectors', value: 5400, qty: 6, code: 'OE-PRJ-004' },
        { name: 'Shredders', value: 1900, qty: 7, code: 'OE-SHR-022' },
      ]),
      group('Security', [
        { name: 'CCTV Cameras', value: 7200, qty: 30, code: 'SC-CAM-001' },
        { name: 'Access Control', value: 5600, qty: 10, code: 'SC-ACC-009' },
        { name: 'Alarm Systems', value: 3100, qty: 5, code: 'SC-ALM-013' },
      ]),
    ],
  },
  {
    id: 'beirut-branch',
    name: 'Beirut Branch',
    groups: [
      group('Fixtures & Fittings', [
        { name: 'Display Racks', value: 9200, qty: 45, code: 'FX-RCK-002' },
        { name: 'Mannequins', value: 4100, qty: 60, code: 'FX-MAN-018' },
        { name: 'Lighting Rigs', value: 6300, qty: 20, code: 'FX-LGT-007' },
        { name: 'Fitting Rooms', value: 5200, qty: 8, code: 'FX-FIT-025' },
      ]),
      group('IT Equipment', [
        { name: 'POS Terminals', value: 14_600, qty: 12, code: 'IT-POS-003' },
        { name: 'Barcode Scanners', value: 3400, qty: 18, code: 'IT-SCN-041' },
        { name: 'Receipt Printers', value: 2600, qty: 14, code: 'IT-RPR-055' },
      ]),
      group('Security', [
        { name: 'CCTV Cameras', value: 5400, qty: 22, code: 'SC-CAM-001' },
        { name: 'EAS Gates', value: 4800, qty: 6, code: 'SC-EAS-017' },
      ]),
      group('Furniture', [
        { name: 'Checkout Counters', value: 6100, qty: 6, code: 'FN-CNT-012' },
        { name: 'Staff Lockers', value: 2200, qty: 20, code: 'FN-LKR-034' },
      ]),
    ],
  },
  {
    id: 'tripoli-branch',
    name: 'Tripoli Branch',
    groups: [
      group('Fixtures & Fittings', [
        { name: 'Display Racks', value: 6800, qty: 34, code: 'FX-RCK-002' },
        { name: 'Wall Shelving', value: 4400, qty: 40, code: 'FX-SHL-021' },
        { name: 'Mannequins', value: 3100, qty: 45, code: 'FX-MAN-018' },
      ]),
      group('IT Equipment', [
        { name: 'POS Terminals', value: 9800, qty: 8, code: 'IT-POS-003' },
        { name: 'Barcode Scanners', value: 2400, qty: 12, code: 'IT-SCN-041' },
      ]),
      group('Office Equipment', [
        { name: 'Printers', value: 2100, qty: 3, code: 'OE-PRN-011' },
        { name: 'Safes', value: 3600, qty: 2, code: 'OE-SAF-008' },
      ]),
    ],
  },
  {
    id: 'central-warehouse',
    name: 'Central Warehouse',
    groups: [
      group('Machinery', [
        { name: 'Forklifts', value: 48_000, qty: 4, code: 'MC-FRK-001' },
        { name: 'Pallet Jacks', value: 6200, qty: 14, code: 'MC-PLJ-010' },
        { name: 'Conveyor Belts', value: 22_000, qty: 3, code: 'MC-CNV-005' },
        { name: 'Shrink Wrappers', value: 7400, qty: 5, code: 'MC-SWR-019' },
      ]),
      group('Vehicles', [
        { name: 'Delivery Vans', value: 96_000, qty: 6, code: 'VH-VAN-002' },
        { name: 'Box Trucks', value: 78_000, qty: 3, code: 'VH-TRK-004' },
        { name: 'Company Cars', value: 54_000, qty: 4, code: 'VH-CAR-011' },
      ]),
      group('IT Equipment', [
        { name: 'Handheld Scanners', value: 5600, qty: 25, code: 'IT-HHS-042' },
        { name: 'Label Printers', value: 4200, qty: 10, code: 'IT-LBP-056' },
      ]),
      group('Fixtures & Fittings', [
        { name: 'Pallet Racking', value: 31_000, qty: 120, code: 'FX-RAC-030' },
        { name: 'Mezzanine Floor', value: 26_000, qty: 1, code: 'FX-MEZ-001' },
      ]),
    ],
  },
  {
    id: 'jounieh-branch',
    name: 'Jounieh Branch',
    groups: [
      group('Fixtures & Fittings', [
        { name: 'Display Racks', value: 5200, qty: 26, code: 'FX-RCK-002' },
        { name: 'Lighting Rigs', value: 3900, qty: 14, code: 'FX-LGT-007' },
        { name: 'Fitting Rooms', value: 3400, qty: 5, code: 'FX-FIT-025' },
      ]),
      group('Kitchen', [
        { name: 'Coffee Machines', value: 2600, qty: 3, code: 'KT-COF-001' },
        { name: 'Refrigerators', value: 3100, qty: 4, code: 'KT-REF-006' },
        { name: 'Microwaves', value: 800, qty: 5, code: 'KT-MCW-012' },
      ]),
      group('IT Equipment', [
        { name: 'POS Terminals', value: 7300, qty: 6, code: 'IT-POS-003' },
        { name: 'Tablets', value: 2800, qty: 10, code: 'IT-TAB-063' },
      ]),
      group('Furniture', [
        { name: 'Checkout Counters', value: 4100, qty: 4, code: 'FN-CNT-012' },
        { name: 'Staff Lockers', value: 1600, qty: 14, code: 'FN-LKR-034' },
      ]),
    ],
  },
]

export const assetLocationOptions = [
  { title: 'All Locations', value: 'all' },
  ...locations.map(l => ({ title: l.name, value: l.id })),
]

// Total book value (USD) and unit count for the given location id ('all' = every location).
export function assetTotals (locationId = 'all') {
  const source = locationId === 'all'
    ? locations
    : locations.filter(l => l.id === locationId)

  let value = 0
  let qty = 0
  for (const l of source) {
    for (const g of l.groups) {
      for (const item of g.items) {
        value += item.value
        qty += item.qty
      }
    }
  }
  return { value, qty }
}

// Build echarts treemap nodes for the given location id ('all' merges every location).
// Group value is the sum of its assets so the coloured block area maps to book value.
export function buildTreemapData (locationId = 'all') {
  const source = locationId === 'all'
    ? locations
    : locations.filter(l => l.id === locationId)

  const merged = new Map()
  for (const l of source) {
    for (const g of l.groups) {
      if (!merged.has(g.name)) {
        merged.set(g.name, { name: g.name, color: g.color, items: [] })
      }
      merged.get(g.name).items.push(...g.items.map(item => ({
        ...item,
        // keep the location name on the leaf when showing everything
        location: l.name,
      })))
    }
  }

  return [...merged.values()].map(g => {
    const items = g.items.map(item => ({
      name: item.name,
      value: [item.value, item.qty],
      code: item.code,
      location: item.location,
    }))
    const value = items.reduce((sum, i) => sum + i.value[0], 0)
    const qty = items.reduce((sum, i) => sum + i.value[1], 0)
    return {
      name: g.name,
      value: [value, qty],
      itemStyle: { color: g.color },
      children: items,
    }
  })
}
