// Self-contained mock data + ECharts option builders for the Financials tab.
// Registers a few components that aren't in the app-wide
// ui/src/plugins/echarts.js set (MarkPoint/MarkLine/MarkArea), scoped to
// this file so the shared plugin stays untouched. Bar/Line chart types are
// already registered there.
import { MarkAreaComponent, MarkLineComponent, MarkPointComponent } from 'echarts/components'
import { use } from 'echarts/core'

use([MarkPointComponent, MarkLineComponent, MarkAreaComponent])

const axis = '#555'
const label = '#666'
const upColor = '#34A853'
const downColor = '#EA4335'
const seasonColor = '#4285F4'
const profitColor = '#673AB7'
const forecastBandColor = 'rgba(66,133,244,0.14)'
const showHighSeason = false

function formatDate (d) {
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' })
}

// Default high season = months where Kashi branches historically earn
// ~1.5x a normal month: Feb, May/Jun, Sep, Nov/Dec (holidays). 1-indexed.
// Configurable per-viewer from Settings -> Financials (stores/settings.js).
export const DEFAULT_HIGH_SEASON_MONTHS = [2, 5, 6, 9, 11, 12]
const NORMAL_DAILY_REVENUE = 4000
const HIGH_SEASON_MULTIPLIER = 1.5

const EVENT_TYPES = [
  { type: 'promotion', label: 'Promotion', color: upColor, symbol: 'pin', glyph: 'P' },
  { type: 'holiday', label: 'Holiday', color: seasonColor, symbol: 'diamond', glyph: 'H' },
  { type: 'stockout', label: 'Stockout', color: downColor, symbol: 'triangle', glyph: 'S' },
  { type: 'expensive', label: 'Unusually expensive day', color: '#FBBC05', symbol: 'roundRect', glyph: '$' },
]

function simpleMovingAverage (dayCount, series) {
  const result = []
  for (let i = 0; i < series.length; i++) {
    if (series[i] == null) {
      result.push(null)
      continue
    }
    const window = series.slice(Math.max(0, i - dayCount + 1), i + 1).filter(v => v != null)
    result.push(window.length < dayCount ? null : Math.round(window.reduce((a, b) => a + b, 0) / window.length))
  }
  return result
}

function generateEvents (categoryData, revenue, todayIndex) {
  const usedIndices = new Set()
  const events = []
  for (let n = 0; n < 8; n++) {
    let idx
    do {
      idx = Math.floor(Math.random() * (todayIndex + 1))
    } while (usedIndices.has(idx))
    usedIndices.add(idx)
    const meta = EVENT_TYPES[Math.floor(Math.random() * EVENT_TYPES.length)]
    events.push({ ...meta, date: categoryData[idx], value: revenue[idx] })
  }
  return events.toSorted((a, b) => categoryData.indexOf(a.date) - categoryData.indexOf(b.date))
}

// Full calendar year of daily financials: actual revenue/profit (only up
// to `todayIndex` -- the rest of the year hasn't happened yet), an AI
// forecast line spanning the whole year with a widening confidence band,
// 7/30-day revenue moving averages, and a handful of annotated events
// (promotions/holidays/stockouts/unusually expensive days). `seasonRanges`
// marks the configured high season month spans for chart shading.
export function generateFinancialsSeries (year = new Date().getFullYear(), highSeasonMonths = DEFAULT_HIGH_SEASON_MONTHS) {
  const highSeasonSet = new Set(highSeasonMonths)
  const start = new Date(year, 0, 1)
  const end = new Date(year, 11, 31)
  const totalDays = Math.round((end - start) / 86_400_000) + 1
  const todayIndex = Math.min(totalDays - 1, Math.round(totalDays * 0.72))

  const categoryData = []
  const revenue = []
  const profit = []
  const forecast = []
  const forecastLow = []
  const forecastHigh = []
  const seasonRanges = []

  let base = NORMAL_DAILY_REVENUE
  let currentMonth = null
  let rangeStart = null

  const d = new Date(start)
  for (let i = 0; i < totalDays; i++, d.setDate(d.getDate() + 1)) {
    const month = d.getMonth() + 1
    const isHighSeason = highSeasonSet.has(month)
    const dateLabel = formatDate(d)
    categoryData.push(dateLabel)

    if (month !== currentMonth) {
      if (currentMonth !== null && highSeasonSet.has(currentMonth)) {
        seasonRanges.push([rangeStart, categoryData.at(-2)])
      }
      currentMonth = month
      rangeStart = dateLabel
    }

    const target = NORMAL_DAILY_REVENUE * (isHighSeason ? HIGH_SEASON_MULTIPLIER : 1)
    base = base + (target - base) * 0.15 + (Math.random() - 0.5) * target * 0.12
    const actual = Math.max(300, base)

    // Deliberately not identical to `actual` so over/under-performance
    // against the forecast is visible in the bar coloring.
    const forecastValue = target * (0.94 + Math.sin(i / 9) * 0.03)
    const bandWidth = forecastValue * (0.06 + (i / totalDays) * 0.16)
    forecast.push(Math.round(forecastValue))
    forecastLow.push(Math.round(forecastValue - bandWidth))
    forecastHigh.push(Math.round(forecastValue + bandWidth))

    if (i <= todayIndex) {
      revenue.push(Math.round(actual))
      profit.push(Math.round(actual * (0.15 + Math.random() * 0.15)))
    } else {
      revenue.push(null)
      profit.push(null)
    }
  }
  if (currentMonth !== null && highSeasonSet.has(currentMonth)) {
    seasonRanges.push([rangeStart, categoryData.at(-1)])
  }

  return {
    categoryData,
    revenue,
    profit,
    forecast,
    forecastLow,
    forecastHigh,
    ma7: simpleMovingAverage(7, revenue),
    ma30: simpleMovingAverage(30, revenue),
    events: generateEvents(categoryData, revenue, todayIndex),
    seasonRanges,
    todayIndex,
  }
}

// The initial view focuses on two months before and one month after today.
export const DEFAULT_ZOOM_WINDOW = { beforeDays: 60, afterDays: 30 }

// Trailing-day window sizes for the Day/Week/Month/Year zoom presets.
export const ZOOM_PRESETS = { day: 14, week: 84, month: 180, year: null }

export function buildRevenueForecastOption (data, zoom = null) {
  const { categoryData, revenue, profit, forecast, forecastLow, forecastHigh, ma7, ma30, events, seasonRanges, todayIndex } = data
  const forecastBandDelta = forecastHigh.map((h, i) => h - forecastLow[i])
  const startIndex = typeof zoom === 'object'
    ? Math.max(0, todayIndex - zoom.beforeDays)
    : null
  const endIndex = typeof zoom === 'object'
    ? Math.min(categoryData.length - 1, todayIndex + zoom.afterDays)
    : null
  let startPct = 0
  let endPct = 100
  if (typeof zoom === 'object') {
    startPct = (startIndex / (categoryData.length - 1)) * 100
    endPct = (endIndex / (categoryData.length - 1)) * 100
  } else if (zoom) {
    startPct = Math.max(0, 100 - (zoom / categoryData.length) * 100)
  }

  return {
    animation: false,
    legend: {
      top: 0,
      data: ['Revenue', 'Net Profit', 'AI Forecast', 'Forecast Range', 'MA7', 'MA30'],
      textStyle: { color: label },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      appendTo: 'body',
    },
    toolbox: {
      feature: {
        dataZoom: { yAxisIndex: false },
        restore: {},
      },
    },
    grid: { left: '6%', right: '4%', top: 56, bottom: 96 },
    xAxis: {
      type: 'category',
      data: categoryData,
      boundaryGap: true,
      axisLine: { lineStyle: { color: axis } },
      axisLabel: { color: label, fontSize: 10 },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: label, formatter: '${value}' },
      splitLine: { lineStyle: { color: axis, type: 'dashed' } },
    },
    dataZoom: [
      { type: 'inside', start: startPct, end: endPct },
      { show: true, type: 'slider', bottom: 4, start: startPct, end: endPct },
    ],
    series: [
      {
        // Invisible baseline the visible band stacks on top of -- the
        // standard ECharts trick for rendering a [low, high] confidence
        // band as a stacked area. Not in legend.data, so it never shows.
        name: '_forecastBandBaseline',
        type: 'line',
        data: forecastLow,
        stack: 'confidence-band',
        symbol: 'none',
        lineStyle: { opacity: 0 },
        areaStyle: { opacity: 0 },
        silent: true,
        tooltip: { show: false },
        z: 1,
      },
      {
        name: 'Forecast Range',
        type: 'line',
        data: forecastBandDelta,
        stack: 'confidence-band',
        symbol: 'none',
        lineStyle: { opacity: 0 },
        itemStyle: { color: forecastBandColor },
        areaStyle: { color: forecastBandColor },
        silent: true,
        tooltip: { show: false },
        z: 1,
        markArea: showHighSeason
          ? {
              silent: true,
              label: {
                show: true,
                position: 'insideTop',
                color: seasonColor,
                fontSize: 10,
                fontWeight: 600,
              },
              itemStyle: {
                color: 'rgba(66,133,244,0.08)',
                borderColor: 'rgba(66,133,244,0.35)',
                borderWidth: 1,
              },
              data: seasonRanges.map(([from, to]) => [
                { name: 'High Season', xAxis: from },
                { xAxis: to },
              ]),
            }
          : undefined,
      },
      {
        name: 'Revenue',
        type: 'bar',
        data: revenue.map((v, i) => ({
          value: v,
          itemStyle: { color: v != null && v >= forecast[i] ? upColor : downColor },
        })),
        barMaxWidth: 14,
        z: 3,
        markPoint: {
          symbolSize: 20,
          label: {
            show: true,
            formatter: p => p.data.glyph,
            fontSize: 10,
            fontWeight: 700,
            color: '#fff',
          },
          tooltip: {
            formatter: p => `${p.data.eventLabel}<br/>${p.data.coord[0]}`,
          },
          data: events.map(e => ({
            name: e.label,
            coord: [e.date, e.value ?? 0],
            symbol: e.symbol,
            glyph: e.glyph,
            eventLabel: e.label,
            itemStyle: { color: e.color, borderColor: '#fff', borderWidth: 1 },
          })),
        },
        markLine: typeof todayIndex === 'number' && categoryData[todayIndex]
          ? {
              silent: true,
              symbol: 'none',
              lineStyle: { color: '#999', type: 'dashed' },
              label: { formatter: 'Today', color: '#999', fontSize: 10 },
              data: [{ xAxis: categoryData[todayIndex] }],
            }
          : undefined,
      },
      {
        name: 'Net Profit',
        type: 'line',
        data: profit,
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2, color: profitColor },
        itemStyle: { color: profitColor },
        z: 4,
      },
      {
        name: 'AI Forecast',
        type: 'line',
        data: forecast,
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 2, type: 'dashed', color: seasonColor },
        itemStyle: { color: seasonColor },
        z: 4,
      },
      {
        name: 'MA7',
        type: 'line',
        data: ma7,
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 1, opacity: 0.7, color: '#888' },
        z: 2,
      },
      {
        name: 'MA30',
        type: 'line',
        data: ma30,
        smooth: true,
        showSymbol: false,
        lineStyle: { width: 1, opacity: 0.7, type: 'dotted', color: '#555' },
        z: 2,
      },
    ],
  }
}

const BRANCHES = ['Downtown', 'Mall Plaza', 'Airport Rd', 'Old Town', 'Harbor View', 'North Gate']

export function buildTopBranchesBarOption () {
  const data = BRANCHES
    .map(name => ({ name, value: Math.round(3000 + Math.random() * 9000) }))
    .toSorted((a, b) => b.value - a.value)

  return {
    grid: { top: 12, right: 16, bottom: 28, left: 90 },
    tooltip: { trigger: 'axis', valueFormatter: v => `$${v}`, appendTo: 'body' },
    xAxis: {
      type: 'value',
      axisLabel: { color: label, formatter: '${value}' },
      splitLine: { lineStyle: { color: axis, type: 'dashed' } },
    },
    yAxis: {
      type: 'category',
      data: data.map(d => d.name),
      axisLine: { lineStyle: { color: axis } },
      axisLabel: { color: label, fontSize: 10 },
    },
    series: [{
      name: 'Revenue',
      type: 'bar',
      data: data.map(d => d.value),
      itemStyle: { color: '#4285F4', borderRadius: [0, 4, 4, 0] },
      barMaxWidth: 18,
    }],
  }
}

const PAYMENT_METHODS = [
  { name: 'Cash', color: '#4285F4' },
  { name: 'Card', color: '#34A853' },
  { name: 'Credit', color: '#FBBC05' },
  { name: 'Exchange', color: '#EA4335' },
]

export function buildPaymentMethodPieOption () {
  const data = PAYMENT_METHODS.map(m => ({
    name: m.name,
    value: Math.round(500 + Math.random() * 4000),
    itemStyle: { color: m.color },
  }))

  return {
    tooltip: { trigger: 'item', valueFormatter: v => `$${v}`, appendTo: 'body' },
    legend: { bottom: 0, textStyle: { color: label, fontSize: 10 } },
    series: [{
      name: 'Payment Method',
      type: 'pie',
      radius: ['45%', '72%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: true,
      itemStyle: { borderColor: '#fff', borderWidth: 1 },
      label: { color: label, fontSize: 10 },
      data,
    }],
  }
}

// Day-of-week x hour-of-day sales intensity.
export function buildHourlyHeatmapOption () {
  const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  const hours = Array.from({ length: 24 }, (_, h) => `${h}:00`)

  const data = []
  let max = 0
  for (let d = 0; d < days.length; d++) {
    for (let h = 0; h < hours.length; h++) {
      const isOpenHours = h >= 8 && h <= 22
      const isWeekend = d >= 5
      const base = isOpenHours ? (isWeekend ? 60 : 40) : 4
      const value = Math.round(Math.max(0, base + (Math.random() - 0.5) * base))
      max = Math.max(max, value)
      data.push([h, d, value])
    }
  }

  return {
    tooltip: { position: 'top', appendTo: 'body' },
    grid: { top: 8, right: 12, bottom: 28, left: 48 },
    xAxis: {
      type: 'category',
      data: hours,
      splitArea: { show: true },
      axisLabel: { color: label, fontSize: 9, interval: 1 },
      axisLine: { lineStyle: { color: axis } },
    },
    yAxis: {
      type: 'category',
      data: days,
      splitArea: { show: true },
      axisLabel: { color: label, fontSize: 10 },
      axisLine: { lineStyle: { color: axis } },
    },
    visualMap: {
      min: 0,
      max: max || 1,
      calculable: true,
      orient: 'horizontal',
      left: 'center',
      bottom: -4,
      textStyle: { color: label, fontSize: 10 },
      inRange: { color: ['#e8f0fe', '#4285F4'] },
    },
    series: [{
      name: 'Sales intensity',
      type: 'heatmap',
      data,
      label: { show: false },
      emphasis: { itemStyle: { shadowBlur: 6, shadowColor: 'rgba(0,0,0,0.3)' } },
    }],
  }
}

export function buildMarginTrendOption (days = 30) {
  const dates = []
  const margin = []
  let m = 22
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    dates.push(formatDate(d))
    m = Math.max(2, Math.min(45, m + (Math.random() - 0.5) * 4))
    margin.push(+m.toFixed(1))
  }

  return {
    grid: { top: 12, right: 16, bottom: 28, left: 40 },
    tooltip: { trigger: 'axis', valueFormatter: v => `${v}%`, appendTo: 'body' },
    xAxis: {
      type: 'category',
      data: dates,
      boundaryGap: false,
      axisLine: { lineStyle: { color: axis } },
      axisLabel: { color: label, fontSize: 10 },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: label, formatter: '{value}%' },
      splitLine: { lineStyle: { color: axis, type: 'dashed' } },
    },
    series: [{
      name: 'Profit Margin',
      type: 'line',
      smooth: true,
      symbol: 'circle',
      symbolSize: 4,
      lineStyle: { width: 2, color: '#34A853' },
      itemStyle: { color: '#34A853' },
      areaStyle: { color: 'rgba(52,168,83,0.15)' },
      data: margin,
    }],
  }
}
