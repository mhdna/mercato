<template>
  <div class="page-root">
    <v-alert
      v-if="loadError"
      class="mb-4"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ loadError }}
    </v-alert>

    <v-card class="coupons-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-ticket-percent" />
        <span>Coupons</span>
        <v-spacer />
        <v-text-field
          v-model="search"
          class="coupon-search"
          clearable
          density="compact"
          hide-details
          label="Search code or reason"
          prepend-inner-icon="mdi-magnify"
          variant="outlined"
        />
        <v-btn-toggle
          v-model="statusFilter"
          color="primary"
          density="comfortable"
          mandatory
          variant="outlined"
        >
          <v-btn value="all">All</v-btn>
          <v-btn value="active">Active</v-btn>
          <v-btn value="inactive">Inactive</v-btn>
        </v-btn-toggle>
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Coupon"
          variant="flat"
          @click="openCreate"
        />
      </v-card-title>
      <v-divider />

      <div class="coupons-layout">
        <CategorySidebar
          v-model="selectedCategory"
          :categories="couponCategories"
          @add="openCreateCategory"
          @delete="removeCategory"
          @edit="openEditCategory"
        />

        <v-divider vertical />

        <div class="coupons-content">
          <div class="pa-4">
            <BulkDeleteBar
              :count="selectedCoupons.length"
              :error="couponSelError"
              :loading="couponDeleting"
              @clear="resetCouponSel"
              @confirm="bulkDeleteCoupons"
            />

            <ServerSideTable
              ref="tableRef"
              :api-u-r-l="apiURL"
              density="comfortable"
              :external-search="search ?? ''"
              flush
              :headers="headers"
              hover
              item-value="code"
              :max-page-size="100"
              :query-params="queryParams"
              root-key="coupons"
              selectable
              :show-search-icon="false"
              @row-click="openEdit"
              @update:selected="v => (selectedCoupons = v)"
            >
              <template #item.category_id="{ item }">
                <CategoryChip :category="categoryFor(item.category_id)" />
              </template>
              <template #item.status="{ item }">
                <v-chip :color="item.status === 'active' ? 'success' : 'default'" size="small">
                  {{ item.status }}
                </v-chip>
              </template>
              <template #item.valid_until="{ item }">
                {{ formatDate(item.valid_until) }}
              </template>
              <template #item.actions="{ item }">
                <v-btn
                  :disabled="item.status !== 'active'"
                  size="small"
                  text="Deactivate"
                  variant="text"
                  @click.stop="deactivate(item)"
                />
              </template>
            </ServerSideTable>
          </div>
        </div>
      </div>
    </v-card>

    <!-- add / edit coupon -->
    <v-dialog v-model="dialog" max-width="560">
      <v-card class="px-4">
        <v-card-title>{{ editingCode ? 'Edit Coupon' : 'Add a New Coupon' }}</v-card-title>
        <v-card-text>
          <form @submit.prevent="submit">
            <v-text-field
              v-model="form.code"
              density="compact"
              :disabled="!!editingCode"
              :error-messages="formErrors.code"
              label="Code"
            />
            <v-select
              v-model="form.categoryId"
              clearable
              density="compact"
              item-title="name"
              item-value="id"
              :items="activeCategories"
              label="Category"
            />
            <v-select
              v-model="form.clientId"
              density="compact"
              :error-messages="formErrors.clientId"
              :item-title="c => `${c.name} (${c.phone})`"
              :item-value="c => c.id"
              :items="clients"
              label="Client"
            />
            <v-row>
              <v-col cols="6">
                <v-select
                  v-model="form.discountType"
                  density="compact"
                  :error-messages="formErrors.discountType"
                  :items="['fixed', 'percentage']"
                  label="Discount Type"
                />
              </v-col>
              <v-col cols="6">
                <v-text-field
                  v-model="form.validUntil"
                  density="compact"
                  :error-messages="formErrors.validUntil"
                  label="Valid Until"
                  type="date"
                />
              </v-col>
            </v-row>
            <v-textarea
              v-model="form.reason"
              density="compact"
              :error-messages="formErrors.reason"
              label="Reason"
              no-resize
              rows="2"
            />

            <v-alert v-if="submitError" class="mb-4" type="error" variant="tonal">
              {{ submitError }}
            </v-alert>

            <v-card-actions class="px-0">
              <v-spacer />
              <v-btn text="Cancel" @click="closeDialog" />
              <v-btn
                color="primary"
                :loading="submitting"
                :text="editingCode ? 'Save' : 'Add Coupon'"
                type="submit"
              />
            </v-card-actions>
          </form>
        </v-card-text>
      </v-card>
    </v-dialog>

    <CategoryDialog
      v-model="categoryDialog"
      :category="editingCategory"
      name-hint="Coupon group, e.g. Seasonal, VIP"
      :on-submit="submitCategory"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import CategoryChip from '@/components/CategoryChip.vue'
  import CategorySidebar from '@/components/CategorySidebar.vue'
  import CategoryDialog from '@/components/Forms/CategoryDialog.vue'
  import BulkDeleteBar from '@/components/Tables/BulkDeleteBar.vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useBulkDelete } from '@/composables/useBulkDelete'
  import { useClients } from '@/composables/useClients'
  import { useCouponCategories } from '@/composables/useCouponCategories'
  import { useCoupons } from '@/composables/useCoupons'
  import { API_BASE } from '@/config'

  const { createCoupon, updateCoupon, deactivateCoupon } = useCoupons()
  const { clients, fetchClients } = useClients()
  const {
    couponCategories,
    fetchCouponCategories,
    createCouponCategory,
    updateCouponCategory,
    deleteCouponCategory,
  } = useCouponCategories()

  const apiURL = `${API_BASE}/coupons`
  const headers = ref([
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Category', key: 'category_id', align: 'start', sortable: false },
    { title: 'Status', key: 'status', align: 'start' },
    { title: 'Discount Type', key: 'discount_type', align: 'start' },
    { title: 'Reason', key: 'reason', align: 'start' },
    { title: 'Client ID', key: 'client_id', align: 'end' },
    { title: 'Valid Until', key: 'valid_until', align: 'end' },
    { title: 'Actions', key: 'actions', align: 'end', sortable: false },
  ])

  const tableRef = ref(null)
  const search = ref('')
  const statusFilter = ref('all')
  const selectedCategory = ref(null)
  const loadError = ref('')

  const queryParams = computed(() => ({
    status: statusFilter.value === 'all' ? '' : statusFilter.value,
    category_id: selectedCategory.value || '',
  }))

  const activeCategories = computed(() => couponCategories.value.filter(c => c.is_active))

  function categoryFor (id) {
    return couponCategories.value.find(c => c.id === id) ?? null
  }

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? value : d.toLocaleDateString()
  }

  function reload () {
    tableRef.value?.reload()
  }

  onMounted(() => {
    fetchClients()
    fetchCouponCategories(true).catch(error => (loadError.value = error.message))
  })

  // ---- bulk delete ---------------------------------------------------
  const {
    selected: selectedCoupons,
    deleting: couponDeleting,
    error: couponSelError,
    reset: resetCouponSel,
    run: runCouponBulk,
  } = useBulkDelete()

  async function bulkDeleteCoupons () {
    try {
      await runCouponBulk('/coupons/bulk_delete', { codes: selectedCoupons.value })
      resetCouponSel()
      reload()
    } catch { /* error shown in the bar */ }
  }

  watch([statusFilter, selectedCategory], () => resetCouponSel())

  // ---- add / edit coupon -------------------------------------------------
  const dialog = ref(false)
  const editingCode = ref('')
  const submitting = ref(false)
  const submitError = ref('')
  const form = reactive({
    code: '',
    categoryId: null,
    clientId: null,
    discountType: 'percentage',
    validUntil: '',
    reason: '',
  })
  const formErrors = reactive({ code: '', clientId: '', discountType: '', validUntil: '', reason: '' })

  function resetForm () {
    Object.assign(form, {
      code: '', categoryId: null, clientId: null,
      discountType: 'percentage', validUntil: '', reason: '',
    })
    Object.assign(formErrors, { code: '', clientId: '', discountType: '', validUntil: '', reason: '' })
    submitError.value = ''
  }

  function openCreate () {
    editingCode.value = ''
    resetForm()
    dialog.value = true
  }

  function openEdit (item) {
    editingCode.value = item.code
    resetForm()
    Object.assign(form, {
      code: item.code,
      categoryId: item.category_id ?? null,
      clientId: item.client_id ?? null,
      discountType: item.discount_type,
      validUntil: item.valid_until?.slice(0, 10) ?? '',
      reason: item.reason,
    })
    dialog.value = true
  }

  function closeDialog () {
    dialog.value = false
    editingCode.value = ''
    resetForm()
  }

  function validate () {
    formErrors.code = form.code?.trim().length >= 1 ? '' : 'Code is required.'
    formErrors.clientId = form.clientId ? '' : 'Client is required.'
    formErrors.discountType = ['fixed', 'percentage'].includes(form.discountType) ? '' : 'Select a discount type.'
    formErrors.validUntil = form.validUntil ? '' : 'Valid until date is required.'
    formErrors.reason = form.reason?.trim().length >= 1 ? '' : 'Reason is required.'
    return !formErrors.code && !formErrors.clientId && !formErrors.discountType && !formErrors.validUntil && !formErrors.reason
  }

  async function submit () {
    if (!validate()) return
    submitting.value = true
    submitError.value = ''
    const payload = {
      status: 'active',
      discount_type: form.discountType,
      reason: form.reason.trim(),
      client_id: form.clientId ?? 0,
      valid_until: new Date(form.validUntil).toISOString(),
      category_id: form.categoryId ?? 0,
    }
    try {
      await (editingCode.value ? updateCoupon(editingCode.value, payload) : createCoupon({ code: form.code.trim(), ...payload }));
      closeDialog()
      reload()
    } catch (error) {
      submitError.value = error.message
    } finally {
      submitting.value = false
    }
  }

  async function deactivate (item) {
    try {
      await deactivateCoupon(item.code)
      reload()
    } catch (error) {
      loadError.value = error.message
    }
  }

  // ---- categories --------------------------------------------------------
  const categoryDialog = ref(false)
  const editingCategory = ref(null)

  function openCreateCategory () {
    editingCategory.value = null
    categoryDialog.value = true
  }

  function openEditCategory (category) {
    editingCategory.value = category
    categoryDialog.value = true
  }

  async function submitCategory (payload) {
    await (editingCategory.value
      ? updateCouponCategory({ id: editingCategory.value.id, ...payload })
      : createCouponCategory(payload))
    await fetchCouponCategories(true)
    reload()
  }

  async function removeCategory (category) {
    if (!confirm(`Delete category "${category.name}"? Categories still used by a coupon can't be deleted.`)) return
    try {
      await deleteCouponCategory(category.id)
      if (selectedCategory.value === category.id) selectedCategory.value = null
      await fetchCouponCategories(true)
    } catch (error) {
      loadError.value = error.message
    }
  }
</script>

<style scoped>
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.coupons-card {
  display: flex;
  flex: 1 1 auto;
  min-height: 0;
  flex-direction: column;
}

.coupons-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}

.coupons-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
}

.coupon-search {
  flex: 0 1 320px;
}
</style>
