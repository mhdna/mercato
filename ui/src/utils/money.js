export function formatMoney (cents) {
  return (Number(cents) / 100).toFixed(2)
}

// Compact money for at-a-glance panels: $840, $12.4K, $1.2M. Negative
// amounts keep the sign outside the symbol: -$1.2K.
export function formatCompactMoney (cents) {
  const n = Number(cents || 0) / 100
  const abs = Math.abs(n)
  const sign = n < 0 ? '-' : ''
  if (abs >= 1e6) {
    return `${sign}$${(abs / 1e6).toFixed(1)}M`
  }
  if (abs >= 1e3) {
    return `${sign}$${(abs / 1e3).toFixed(1)}K`
  }
  return `${sign}$${abs.toFixed(0)}`
}
