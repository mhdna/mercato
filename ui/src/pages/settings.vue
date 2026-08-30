<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2 class="text-h6">Settings</h2>
    </div>

    <v-card max-width="480">
      <v-card-title class="text-subtitle-1">Branch activity</v-card-title>
      <v-card-subtitle>How live branch sales/returns/expenses are shown in the appbar</v-card-subtitle>
      <v-card-text>
        <div class="text-body-2 mb-2">Display mode</div>
        <v-btn-toggle
          v-model="displayMode"
          class="mb-4"
          color="primary"
          divided
          mandatory
        >
          <v-btn value="notification">Notification</v-btn>
          <v-btn value="appbar">Appbar messages</v-btn>
        </v-btn-toggle>

        <v-text-field
          v-model.number="messageSeconds"
          density="compact"
          hint="How long each message is shown before moving to the next one queued behind it"
          label="Message duration (seconds)"
          min="1"
          persistent-hint
          type="number"
        />
      </v-card-text>
    </v-card>

    <v-card class="mt-4" max-width="480">
      <v-card-title class="text-subtitle-1">Financials</v-card-title>
      <v-card-subtitle>Dashboard "Financials" tab revenue chart</v-card-subtitle>
      <v-card-text>
        <v-select
          v-model="highSeasonMonths"
          chips
          density="compact"
          hint="Months that earn ~1.5x a normal month, shaded on the revenue candles chart"
          :items="monthItems"
          label="High season months"
          multiple
          persistent-hint
        />
      </v-card-text>
    </v-card>

    <v-card class="mt-4" max-width="480">
      <v-card-title class="text-subtitle-1">Barcode labels</v-card-title>
      <v-card-subtitle>Default page size for generated barcode-label PDFs</v-card-subtitle>
      <v-card-text>
        <div class="d-flex ga-3">
          <v-text-field
            v-model.number="labelWidth"
            density="compact"
            hide-details
            label="Width (pt)"
            type="number"
            variant="outlined"
          />
          <v-text-field
            v-model.number="labelHeight"
            density="compact"
            hide-details
            label="Height (pt)"
            type="number"
            variant="outlined"
          />
        </div>
        <div class="text-caption text-medium-emphasis mt-1">72 pt = 1 inch. Common thermal label: 288 × 144 (4 × 2 in).</div>
        <v-btn
          class="mt-3"
          :disabled="!labelSizeDirty"
          :loading="savingLabel"
          size="small"
          text="Save"
          variant="tonal"
          @click="saveLabelSize"
        />
      </v-card-text>
    </v-card>

    <v-card class="mt-4" max-width="640">
      <v-card-title class="d-flex align-center justify-space-between">
        <span class="text-subtitle-1">Users</span>
        <v-btn
          color="primary"
          prepend-icon="mdi-plus"
          size="small"
          variant="tonal"
          @click="openCreate"
        >
          Add user
        </v-btn>
      </v-card-title>
      <v-card-subtitle>POS operators and their login PINs. Changes take effect on the next login.</v-card-subtitle>
      <v-card-text>
        <v-alert
          v-if="usersError"
          class="mb-3"
          density="compact"
          type="error"
          variant="tonal"
        >
          {{ usersError }}
        </v-alert>

        <v-table density="compact">
          <thead>
            <tr>
              <th class="text-left">Name</th>
              <th class="text-left">Status</th>
              <th class="text-left">Created</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td>{{ u.name }}</td>
              <td>
                <v-chip :color="u.activated ? 'success' : 'default'" size="x-small" variant="tonal">
                  {{ u.activated ? 'Active' : 'Inactive' }}
                </v-chip>
              </td>
              <td>{{ formatDate(u.created_at) }}</td>
              <td class="text-right">
                <v-btn
                  icon="mdi-key"
                  size="small"
                  title="Change PIN"
                  variant="text"
                  @click="openPin(u)"
                />
                <v-btn
                  color="error"
                  icon="mdi-delete"
                  size="small"
                  title="Delete user"
                  variant="text"
                  @click="confirmDelete(u)"
                />
              </td>
            </tr>
            <tr v-if="users.length === 0">
              <td class="text-medium-emphasis py-4" colspan="4">No users yet.</td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
    </v-card>

    <!-- Add user -->
    <v-dialog v-model="createDialog" max-width="420">
      <v-card>
        <v-card-title class="text-subtitle-1">Add user</v-card-title>
        <v-card-text>
          <v-alert
            v-if="formError"
            class="mb-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ formError }}
          </v-alert>
          <v-text-field
            v-model="form.name"
            density="compact"
            label="Name"
            :rules="[nameRule]"
          />
          <v-text-field
            v-model="form.pin"
            density="compact"
            hint="4 to 6 digits"
            inputmode="numeric"
            label="PIN"
            maxlength="6"
            persistent-hint
            :rules="[pinRule]"
          />
          <v-switch
            v-model="form.activated"
            color="primary"
            hide-details
            label="Active"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="createDialog = false">Cancel</v-btn>
          <v-btn
            color="primary"
            :disabled="!canSubmitCreate"
            :loading="saving"
            variant="tonal"
            @click="submitCreate"
          >
            Create
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Change PIN -->
    <v-dialog v-model="pinDialog" max-width="420">
      <v-card>
        <v-card-title class="text-subtitle-1">Change PIN — {{ pinTarget?.name }}</v-card-title>
        <v-card-text>
          <v-alert
            v-if="formError"
            class="mb-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ formError }}
          </v-alert>
          <v-text-field
            v-model="newPin"
            autofocus
            density="compact"
            hint="4 to 6 digits"
            inputmode="numeric"
            label="New PIN"
            maxlength="6"
            persistent-hint
            :rules="[pinRule]"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="pinDialog = false">Cancel</v-btn>
          <v-btn
            color="primary"
            :disabled="!isPin(newPin)"
            :loading="saving"
            variant="tonal"
            @click="submitPin"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete confirm -->
    <v-dialog v-model="deleteDialog" max-width="380">
      <v-card>
        <v-card-title class="text-subtitle-1">Delete user</v-card-title>
        <v-card-text>
          Delete <strong>{{ deleteTarget?.name }}</strong>? This cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="saving" variant="tonal" @click="submitDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useUsers } from '@/composables/useUsers'
  import { useSettingsStore } from '@/stores/settings'

  const settingsStore = useSettingsStore()
  onMounted(async () => {
    await settingsStore.init()
    labelWidth.value = settingsStore.barcodeLabelWidth
    labelHeight.value = settingsStore.barcodeLabelHeight
  })

  // ---- Barcode label size ----
  const labelWidth = ref(288)
  const labelHeight = ref(144)
  const savingLabel = ref(false)
  const labelSizeDirty = computed(() =>
    Number(labelWidth.value) !== settingsStore.barcodeLabelWidth
    || Number(labelHeight.value) !== settingsStore.barcodeLabelHeight,
  )
  async function saveLabelSize () {
    const w = Number(labelWidth.value)
    const h = Number(labelHeight.value)
    if (!(w > 0) || !(h > 0)) {
      return
    }
    savingLabel.value = true
    try {
      await settingsStore.setBarcodeLabelSize(w, h)
    } finally {
      savingLabel.value = false
    }
  }

  const displayMode = computed({
    get: () => settingsStore.activityDisplayMode,
    set: value => settingsStore.setActivityDisplayMode(value),
  })

  const messageSeconds = computed({
    get: () => settingsStore.activityMessageSeconds,
    set: value => {
      const seconds = Number(value)
      if (Number.isFinite(seconds) && seconds > 0) {
        settingsStore.setActivityMessageSeconds(seconds)
      }
    },
  })

  const monthItems = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
  ].map((title, i) => ({ title, value: i + 1 }))

  const highSeasonMonths = computed({
    get: () => settingsStore.financialsHighSeasonMonths,
    set: value => settingsStore.setFinancialsHighSeasonMonths(value),
  })

  // ---- Users ----
  const { users, fetchUsers, createUser, updateUser, deleteUser } = useUsers()

  const usersError = ref('')
  const formError = ref('')
  const saving = ref(false)

  onMounted(async () => {
    try {
      await fetchUsers()
    } catch (error) {
      usersError.value = error.message
    }
  })

  function formatDate (value) {
    if (!value) return '—'
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? '—' : d.toLocaleDateString()
  }

  const isPin = v => /^\d{4,6}$/.test(String(v ?? ''))
  const pinRule = v => isPin(v) || 'PIN must be 4 to 6 digits'
  const nameRule = v => /^[a-zA-Z0-9]+$/.test(String(v ?? '')) || 'Letters and numbers only'

  // Add user
  const createDialog = ref(false)
  const form = reactive({ name: '', pin: '', activated: true })
  const canSubmitCreate = computed(() => nameRule(form.name) === true && isPin(form.pin))

  function openCreate () {
    form.name = ''
    form.pin = ''
    form.activated = true
    formError.value = ''
    createDialog.value = true
  }

  async function submitCreate () {
    saving.value = true
    formError.value = ''
    try {
      await createUser({ name: form.name, password: form.pin, activated: form.activated })
      await fetchUsers(true)
      createDialog.value = false
    } catch (error) {
      formError.value = error.message
    } finally {
      saving.value = false
    }
  }

  // Change PIN
  const pinDialog = ref(false)
  const pinTarget = ref(null)
  const newPin = ref('')

  function openPin (user) {
    pinTarget.value = user
    newPin.value = ''
    formError.value = ''
    pinDialog.value = true
  }

  async function submitPin () {
    saving.value = true
    formError.value = ''
    try {
      await updateUser({
        id: pinTarget.value.id,
        name: pinTarget.value.name,
        activated: pinTarget.value.activated,
        password: newPin.value,
      })
      await fetchUsers(true)
      pinDialog.value = false
    } catch (error) {
      formError.value = error.message
    } finally {
      saving.value = false
    }
  }

  // Delete
  const deleteDialog = ref(false)
  const deleteTarget = ref(null)

  function confirmDelete (user) {
    deleteTarget.value = user
    formError.value = ''
    deleteDialog.value = true
  }

  async function submitDelete () {
    saving.value = true
    try {
      await deleteUser(deleteTarget.value.id)
      await fetchUsers(true)
      deleteDialog.value = false
    } catch (error) {
      usersError.value = error.message
      deleteDialog.value = false
    } finally {
      saving.value = false
    }
  }
</script>
