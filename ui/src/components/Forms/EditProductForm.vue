<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <span class="text-h6">Edit Product</span>
      <v-spacer />
      <v-btn icon="mdi-close" size="small" variant="text" @click="$emit('close')" />
    </v-card-title>
    <v-divider />

    <v-card-text class="py-4">
      <div v-if="loading" class="py-8 text-center">
        <v-progress-circular color="primary" indeterminate />
      </div>

      <v-alert
        v-else-if="loadError"
        density="compact"
        type="error"
        variant="tonal"
      >
        {{ loadError }}
        <template #append>
          <v-btn size="small" text="Retry" variant="text" @click="load" />
        </template>
      </v-alert>

      <template v-else>
        <v-row dense>
          <v-col cols="12" sm="6">
            <v-text-field
              v-model="form.code"
              density="compact"
              :error-messages="errors.code"
              label="Code"
              variant="outlined"
            />
          </v-col>
          <v-col cols="12" sm="6">
            <v-text-field
              v-model="form.name"
              density="compact"
              :error-messages="errors.name"
              label="Name"
              variant="outlined"
            />
          </v-col>
        </v-row>

        <v-textarea
          v-model="form.description"
          auto-grow
          density="compact"
          label="Description"
          rows="2"
          variant="outlined"
        />

        <v-switch
          v-model="form.is_active"
          color="primary"
          density="compact"
          hide-details
          label="Active"
        />

        <div class="text-overline text-medium-emphasis mt-2 mb-1">Attributes</div>
        <v-row dense>
          <v-col v-for="field in attributeFields" :key="field.key" cols="12" sm="6">
            <v-combobox
              v-model="form.attributes[field.key]"
              clearable
              density="compact"
              hide-details="auto"
              :items="optionsFor(field.attrName)"
              :label="field.label"
              variant="outlined"
            />
          </v-col>
        </v-row>

        <div class="d-flex align-center mt-4 mb-1">
          <div class="text-overline text-medium-emphasis">Variants</div>
          <v-spacer />
          <v-btn
            prepend-icon="mdi-plus"
            size="small"
            text="Add variant"
            variant="text"
            @click="addVariant"
          />
        </div>

        <v-alert
          v-if="variantError"
          class="mb-2"
          density="compact"
          type="error"
          variant="tonal"
          @click:close="variantError = ''"
        >
          {{ variantError }}
        </v-alert>

        <v-table v-if="variants.length > 0" density="compact">
          <thead>
            <tr>
              <th>Color</th>
              <th>Size</th>
              <th>Barcode</th>
              <th style="width: 96px">Price</th>
              <th style="width: 64px">Active</th>
              <th style="width: 88px" />
            </tr>
          </thead>
          <tbody>
            <tr v-for="v in variants" :key="v.id">
              <td>
                <v-autocomplete
                  v-model="v.color_id"
                  clearable
                  density="compact"
                  hide-details
                  item-title="name"
                  item-value="id"
                  :items="colors"
                  variant="plain"
                />
              </td>
              <td>
                <v-autocomplete
                  v-model="v.size_id"
                  clearable
                  density="compact"
                  hide-details
                  :item-title="s => `${s.type}: ${s.name}`"
                  item-value="id"
                  :items="sizes"
                  variant="plain"
                />
              </td>
              <td>
                <v-text-field
                  v-model="v.barcode"
                  density="compact"
                  hide-details
                  variant="plain"
                />
              </td>
              <td>
                <v-text-field
                  v-model.number="v.price"
                  density="compact"
                  hide-details
                  type="number"
                  variant="plain"
                />
              </td>
              <td>
                <v-checkbox v-model="v.is_active" density="compact" hide-details />
              </td>
              <td class="text-right">
                <v-btn
                  icon="mdi-content-save"
                  :loading="v._saving"
                  size="small"
                  variant="text"
                  @click="saveVariant(v)"
                />
                <v-btn
                  color="error"
                  icon="mdi-delete"
                  :loading="v._deleting"
                  size="small"
                  variant="text"
                  @click="removeVariant(v)"
                />
              </td>
            </tr>
          </tbody>
        </v-table>
        <div v-else class="text-caption text-medium-emphasis mb-2">No variants.</div>

        <v-alert
          v-if="saveError"
          class="mt-4"
          density="compact"
          type="error"
          variant="tonal"
        >
          {{ saveError }}
        </v-alert>
      </template>
    </v-card-text>

    <v-divider />
    <v-card-actions class="px-4 py-3">
      <v-btn
        color="error"
        :disabled="loading || !!loadError"
        text="Delete product"
        variant="text"
        @click="deleteConfirm = true"
      />
      <v-spacer />
      <v-btn text="Cancel" variant="text" @click="$emit('close')" />
      <v-btn
        color="primary"
        :disabled="loading || !!loadError"
        :loading="saving"
        text="Save"
        variant="flat"
        @click="save"
      />
    </v-card-actions>

    <v-dialog v-model="deleteConfirm" max-width="440">
      <v-card>
        <v-card-title class="text-h6">Delete product</v-card-title>
        <v-card-text>
          Delete <strong>{{ form.name || 'this product' }}</strong> and all its variants?
          This can't be undone.
          <v-alert
            v-if="deleteError"
            class="mt-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ deleteError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="deleteConfirm = false" />
          <v-btn
            color="error"
            :loading="deleting"
            text="Delete"
            variant="flat"
            @click="doDelete"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup>
  import { reactive, ref, watch } from 'vue'
  import { useAttributes } from '@/composables/useAttributes'
  import { useColorsAndSizes } from '@/composables/useColorsAndSizes'
  import { useProducts } from '@/composables/useProducts'

  const props = defineProps({
    productId: {
      type: [Number, String],
      required: true,
    },
  })

  const emit = defineEmits(['close', 'saved', 'deleted'])

  const {
    fetchProduct,
    updateProduct,
    deleteProduct,
    createVariant,
    updateVariant,
    deleteVariant,
  } = useProducts()
  const { colors, sizes, fetchColors, fetchSizes } = useColorsAndSizes()
  const { values: attributeValues, fetchAll: fetchAttributes } = useAttributes()

  // key -> payload key (matches the API's Attributes json); attrName -> the
  // seeded attribute type name used to look options up in valuesByType.
  const attributeFields = [
    { key: 'category', attrName: 'category', label: 'Category', id: 0 },
    { key: 'subcategory', attrName: 'sub-category', label: 'Sub-Category', id: 1 },
    { key: 'brand', attrName: 'brand', label: 'Brand', id: 2 },
    { key: 'kind', attrName: 'kind', label: 'Kind', id: 3 },
    { key: 'type', attrName: 'type', label: 'Type', id: 4 },
    { key: 'unit', attrName: 'unit', label: 'Unit', id: 5 },
    { key: 'year', attrName: 'year', label: 'Year', id: 6 },
    { key: 'season', attrName: 'season', label: 'Season', id: 7 },
    { key: 'origin', attrName: 'origin', label: 'Origin', id: 8 },
  ]

  const loading = ref(true)
  const loadError = ref('')
  const saving = ref(false)
  const saveError = ref('')
  const variantError = ref('')

  const form = reactive({
    code: '',
    name: '',
    description: '',
    is_active: true,
    attributes: {},
  })
  const errors = reactive({ code: '', name: '' })
  const variants = ref([])

  function optionsFor (attrName) {
    return attributeValues.value
      .filter(v => v.attribute_name === attrName)
      .map(v => v.value)
      .toSorted((a, b) => a.localeCompare(b))
  }

  async function load () {
    loading.value = true
    loadError.value = ''
    try {
      await Promise.all([fetchColors(), fetchSizes(), fetchAttributes()])
      const data = await fetchProduct(props.productId)
      const p = data.product ?? data
      form.code = p.code ?? ''
      form.name = p.name ?? ''
      form.description = p.description ?? ''
      form.is_active = p.is_active ?? true

      const valueById = new Map(attributeValues.value.map(v => [v.id, v.value]))
      const byAttrId = new Map((data.attributes ?? []).map(a => [a.attribute_id, a.attribute_value_id]))
      const attrs = {}
      for (const field of attributeFields) {
        attrs[field.key] = valueById.get(byAttrId.get(field.id)) ?? ''
      }
      form.attributes = attrs

      variants.value = (data.variants ?? []).map(v => ({
        id: v.id,
        color_id: v.color_id ?? null,
        size_id: v.size_id ?? null,
        barcode: v.barcode ?? '',
        price: v.price ?? null,
        is_active: v.is_active ?? true,
        _saving: false,
        _deleting: false,
      }))
    } catch (error) {
      loadError.value = error.message
    } finally {
      loading.value = false
    }
  }

  watch(() => props.productId, load, { immediate: true })

  function validate () {
    errors.code = form.code.trim() ? '' : 'Code is required.'
    errors.name = form.name.trim().length >= 2 ? '' : 'Name needs to be at least 2 characters.'
    return !errors.code && !errors.name
  }

  async function save () {
    if (!validate()) return
    saving.value = true
    saveError.value = ''
    try {
      await updateProduct({
        id: Number(props.productId),
        code: form.code.trim(),
        name: form.name.trim(),
        description: form.description,
        is_active: form.is_active,
        attributes: { ...form.attributes },
      })
      emit('saved')
    } catch (error) {
      saveError.value = error.message
    } finally {
      saving.value = false
    }
  }

  async function saveVariant (v) {
    v._saving = true
    variantError.value = ''
    try {
      await updateVariant({
        id: v.id,
        color_id: v.color_id,
        size_id: v.size_id,
        barcode: String(v.barcode ?? '').trim(),
        price: v.price === '' || v.price == null ? null : Number(v.price),
        is_active: v.is_active,
      })
    } catch (error) {
      variantError.value = error.message
    } finally {
      v._saving = false
    }
  }

  async function addVariant () {
    variantError.value = ''
    try {
      await createVariant({ product_id: Number(props.productId) })
      await load()
    } catch (error) {
      variantError.value = error.message
    }
  }

  async function removeVariant (v) {
    v._deleting = true
    variantError.value = ''
    try {
      await deleteVariant(v.id)
      variants.value = variants.value.filter(x => x.id !== v.id)
    } catch (error) {
      variantError.value = error.message === 'still referenced by one or more products'
        ? 'This variant is used on an invoice or transfer and can\'t be deleted.'
        : error.message
    } finally {
      v._deleting = false
    }
  }

  // ---- delete product --------------------------------------------------
  const deleteConfirm = ref(false)
  const deleting = ref(false)
  const deleteError = ref('')

  async function doDelete () {
    deleting.value = true
    deleteError.value = ''
    try {
      await deleteProduct(props.productId)
      deleteConfirm.value = false
      emit('deleted')
    } catch (error) {
      deleteError.value = error.message === 'still referenced by one or more products'
        ? 'This product is referenced elsewhere (invoices, purchases, or transfers) and can\'t be deleted.'
        : error.message
    } finally {
      deleting.value = false
    }
  }
</script>

<style scoped>
:deep(.v-table) th {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
</style>
