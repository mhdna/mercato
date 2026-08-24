import { buildChartOptions, generateMockData } from '../chartOptions'

const { dates, revenue, expenses } = generateMockData()
const charts = buildChartOptions(dates, revenue, expenses)

export function useChartData () {
  return { charts, dates, revenue, expenses }
}
