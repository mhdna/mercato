import {
  BarChart,
  CustomChart,
  GaugeChart,
  HeatmapChart,
  LineChart,
  PieChart,
  ScatterChart,
  TreemapChart,
} from 'echarts/charts'
import {
  CalendarComponent,
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  TitleComponent,
  ToolboxComponent,
  TooltipComponent,
  VisualMapComponent,
} from 'echarts/components'
import { registerTheme, use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'

use([
  CanvasRenderer,
  LineChart,
  BarChart,
  PieChart,
  ScatterChart,
  HeatmapChart,
  TreemapChart,
  CustomChart,
  GaugeChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
  MarkLineComponent,
  ToolboxComponent,
  VisualMapComponent,
  CalendarComponent,
  DataZoomComponent,
])

// Two app themes so every <VChart> can flip with Vuetify's light/dark mode
// via vue-echarts' THEME_KEY (provided app-wide in App.vue). Both keep a
// transparent canvas so charts inherit whatever card/page they sit on --
// echarts' own built-in "dark" theme paints a solid near-black rectangle,
// which never lines up with our surface colours. Series colours are left to
// echarts' default palette (or whatever a chart sets explicitly); only the
// chrome -- axes, grid lines, labels, legend, tooltip -- is themed here.
const darkAxis = {
  axisLine: { show: true, lineStyle: { color: 'rgba(255, 255, 255, 0.25)' } },
  axisTick: { show: true, lineStyle: { color: 'rgba(255, 255, 255, 0.25)' } },
  axisLabel: { show: true, color: 'rgba(255, 255, 255, 0.7)' },
  splitLine: { show: true, lineStyle: { color: ['rgba(255, 255, 255, 0.08)'] } },
  splitArea: {
    show: false,
    areaStyle: { color: ['rgba(255, 255, 255, 0.02)', 'rgba(255, 255, 255, 0.05)'] },
  },
}

registerTheme('kashi-light', {
  backgroundColor: 'transparent',
})

registerTheme('kashi-dark', {
  darkMode: true,
  backgroundColor: 'transparent',
  textStyle: { color: 'rgba(255, 255, 255, 0.85)' },
  title: {
    textStyle: { color: 'rgba(255, 255, 255, 0.9)' },
    subtextStyle: { color: 'rgba(255, 255, 255, 0.55)' },
  },
  legend: { textStyle: { color: 'rgba(255, 255, 255, 0.75)' } },
  categoryAxis: darkAxis,
  valueAxis: darkAxis,
  logAxis: darkAxis,
  timeAxis: darkAxis,
  tooltip: {
    backgroundColor: 'rgba(30, 30, 30, 0.95)',
    borderColor: 'rgba(255, 255, 255, 0.12)',
    textStyle: { color: 'rgba(255, 255, 255, 0.9)' },
  },
  visualMap: { textStyle: { color: 'rgba(255, 255, 255, 0.7)' } },
  dataZoom: {
    textStyle: { color: 'rgba(255, 255, 255, 0.6)' },
    borderColor: 'rgba(255, 255, 255, 0.15)',
  },
})

export { graphic } from 'echarts/core'
