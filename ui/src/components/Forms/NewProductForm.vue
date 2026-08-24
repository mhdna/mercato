<template>
  <form @submit.prevent="submit">
    <v-row>
      <v-col cols="8">
        <div class="w-100">
          <v-row>
            <v-col cols="6">
              <v-text-field
                v-model="code.value.value"
                :counter="20"
                density="compact"
                :error-messages="code.errorMessage.value"
                label="Code"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model="name.value.value"
                :counter="100"
                density="compact"
                :error-messages="name.errorMessage.value"
                label="Name"
              />
            </v-col>
          </v-row>

          <v-textarea
            v-model="description.value.value"
            :counter="100"
            :error-messages="description.errorMessage.value"
            label="Description"
            no-resize
            rows="2"
          />

          <v-row>
            <v-col cols="4">
              <v-text-field v-model="category.value.value" density="compact" label="Category" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="subCategory.value.value" density="compact" label="Sub-Category" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="brand.value.value" density="compact" label="Brand" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="kind.value.value" density="compact" label="Kind" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="type.value.value" density="compact" label="Type" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="unit.value.value" density="compact" label="Unit" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="year.value.value" density="compact" label="Year" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="season.value.value" density="compact" label="Season" />
            </v-col>
            <v-col cols="4">
              <v-text-field v-model="origin.value.value" density="compact" label="Origin" />
            </v-col>
          </v-row>

          <v-text-field
            v-model.number="barcode.value.value"
            clearable
            density="compact"
            :error-messages="barcode.errorMessage.value"
            hint="Leave empty to auto-generate a barcode"
            label="Barcode (optional)"
            persistent-hint
            type="number"
          />

          <div class="d-flex align-center mt-4" style="gap: 12px;">
            <v-select
              v-model="colorId.value.value"
              clearable
              density="compact"
              :error-messages="colorId.errorMessage.value"
              :item-title="c => c.name"
              :item-value="c => c.id"
              :items="colors"
              label="Color"
            >
              <template #item="{ props: itemProps, item }">
                <v-list-item v-bind="itemProps">
                  <template #prepend>
                    <div
                      class="rounded me-2"
                      :style="{ backgroundColor: item.raw.hexValue, width: '16px', height: '16px', border: '1px solid #999' }"
                    />
                  </template>
                </v-list-item>
              </template>
            </v-select>
            <v-select
              v-model="sizeId.value.value"
              clearable
              density="compact"
              :error-messages="sizeId.errorMessage.value"
              :item-title="s => `${s.type}: ${s.name}`"
              :item-value="s => s.id"
              :items="sizes"
              label="Size"
            />
            <ColorsSizesDialog />
          </div>
        </div>
      </v-col>
      <v-col class="4">
        <v-card class="pa-5" flat height="650" width="100%">
          <FileUploadCard />
        </v-card>
      </v-col>
    </v-row>

    <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">{{ submitError }}</v-alert>

    <v-row>
      <v-col>
        <v-btn class="w-100" variant="solid" @click="handleReset"> clear </v-btn>
      </v-col>
      <v-col>
        <v-btn class="me-4 w-100" :loading="submitting" type="submit" variant="solid"> Add Product </v-btn>
      </v-col>
    </v-row>
  </form>
</template>
<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import ColorsSizesDialog from '@/components/Forms/ColorsSizesDialog.vue'
  import { useColorsAndSizes } from '@/composables/useColorsAndSizes'
  import { API_BASE } from '@/config'
  import { authFetch } from '@/composables/useApi'

  const emit = defineEmits(['created'])

  const { colors, sizes, fetchColors, fetchSizes } = useColorsAndSizes()
  fetchColors()
  fetchSizes()

  const { handleSubmit, handleReset } = useForm({
    validationSchema: {
      code (value) {
        if (value?.length >= 1) return true
        return 'Code is required.'
      },
      name (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
      description (value) {
        if (value?.length >= 1) return true
        return 'Description is required.'
      },
      barcode () {
        return true
      },
      colorId () {
        return true
      },
      sizeId () {
        return true
      },
    },
  })

  const code = useField('code')
  const name = useField('name')
  const description = useField('description')
  const category = useField('category')
  const subCategory = useField('subCategory')
  const brand = useField('brand')
  const kind = useField('kind')
  const type = useField('type')
  const unit = useField('unit')
  const year = useField('year')
  const season = useField('season')
  const origin = useField('origin')
  const barcode = useField('barcode')
  const colorId = useField('colorId')
  const sizeId = useField('sizeId')

  const submitting = ref(false)
  const submitError = ref('')

  const submit = handleSubmit(async values => {
    submitting.value = true
    submitError.value = ''
    try {
      const res = await authFetch(`${API_BASE}/products`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: values.code,
          name: values.name,
          description: values.description,
          attributes: {
            category: values.category ?? '',
            subcategory: values.subCategory ?? '',
            brand: values.brand ?? '',
            kind: values.kind ?? '',
            type: values.type ?? '',
            unit: values.unit ?? '',
            year: values.year ?? '',
            season: values.season ?? '',
            origin: values.origin ?? '',
          },
          barcode: values.barcode || null,
          color_id: values.colorId || null,
          size_id: values.sizeId || null,
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
