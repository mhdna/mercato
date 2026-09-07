<script setup lang="ts">
  import { computed } from 'vue'
  import VChart from 'vue-echarts'
  import SparklineCard from '@/components/Cards/SparklineCard.vue'
  import ClientRankTag from '@/components/ClientRankTag.vue'
  import Heatmap from '@/components/Stats/Heatmap.vue'

  // ---------------------------------------------------------------------------
  // Mock client. There's no /clients/:id endpoint yet, so this page renders a
  // representative profile; swap `client` and the series below for real data
  // once the detail route lands.
  // ---------------------------------------------------------------------------
  const client = {
    name: 'Aisha Miliki',
    initials: 'AM',
    id: 'dfkj3-fjie2-jkfdi2-klio9',
    phone: '+964 772 837 421',
    branch: 'Karrada Flagship',
    firstPurchase: '13 Apr 2026',
    rank: 'Gold Tier',
  }

  // Trailing 12 months, oldest first.
  const kpis = [
    {
      title: 'Money Spent',
      unit: '$',
      value: 728_000,
      prevValue: 610_000,
      series: [42_000, 38_000, 51_000, 47_000, 55_000, 61_000, 58_000, 67_000, 72_000, 69_000, 81_000, 88_000],
    },
    {
      title: 'Items Bought',
      unit: '',
      value: 148,
      prevValue: 121,
      series: [6, 5, 8, 7, 9, 11, 10, 12, 13, 12, 15, 17],
    },
    {
      title: 'Points Remaining',
      unit: '',
      value: 320,
      prevValue: 280,
      series: [120, 140, 160, 150, 180, 200, 240, 220, 260, 300, 280, 320],
    },
    {
      title: 'Points Redeemed',
      unit: '',
      value: 240,
      prevValue: 180,
      series: [0, 20, 20, 60, 60, 60, 100, 140, 140, 180, 180, 240],
    },
  ]

  // Lifetime spend vs. the next loyalty tier.
  const lifetimeSpent = 7280
  const nextTier = { name: 'Platinum Tier', threshold: 10_000 }
  const tierProgress = computed(() => Math.min(100, Math.round((lifetimeSpent / nextTier.threshold) * 100)))
  const tierRemaining = computed(() => nextTier.threshold - lifetimeSpent)

  const money = n => '$ ' + Number(n || 0).toLocaleString('en-US')

  // --- Spend by category (treemap) -----------------------------------------
  const categorySpend = [
    {
      name: 'Apparel',
      children: [
        { name: 'Dresses', value: 1450 },
        { name: 'Tops', value: 980 },
        { name: 'Outerwear', value: 690 },
      ],
    },
    {
      name: 'Footwear',
      children: [
        { name: 'Heels', value: 720 },
        { name: 'Sneakers', value: 640 },
        { name: 'Flats', value: 480 },
      ],
    },
    {
      name: 'Accessories',
      children: [
        { name: 'Bags', value: 720 },
        { name: 'Jewelry', value: 360 },
        { name: 'Belts', value: 180 },
      ],
    },
    {
      name: 'Beauty',
      children: [
        { name: 'Skincare', value: 430 },
        { name: 'Fragrance', value: 330 },
      ],
    },
  ]

  const treemapOption = {
    backgroundColor: 'transparent',
    tooltip: {
      appendTo: 'body',
      formatter: info => `<b>${info.name}</b><br>${money((info.value || [])[0] ?? info.value)}`,
    },
    series: [
      {
        type: 'treemap',
        roam: false,
        nodeClick: 'zoomToNode',
        top: 8,
        left: 4,
        right: 4,
        bottom: 4,
        data: categorySpend,
        breadcrumb: { show: true, top: 'bottom', left: 'center', height: 20 },
        label: {
          color: '#fff',
          fontSize: 12,
          textBorderColor: 'rgba(0, 0, 0, 0.5)',
          textBorderWidth: 2,
        },
        upperLabel: {
          show: true,
          height: 22,
          color: '#fff',
          fontWeight: 'bold',
          backgroundColor: 'rgba(0, 0, 0, 0.45)',
        },
        itemStyle: { borderColor: 'rgba(0, 0, 0, 0.35)', borderWidth: 1, gapWidth: 2 },
        levels: [
          { itemStyle: { borderWidth: 3, gapWidth: 3, borderColor: 'rgba(0, 0, 0, 0.35)' }, upperLabel: { show: true } },
          { colorSaturation: [0.35, 0.55], itemStyle: { gapWidth: 1, borderColorSaturation: 0.6 } },
        ],
      },
    ],
  }

  // --- Payment mix (donut) -----------------------------------------------------
  const paymentOption = {
    backgroundColor: 'transparent',
    tooltip: { trigger: 'item', appendTo: 'body', formatter: '{b}: {c}%' },
    legend: { bottom: 0, left: 'center' },
    series: [
      {
        name: 'Payment mix',
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '44%'],
        avoidLabelOverlap: false,
        itemStyle: { borderColor: 'rgba(var(--v-theme-surface), 1)', borderWidth: 2 },
        label: { show: false },
        labelLine: { show: false },
        data: [
          { value: 58, name: 'Card' },
          { value: 24, name: 'Cash' },
          { value: 12, name: 'Store Credit' },
          { value: 6, name: 'Loyalty Points' },
        ],
      },
    ],
  }

  // --- Recent purchases ------------------------------------------------------
  const purchaseHeaders = [
    { title: 'Date', key: 'date' },
    { title: 'Invoice', key: 'invoice' },
    { title: 'Branch', key: 'branch' },
    { title: 'Items', key: 'items', align: 'end' },
    { title: 'Payment', key: 'payment' },
    { title: 'Total', key: 'total', align: 'end' },
  ]
  const purchases = [
    { date: '02 Sep 2026', invoice: 'INV-20418', branch: 'Karrada Flagship', items: 4, payment: 'Card', total: 18_400 },
    { date: '21 Aug 2026', invoice: 'INV-20255', branch: 'Karrada Flagship', items: 2, payment: 'Cash', total: 6900 },
    { date: '09 Aug 2026', invoice: 'INV-20099', branch: 'Mansour', items: 6, payment: 'Card', total: 24_150 },
    { date: '27 Jul 2026', invoice: 'INV-19881', branch: 'Karrada Flagship', items: 1, payment: 'Loyalty Points', total: 0 },
    { date: '14 Jul 2026', invoice: 'INV-19702', branch: 'Mansour', items: 3, payment: 'Store Credit', total: 11_300 },
    { date: '30 Jun 2026', invoice: 'INV-19540', branch: 'Karrada Flagship', items: 5, payment: 'Card', total: 20_750 },
    { date: '18 Jun 2026', invoice: 'INV-19388', branch: 'Karrada Flagship', items: 2, payment: 'Cash', total: 7600 },
    { date: '05 Jun 2026', invoice: 'INV-19221', branch: 'Mansour', items: 3, payment: 'Card', total: 13_900 },
  ]

  const paymentColor = {
    'Card': 'indigo',
    'Cash': 'teal',
    'Store Credit': 'deep-purple',
    'Loyalty Points': 'amber-darken-2',
  }

  const meta = [
    { label: 'Client ID', value: client.id },
    { label: 'Phone', value: client.phone },
    { label: 'Home branch', value: client.branch },
    { label: 'First purchase', value: client.firstPurchase },
  ]
</script>

<template>
  <div class="client-page pa-4">
    <!-- Header: avatar + identity + trailing-12-month KPIs. -->
    <v-sheet class="client-header pa-6 mb-4" color="surface-light" rounded="lg">
      <div class="d-flex flex-wrap ga-6 align-center">
        <v-avatar class="text-white flex-shrink-0" color="blue-darken-4" size="112">
          <span class="text-h3">{{ client.initials }}</span>
        </v-avatar>

        <div class="flex-grow-1" style="min-width: 260px">
          <div class="d-flex flex-wrap align-center ga-3 mb-3">
            <span class="text-h5 font-weight-bold">{{ client.name }}</span>
            <ClientRankTag :rank="client.rank" />
          </div>
          <div class="meta-grid">
            <template v-for="row in meta" :key="row.label">
              <span class="text-body-2 text-medium-emphasis">{{ row.label }}</span>
              <span class="text-body-2">{{ row.value }}</span>
            </template>
          </div>
        </div>
      </div>
    </v-sheet>

    <v-row dense>
      <v-col v-for="kpi in kpis" :key="kpi.title" cols="6" md="3">
        <SparklineCard
          :prev-value="kpi.prevValue"
          :series="kpi.series"
          :title="kpi.title"
          :unit="kpi.unit"
          :value="kpi.value"
        />
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="7">
        <v-card class="fill-height" flat>
          <v-card-title class="card-title">Spend by Category</v-card-title>
          <VChart autoresize class="treemap" :option="treemapOption" />
        </v-card>
      </v-col>

      <v-col cols="12" md="5">
        <v-card class="fill-height" flat>
          <v-card-title class="card-title">Payment Mix</v-card-title>
          <VChart autoresize class="donut" :option="paymentOption" />
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="7">
        <v-card flat>
          <v-card-title class="card-title">Recent Purchases</v-card-title>
          <v-data-table
            density="compact"
            :headers="purchaseHeaders"
            :items="purchases"
            :items-per-page="8"
          >
            <template #item.payment="{ value }">
              <v-chip :color="paymentColor[value]" label size="small" variant="tonal">{{ value }}</v-chip>
            </template>
            <template #item.total="{ value }">
              {{ value === 0 ? '—' : money(value) }}
            </template>
          </v-data-table>
        </v-card>
      </v-col>

      <v-col cols="12" md="5">
        <v-card class="fill-height pa-4" flat>
          <v-card-title class="card-title px-0">Loyalty</v-card-title>

          <div class="d-flex align-baseline justify-space-between mb-1">
            <span class="text-body-2 text-medium-emphasis">{{ client.rank }}</span>
            <span class="text-body-2 text-medium-emphasis">{{ nextTier.name }}</span>
          </div>
          <v-progress-linear
            color="amber-darken-2"
            height="10"
            :model-value="tierProgress"
            rounded
          />
          <div class="text-caption text-medium-emphasis mt-1">
            {{ money(tierRemaining) }} more spend to reach {{ nextTier.name }}
          </div>

          <v-divider class="my-4" />

          <v-row dense>
            <v-col cols="6">
              <div class="text-h6 font-weight-bold">{{ money(lifetimeSpent) }}</div>
              <div class="text-caption text-medium-emphasis">Lifetime spend</div>
            </v-col>
            <v-col cols="6">
              <div class="text-h6 font-weight-bold">320</div>
              <div class="text-caption text-medium-emphasis">Points balance</div>
            </v-col>
            <v-col cols="6">
              <div class="text-h6 font-weight-bold">148</div>
              <div class="text-caption text-medium-emphasis">Items bought</div>
            </v-col>
            <v-col cols="6">
              <div class="text-h6 font-weight-bold">$ 4,919</div>
              <div class="text-caption text-medium-emphasis">Avg. basket</div>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12">
        <v-card flat>
          <v-card-title class="card-title">Activity Over the Year</v-card-title>
          <Heatmap />
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<style scoped>
.meta-grid {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 4px 24px;
}

.treemap {
  height: 320px;
  width: 100%;
}

.donut {
  height: 320px;
  width: 100%;
}
</style>
