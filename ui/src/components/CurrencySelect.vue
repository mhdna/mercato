<template>
  <v-autocomplete
    v-model="model"
    :auto-select-first="true"
    density="compact"
    :error-messages="mergedErrors"
    :item-props="itemProps"
    item-title="code"
    item-value="code"
    :items="sortedCurrencies"
    :label="label"
    :loading="loading"
    :no-data-text="loadError || 'No currencies found'"
    :return-object="false"
  />
</template>

<script setup lang="ts">
  import { computed, onMounted, ref } from 'vue'
  import { useCurrencies } from '@/composables/useCurrencies'

  const props = defineProps({
    label: { type: String, default: 'Currency' },
    errorMessages: { type: [String, Array], default: () => [] },
  })

  // Held as the plain currency code string.
  const model = defineModel({ type: String, default: '' })

  const { currencies, fetchCurrencies } = useCurrencies()
  const loading = ref(false)
  const loadError = ref('')

  function itemProps (c) {
    return {
      title: c.code,
      subtitle: [c.symbol, c.name].filter(Boolean).join(' · '),
    }
  }

  // Real ISO currency codes are 3 chars -- float those (and the default /
  // active ones) to the top so they're not buried under any test data.
  const sortedCurrencies = computed(() => {
    return currencies.value.toSorted((a, b) => {
      const rank = c => (c.is_default ? 0 : 1) + (c.code?.length <= 3 ? 0 : 2) + (c.is_active ? 0 : 4)
      return rank(a) - rank(b) || String(a.code).localeCompare(String(b.code))
    })
  })

  const mergedErrors = computed(() => {
    const base = Array.isArray(props.errorMessages)
      ? [...props.errorMessages]
      : (props.errorMessages ? [props.errorMessages] : [])
    if (loadError.value) base.push(loadError.value)
    return base
  })

  onMounted(async () => {
    loading.value = true
    loadError.value = ''
    try {
      await fetchCurrencies(true)
    } catch (error) {
      loadError.value = `Could not load currencies: ${error.message}`
      console.error('CurrencySelect: failed to load currencies', error)
    } finally {
      loading.value = false
    }
  })
</script>
