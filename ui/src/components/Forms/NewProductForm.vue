<template>
  <form class="new-product-form" @submit.prevent="submit">
    <v-row density="compact">
      <v-col cols="12" sm="6">
        <v-text-field
          v-model="code.value.value"
          density="compact"
          :error-messages="code.errorMessage.value"
          label="Code"
          variant="outlined"
        />
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          v-model="name.value.value"
          density="compact"
          :error-messages="name.errorMessage.value"
          label="Name"
          variant="outlined"
        />
      </v-col>
    </v-row>

    <v-textarea
      v-model="description.value.value"
      auto-grow
      density="compact"
      :error-messages="description.errorMessage.value"
      label="Description"
      rows="2"
      variant="outlined"
    />

    <div class="text-overline text-medium-emphasis mt-1 mb-1">Attributes</div>
    <v-row density="compact">
      <v-col v-for="field in attributeFields" :key="field.key" cols="12" sm="6">
        <v-combobox
          v-model="field.model.value.value"
          clearable
          density="compact"
          hide-details="auto"
          :items="optionsFor(field.attribute)"
          :label="field.label"
          variant="outlined"
        />
      </v-col>
    </v-row>

    <div class="text-overline text-medium-emphasis mt-3 mb-1">Variants</div>
    <v-row density="compact">
      <v-col cols="12" sm="6">
        <v-autocomplete
          v-model="colorIds.value.value"
          chips
          clearable
          closable-chips
          density="compact"
          hide-details="auto"
          item-title="name"
          item-value="id"
          :items="colors"
          label="Colors"
          :menu-props="{ maxHeight: 320 }"
          multiple
          variant="outlined"
        >
          <!-- Vuetify 4: #item slot's `item` is the raw record, not an InternalItem. -->
          <template #item="{ props: itemProps, item }">
            <v-list-item v-bind="itemProps" :title="item.name">
              <template #prepend>
                <div
                  class="rounded me-2 swatch"
                  :style="{ backgroundColor: item.hex_value || 'transparent' }"
                />
              </template>
            </v-list-item>
          </template>
        </v-autocomplete>
      </v-col>
      <v-col cols="12" sm="6">
        <v-autocomplete
          v-model="sizeIds.value.value"
          chips
          clearable
          closable-chips
          density="compact"
          hide-details="auto"
          :item-title="s => `${s.type}: ${s.name}`"
          item-value="id"
          :items="sizes"
          label="Sizes"
          :menu-props="{ maxHeight: 320 }"
          multiple
          variant="outlined"
        />
      </v-col>
    </v-row>
    <div class="text-caption text-medium-emphasis mt-1">
      {{ variantCount }} variant{{ variantCount === 1 ? '' : 's' }} will be created,
      each with its own barcode.
    </div>

    <v-text-field
      v-model.number="barcode.value.value"
      class="mt-3"
      clearable
      density="compact"
      :disabled="variantCount > 1"
      :error-messages="barcode.errorMessage.value"
      :hint="variantCount > 1 ? 'Barcodes are auto-generated when creating multiple variants' : 'Leave empty to auto-generate'"
      label="Barcode (optional)"
      persistent-hint
      type="number"
      variant="outlined"
    />

    <v-alert
      v-if="submitError"
      class="mt-4"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ submitError }}
    </v-alert>

    <div class="d-flex justify-end ga-2 mt-4">
      <v-btn text="Clear" variant="text" @click="handleReset" />
      <v-btn
        color="primary"
        :loading="submitting"
        text="Add Product"
        type="submit"
        variant="flat"
      />
    </div>
  </form>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { computed, ref } from 'vue'
  import { authFetch } from '@/composables/useApi'
  import { useAttributes } from '@/composables/useAttributes'
  import { useColorsAndSizes } from '@/composables/useColorsAndSizes'
  import { API_BASE } from '@/config'

  const emit = defineEmits(['created'])

  const { colors, sizes, fetchColors, fetchSizes } = useColorsAndSizes()
  const { valuesByType, fetchAll: fetchAttributes } = useAttributes()
  fetchColors()
  fetchSizes()
  fetchAttributes()

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      code (value) {
        return value?.length >= 1 ? true : 'Code is required.'
      },
      name (value) {
        return value?.length >= 2 ? true : 'Name needs to be at least 2 characters.'
      },
      description (value) {
        return value?.length >= 1 ? true : 'Description is required.'
      },
      barcode () {
        return true
      },
      colorIds () {
        return true
      },
      sizeIds () {
        return true
      },
    },
  })

  const code = useField('code')
  const name = useField('name')
  const description = useField('description')
  const barcode = useField('barcode')
  const colorIds = useField('colorIds', undefined, { initialValue: [] })
  const sizeIds = useField('sizeIds', undefined, { initialValue: [] })

  // One variant per colour x size; at least one either way.
  const variantCount = computed(() => {
    const c = colorIds.value.value?.length || 0
    const s = sizeIds.value.value?.length || 0
    return Math.max(1, c || 1) * Math.max(1, s || 1)
  })

  // key -> payload key sent to the API; attribute -> seeded attribute type name.
  const attributeFields = [
    { key: 'category', attribute: 'category', label: 'Category', model: useField('category') },
    { key: 'subcategory', attribute: 'sub-category', label: 'Sub-Category', model: useField('subcategory') },
    { key: 'brand', attribute: 'brand', label: 'Brand', model: useField('brand') },
    { key: 'kind', attribute: 'kind', label: 'Kind', model: useField('kind') },
    { key: 'type', attribute: 'type', label: 'Type', model: useField('type') },
    { key: 'unit', attribute: 'unit', label: 'Unit', model: useField('unit') },
    { key: 'year', attribute: 'year', label: 'Year', model: useField('year') },
    { key: 'season', attribute: 'season', label: 'Season', model: useField('season') },
    { key: 'origin', attribute: 'origin', label: 'Origin', model: useField('origin') },
  ]

  function optionsFor (attribute) {
    return (valuesByType.value[attribute] ?? []).map(v => v.value)
  }

  const submitting = ref(false)
  const submitError = ref('')

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      const attributes = {}
      for (const field of attributeFields) {
        attributes[field.key] = values[field.key] ?? ''
      }

      const res = await authFetch(`${API_BASE}/products`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: values.code,
          name: values.name,
          description: values.description,
          attributes,
          barcode: variantCount.value > 1 ? null : (values.barcode || null),
          color_ids: values.colorIds ?? [],
          size_ids: values.sizeIds ?? [],
        }),
      })

      if (!res.ok) {
        const err = await res.json().catch(() => ({}))
        throw new Error(err.error || `Request failed with status ${res.status}`)
      }

      handleReset()
      emit('created')
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  })
</script>

<style scoped>
.new-product-form {
  width: 100%;
}
.swatch {
  width: 16px;
  height: 16px;
  border: 1px solid rgba(128, 128, 128, 0.6);
}
</style>
