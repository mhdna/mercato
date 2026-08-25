<template>
  <v-chart
    autoresize
    class="chart"
    :option="option"
    :theme="currentTheme"
  />
  <!-- :loading="loading" -->
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import VChart from 'vue-echarts'
  import { useTheme } from 'vuetify'

  const theme = useTheme()
  const currentTheme = computed(() => theme.global.current.value?.dark ? 'dark' : 'light')

  // const loading = ref(true)
  const option = ref({})

  function generateInventoryData () {
    return [
      {
        name: 'Turkish CFS -40',
        code: 'T12830',
        color: '#9B59B6',
        children: [
          { name: 'Lady Sweatshirt', value: 40, code: 'T12830-LS' },
          { name: 'Lady Sweater', value: 30, code: 'T12830-LSW' },
          { name: 'Kids Sweatshirt', value: 25, code: 'T12830-KS' },
          { name: 'Lady Pants', value: 18, code: 'T12830-LP' },
          { name: 'Men Pants', value: 12, code: 'T12830-MP' },
          { name: 'Men Sweater', value: 10, code: 'T12830-MS' },
        ],
      },
      {
        name: 'European Collection',
        code: 'EU8920',
        color: '#E67E22',
        children: [
          { name: 'Premium Jacket', value: 22, code: 'EU8920-PJ' },
          { name: 'Wool Coat', value: 18, code: 'EU8920-WC' },
          { name: 'Designer Dress', value: 15, code: 'EU8920-DD' },
          { name: 'Leather Jacket', value: 14, code: 'EU8920-LJ' },
          { name: 'Cashmere Scarf', value: 12, code: 'EU8920-CS' },
          { name: 'Silk Blouse', value: 10, code: 'EU8920-SB' },
          { name: 'Tailored Blazer', value: 7, code: 'EU8920-TB' },
        ],
      },
      {
        name: 'Lebanese VIP',
        code: 'LV5640',
        color: '#ECF0F1',
        children: [
          { name: 'Evening Gown', value: 18, code: 'LV5640-EG' },
          { name: 'Luxury Handbag', value: 15, code: 'LV5640-LH' },
          { name: 'Designer Shoes', value: 13, code: 'LV5640-DS' },
          { name: 'Premium Suit', value: 12, code: 'LV5640-PS' },
          { name: 'Gold Accessories', value: 10, code: 'LV5640-GA' },
          { name: 'Silk Kimono', value: 8, code: 'LV5640-SK' },
        ],
      },
      {
        name: 'Lebanese Regular',
        code: 'LR3450',
        color: '#1A237E',
        children: [
          { name: 'Casual Jeans', value: 35, code: 'LR3450-CJ' },
          { name: 'T-Shirts', value: 32, code: 'LR3450-TS' },
          { name: 'Summer Dress', value: 28, code: 'LR3450-SD' },
          { name: 'Shorts', value: 22, code: 'LR3450-SH' },
          { name: 'Polo Shirts', value: 20, code: 'LR3450-PS' },
          { name: 'Casual Shoes', value: 19, code: 'LR3450-CS' },
        ],
      },
      {
        name: 'Asian Import',
        code: 'AI7830',
        color: '#C62828',
        children: [
          { name: 'Silk Scarves', value: 28, code: 'AI7830-SS' },
          { name: 'Cotton Shirts', value: 25, code: 'AI7830-CS' },
          { name: 'Bamboo Socks', value: 20, code: 'AI7830-BS' },
          { name: 'Linen Pants', value: 18, code: 'AI7830-LP' },
          { name: 'Traditional Wear', value: 12, code: 'AI7830-TW' },
          { name: 'Accessories', value: 9, code: 'AI7830-AC' },
        ],
      },
      {
        name: 'American Casual',
        code: 'AC9210',
        color: '#2E7D32',
        children: [
          { name: 'Denim Jacket', value: 20, code: 'AC9210-DJ' },
          { name: 'Hoodies', value: 18, code: 'AC9210-HD' },
          { name: 'Cargo Pants', value: 16, code: 'AC9210-CP' },
          { name: 'Baseball Caps', value: 14, code: 'AC9210-BC' },
          { name: 'Sneakers', value: 12, code: 'AC9210-SN' },
          { name: 'Backpacks', value: 9, code: 'AC9210-BP' },
        ],
      },
      {
        name: 'Winter Collection',
        code: 'WC4560',
        color: '#00838F',
        children: [
          { name: 'Puffer Jackets', value: 22, code: 'WC4560-PJ' },
          { name: 'Thermal Wear', value: 20, code: 'WC4560-TW' },
          { name: 'Wool Sweaters', value: 18, code: 'WC4560-WS' },
          { name: 'Winter Boots', value: 15, code: 'WC4560-WB' },
          { name: 'Gloves & Hats', value: 12, code: 'WC4560-GH' },
          { name: 'Fleece Blankets', value: 7, code: 'WC4560-FB' },
        ],
      },
      {
        name: 'Sports & Active',
        code: 'SA6780',
        color: '#F57F17',
        children: [
          { name: 'Yoga Pants', value: 30, code: 'SA6780-YP' },
          { name: 'Running Shoes', value: 26, code: 'SA6780-RS' },
          { name: 'Sports Bras', value: 22, code: 'SA6780-SB' },
          { name: 'Gym Shorts', value: 20, code: 'SA6780-GS' },
          { name: 'Track Jackets', value: 18, code: 'SA6780-TJ' },
          { name: 'Water Bottles', value: 12, code: 'SA6780-WB' },
        ],
      },
      {
        name: 'Kids Collection',
        code: 'KC2340',
        color: '#FF6F00',
        children: [
          { name: 'School Uniforms', value: 25, code: 'KC2340-SU' },
          { name: 'Play Clothes', value: 20, code: 'KC2340-PC' },
          { name: 'Kids Shoes', value: 16, code: 'KC2340-KS' },
          { name: 'Pajamas', value: 14, code: 'KC2340-PJ' },
          { name: 'Backpacks', value: 12, code: 'KC2340-BP' },
        ],
      },
      {
        name: 'Premium Luxury',
        code: 'PL9870',
        color: '#4A148C',
        children: [
          { name: 'Designer Bags', value: 15, code: 'PL9870-DB' },
          { name: 'Luxury Watches', value: 12, code: 'PL9870-LW' },
          { name: 'Diamond Jewelry', value: 11, code: 'PL9870-DJ' },
          { name: 'Fur Coats', value: 10, code: 'PL9870-FC' },
          { name: 'Limited Edition', value: 9, code: 'PL9870-LE' },
          { name: 'Collectibles', value: 7, code: 'PL9870-CL' },
        ],
      },
    ]
  }

  function createChartOption (data) {
    return {
      tooltip: {
        appendTo: 'body',
        formatter: info => {
          const { name, value, data } = info
          let content = `<strong>${name}</strong>`
          if (data.code) {
            content += `<br/>Code: ${data.code}`
          }
          if (value) {
            content += `<br/>Quantity: ${value}`
          }
          return content
        },
      },
      series: [
        {
          name: 'Inventory',
          type: 'treemap',
          visibleMin: 300,
          data: data,
          leafDepth: null,
          width: '100%',
          height: '100%',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          roam: false,
          nodeClick: 'link',
          label: {
            show: true,
            formatter: params => {
              if (params.value) {
                return `${params.name}\n${params.value}`
              }
              return params.name
            },
            fontSize: 14,
            fontWeight: 'bold',
          },
          upperLabel: {
            backgroundColor: '#000',
            show: true,
            height: 18,
            color: '#fff',
            fontSize: 13,
            fontWeight: 'bold',
          },
          itemStyle: {
            borderColor: '#fff',
            borderWidth: 0,
            gapWidth: 2,
          },
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
              borderColor: '#FFD700',
              borderWidth: 3,
            },
            label: {
              fontSize: 16,
              fontWeight: 'bold',
            },
          },
          levels: [
            {
              itemStyle: {
                borderColor: '#000',
                borderWidth: 0,
                gapWidth: 2,
              },
              upperLabel: {
                show: true,
              },
            },
            {
              itemStyle: {
                borderColor: '#000',
                borderWidth: 0,
                gapWidth: 2,
              },
              label: {
                fontSize: 12,
              },
            },
          ],
        },
      ],
    }
  }

  function loadData () {
    setTimeout(() => {
      const inventoryData = generateInventoryData()
      option.value = createChartOption(inventoryData)
      // loading.value = false
    }, 500)
  }

  onMounted(() => {
    loadData()
  })
</script>

<style scoped>
.chart {
  height: 100%;
  cursor: pointer;
}
</style>
