// Fake cost-of-goods data for the warehouses treemap.
// Later this will be replaced by real data coming from the cost-of-goods API.
// Each warehouse holds a list of colored groups; every group holds items with
// a cost (USD), a weight (KG) and a stock code.

const GROUP_COLORS = {
  'Lebanese Outlets': '#1A237E',
  'Turkish Clothes': '#C62828',
  'European Collection': '#E67E22',
  'Asian Import': '#00838F',
  'American Casual': '#2E7D32',
  'Sports & Active': '#F57F17',
  'Premium Luxury': '#4A148C',
  'Winter Collection': '#0277BD',
  'Kids Collection': '#AD1457',
}

function group (name, items) {
  return { name, color: GROUP_COLORS[name] || '#546E7A', items }
}

export const warehouses = [
  {
    id: 'beirut-central',
    name: 'Beirut Central',
    groups: [
      group('Lebanese Outlets', [
        { name: 'Lebanese Shirts', cost: 5000, weight: 300, code: 'L12FDKF8' },
        { name: 'Lebanese Pants', cost: 3000, weight: 210, code: 'L83DJFEJ' },
        { name: 'Lebanese Jackets', cost: 4200, weight: 260, code: 'L55KDIE1' },
        { name: 'Lebanese Skirts', cost: 1800, weight: 90, code: 'L20QWME4' },
        { name: 'Lebanese Abaya', cost: 2600, weight: 130, code: 'L44TRGH6' },
      ]),
      group('Turkish Clothes', [
        { name: 'Turkish Sweatshirts', cost: 6400, weight: 420, code: 'T90ALKD2' },
        { name: 'Turkish Coats', cost: 7800, weight: 510, code: 'T14ZXPQ9' },
        { name: 'Turkish Denim', cost: 5200, weight: 380, code: 'T61MNBV7' },
        { name: 'Turkish Knitwear', cost: 3900, weight: 240, code: 'T27HGFD5' },
      ]),
      group('European Collection', [
        { name: 'Wool Coats', cost: 9100, weight: 240, code: 'E33LWKA5' },
        { name: 'Cashmere Scarves', cost: 3400, weight: 40, code: 'E71PLMD8' },
        { name: 'Tailored Blazers', cost: 6600, weight: 150, code: 'E12OKSD3' },
        { name: 'Silk Blouses', cost: 4100, weight: 60, code: 'E58MKAP2' },
      ]),
      group('Premium Luxury', [
        { name: 'Designer Bags', cost: 12_400, weight: 80, code: 'P01LUXA1' },
        { name: 'Luxury Watches', cost: 15_800, weight: 20, code: 'P02LUXW2' },
        { name: 'Fur Coats', cost: 9800, weight: 130, code: 'P03LUXF3' },
      ]),
    ],
  },
  {
    id: 'tripoli-depot',
    name: 'Tripoli Depot',
    groups: [
      group('Lebanese Outlets', [
        { name: 'Lebanese Shirts', cost: 3200, weight: 190, code: 'L12FDKF8' },
        { name: 'Lebanese Abaya', cost: 2600, weight: 120, code: 'L44TRGH6' },
        { name: 'Lebanese Shoes', cost: 4100, weight: 300, code: 'L09PLKJ2' },
        { name: 'Lebanese Belts', cost: 1200, weight: 50, code: 'L61XCVB3' },
      ]),
      group('Asian Import', [
        { name: 'Cotton Shirts', cost: 2800, weight: 220, code: 'A21WERT9' },
        { name: 'Silk Scarves', cost: 3900, weight: 60, code: 'A88LKMN4' },
        { name: 'Linen Pants', cost: 2300, weight: 170, code: 'A53QAZX1' },
        { name: 'Bamboo Socks', cost: 900, weight: 45, code: 'A17MJUY7' },
        { name: 'Traditional Wear', cost: 3300, weight: 150, code: 'A42POIU8' },
      ]),
      group('Winter Collection', [
        { name: 'Puffer Jackets', cost: 6800, weight: 320, code: 'W10PLKM2' },
        { name: 'Thermal Wear', cost: 3100, weight: 180, code: 'W22OKJI7' },
        { name: 'Wool Sweaters', cost: 4400, weight: 260, code: 'W35MNBV1' },
        { name: 'Winter Boots', cost: 5200, weight: 400, code: 'W47ZAQW9' },
      ]),
    ],
  },
  {
    id: 'saida-store',
    name: 'Saida Store',
    groups: [
      group('Turkish Clothes', [
        { name: 'Turkish Kids Wear', cost: 3100, weight: 180, code: 'T77KDLA3' },
        { name: 'Turkish Pyjamas', cost: 2200, weight: 140, code: 'T39SLDK8' },
        { name: 'Turkish Towels', cost: 2900, weight: 300, code: 'T51QWER6' },
      ]),
      group('Sports & Active', [
        { name: 'Yoga Pants', cost: 4300, weight: 210, code: 'S12PLKO9' },
        { name: 'Running Shoes', cost: 5600, weight: 260, code: 'S45MNBH2' },
        { name: 'Track Jackets', cost: 3700, weight: 190, code: 'S88QWER4' },
        { name: 'Gym Shorts', cost: 2100, weight: 120, code: 'S63ZXCV7' },
      ]),
      group('Premium Luxury', [
        { name: 'Designer Bags', cost: 12_400, weight: 80, code: 'P01LUXA1' },
        { name: 'Luxury Watches', cost: 15_800, weight: 20, code: 'P02LUXW2' },
        { name: 'Fur Coats', cost: 9800, weight: 130, code: 'P03LUXF3' },
        { name: 'Gold Accessories', cost: 8700, weight: 15, code: 'P04LUXG4' },
      ]),
      group('Kids Collection', [
        { name: 'School Uniforms', cost: 3600, weight: 240, code: 'K11ASDF2' },
        { name: 'Play Clothes', cost: 2400, weight: 160, code: 'K22GHJK8' },
        { name: 'Kids Shoes', cost: 3000, weight: 200, code: 'K33LZXC1' },
      ]),
    ],
  },
  {
    id: 'zahle-hub',
    name: 'Zahle Hub',
    groups: [
      group('American Casual', [
        { name: 'Denim Jackets', cost: 5100, weight: 300, code: 'C10DENJ2' },
        { name: 'Hoodies', cost: 4300, weight: 280, code: 'C21HOOD7' },
        { name: 'Cargo Pants', cost: 3800, weight: 250, code: 'C32CARG1' },
        { name: 'Baseball Caps', cost: 1500, weight: 60, code: 'C43CAPS9' },
        { name: 'Sneakers', cost: 4900, weight: 220, code: 'C54SNKR4' },
      ]),
      group('European Collection', [
        { name: 'Leather Jackets', cost: 8800, weight: 280, code: 'E80LEAT3' },
        { name: 'Designer Dresses', cost: 5400, weight: 90, code: 'E91DRES6' },
        { name: 'Wool Coats', cost: 9100, weight: 240, code: 'E33LWKA5' },
      ]),
      group('Winter Collection', [
        { name: 'Fleece Blankets', cost: 2600, weight: 200, code: 'W55FLEE2' },
        { name: 'Gloves & Hats', cost: 1900, weight: 70, code: 'W66GLOV8' },
        { name: 'Ski Wear', cost: 7300, weight: 340, code: 'W77SKIW1' },
      ]),
    ],
  },
  {
    id: 'jounieh-outlet',
    name: 'Jounieh Outlet',
    groups: [
      group('Lebanese Outlets', [
        { name: 'Lebanese Shirts', cost: 4200, weight: 250, code: 'L12FDKF8' },
        { name: 'Lebanese Pants', cost: 3400, weight: 230, code: 'L83DJFEJ' },
        { name: 'Lebanese Dresses', cost: 3900, weight: 140, code: 'L77MKOP5' },
      ]),
      group('Sports & Active', [
        { name: 'Sports Bras', cost: 2700, weight: 80, code: 'S71BRAS3' },
        { name: 'Compression Wear', cost: 3300, weight: 110, code: 'S82COMP9' },
        { name: 'Water Bottles', cost: 1100, weight: 90, code: 'S93BOTL5' },
      ]),
      group('Kids Collection', [
        { name: 'Pajamas', cost: 2200, weight: 130, code: 'K44PJMS7' },
        { name: 'Backpacks', cost: 2600, weight: 150, code: 'K55BPAK2' },
        { name: 'Winter Coats', cost: 3500, weight: 210, code: 'K66WCOT8' },
      ]),
      group('Turkish Clothes', [
        { name: 'Turkish Denim', cost: 5200, weight: 380, code: 'T61MNBV7' },
        { name: 'Turkish Shirts', cost: 4000, weight: 300, code: 'T12SHRT4' },
      ]),
    ],
  },
]

export const warehouseOptions = [
  { title: 'All Warehouses', value: 'all' },
  ...warehouses.map(w => ({ title: w.name, value: w.id })),
]

// Total cost (USD) and weight (KG) for the given warehouse id ('all' = every warehouse).
export function warehouseTotals (warehouseId = 'all') {
  const source = warehouseId === 'all'
    ? warehouses
    : warehouses.filter(w => w.id === warehouseId)

  let cost = 0
  let weight = 0
  for (const w of source) {
    for (const g of w.groups) {
      for (const item of g.items) {
        cost += item.cost
        weight += item.weight
      }
    }
  }
  return { cost, weight }
}

// Build echarts treemap nodes for the given warehouse id ('all' merges every warehouse).
// Group cost is the sum of its items so the coloured block area maps to cost of goods.
export function buildTreemapData (warehouseId = 'all') {
  const source = warehouseId === 'all'
    ? warehouses
    : warehouses.filter(w => w.id === warehouseId)

  const merged = new Map()
  for (const w of source) {
    for (const g of w.groups) {
      if (!merged.has(g.name)) {
        merged.set(g.name, { name: g.name, color: g.color, items: [] })
      }
      merged.get(g.name).items.push(...g.items.map(item => ({
        ...item,
        // keep the warehouse name on the leaf when showing everything
        warehouse: w.name,
      })))
    }
  }

  return [...merged.values()].map(g => {
    const items = g.items.map(item => ({
      name: item.name,
      value: [item.cost, item.weight],
      code: item.code,
      warehouse: item.warehouse,
    }))
    const cost = items.reduce((sum, i) => sum + i.value[0], 0)
    const weight = items.reduce((sum, i) => sum + i.value[1], 0)
    return {
      name: g.name,
      value: [cost, weight],
      itemStyle: { color: g.color },
      children: items,
    }
  })
}
