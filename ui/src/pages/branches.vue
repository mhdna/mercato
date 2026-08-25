<template>
  <v-dialog v-model="createDialog" max-width="480">
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
  <v-dialog v-model="keyDialog" max-width="520">
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
       it never creates or edits them. Fixed-height + scrollable so the
       header/tabs/footer stay put while a long target list scrolls in
       the middle, instead of the whole dialog growing with content. -->
  <v-dialog v-model="targetsDialog" max-width="720" scrollable>
    <v-card class="targets-dialog-card">
      <v-card-title class="d-flex align-center">
        Targets -- {{ targetsBranch?.name }}
        <v-spacer />
        <v-icon-btn icon="mdi-close" size="small" variant="text" @click="targetsDialog = false" />
      </v-card-title>

      <v-tabs v-model="targetsTab" color="primary">
        <v-tab value="targets">Targets</v-tab>
        <v-tab value="recurring">Recurring</v-tab>
      </v-tabs>
      <v-divider />

      <v-card-text class="targets-dialog-body">
        <v-alert v-if="targetsError" class="mb-4" type="error" variant="tonal" closable @click:close="targetsError = ''">
          {{ targetsError }}
        </v-alert>
        <div v-if="targetsLoading" class="d-flex justify-center pa-8">
          <v-progress-circular color="primary" indeterminate />
        </div>

        <v-window v-else v-model="targetsTab">
          <v-window-item value="targets">
            <TargetProgressBars v-if="targetsList.length" class="mb-4" :targets="targetsList" />
            <div v-if="targetsList.length === 0" class="targets-empty">
              <v-icon icon="mdi-flag-outline" size="32" class="mb-2" />
              <div>No targets set for this branch yet.</div>
            </div>
            <v-list v-else class="target-list" lines="two">
              <v-list-item v-for="target in targetsList" :key="target.id" class="target-list-item">
                <template #prepend>
                  <span class="target-color-swatch" :style="{ background: target.color }" />
                </template>
                <v-list-item-title class="d-flex align-center ga-1">
                  ${{ (target.target_amount / 100).toLocaleString() }}
                  <v-icon
                    v-if="target.series_id?.Valid"
                    icon="mdi-repeat"
                    size="14"
                    title="Generated by a recurring target"
                  />
                </v-list-item-title>
                <v-list-item-subtitle>
                  {{ new Date(target.date_from).toLocaleDateString() }} -- {{ new Date(target.date_to).toLocaleDateString() }}
                </v-list-item-subtitle>
                <template #append>
                  <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEditTarget(target)" />
                  <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="removeTarget(target)" />
                </template>
              </v-list-item>
            </v-list>
          </v-window-item>

          <v-window-item value="recurring">
            <div v-if="!seriesList.length" class="targets-empty">
              <v-icon icon="mdi-repeat" size="32" class="mb-2" />
              <div>No recurring targets set up yet.</div>
              <div class="text-caption text-medium-emphasis mt-1">
                Each period (e.g. every month) creates its own target automatically, starting fresh at $0.
              </div>
            </div>
            <v-list v-else class="target-list" lines="two">
              <v-list-item v-for="s in seriesList" :key="s.id" class="target-list-item">
                <template #prepend>
                  <span class="target-color-swatch" :style="{ background: s.color }" />
                </template>
                <v-list-item-title>${{ (s.target_amount / 100).toLocaleString() }}</v-list-item-title>
                <v-list-item-subtitle>{{ describeSeries(s) }}</v-list-item-subtitle>
                <template #append>
                  <v-switch
                    class="me-1"
                    color="primary"
                    density="compact"
                    hide-details
                    :model-value="s.active"
                    title="Pause/resume -- keeps already-generated targets"
                    @update:model-value="value => toggleSeriesActive(s, value)"
                  />
                  <v-icon-btn icon="mdi-pencil" size="small" variant="text" @click="openEditSeries(s)" />
                  <v-icon-btn icon="mdi-delete" size="small" variant="text" @click="openDeleteSeriesConfirm(s)" />
                </template>
              </v-list-item>
            </v-list>
          </v-window-item>
        </v-window>
      </v-card-text>

      <v-divider />
      <v-card-actions>
        <v-btn
          v-if="targetsTab === 'targets'"
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Target"
          variant="tonal"
          @click="openAddTarget"
        />
        <v-btn
          v-else
          color="primary"
          prepend-icon="mdi-plus"
          text="Add Recurring Target"
          variant="tonal"
          @click="openAddSeries"
        />
        <v-spacer />
        <v-btn text="Close" @click="targetsDialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>

  <v-dialog v-model="targetFormDialog" max-width="440" :persistent="true">
    <v-card class="px-4">
      <v-card-title>{{ editingTargetId ? "Edit Target" : "Add Target" }}</v-card-title>
      <v-card-text>
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
            prefix="$"
            step="0.01"
            type="number"
            variant="outlined"
            class="field-full"
          />
          <div class="d-flex align-center ga-2 field-full">
            <v-checkbox
              v-model="targetForm.autoColor"
              density="compact"
              hide-details
              :label="editingTargetId ? 'Keep current color' : 'Auto-assign color'"
            />
            <v-text-field
              v-if="!targetForm.autoColor"
              v-model="targetForm.color"
              density="compact"
              hide-details
              label="Color"
              style="max-width: 100px"
              type="color"
              variant="outlined"
            />
          </div>

          <v-alert v-if="targetFormError" class="mb-2 mt-2 field-full" type="error" variant="tonal">
            {{ targetFormError }}
          </v-alert>

          <v-card-actions class="px-0 field-full">
            <v-spacer />
            <v-btn text="Cancel" variant="text" @click="targetFormDialog = false" />
            <v-btn
              color="primary"
              :loading="targetSaving"
              :text="editingTargetId ? 'Save' : 'Add'"
              type="submit"
            />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="seriesFormDialog" max-width="440" :persistent="true">
    <v-card class="px-4">
      <v-card-title>{{ editingSeriesId ? "Edit Recurring Target" : "Add Recurring Target" }}</v-card-title>
      <v-card-text>
        <form class="field-grid" @submit.prevent="submitSeries">
          <v-text-field
            v-model.number="seriesForm.targetAmount"
            density="compact"
            label="Target Amount ($)"
            min="0"
            prefix="$"
            step="0.01"
            type="number"
            variant="outlined"
            class="field-full"
          />
          <v-text-field
            v-model.number="seriesForm.startDay"
            density="compact"
            hint="Day of month a period starts on -- clamped in short months"
            label="Starts on day"
            max="31"
            min="1"
            persistent-hint
            type="number"
            variant="outlined"
          />
          <v-text-field
            v-model.number="seriesForm.intervalCount"
            density="compact"
            hint="1 = every month, 3 = quarterly, etc."
            label="Repeat every N month(s)"
            min="1"
            persistent-hint
            type="number"
            variant="outlined"
          />
          <div class="d-flex align-center ga-2 field-full">
            <v-checkbox
              v-model="seriesForm.autoColor"
              density="compact"
              hide-details
              :label="editingSeriesId ? 'Keep current color' : 'Auto-assign color'"
            />
            <v-text-field
              v-if="!seriesForm.autoColor"
              v-model="seriesForm.color"
              density="compact"
              hide-details
              label="Color"
              style="max-width: 100px"
              type="color"
              variant="outlined"
            />
          </div>

          <div v-if="editingSeriesId" class="text-caption text-medium-emphasis field-full">
            Changes only apply to periods generated from now on -- the current period keeps its
            already-set amount and color.
          </div>

          <v-alert v-if="seriesFormError" class="mb-2 mt-2 field-full" type="error" variant="tonal">
            {{ seriesFormError }}
          </v-alert>

          <v-card-actions class="px-0 field-full">
            <v-spacer />
            <v-btn text="Cancel" variant="text" @click="seriesFormDialog = false" />
            <v-btn
              color="primary"
              :loading="seriesSaving"
              :text="editingSeriesId ? 'Save' : 'Add'"
              type="submit"
            />
          </v-card-actions>
        </form>
      </v-card-text>
    </v-card>
  </v-dialog>

  <v-dialog v-model="deleteSeriesDialog" max-width="420">
    <v-card>
      <v-card-title>Delete Recurring Target</v-card-title>
      <v-card-text>
        This removes the recurring schedule and every target it has generated (past and
        current), for every branch install syncing it. This can't be undone.
      </v-card-text>
      <v-alert v-if="deleteSeriesError" class="mx-4 mb-2" type="error" variant="tonal">{{ deleteSeriesError }}</v-alert>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Cancel" @click="deleteSeriesDialog = false" />
        <v-btn color="error" :loading="deleteSeriesSaving" text="Delete" @click="confirmDeleteSeries" />
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
    updateBranchTarget,
    deleteBranchTarget,
    listBranchTargetSeries,
    createBranchTargetSeries,
    updateBranchTargetSeries,
    deleteBranchTargetSeries,
    setBranchTargetSeriesActive,
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
  const targetsTab = ref('targets')
  const targetsBranch = ref(null)
  const targetsList = ref([])
  const targetsLoading = ref(false)
  const targetsError = ref('')

  async function openTargets (item) {
    targetsBranch.value = item
    targetsTab.value = 'targets'
    targetsDialog.value = true
    targetsError.value = ''
    await Promise.all([loadTargets(), loadSeries()])
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

  const targetFormDialog = ref(false)
  const targetFormError = ref('')
  const targetSaving = ref(false)
  const targetForm = ref({ dateFrom: '', dateTo: '', targetAmount: null, autoColor: true, color: '#4C6EF5' })
  const editingTargetId = ref(null)

  function openAddTarget () {
    editingTargetId.value = null
    targetForm.value = { dateFrom: '', dateTo: '', targetAmount: null, autoColor: true, color: '#4C6EF5' }
    targetFormError.value = ''
    targetFormDialog.value = true
  }

  function openEditTarget (target) {
    editingTargetId.value = target.id
    targetForm.value = {
      dateFrom: target.date_from.slice(0, 10),
      dateTo: target.date_to.slice(0, 10),
      targetAmount: target.target_amount / 100,
      autoColor: true,
      color: target.color || '#4C6EF5',
    }
    targetFormError.value = ''
    targetFormDialog.value = true
  }

  async function submitTarget () {
    targetSaving.value = true
    targetFormError.value = ''
    try {
      const payload = {
        dateFrom: targetForm.value.dateFrom,
        dateTo: targetForm.value.dateTo,
        targetAmount: Math.round(Number(targetForm.value.targetAmount) * 100),
        color: targetForm.value.autoColor ? '' : targetForm.value.color,
      }
      if (editingTargetId.value) {
        await updateBranchTarget(editingTargetId.value, payload)
      } else {
        await createBranchTarget(targetsBranch.value.id, payload)
      }
      targetFormDialog.value = false
      await loadTargets()
    } catch (error) {
      targetFormError.value = error.message
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

  const seriesList = ref([])

  async function loadSeries () {
    try {
      seriesList.value = await listBranchTargetSeries(targetsBranch.value.id)
    } catch (error) {
      targetsError.value = error.message
    }
  }

  function describeSeries (s) {
    const every = s.interval_count === 1 ? 'month' : `${s.interval_count} months`
    return `Every ${every}, starting day ${s.start_day}`
  }

  const seriesFormDialog = ref(false)
  const seriesFormError = ref('')
  const seriesSaving = ref(false)
  const seriesForm = ref({ targetAmount: null, startDay: 1, intervalCount: 1, autoColor: true, color: '#4C6EF5' })
  const editingSeriesId = ref(null)

  function openAddSeries () {
    editingSeriesId.value = null
    seriesForm.value = { targetAmount: null, startDay: 1, intervalCount: 1, autoColor: true, color: '#4C6EF5' }
    seriesFormError.value = ''
    seriesFormDialog.value = true
  }

  function openEditSeries (s) {
    editingSeriesId.value = s.id
    seriesForm.value = {
      targetAmount: s.target_amount / 100,
      startDay: s.start_day,
      intervalCount: s.interval_count,
      autoColor: true,
      color: s.color || '#4C6EF5',
    }
    seriesFormError.value = ''
    seriesFormDialog.value = true
  }

  async function submitSeries () {
    seriesSaving.value = true
    seriesFormError.value = ''
    try {
      const payload = {
        targetAmount: Math.round(Number(seriesForm.value.targetAmount) * 100),
        startDay: seriesForm.value.startDay,
        intervalCount: seriesForm.value.intervalCount,
        color: seriesForm.value.autoColor ? '' : seriesForm.value.color,
      }
      if (editingSeriesId.value) {
        await updateBranchTargetSeries(editingSeriesId.value, payload)
      } else {
        await createBranchTargetSeries(targetsBranch.value.id, payload)
      }
      seriesFormDialog.value = false
      await Promise.all([loadSeries(), loadTargets()])
    } catch (error) {
      seriesFormError.value = error.message
    } finally {
      seriesSaving.value = false
    }
  }

  async function toggleSeriesActive (s, value) {
    try {
      await setBranchTargetSeriesActive(s.id, value)
      await loadSeries()
    } catch (error) {
      targetsError.value = error.message
    }
  }

  const deleteSeriesDialog = ref(false)
  const deleteSeriesError = ref('')
  const deleteSeriesSaving = ref(false)
  const deleteSeriesTarget = ref(null)

  function openDeleteSeriesConfirm (s) {
    deleteSeriesTarget.value = s
    deleteSeriesError.value = ''
    deleteSeriesDialog.value = true
  }

  async function confirmDeleteSeries () {
    deleteSeriesSaving.value = true
    deleteSeriesError.value = ''
    try {
      await deleteBranchTargetSeries(deleteSeriesTarget.value.id)
      deleteSeriesDialog.value = false
      await Promise.all([loadSeries(), loadTargets()])
    } catch (error) {
      deleteSeriesError.value = error.message
    } finally {
      deleteSeriesSaving.value = false
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
  flex-shrink: 0;
}

.targets-dialog-card {
  height: 640px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.targets-dialog-body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
}

.targets-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 16px;
  color: rgba(var(--v-theme-on-surface), 0.6);
  text-align: center;
}

.target-list {
  padding: 0;
}

.target-list-item {
  border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
  border-radius: 8px;
  margin-bottom: 8px;
}

.target-list-item :deep(.v-list-item__prepend) {
  margin-right: 12px;
  align-self: center;
}
</style>
