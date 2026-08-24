const baseGrid = { top: 8, right: 12, bottom: 4, left: 12, containLabel: true }

export function buildChartOptions (dates, revenue, expenses) {
  const axis = '#555'
  const label = '#666'

  return {
    revenue: {
      grid: { ...baseGrid, left: 8, right: 16, containLabel: true },
      tooltip: { trigger: 'axis', valueFormatter: v => `$${v}` },
      xAxis: { type: 'category', data: dates, boundaryGap: false, axisLine: { lineStyle: { color: axis } }, axisLabel: { color: label, fontSize: 10 } },
      yAxis: { type: 'value', axisLabel: { color: label, formatter: '${value}' }, splitLine: { lineStyle: { color: axis, type: 'dashed' } } },
      series: [{ name: 'Revenue', type: 'line', smooth: true, symbol: 'circle', symbolSize: 4, lineStyle: { width: 2, color: '#4285F4' }, itemStyle: { color: '#4285F4' }, areaStyle: { color: 'rgba(66,133,244,0.15)' }, data: revenue }],
    },
    expenses: {
      grid: { ...baseGrid, left: 8, right: 16, containLabel: true },
      tooltip: { trigger: 'axis', valueFormatter: v => `$${v}` },
      xAxis: { type: 'category', data: dates, boundaryGap: false, axisLine: { lineStyle: { color: axis } }, axisLabel: { color: label, fontSize: 10 } },
      yAxis: { type: 'value', axisLabel: { color: label, formatter: '${value}' }, splitLine: { lineStyle: { color: axis, type: 'dashed' } } },
      series: [{ name: 'Expenses', type: 'line', smooth: true, symbol: 'circle', symbolSize: 4, lineStyle: { width: 2, color: '#EA4335' }, itemStyle: { color: '#EA4335' }, areaStyle: { color: 'rgba(234,67,53,0.15)' }, data: expenses }],
    },
    profit: {
      grid: { ...baseGrid, left: 8, right: 16, containLabel: true },
      tooltip: { trigger: 'axis', valueFormatter: v => `$${v}` },
      xAxis: { type: 'category', data: dates, boundaryGap: false, axisLine: { lineStyle: { color: axis } }, axisLabel: { color: label, fontSize: 10 } },
      yAxis: { type: 'value', axisLabel: { color: label, formatter: '${value}' }, splitLine: { lineStyle: { color: axis, type: 'dashed' } } },
      series: [{ name: 'Profit', type: 'line', smooth: true, symbol: 'circle', symbolSize: 4, lineStyle: { width: 2, color: '#34A853' }, itemStyle: { color: '#34A853' }, areaStyle: { color: 'rgba(52,168,83,0.15)' }, data: revenue.map((r, i) => Math.max(r - expenses[i], 0)) }],
    },
    margin: {
      grid: { ...baseGrid, left: 8, right: 16, containLabel: true },
      tooltip: { trigger: 'axis', valueFormatter: v => `${v}%` },
      xAxis: { type: 'category', data: dates, boundaryGap: false, axisLine: { lineStyle: { color: axis } }, axisLabel: { color: label, fontSize: 10 } },
      yAxis: { type: 'value', axisLabel: { color: label, formatter: '{value}%' }, splitLine: { lineStyle: { color: axis, type: 'dashed' } } },
      series: [{ name: 'Profit Margin', type: 'line', smooth: true, symbol: 'circle', symbolSize: 4, lineStyle: { width: 2, color: '#34A853' }, itemStyle: { color: '#34A853' }, areaStyle: { color: 'rgba(52,168,83,0.15)' }, data: revenue.map((r, i) => (r > 0 ? Math.max(0, Math.min(100, Math.round(((r - expenses[i]) / r) * 1000) / 10)) : 0)) }],
    },
  }
}

export function generateMockData () {
  const dates = []
  const revenue = []
  const expenses = []

  for (let i = 29; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    dates.push(d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' }))
    const r = Math.round(Math.random() * 5000 + 500)
    revenue.push(r)
    expenses.push(Math.round(Math.random() * (r * 0.6) + 50))
  }

  return { dates, revenue, expenses }
}
