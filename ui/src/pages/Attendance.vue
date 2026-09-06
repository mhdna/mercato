<template>
  <div class="pa-4">
    <v-card flat>
      <v-card-title class="d-flex flex-wrap align-center ga-3 px-2 py-3">
        <v-icon icon="mdi-clock-check-outline" />
        <span>Attendance</span>
        <v-select
          v-model="branchId"
          class="flex-grow-0"
          clearable
          density="compact"
          hide-details
          item-title="name"
          item-value="id"
          :items="branches"
          label="Branch"
          style="max-width: 220px"
          variant="outlined"
        />
        <v-spacer />
        <v-btn
          prepend-icon="mdi-refresh"
          size="small"
          text="Sync from POS"
          variant="text"
          @click="reload"
        />
        <v-btn
          color="primary"
          prepend-icon="mdi-file-delimited-outline"
          size="small"
          text="Import CSV"
          variant="flat"
          @click="fileInput?.click()"
        />
        <input
          ref="fileInput"
          accept=".csv,text/csv"
          hidden
          type="file"
          @change="onFilePicked"
        >
      </v-card-title>
      <v-divider />

      <v-alert
        v-if="message"
        class="ma-2"
        closable
        density="compact"
        :type="messageType"
        variant="tonal"
        @click:close="message = ''"
      >
        {{ message }}
      </v-alert>

      <p v-if="!branchId" class="text-medium-emphasis text-caption pa-4">
        Attendance punches sync from kashi-pos automatically. Pick a branch to review them,
        or import a CSV (<code>date,time,name,type,status,user_id</code>) for a branch with no device.
      </p>

      <ServerSideTable
        v-else
        ref="tableRef"
        :api-u-r-l="apiURL"
        :headers="headers"
        hover
        item-value="id"
        :query-params="{ branch_id: branchId }"
        root-key="branch_attendance_events"
        :show-search-icon="false"
      >
        <template #item.type="{ item }">
          <v-chip
            :color="item.type?.toLowerCase().includes('out') ? 'deep-orange' : 'green'"
            size="x-small"
            :text="item.type || '—'"
          />
        </template>
        <template #item.actions="{ item }">
          <v-btn icon="mdi-pencil" size="x-small" variant="text" @click.stop="openEdit(item)" />
          <v-btn
            color="error"
            icon="mdi-delete"
            size="x-small"
            variant="text"
            @click.stop="remove(item)"
          />
        </template>
      </ServerSideTable>
    </v-card>

    <v-dialog v-model="dialog" max-width="460">
      <v-card class="px-4 py-2">
        <v-card-title>Edit attendance punch</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.salesperson_name" density="compact" label="Name" />
          <v-row dense>
            <v-col cols="6">
              <v-text-field v-model="form.event_date" density="compact" label="Date" type="date" />
            </v-col>
            <v-col cols="6">
              <v-text-field v-model="form.event_time" density="compact" label="Time" type="time" />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="6">
              <v-select v-model="form.type" density="compact" :items="['IN', 'OUT']" label="Type" />
            </v-col>
            <v-col cols="6">
              <v-text-field v-model="form.status" density="compact" label="Status" />
            </v-col>
          </v-row>
          <v-alert
            v-if="formError"
            class="mt-2"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ formError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" @click="dialog = false" />
          <v-btn color="primary" :loading="submitting" text="Save" @click="submit" />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'
  import { useAttendance } from '@/composables/useAttendance'
  import { useBranches } from '@/composables/useBranches'
  import { API_BASE } from '@/config'

  const { importCsv, updateEvent, deleteEvent } = useAttendance()
  const { branches, fetchBranches } = useBranches()

  const apiURL = `${API_BASE}/branch_attendance_events`
  const headers = [
    { title: 'Date', key: 'event_date', align: 'start' },
    { title: 'Time', key: 'event_time' },
    { title: 'Name', key: 'salesperson_name', align: 'start' },
    { title: 'Type', key: 'type' },
    { title: 'Status', key: 'status' },
    { title: 'User ID', key: 'attendance_user_id' },
    { title: '', key: 'actions', align: 'end', sortable: false },
  ]

  const branchId = ref(null)
  const tableRef = ref(null)
  const fileInput = ref(null)
  const message = ref('')
  const messageType = ref('success')

  function reload () {
    tableRef.value?.reload()
  }

  function flash (text, type = 'success') {
    message.value = text
    messageType.value = type
  }

  onMounted(() => {
    fetchBranches().catch(() => {})
  })

  async function onFilePicked (event) {
    const file = event.target.files?.[0]
    event.target.value = ''
    if (!file) return
    if (!branchId.value) {
      flash('Pick a branch before importing.', 'warning')
      return
    }
    try {
      const res = await importCsv(file, branchId.value)
      flash(`Imported ${res?.inserted ?? 0} punches (${res?.skipped ?? 0} already present).`)
      reload()
    } catch (error) {
      flash(error.message, 'error')
    }
  }

  // ---- edit / delete ---------------------------------------------------
  const dialog = ref(false)
  const submitting = ref(false)
  const formError = ref('')
  const editingId = ref(null)
  const form = reactive({ salesperson_name: '', event_date: '', event_time: '', type: '', status: '', attendance_user_id: '', event_at: '' })

  function openEdit (item) {
    editingId.value = item.id
    Object.assign(form, {
      salesperson_name: item.salesperson_name ?? '',
      event_date: item.event_date ?? '',
      event_time: item.event_time ?? '',
      type: item.type ?? '',
      status: item.status ?? '',
      attendance_user_id: item.attendance_user_id ?? '',
      event_at: item.event_at ?? '',
    })
    formError.value = ''
    dialog.value = true
  }

  async function submit () {
    submitting.value = true
    formError.value = ''
    try {
      await updateEvent(editingId.value, {
        salesperson_name: form.salesperson_name,
        attendance_user_id: form.attendance_user_id,
        event_date: form.event_date,
        event_time: form.event_time,
        event_at: `${form.event_date} ${form.event_time}`.trim(),
        type: form.type,
        status: form.status,
      })
      dialog.value = false
      reload()
    } catch (error) {
      formError.value = error.message
    } finally {
      submitting.value = false
    }
  }

  async function remove (item) {
    if (!confirm(`Delete the ${item.event_date} ${item.event_time} punch for ${item.salesperson_name || 'this employee'}?`)) return
    try {
      await deleteEvent(item.id)
      reload()
    } catch (error) {
      flash(error.message, 'error')
    }
  }
</script>
