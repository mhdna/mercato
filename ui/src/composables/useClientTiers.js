import { ref, watch } from 'vue'

// Shared across both the retail and wholesale clients pages -- a client's
// spend tier means the same thing regardless of which tab they're viewed
// from, so one set of thresholds/colors is edited once and applies
// everywhere. Persisted to localStorage since there's no backend concept
// of "tiers" (they're a display-only bucketing over total_spent).
const STORAGE_KEY = 'kashi.clientTiers'

// minAmount is in whole currency units (dollars), not cents -- total_spent
// from the API is in cents and gets divided by 100 before comparing.
const defaultTiers = [
  { id: 'diamond', label: 'Diamond', color: '#7C4DFF', minAmount: 5000 },
  { id: 'platinum', label: 'Platinum', color: '#00BCD4', minAmount: 2000 },
  { id: 'gold', label: 'Gold', color: '#FFC107', minAmount: 800 },
  { id: 'silver', label: 'Silver', color: '#9E9E9E', minAmount: 300 },
  { id: 'bronze', label: 'Bronze', color: '#8D6E63', minAmount: 0 },
]

function loadTiers () {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return structuredClone(defaultTiers)
    }
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed) || parsed.length === 0) {
      return structuredClone(defaultTiers)
    }
    return parsed
  } catch {
    return structuredClone(defaultTiers)
  }
}

const tiers = ref(loadTiers())

watch(tiers, value => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
}, { deep: true })

function sortedTiers () {
  return tiers.value.toSorted((a, b) => b.minAmount - a.minAmount)
}

// centsSpent -> the matching tier (highest minAmount the spend clears), or
// the lowest tier if somehow none match (shouldn't happen once a 0-floor
// tier exists, but tiers are user-editable so nothing guarantees that).
function tierFor (centsSpent) {
  const dollars = (centsSpent ?? 0) / 100
  const sorted = sortedTiers()
  return sorted.find(t => dollars >= t.minAmount) ?? sorted.at(-1) ?? null
}

function addTier () {
  tiers.value.push({
    id: `tier-${Date.now()}`,
    label: 'New Tier',
    color: '#607D8B',
    minAmount: 0,
  })
}

function removeTier (id) {
  tiers.value = tiers.value.filter(t => t.id !== id)
}

function resetTiers () {
  tiers.value = structuredClone(defaultTiers)
}

export function useClientTiers () {
  return { tiers, sortedTiers, tierFor, addTier, removeTier, resetTiers }
}
