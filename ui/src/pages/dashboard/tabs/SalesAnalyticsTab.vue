<template>
  <v-container class="pa-0 fill-height mx-2 mb-4" fluid>
    <v-row class="fill-height" no-gutters>
      <v-col v-for="panel in panels" :key="panel.key" class="d-flex flex-column" cols="6">
        <v-card border class="d-flex flex-column fill-height" flat tile>
          <v-card-title class="text-caption text-uppercase text-grey py-1 px-3">{{ panel.title }}</v-card-title>
          <v-divider />
          <v-list class="pa-0 flex-grow-1" density="compact">
            <v-list-item v-for="row in tableItems(panel.key)" :key="row.label" class="px-3">
              <template #title>{{ row.label }}</template>
              <template #append>
                <span class="text-caption mr-3">{{ row.value }}</span>
                <span class="text-caption text-grey">{{ row.pct }}</span>
              </template>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
  const panels = [
    { key: 'categories', title: 'Top Categories' },
    { key: 'subcategories', title: 'Top SubCategories' },
    { key: 'kinds', title: 'Top Kinds' },
    { key: 'brands', title: 'Top Brands' },
  ]

  const data = {
    categories: [{ label: "Men's Wear", value: 340 }, { label: "Women's Wear", value: 280 }, { label: "Kids' Wear", value: 175 }, { label: 'Footwear', value: 155 }, { label: 'Accessories', value: 110 }, { label: 'Sportswear', value: 85 }],
    subcategories: [{ label: 'Shirts', value: 190 }, { label: 'Pants', value: 160 }, { label: 'Dresses', value: 145 }, { label: 'Jackets', value: 120 }, { label: 'Shoes', value: 100 }, { label: 'Bags', value: 85 }, { label: 'Hats', value: 55 }],
    kinds: [{ label: 'Casual', value: 310 }, { label: 'Formal', value: 220 }, { label: 'Athletic', value: 175 }, { label: 'Luxury', value: 120 }, { label: 'Seasonal', value: 80 }],
    brands: [{ label: 'Nike', value: 280 }, { label: 'Adidas', value: 240 }, { label: 'Zara', value: 210 }, { label: 'H&M', value: 185 }, { label: "Levi's", value: 140 }, { label: 'Gucci', value: 110 }, { label: 'Prada', value: 85 }],
  }

  function tableItems (key) {
    const rows = data[key]
    const total = rows.reduce((s, r) => s + r.value, 0)
    return rows.map(r => ({ ...r, pct: ((r.value / total) * 100).toFixed(1) + '%' }))
  }
</script>

<style scoped>
.v-row {
  flex-wrap: wrap;
}

.v-col {
  height: 50%;
}
</style>
