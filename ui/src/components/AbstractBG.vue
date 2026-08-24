<template>
  <svg ref="svgEl" style="position:fixed;top:0;left:0;width:100%;height:100%;display:block;" version="1.1" />
</template>

<script setup>
  import { onBeforeUnmount, onMounted, ref } from 'vue'

  const svgEl = ref(null)

  class Vector {
    constructor (x = 0, y = 0) {
      this.x = x; this.y = y
    }

    clone () {
      return new Vector(this.x, this.y)
    }
  }

  function jitter (point, attr, dimension) {
    point[attr] += Math.random() * (Math.random() > 0.5 ? -1 : 1) * dimension * 0.175
  }

  function generatePoints (cols, rows) {
    const rowH = 100 / rows, colW = 100 / cols
    const grid = Array.from({ length: rows }, () => new Array(cols))

    for (let r = 0; r < rows; r++) {
      for (let c = 0; c < cols; c++) {
        let topLeft, topRight, bottomLeft
        const bottomRight = new Vector((c + 1) * colW, (r + 1) * rowH)

        topRight = r === 0 ? new Vector((c + 1) * colW, 0) : grid[r - 1][c].bottomRight

        if (c === 0) {
          topLeft = r === 0 ? new Vector(0, 0) : grid[r - 1][0].bottomLeft
          bottomLeft = new Vector(0, (r + 1) * rowH)
        } else {
          topLeft = grid[r][c - 1].topRight
          bottomLeft = grid[r][c - 1].bottomRight
        }

        if (!c && r < rows - 1) jitter(bottomLeft, 'y', rowH)
        if (!r && c < cols - 1) jitter(topRight, 'x', colW)
        if (r < rows - 1) jitter(bottomRight, 'y', rowH)
        if (c < cols - 1) jitter(bottomRight, 'x', colW)

        grid[r][c] = { topLeft, topRight, bottomLeft, bottomRight }
      }
    }
    return grid
  }

  function render (svg, grid, rows, cols) {
    svg.innerHTML = ''
    const w = window.innerWidth, h = window.innerHeight
    const px = p => `${Math.floor((p.x / 100) * w)},${Math.floor((p.y / 100) * h)}`

    for (let r = 0; r < rows; r++) {
      for (let c = 0; c < cols; c++) {
        const { topLeft, topRight, bottomLeft, bottomRight } = grid[r][c]
        const el = document.createElementNS('http://www.w3.org/2000/svg', 'polygon')
        el.setAttribute('points', [topLeft, topRight, bottomRight, bottomLeft].map(px).join(' '))
        el.setAttribute('style', 'fill:#333;fill-opacity:1;stroke:#222;stroke-width:1px;')
        svg.append(el)
      }
    }
  }

  const ROWS = 5, COLS = 20
  let grid, onResize

  onMounted(() => {
    grid = generatePoints(COLS, ROWS)
    onResize = () => render(svgEl.value, grid, ROWS, COLS)
    onResize()
    window.addEventListener('resize', onResize)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', onResize)
  })
</script>
