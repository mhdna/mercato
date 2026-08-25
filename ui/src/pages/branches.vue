<template>
  <v-dialog v-model="createDialog" max-width="480" :persistent="true">
    <v-card class="px-4">
      <v-card-title>Add a New Branch</v-card-title>
      <v-card-text>
        <form @submit.prevent="submitCreate">
          <v-text-field
            v-model="createName"
            density="compact"
            label="Name"
          />
          <v-text-field
            v-model="createCode"
            density="compact"
            hint="Short unique code, e.g. MAIN"
            label="Code"
            persistent-hint
            @update:model-value="createCode = createCode.toUpperCase()"
          />

          <v-alert v-if="createError" class="mb-4 mt-4" type="error" variant="tonal">
            {{ createError }}
          </v-alert>

          <v-card-actions class="px-0">
            <v-spacer />
            <v-btn text="Cancel" @click="createDialog = false" />
            <v-btn color="primary" :loading="creating" text="Add Branch" type="submit" />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <!-- Shown exactly once, right after creating a branch or rotating its
       key -- the plaintext API key is never retrievable again afterward,
       only its bcrypt hash is stored server-side. -->
  <v-dialog v-model="keyDialog" max-width="520" :persistent="true">
    <v-card class="px-4">
      <v-card-title>API Key</v-card-title>
      <v-card-text>
        <v-alert class="mb-4" type="warning" variant="tonal">
          This key is shown once. Copy it into the branch's kashi-pos settings now --
          it cannot be retrieved again after closing this dialog.
        </v-alert>
        <v-text-field
          v-model="revealedKey"
          density="compact"
          label="API key"
          readonly
          variant="outlined"
        >
          <template #append-inner>
            <v-icon-btn icon="mdi-content-copy" size="small" variant="text" @click="copyKey" />
          </template>
        </v-text-field>
        <v-alert v-if="copied" class="mt-2" type="success" variant="tonal">Copied to clipboard</v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn color="primary" text="Done" @click="keyDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog v-model="rotateDialog" max-width="420">
    <v-card>
      <v-card-title>Rotate API Key</v-card-title>
      <v-card-text>
        Rotating "{{ rotateTarget?.name }}"'s key immediately invalidates its current key.
        The branch's kashi-pos install will need the new key entered before it can sync again.
      </v-card-text>
      <v-alert v-if="rotateError" class="mx-4 mb-2" type="error" variant="tonal">{{ rotateError }}</v-alert>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Cancel" @click="rotateDialog = false" />
        <v-btn color="warning" :loading="rotating" text="Rotate Key" @click="confirmRotate" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Editing here never touches the branch's database directly -- Save
       enqueues a remote command (POST /branches/:id/commands) that the
       branch executes itself via its own real, validated local functions.
       See api/branch_command.go and kashi-pos's commands.go. -->
  <v-dialog v-model="settingsDialog" max-width="640">
    <v-card>
      <v-card-title>Settings -- {{ settingsTarget?.name }}</v-card-title>
      <v-card-text>
        <v-alert v-if="settingsError" class="mb-4" type="error" variant="tonal">{{ settingsError }}</v-alert>
        <v-alert v-if="saveOutcome" class="mb-4" :type="saveOutcome.ok ? 'success' : 'error'" variant="tonal">
          {{ saveOutcome.message }}
        </v-alert>
        <div v-if="settingsLoading" class="d-flex justify-center pa-4">
          <v-progress-circular color="primary" indeterminate />
        </div>
        <template v-else-if="!settingsData && !settingsForm">
          <v-alert type="info" variant="tonal">
            This branch hasn't pushed its settings yet -- it needs to sync at least once before you can view or edit them.
          </v-alert>
        </template>
        <form v-else class="field-grid" @submit.prevent="saveSettings">
          <v-text-field v-model="settingsForm.branch_name" density="compact" label="Branch Name" variant="outlined" />
          <v-text-field
            v-model.number="settingsForm.tax_rate"
            density="compact"
            label="Tax Rate (0-1)"
            max="1"
            min="0"
            step="0.01"
            type="number"
            variant="outlined"
          />
          <v-select
            v-model="settingsForm.rounding_mode"
            density="compact"
            :items="['half_up', 'up', 'down']"
            label="Rounding Mode"
            variant="outlined"
          />
          <v-text-field v-model="settingsForm.rounding_currency" density="compact" label="Rounding Currency" variant="outlined" />
          <v-text-field
            v-model.number="settingsForm.exchange_rate"
            density="compact"
            label="Exchange Rate"
            type="number"
            variant="outlined"
          />
          <v-text-field
            v-model.number="settingsForm.exchange_window_hours"
            density="compact"
            label="Exchange Window (hours)"
            type="number"
            variant="outlined"
          />
          <v-text-field v-model="settingsForm.market_name" density="compact" label="Market Name" variant="outlined" />
          <v-text-field v-model="settingsForm.market_phone" density="compact" label="Market Phone" variant="outlined" />
          <v-textarea
            v-model="settingsForm.market_description"
            density="compact"
            label="Market Description"
            rows="2"
            variant="outlined"
          />
          <v-textarea
            v-model="settingsForm.return_policy"
            density="compact"
            label="Return Policy"
            rows="2"
            variant="outlined"
          />
          <v-text-field v-model="settingsForm.website" density="compact" label="Website" variant="outlined" />
          <v-text-field v-model="settingsForm.instagram" density="compact" label="Instagram" variant="outlined" />
          <v-textarea
            v-model="socialPlatformsText"
            density="compact"
            hint="JSON array, e.g. [&quot;facebook&quot;, &quot;instagram&quot;]"
            label="Social Platforms"
            persistent-hint
            rows="2"
            variant="outlined"
          />
          <v-textarea
            v-model="socialHandlesText"
            density="compact"
            hint="JSON object, e.g. {&quot;instagram&quot;: &quot;@handle&quot;}"
            label="Social Handles"
            persistent-hint
            rows="2"
            variant="outlined"
          />
          <div class="d-flex justify-end field-full">
            <v-btn color="primary" :loading="saving" text="Save" type="submit" />
          </div>
        </form>

        <v-divider class="my-4" />
        <div class="d-flex justify-space-between align-center mb-2">
          <span class="text-subtitle-2">Recent Commands</span>
          <v-icon-btn icon="mdi-refresh" size="small" variant="text" @click="loadCommandHistory" />
        </div>
        <v-table density="compact">
          <thead>
            <tr><th>Type</th><th>Status</th><th>Issued</th><th>Error</th></tr>
          </thead>
          <tbody>
            <tr v-for="cmd in commandHistory" :key="cmd.id">
              <td>{{ cmd.type }}</td>
              <td>
                <v-chip :color="cmd.status === 'success' ? 'success' : cmd.status === 'failed' ? 'error' : 'warning'" size="small">
                  {{ cmd.status }}
                </v-chip>
              </td>
              <td>{{ new Date(cmd.created_at).toLocaleString() }}</td>
              <td>{{ cmd.error?.Valid ? cmd.error.String : '' }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="settingsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Targets are admin-managed only here -- kashi-pos pulls them down
       read-only via the branch-sync mechanism (GET /branch/sync/changes),
       it never creates or edits them. -->
  <v-dialog v-model="targetsDialog" max-width="640">
    <v-card>
      <v-card-title>Targets -- {{ targetsBranch?.name }}</v-card-title>
      <v-card-text>
        <v-alert v-if="targetsError" class="mb-4" type="error" variant="tonal">{{ targetsError }}</v-alert>
        <div v-if="targetsLoading" class="d-flex justify-center pa-4">
          <v-progress-circular color="primary" indeterminate />
        </div>
        <template v-else>
          <TargetProgressBars class="mb-4" :targets="targetsList" />
          <v-alert v-if="targetsList.length === 0" class="mb-4" type="info" variant="tonal">
            No targets set for this branch yet.
          </v-alert>

          <v-table class="mb-4" density="compact">
            <thead>
              <tr><th /><th>From</th><th>To</th><th>Target</th><th /></tr>
            </thead>
            <tbody>
              <tr v-for="target in targetsList" :key="target.id">
                <td><span class="target-color-swatch" :style="{ background: target.color }" /></td>
                <td>{{ new Date(target.date_from).toLocaleDateString() }}</td>
                <td>{{ new Date(target.date_to).toLocaleDateString() }}</td>
                <td>${{ (target.target_amount / 100).toLocaleString() }}</td>
                <td class="text-end">
                  <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="removeTarget(target)" />
                </td>
              </tr>
            </tbody>
          </v-table>

          <form class="field-grid" @submit.prevent="submitTarget">
            <v-text-field
              v-model="targetForm.dateFrom"
              density="compact"
              label="From"
              type="date"
              variant="outlined"
            />
            <v-text-field
              v-model="targetForm.dateTo"
              density="compact"
              label="To"
              type="date"
              variant="outlined"
            />
            <v-text-field
              v-model.number="targetForm.targetAmount"
              density="compact"
              label="Target Amount ($)"
              min="0"
              step="0.01"
              type="number"
              variant="outlined"
            />
            <div class="d-flex align-center ga-2">
              <v-checkbox
                v-model="targetForm.autoColor"
                density="compact"
                hide-details
                label="Auto-assign color"
              />
              <v-text-field
                v-if="!targetForm.autoColor"
                v-model="targetForm.color"
                density="compact"
                hide-details
                label="Color"
                type="color"
                variant="outlined"
              />
            </div>
            <div class="d-flex justify-end field-full">
              <v-btn color="primary" :loading="targetSaving" text="Add Target" type="submit" />
            </div>
          </form>
        </template>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="targetsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <!-- Salespersons are admin-managed only here -- kashi-pos pulls the
       roster down read-only via the branch-sync mechanism (GET
       /branch/sync/changes), it only edits the local attendance-device
       pairing id, never the name or active state. -->
  <v-dialog v-model="salespersonsDialog" max-width="560">
    <v-card>
      <v-card-title>Salespersons -- {{ salespersonsBranch?.name }}</v-card-title>
      <v-card-text>
        <v-alert v-if="salespersonsError" class="mb-4" type="error" variant="tonal">{{ salespersonsError }}</v-alert>
        <div v-if="salespersonsLoading" class="d-flex justify-center pa-4">
          <v-progress-circular color="primary" indeterminate />
        </div>
        <template v-else>
          <v-alert v-if="salespersonsList.length === 0" class="mb-4" type="info" variant="tonal">
            No salespersons for this branch yet.
          </v-alert>

          <v-table class="mb-4" density="compact">
            <thead>
              <tr><th>Name</th><th>Active</th><th /></tr>
            </thead>
            <tbody>
              <tr v-for="person in salespersonsList" :key="person.id">
                <td>
                  <v-text-field
                    v-if="editingSalespersonId === person.id"
                    v-model="editingSalespersonName"
                    density="compact"
                    hide-details
                    variant="outlined"
                    @keyup.enter="submitEditSalesperson(person)"
                  />
                  <template v-else>{{ person.name }}</template>
                </td>
                <td>
                  <v-switch
                    color="primary"
                    density="compact"
                    hide-details
                    :model-value="person.is_active"
                    @update:model-value="value => toggleSalespersonActive(person, value)"
                  />
                </td>
                <td class="text-end">
                  <v-icon-btn
                    v-if="editingSalespersonId === person.id"
                    icon="mdi-check"
                    size="small"
                    variant="text"
                    @click="submitEditSalesperson(person)"
                  />
                  <v-icon-btn
                    v-else
                    icon="mdi-pencil"
                    size="small"
                    variant="text"
                    @click="openEditSalesperson(person)"
                  />
                </td>
              </tr>
            </tbody>
          </v-table>

          <form class="d-flex ga-2 align-start" @submit.prevent="submitNewSalesperson">
            <v-text-field
              v-model="newSalespersonName"
              density="compact"
              label="Full name"
              variant="outlined"
              hide-details
            />
            <v-btn color="primary" :loading="salespersonSaving" text="Add" type="submit" />
          </form>
        </template>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="salespersonsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <div class="d-flex justify-space-between align-center mb-2">
    <h2 class="text-h6">Branches</h2>
    <v-btn color="primary" prepend-icon="mdi-plus" text="Add Branch" @click="openCreate" />
  </div>

  <v-alert v-if="listError" class="mb-4" type="error" variant="tonal">{{ listError }}</v-alert>

  <v-data-table :headers="headers" :items="branches" :loading="listLoading">
    <template #item.is_active="{ item }">
      <v-switch
        color="primary"
        density="compact"
        hide-details
        :model-value="item.is_active"
        @update:model-value="value => toggleActive(item, value)"
      />
    </template>
    <template #item.last_seen_at="{ item }">
      {{ item.last_seen_at?.Valid ? new Date(item.last_seen_at.Time).toLocaleString() : 'Never' }}
    </template>
    <template #item.actions="{ item }">
      <v-icon-btn icon="mdi-flag-outline" size="small" variant="text" @click="openTargets(item)" />
      <v-icon-btn icon="mdi-account-tie-outline" size="small" variant="text" @click="openSalespersons(item)" />
      <v-icon-btn icon="mdi-cog-outline" size="small" variant="text" @click="openSettings(item)" />
      <v-icon-btn icon="mdi-key" size="small" variant="text" @click="openRotate(item)" />
    </template>
  </v-data-table>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchSettings } from '@/composables/useBranchSettings'
  import { useBranchTargets } from '@/composables/useBranchTargets'
  import { useBranchSalespersons } from '@/composables/useBranchSalespersons'

  const { branches, fetchBranches, createBranch, setBranchActive, rotateBranchKey } = useBranches()
  const { getBranchSettings, updateBranchSettings, listBranchCommands } = useBranchSettings()
  const {
    listBranchTargets,
    createBranchTarget,
    deleteBranchTarget,
  } = useBranchTargets()
  const {
    listBranchSalespersons,
    createBranchSalesperson,
    updateBranchSalesperson,
    setBranchSalespersonActive,
  } = useBranchSalespersons()

  const headers = ref([
    { title: 'Name', key: 'name', align: 'start' },
    { title: 'Code', key: 'code', align: 'start' },
    { title: 'Active', key: 'is_active', align: 'center', sortable: false },
    { title: 'Last Seen', key: 'last_seen_at', align: 'start', sortable: false },
    { title: '', key: 'actions', align: 'end', sortable: false },
  ])

  const listLoading = ref(false)
  const listError = ref('')

  async function load () {
    listLoading.value = true
    listError.value = ''
    try {
      await fetchBranches(true)
    } catch (error) {
      listError.value = error.message
    } finally {
      listLoading.value = false
    }
  }

  onMounted(load)

  const createDialog = ref(false)
  const createName = ref('')
  const createCode = ref('')
  const creating = ref(false)
  const createError = ref('')

  function openCreate () {
    createName.value = ''
    createCode.value = ''
    createError.value = ''
    createDialog.value = true
  }

  const keyDialog = ref(false)
  const revealedKey = ref('')
  const copied = ref(false)

  async function submitCreate () {
    creating.value = true
    createError.value = ''
    try {
      const { api_key: apiKey } = await createBranch({ name: createName.value, code: createCode.value })
      createDialog.value = false
      revealedKey.value = apiKey
      copied.value = false
      keyDialog.value = true
    } catch (error) {
      createError.value = error.message
    } finally {
      creating.value = false
    }
  }

  async function copyKey () {
    try {
      await navigator.clipboard.writeText(revealedKey.value)
      copied.value = true
    } catch {
      copied.value = false
    }
  }

  async function toggleActive (item, value) {
    try {
      await setBranchActive(item.id, value)
    } catch (error) {
      listError.value = error.message
    }
  }

  const settingsDialog = ref(false)
  const settingsTarget = ref(null)
  const settingsData = ref(null)
  const settingsForm = ref(null)
  const settingsLoading = ref(false)
  const settingsError = ref('')
  const saving = ref(false)
  const saveOutcome = ref(null)
  const socialPlatformsText = ref('[]')
  const socialHandlesText = ref('{}')
  const commandHistory = ref([])

  function formFromSettings (data) {
    return {
      branch_name: data.branch_name,
      tax_rate: data.tax_rate,
      rounding_mode: data.rounding_mode,
      rounding_currency: data.rounding_currency,
      exchange_rate: data.exchange_rate,
      exchange_window_hours: data.exchange_window_hours,
      market_name: data.market_name,
      market_phone: data.market_phone,
      market_description: data.market_description,
      return_policy: data.return_policy,
      website: data.website,
      instagram: data.instagram,
    }
  }

  async function openSettings (item) {
    settingsTarget.value = item
    settingsDialog.value = true
    settingsLoading.value = true
    settingsError.value = ''
    saveOutcome.value = null
    settingsData.value = null
    settingsForm.value = null
    try {
      const data = await getBranchSettings(item.id)
      settingsData.value = data
      if (data) {
        settingsForm.value = formFromSettings(data)
        socialPlatformsText.value = JSON.stringify(data.social_platforms ?? [])
        socialHandlesText.value = JSON.stringify(data.social_handles ?? {})
      }
    } catch (error) {
      settingsError.value = error.message
    } finally {
      settingsLoading.value = false
    }
    loadCommandHistory()
  }

  async function loadCommandHistory () {
    if (!settingsTarget.value) return
    try {
      commandHistory.value = await listBranchCommands(settingsTarget.value.id)
    } catch (error) {
      settingsError.value = error.message
    }
  }

  async function saveSettings () {
    saving.value = true
    saveOutcome.value = null
    settingsError.value = ''
    try {
      let socialPlatforms
      let socialHandles
      try {
        socialPlatforms = JSON.parse(socialPlatformsText.value)
        socialHandles = JSON.parse(socialHandlesText.value)
      } catch {
        throw new Error('Social Platforms/Handles must be valid JSON')
      }

      await updateBranchSettings(settingsTarget.value.id, {
        ...settingsForm.value,
        social_platforms: socialPlatforms,
        social_handles: socialHandles,
      })
      saveOutcome.value = { ok: true, message: 'Command sent -- the branch will apply it within a few seconds if online.' }
      await loadCommandHistory()
    } catch (error) {
      saveOutcome.value = { ok: false, message: error.message }
    } finally {
      saving.value = false
    }
  }

  const targetsDialog = ref(false)
  const targetsBranch = ref(null)
  const targetsList = ref([])
  const targetsLoading = ref(false)
  const targetsError = ref('')
  const targetSaving = ref(false)
  const targetForm = ref({ dateFrom: '', dateTo: '', targetAmount: null, autoColor: true, color: '#4C6EF5' })

  async function openTargets (item) {
    targetsBranch.value = item
    targetsDialog.value = true
    targetsError.value = ''
    targetForm.value = { dateFrom: '', dateTo: '', targetAmount: null, autoColor: true, color: '#4C6EF5' }
    await loadTargets()
  }

  async function loadTargets () {
    targetsLoading.value = true
    targetsError.value = ''
    try {
      targetsList.value = await listBranchTargets(targetsBranch.value.id)
    } catch (error) {
      targetsError.value = error.message
    } finally {
      targetsLoading.value = false
    }
  }

  async function submitTarget () {
    targetSaving.value = true
    targetsError.value = ''
    try {
      await createBranchTarget(targetsBranch.value.id, {
        dateFrom: targetForm.value.dateFrom,
        dateTo: targetForm.value.dateTo,
        targetAmount: Math.round(Number(targetForm.value.targetAmount) * 100),
        color: targetForm.value.autoColor ? '' : targetForm.value.color,
      })
      targetForm.value = { dateFrom: '', dateTo: '', targetAmount: null, autoColor: true, color: '#4C6EF5' }
      await loadTargets()
    } catch (error) {
      targetsError.value = error.message
    } finally {
      targetSaving.value = false
    }
  }

  async function removeTarget (target) {
    try {
      await deleteBranchTarget(target.id)
      await loadTargets()
    } catch (error) {
      targetsError.value = error.message
    }
  }

  const salespersonsDialog = ref(false)
  const salespersonsBranch = ref(null)
  const salespersonsList = ref([])
  const salespersonsLoading = ref(false)
  const salespersonsError = ref('')
  const salespersonSaving = ref(false)
  const newSalespersonName = ref('')
  const editingSalespersonId = ref(null)
  const editingSalespersonName = ref('')

  async function openSalespersons (item) {
    salespersonsBranch.value = item
    salespersonsDialog.value = true
    salespersonsError.value = ''
    newSalespersonName.value = ''
    editingSalespersonId.value = null
    await loadSalespersons()
  }

  async function loadSalespersons () {
    salespersonsLoading.value = true
    salespersonsError.value = ''
    try {
      salespersonsList.value = await listBranchSalespersons(salespersonsBranch.value.id)
    } catch (error) {
      salespersonsError.value = error.message
    } finally {
      salespersonsLoading.value = false
    }
  }

  async function submitNewSalesperson () {
    if (!newSalespersonName.value.trim()) return
    salespersonSaving.value = true
    salespersonsError.value = ''
    try {
      await createBranchSalesperson(salespersonsBranch.value.id, newSalespersonName.value.trim())
      newSalespersonName.value = ''
      await loadSalespersons()
    } catch (error) {
      salespersonsError.value = error.message
    } finally {
      salespersonSaving.value = false
    }
  }

  function openEditSalesperson (person) {
    editingSalespersonId.value = person.id
    editingSalespersonName.value = person.name
  }

  async function submitEditSalesperson (person) {
    if (!editingSalespersonName.value.trim()) return
    salespersonsError.value = ''
    try {
      await updateBranchSalesperson(person.id, editingSalespersonName.value.trim())
      editingSalespersonId.value = null
      await loadSalespersons()
    } catch (error) {
      salespersonsError.value = error.message
    }
  }

  async function toggleSalespersonActive (person, value) {
    salespersonsError.value = ''
    try {
      await setBranchSalespersonActive(person.id, value)
      await loadSalespersons()
    } catch (error) {
      salespersonsError.value = error.message
    }
  }

  const rotateDialog = ref(false)
  const rotateTarget = ref(null)
  const rotating = ref(false)
  const rotateError = ref('')

  function openRotate (item) {
    rotateTarget.value = item
    rotateError.value = ''
    rotateDialog.value = true
  }

  async function confirmRotate () {
    rotating.value = true
    rotateError.value = ''
    try {
      const { api_key: apiKey } = await rotateBranchKey(rotateTarget.value.id)
      rotateDialog.value = false
      revealedKey.value = apiKey
      copied.value = false
      keyDialog.value = true
    } catch (error) {
      rotateError.value = error.message
    } finally {
      rotating.value = false
    }
  }
</script>

<style scoped>
.field-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 12px 16px;
  align-items: start;
}

.field-grid > .field-full {
  grid-column: 1 / -1;
}

.target-color-swatch {
  display: inline-block;
  width: 14px;
  height: 14px;
  border-radius: 4px;
  border: 1px solid rgba(var(--v-theme-on-surface), 0.2);
}
</style>
