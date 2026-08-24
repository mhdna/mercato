<template>
  <div class="pie-chart">
    <div ref="container" />
  </div>
</template>

<script setup>
  import * as d3 from 'd3'
  import { computed, onMounted, ref, watch } from 'vue'

  const props = defineProps({
    data: {
      type: Array,
      default: () => [
        { label: 'Alpha', value: 40 },
        { label: 'Beta', value: 25 },
        { label: 'Gamma', value: 20 },
        { label: 'Delta', value: 15 },
      ],
    },
    width: { type: Number, default: 220 },
    height: { type: Number, default: 220 },
  })

  const emit = defineEmits(['hover'])

  const container = ref(null)
  const hovered = ref(null)
  const colorScale = ref(null)

  const total = computed(() => props.data.reduce((s, d) => s + d.value, 0))

  function draw () {
    d3.select(container.value).selectAll('*').remove()

    const radius = Math.min(props.width, props.height) / 2
    const color = d3.scaleOrdinal([
      'rgba(255, 30,  30,  0.85)',
      'rgba(30,  136, 229, 0.85)',
      'rgba(67,  160, 71,  0.85)',
      'rgba(255, 193, 7,   0.85)',
      'rgba(156, 39,  176, 0.85)',
      'rgba(0,   188, 212, 0.85)',
      'rgba(255, 87,  34,  0.85)',
      'rgba(63,  81,  181, 0.85)',
    ])
    colorScale.value = color

    // emit full summary so parent table can use colors
    emit('hover', null)

    const svg = d3.select(container.value)
      .append('svg')
      .attr('width', props.width)
      .attr('height', props.height)
      .append('g')
      .attr('transform', `translate(${props.width / 2}, ${props.height / 2})`)

    const pie = d3.pie().value(d => d.value).sort(null)
    const arc = d3.arc().innerRadius(0).outerRadius(radius - 20)
    const arcHover = d3.arc().innerRadius(0).outerRadius(radius - 10)
    const labelArc = d3.arc().innerRadius(radius * 0.55).outerRadius(radius * 0.55)

    svg.selectAll('path')
      .data(pie(props.data))
      .enter()
      .append('path')
      .attr('d', arc)
      .attr('fill', d => color(d.data.label))
      .attr('stroke', '#fff')
      .style('stroke-width', '2px')
      .style('cursor', 'pointer')
      .on('mouseenter', function (event, d) {
        d3.select(this).transition().duration(150).attr('d', arcHover)
        hovered.value = d.data.label
        emit('hover', d.data.label)
      })
      .on('mouseleave', function (event, d) {
        d3.select(this).transition().duration(150).attr('d', arc)
        hovered.value = null
        emit('hover', null)
      })

    svg.selectAll('text')
      .data(pie(props.data))
      .enter()
      .append('text')
      .attr('transform', d => `translate(${labelArc.centroid(d)})`)
      .attr('text-anchor', 'middle')
      .attr('dy', '0.35em')
      .style('font-size', '13px')
      .style('fill', '#fff')
      .style('pointer-events', 'none')
      .text(d => d.data.label)
  }

  // expose color lookup for parent
  defineExpose({
    getColor: label => colorScale.value ? colorScale.value(label) : '#ccc',
  })

  onMounted(draw)
  watch(() => props.data, draw, { deep: true })
  watch(() => [props.width, props.height], draw)
</script>

<style scoped>
.pie-chart {
  display: inline-flex;
  flex-shrink: 0;
}
</style>
