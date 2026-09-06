<template>
  <div class="pa-2" style="width: 100%">
    <div class="d-flex flex-wrap align-center ga-2 mb-2">
      <v-icon icon="mdi-account-cash" />
      <span class="text-subtitle-2">Employees &amp; Payroll</span>
      <v-text-field
        v-model="search"
        class="flex-grow-0"
        clearable
        density="compact"
        hide-details
        label="Search name or role"
        prepend-inner-icon="mdi-magnify"
        style="max-width: 260px"
        variant="outlined"
        @update:model-value="debouncedReload"
      />
      <v-select
        v-model="branchFilter"
        class="flex-grow-0"
        clearable
        density="compact"
        hide-details
        item-title="name"
        item-value="id"
        :items="branches"
        label="Branch"
        style="max-width: 200px"
        variant="outlined"
        @update:model-value="reload"
      />
      <v-spacer />
      <template v-if="editing">
        <v-btn size="small" text="Cancel" variant="text" @click="cancelEdit" />
        <v-btn
          color="primary"
          :disabled="!dirtyCount"
          :loading="saving"
          prepend-icon="mdi-content-save-all"
          size="small"
          :text="`Save all${dirtyCount ? ` (${dirtyCount})` : ''}`"
          variant="flat"
          @click="saveAll"
        />
      </template>
      <template v-else>
        <v-btn
          prepend-icon="mdi-pencil"
          size="small"
          text="Edit salaries"
          variant="outlined"
          @click="startEdit"
        />
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          size="small"
          text="Add employee"
          variant="flat"
          @click="openCreate"
        />
      </template>
    </div>

    <v-alert
      v-if="error"
      class="mb-2"
      closable
      density="compact"
      type="error"
      variant="tonal"
      @click:close="error = ''"
    >
      {{ error }}
    </v-alert>

    <v-data-table
      class="text-caption text-center"
      density="compact"
      :headers="headers"
      hide-default-footer
      hover
      :items="rows"
      :items-per-page="-1"
      :loading="loading"
    >
      <template #item="{ item }">
        <tr class="text-no-wrap text-center">
          <td class="text-start">{{ item.name }}</td>
          <td>{{ branchName(item.branch_id) }}</td>
          <td>{{ item.role || '—' }}</td>

          <!-- Base salary: never colorized (per requirements) -->
          <td class="text-end">
            <v-text-field
              v-if="editing"
              v-model="draft[item.id].base"
              density="compact"
              hide-details
              prefix="$"
              style="min-width: 90px"
              type="number"
              variant="plain"
            />
            <template v-else>{{ money(item.base_salary_cents) }}</template>
          </td>

          <!-- Commission: green when positive -->
          <td class="text-end" :class="commissionClass(item)">
            <v-text-field
              v-if="editing"
              v-model="draft[item.id].commission"
              density="compact"
              hide-details
              prefix="$"
              style="min-width: 90px"
              type="number"
              variant="plain"
            />
            <template v-else>{{ money(item.commission_cents) }}</template>
          </td>

          <!-- Late deduction rate x units: amber when any lateness -->
          <td :class="lateClass(item)">
            <div v-if="editing" class="d-flex align-center ga-1 justify-center">
              <v-text-field
                v-model="draft[item.id].rate"
                density="compact"
                hide-details
                prefix="$"
                style="min-width: 70px"
                type="number"
                variant="plain"
              />
              <span>/</span>
              <v-select
                v-model="draft[item.id].unit"
                density="compact"
                hide-details
                :items="['hour', 'day']"
                style="min-width: 74px"
                variant="plain"
              />
              <span class="mx-1">×</span>
              <v-text-field
                v-model="draft[item.id].units"
                density="compact"
                hide-details
                style="min-width: 60px"
                :suffix="draft[item.id].unit === 'day' ? 'd' : 'h'"
                type="number"
                variant="plain"
              />
            </div>
            <template v-else>
              {{ money(item.late_deduction_rate_cents) }}/{{ item.late_deduction_unit }}
              × {{ unitsLabel(item) }}
            </template>
          </td>

          <!-- Deduction: red when positive -->
          <td class="text-end" :class="deductionClass(item)" v-text="`-${money(deductionOf(item))}`" />

          <!-- Net salary: colorized background by ratio to base -->
          <td class="text-end font-weight-medium" :class="netClass(item)" v-text="money(netOf(item))" />

          <td>
            <v-chip
              :color="item.status === 'active' ? 'success' : 'default'"
              size="x-small"
              :text="item.status"
            />
          </td>
          <td class="text-end">
            <v-btn
              icon="mdi-pencil"
              size="x-small"
              variant="text"
              @click="openEdit(item)"
            />
            <v-btn
              color="error"
              icon="mdi-delete"
              size="x-small"
              variant="text"
              @click="remove(item)"
            />
          </td>
        </tr>
      </template>
    </v-data-table>

    <!-- add / edit employee -->
    <v-dialog v-model="dialog" max-width="520">
      <v-card class="px-4 py-2">
        <v-card-title>{{ form.id ? 'Edit employee' : 'Add employee' }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="form.name"
            density="compact"
            :error-messages="formError.name"
            label="Name"
          />
          <v-row dense>
            <v-col cols="6">
              <v-text-field v-model="form.role" density="compact" label="Role" />
            </v-col>
            <v-col cols="6">
              <v-select
                v-model="form.branchId"
                clearable
                density="compact"
                item-title="name"
                item-value="id"
                :items="branches"
                label="Branch (optional)"
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="6">
              <v-select
                v-model="form.status"
                density="compact"
                :items="['active', 'inactive']"
                label="Status"
              />
            </v-col>
            <v-col cols="6">
              <v-text-field
                v-model="form.hiredOn"
                density="compact"
                label="Hired on"
                type="date"
              />
            </v-col>
          </v-row>
          <v-text-field
            v-model="form.base"
            density="compact"
            label="Base salary"
            prefix="$"
            type="number"
          />
          <v-alert
            v-if="formError.submit"
            class="mt-2"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ formError.submit }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" @click="dialog = false" />
          <v-btn
            color="primary"
            :loading="submitting"
            :text="form.id ? 'Save' : 'Add'"
            @click="submit"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useBranches } from '@/composables/useBranches'
  import { useEmployees } from '@/composables/useEmployees'

  const { listEmployees, createEmployee, updateEmployee, deleteEmployee, batchSetSalaries } = useEmployees()
  const { branches, fetchBranches } = useBranches()

  const headers = [
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Branch', key: 'branch' },
    { title: 'Role', key: 'role' },
    { title: 'Base', key: 'base', align: 'end' },
    { title: 'Commission', key: 'commission', align: 'end' },
    { title: 'Late حسم (rate × units)', key: 'late' },
    { title: 'Deduction', key: 'deduction', align: 'end' },
    { title: 'Net salary', key: 'net', align: 'end' },
    { title: 'Status', key: 'status' },
    { title: '', key: 'actions', align: 'end', sortable: false },
  ]

  const rows = ref([])
  const loading = ref(false)
  const error = ref('')
  const search = ref('')
  const branchFilter = ref(null)

  function money (cents) {
    return `$${(Number(cents || 0) / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
  }

  function branchName (id) {
    if (!id) return '—'
    return branches.value.find(b => b.id === id)?.name ?? `#${id}`
  }

  function unitsLabel (item) {
    const n = Number(item.late_units_centi || 0) / 100
    return `${n} ${item.late_deduction_unit === 'day' ? 'd' : 'h'}`
  }

  // Mirror the server's arithmetic so edit-mode cells update live before a save.
  function deductionOf (item) {
    if (editing.value && draft.value[item.id]) {
      const d = draft.value[item.id]
      return Math.round(toCents(d.rate) * toCenti(d.units) / 100)
    }
    return item.deduction_cents
  }
  function netOf (item) {
    if (editing.value && draft.value[item.id]) {
      const d = draft.value[item.id]
      return toCents(d.base) + toCents(d.commission) - deductionOf(item)
    }
    return item.net_salary_cents
  }
  function baseOf (item) {
    if (editing.value && draft.value[item.id]) return toCents(draft.value[item.id].base)
    return item.base_salary_cents
  }
  function commissionOf (item) {
    if (editing.value && draft.value[item.id]) return toCents(draft.value[item.id].commission)
    return item.commission_cents
  }

  const commissionClass = item => (commissionOf(item) > 0 ? 'bg-success' : '')
  const deductionClass = item => (deductionOf(item) > 0 ? 'bg-error' : '')
  const lateClass = item => (deductionOf(item) > 0 ? 'bg-warning' : '')
  function netClass (item) {
    const base = baseOf(item)
    if (base <= 0) return ''
    const net = netOf(item)
    if (net >= base) return 'bg-success'
    if (net >= base * 0.8) return 'bg-warning'
    return 'bg-error'
  }

  let searchTimer
  function debouncedReload () {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(reload, 250)
  }

  async function reload () {
    loading.value = true
    error.value = ''
    try {
      const { employees } = await listEmployees({
        search: search.value || '',
        branchId: branchFilter.value || 0,
      })
      rows.value = employees
      if (editing.value) seedDraft()
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    fetchBranches().catch(() => {})
    reload()
  })

  // ---- batch salary editing -------------------------------------------
  const editing = ref(false)
  const saving = ref(false)
  const draft = ref({})

  function toCents (dollars) {
    return Math.round(Number.parseFloat(dollars || 0) * 100) || 0
  }
  function toCenti (units) {
    return Math.round(Number.parseFloat(units || 0) * 100) || 0
  }

  function seedDraft () {
    const next = {}
    for (const r of rows.value) {
      next[r.id] = {
        base: (r.base_salary_cents / 100).toString(),
        commission: (r.commission_cents / 100).toString(),
        rate: (r.late_deduction_rate_cents / 100).toString(),
        unit: r.late_deduction_unit,
        units: (r.late_units_centi / 100).toString(),
      }
    }
    draft.value = next
  }

  function startEdit () {
    seedDraft()
    editing.value = true
  }
  function cancelEdit () {
    editing.value = false
    draft.value = {}
  }

  function rowChanged (r) {
    const d = draft.value[r.id]
    if (!d) return false
    return toCents(d.base) !== r.base_salary_cents
      || toCents(d.commission) !== r.commission_cents
      || toCents(d.rate) !== r.late_deduction_rate_cents
      || d.unit !== r.late_deduction_unit
      || toCenti(d.units) !== r.late_units_centi
  }

  const dirtyCount = computed(() => rows.value.filter(r => rowChanged(r)).length)

  async function saveAll () {
    const items = rows.value.filter(r => rowChanged(r)).map(r => {
      const d = draft.value[r.id]
      return {
        id: r.id,
        base_salary_cents: toCents(d.base),
        commission_cents: toCents(d.commission),
        late_deduction_rate_cents: toCents(d.rate),
        late_deduction_unit: d.unit,
        late_units_centi: toCenti(d.units),
      }
    })
    if (items.length === 0) return
    saving.value = true
    error.value = ''
    try {
      await batchSetSalaries(items)
      editing.value = false
      draft.value = {}
      await reload()
    } catch (error_) {
      error.value = error_.message
    } finally {
      saving.value = false
    }
  }

  // ---- add / edit employee ------------------------------------------------
  const dialog = ref(false)
  const submitting = ref(false)
  // payroll.* are carried unchanged through a profile edit so saving the
  // dialog never zeroes the salary columns (those are edited via "Edit
  // salaries" / the batch endpoint instead).
  const form = reactive({
    id: null, name: '', role: '', branchId: null, status: 'active', hiredOn: '', base: '0',
    payroll: { commission_cents: 0, late_deduction_rate_cents: 0, late_deduction_unit: 'hour', late_units_centi: 0 },
  })
  const formError = reactive({ name: '', submit: '' })

  function resetForm () {
    Object.assign(form, {
      id: null, name: '', role: '', branchId: null, status: 'active', hiredOn: '', base: '0',
      payroll: { commission_cents: 0, late_deduction_rate_cents: 0, late_deduction_unit: 'hour', late_units_centi: 0 },
    })
    Object.assign(formError, { name: '', submit: '' })
  }

  function openCreate () {
    resetForm()
    dialog.value = true
  }

  function openEdit (item) {
    resetForm()
    Object.assign(form, {
      id: item.id,
      name: item.name,
      role: item.role,
      branchId: item.branch_id ?? null,
      status: item.status,
      hiredOn: item.hired_on ?? '',
      base: (item.base_salary_cents / 100).toString(),
      payroll: {
        commission_cents: item.commission_cents,
        late_deduction_rate_cents: item.late_deduction_rate_cents,
        late_deduction_unit: item.late_deduction_unit,
        late_units_centi: item.late_units_centi,
      },
    })
    dialog.value = true
  }

  async function submit () {
    formError.name = form.name.trim() ? '' : 'Name is required.'
    if (formError.name) return
    submitting.value = true
    formError.submit = ''
    const payload = {
      name: form.name.trim(),
      role: form.role.trim(),
      branch_id: form.branchId ?? 0,
      status: form.status,
      hired_on: form.hiredOn || '',
      base_salary_cents: toCents(form.base),
      ...form.payroll,
    }
    try {
      await (form.id ? updateEmployee(form.id, payload) : createEmployee(payload))
      dialog.value = false
      await reload()
    } catch (error_) {
      formError.submit = error_.message
    } finally {
      submitting.value = false
    }
  }

  async function remove (item) {
    if (!confirm(`Delete employee "${item.name}"?`)) return
    try {
      await deleteEmployee(item.id)
      await reload()
    } catch (error_) {
      error.value = error_.message
    }
  }
</script>
