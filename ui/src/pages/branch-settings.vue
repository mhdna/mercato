<template>
  <div class="page-root">
    <v-alert
      v-if="listError"
      class="mb-4"
      density="compact"
      type="error"
      variant="tonal"
    >
      {{ listError }}
    </v-alert>

    <v-card class="bs-card" flat>
      <v-card-title class="page-heading d-flex flex-wrap align-center ga-3 px-4 py-3">
        <v-icon icon="mdi-store-cog-outline" />
        <span>Branch Settings</span>
        <v-spacer />
        <v-btn
          :disabled="!activeBranch"
          prepend-icon="mdi-refresh"
          text="Reload"
          variant="text"
          @click="reloadActive"
        />
      </v-card-title>
      <v-divider />

      <div class="bs-layout">
        <PageSidebar
          v-model="selectedId"
          empty-text="No branches yet."
          :items="branchItems"
          :loading="listLoading"
        >
          <template #prepend="{ item }">
            <span class="bs-dot me-2" :class="onlineClass(item)" />
          </template>
        </PageSidebar>

        <v-divider vertical />

        <div class="bs-content">
          <!-- ------------------------------------------------- no selection -->
          <div v-if="!activeBranch" class="bs-placeholder">
            <v-icon icon="mdi-store-cog-outline" size="40" />
            <div>Pick a branch to view its settings.</div>
          </div>

          <!-- --------------------------------------------------- a branch -->
          <div v-else class="bs-detail-root pa-5">
            <div class="d-flex align-center ga-2 mb-1">
              <h3 class="text-subtitle-1">{{ activeBranch.name }}</h3>
              <v-chip size="x-small" variant="tonal">{{ activeBranch.code }}</v-chip>
              <span class="text-caption text-medium-emphasis">
                Last seen {{ lastSeenLabel(activeBranch) }}
              </span>
            </div>
            <p class="text-body-2 text-medium-emphasis mb-4">
              Saving enqueues a remote command &mdash; the branch applies it itself within a few
              seconds if it's online, or on its next sync otherwise. Nothing here writes to the
              branch database directly.
            </p>

            <v-alert
              v-if="settingsError"
              class="mb-4"
              density="compact"
              type="error"
              variant="tonal"
            >
              {{ settingsError }}
            </v-alert>
            <v-alert
              v-if="saveOutcome"
              class="mb-4"
              density="compact"
              :type="saveOutcome.ok ? 'success' : 'error'"
              variant="tonal"
            >
              {{ saveOutcome.message }}
            </v-alert>

            <div v-if="settingsLoading" class="d-flex justify-center pa-8">
              <v-progress-circular color="primary" indeterminate />
            </div>

            <v-alert
              v-else-if="!settingsForm"
              type="info"
              variant="tonal"
            >
              This branch hasn't pushed its settings yet &mdash; it needs to sync at least once
              before you can view or edit them.
            </v-alert>

            <div v-else class="bs-detail">
              <v-tabs
                v-model="sectionTab"
                class="bs-section-tabs"
                color="primary"
                direction="vertical"
              >
                <v-tab class="text-none" text="General" value="general" />
                <v-tab class="text-none" text="Behaviour" value="behaviour" />
                <v-tab class="text-none" text="Page access" value="pages" />
                <v-tab class="text-none" text="Printing" value="printing" />
                <v-tab class="text-none" text="Attendance" value="attendance" />
                <v-tab class="text-none" text="Users" value="users" />
                <v-tab class="text-none" text="Commands" value="commands" />
              </v-tabs>

              <v-divider vertical />

              <div class="bs-section-pane">
                <form @submit.prevent="saveSettings">
                  <div v-show="sectionTab === 'general'">
                    <section class="bs-section">
                      <div class="bs-section-title">General</div>
                      <div class="field-grid">
                        <v-text-field
                          v-model="settingsForm.branch_name"
                          density="compact"
                          hide-details="auto"
                          label="Branch Name"
                          variant="outlined"
                        />
                      </div>
                    </section>

                    <section class="bs-section">
                      <div class="bs-section-title">Pricing &amp; Exchange</div>
                      <div class="field-grid">
                        <v-text-field
                          v-model.number="settingsForm.tax_rate"
                          density="compact"
                          hide-details="auto"
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
                          hide-details="auto"
                          :items="['half_up', 'up', 'down']"
                          label="Rounding Mode"
                          variant="outlined"
                        />
                        <CurrencySelect
                          v-model="settingsForm.rounding_currency"
                          label="Rounding Currency"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model.number="settingsForm.exchange_rate"
                          density="compact"
                          hide-details="auto"
                          label="Exchange Rate"
                          type="number"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model.number="settingsForm.exchange_window_hours"
                          density="compact"
                          hide-details="auto"
                          label="Exchange Window (hours)"
                          type="number"
                          variant="outlined"
                        />
                      </div>
                    </section>

                    <section class="bs-section">
                      <div class="bs-section-title">Storefront</div>
                      <div class="field-grid">
                        <v-text-field
                          v-model="settingsForm.market_name"
                          density="compact"
                          hide-details="auto"
                          label="Market Name"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="settingsForm.market_phone"
                          density="compact"
                          hide-details="auto"
                          label="Market Phone"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="settingsForm.website"
                          density="compact"
                          hide-details="auto"
                          label="Website"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="settingsForm.instagram"
                          density="compact"
                          hide-details="auto"
                          label="Instagram"
                          variant="outlined"
                        />
                        <v-textarea
                          v-model="settingsForm.market_description"
                          class="field-full"
                          density="compact"
                          hide-details="auto"
                          label="Market Description"
                          rows="2"
                          variant="outlined"
                        />
                        <v-textarea
                          v-model="settingsForm.return_policy"
                          class="field-full"
                          density="compact"
                          hide-details="auto"
                          label="Return Policy"
                          rows="2"
                          variant="outlined"
                        />
                      </div>
                    </section>

                    <section class="bs-section">
                      <div class="bs-section-title">Social</div>
                      <div class="field-grid">
                        <v-textarea
                          v-model="socialPlatformsText"
                          class="field-full"
                          density="compact"
                          hint="JSON array, e.g. [&quot;facebook&quot;, &quot;instagram&quot;]"
                          label="Social Platforms"
                          persistent-hint
                          rows="2"
                          variant="outlined"
                        />
                        <v-textarea
                          v-model="socialHandlesText"
                          class="field-full"
                          density="compact"
                          hint="JSON object, e.g. {&quot;instagram&quot;: &quot;@handle&quot;}"
                          label="Social Handles"
                          persistent-hint
                          rows="2"
                          variant="outlined"
                        />
                      </div>
                    </section>

                  </div>

                  <div v-show="sectionTab === 'behaviour'">
                    <section class="bs-section">
                      <div class="bs-section-title">Behaviour</div>
                      <div class="bs-switch-grid">
                        <v-switch
                          v-model="settingsForm.search_button_enabled"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Product search button on sale screen"
                        />
                        <v-switch
                          v-model="settingsForm.custom_item_discounts_enabled"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Custom per-line discounts"
                        />
                        <v-switch
                          v-model="settingsForm.custom_item_prices_enabled"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Custom item prices"
                        />
                        <v-switch
                          v-model="settingsForm.per_unit_item_prices_enabled"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Per-unit item pricing"
                        />
                        <v-switch
                          v-model="settingsForm.price_change_manual_override_mode"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Price-change manual override mode"
                        />
                        <v-switch
                          v-model="settingsForm.invoice_keyboard_mode"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Invoice keyboard (number-pad) mode"
                        />
                        <v-switch
                          v-model="settingsForm.client_required"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Require a client on every invoice"
                        />
                      </div>
                    </section>

                  </div>

                  <div v-show="sectionTab === 'pages'">
                    <section class="bs-section">
                      <div class="bs-section-title">Page access (non-admin users)</div>
                      <div class="bs-switch-grid">
                        <v-switch
                          v-model="settingsForm.page_unlock_clients"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Clients"
                        />
                        <v-switch
                          v-model="settingsForm.page_unlock_inventory"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Inventory"
                        />
                        <v-switch
                          v-model="settingsForm.page_unlock_transfers"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Transfers"
                        />
                        <v-switch
                          v-model="settingsForm.page_unlock_attendance"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Attendance"
                        />
                        <v-switch
                          v-model="settingsForm.page_unlock_salespersons"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Salespersons"
                        />
                      </div>
                    </section>

                  </div>

                  <div v-show="sectionTab === 'printing'">
                    <section class="bs-section">
                      <div class="bs-section-title">Printing</div>
                      <div class="field-grid">
                        <v-select
                          v-model="settingsForm.printer_size"
                          density="compact"
                          hide-details="auto"
                          :items="['58mm', '80mm']"
                          label="Printer Size"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model.number="settingsForm.receipt_width"
                          density="compact"
                          hide-details="auto"
                          label="Receipt Width (mm)"
                          type="number"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model.number="settingsForm.receipt_height"
                          density="compact"
                          hide-details="auto"
                          label="Receipt Height (mm)"
                          type="number"
                          variant="outlined"
                        />
                        <v-select
                          v-model="settingsForm.receipt_font"
                          density="compact"
                          hide-details="auto"
                          :items="['Belleza', 'Helvetica', 'Times', 'Courier']"
                          label="Title Font"
                          variant="outlined"
                        />
                        <v-select
                          v-model="settingsForm.receipt_body_font"
                          density="compact"
                          hide-details="auto"
                          :items="['Printer Font A', 'Courier', 'Helvetica', 'Times', 'Belleza']"
                          label="Body Font"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model.number="settingsForm.receipt_title_size"
                          density="compact"
                          hide-details="auto"
                          label="Title Size"
                          type="number"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model.number="settingsForm.receipt_body_size"
                          density="compact"
                          hide-details="auto"
                          label="Body Size"
                          type="number"
                          variant="outlined"
                        />
                      </div>
                      <div class="bs-switch-grid mt-3">
                        <v-switch
                          v-model="settingsForm.receipt_enabled"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Print receipt after checkout"
                        />
                        <v-switch
                          v-model="settingsForm.receipt_cutoff"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Cut paper after printing"
                        />
                      </div>
                    </section>

                    <section class="bs-section">
                      <div class="bs-group-head">
                        <div class="bs-section-title mb-0">Device identifiers</div>
                        <v-btn-toggle
                          v-model="deviceIdsSource"
                          color="primary"
                          density="compact"
                          divided
                          mandatory
                          @update:model-value="v => setGroupSource('device_ids', v)"
                        >
                          <v-btn size="small" value="central">Centrally</v-btn>
                          <v-btn size="small" value="local">At the branch</v-btn>
                        </v-btn-toggle>
                      </div>
                      <p class="text-caption text-medium-emphasis mb-3">
                        Windows printer name and customer-screen COM port are specific to the physical
                        till. Leave these managed at the branch unless it runs a single machine.
                      </p>
                      <div class="field-grid">
                        <v-text-field
                          v-model="settingsForm.printer_id"
                          density="compact"
                          :disabled="isLocal('device_ids')"
                          hide-details="auto"
                          label="Printer ID / name"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="settingsForm.receipt_printer"
                          density="compact"
                          :disabled="isLocal('device_ids')"
                          hide-details="auto"
                          label="Receipt printer"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="settingsForm.screen_port"
                          density="compact"
                          :disabled="isLocal('device_ids')"
                          hide-details="auto"
                          label="Customer screen COM port"
                          variant="outlined"
                        />
                      </div>
                    </section>

                  </div>

                  <div v-show="sectionTab === 'attendance'">
                    <section class="bs-section">
                      <div class="bs-section-title">Attendance</div>
                      <div class="field-grid">
                        <v-text-field
                          v-model="settingsForm.akuvox_ip"
                          density="compact"
                          hide-details="auto"
                          label="Akuvox IP"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="settingsForm.akuvox_username"
                          density="compact"
                          hide-details="auto"
                          label="Akuvox Username"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model.number="settingsForm.attendance_duplicate_interval_seconds"
                          density="compact"
                          hide-details="auto"
                          label="Duplicate interval (sec)"
                          min="0"
                          type="number"
                          variant="outlined"
                        />
                      </div>
                      <div class="bs-switch-grid mt-3">
                        <v-switch
                          v-model="settingsForm.attendance_enabled"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Attendance enabled"
                        />
                        <v-switch
                          v-model="settingsForm.attendance_cashier_history"
                          color="primary"
                          density="compact"
                          hide-details
                          label="Cashiers can browse previous days"
                        />
                      </div>

                      <div class="bs-group-head mt-4">
                        <div class="bs-section-title mb-0">Secrets</div>
                        <v-btn-toggle
                          v-model="secretsSource"
                          color="primary"
                          density="compact"
                          divided
                          mandatory
                          @update:model-value="v => setGroupSource('secrets', v)"
                        >
                          <v-btn size="small" value="central">Centrally</v-btn>
                          <v-btn size="small" value="local">At the branch</v-btn>
                        </v-btn-toggle>
                      </div>
                      <p class="text-caption text-medium-emphasis mb-3">
                        Write-only &mdash; kashi never displays these back. Leave blank to keep the
                        branch's current value; type a new one to push it.
                      </p>
                      <div class="field-grid">
                        <v-text-field
                          v-model="secretForm.akuvox_password"
                          autocomplete="new-password"
                          density="compact"
                          :disabled="isLocal('secrets')"
                          hide-details="auto"
                          label="Akuvox device password"
                          type="password"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="secretForm.attendance_approver_password"
                          autocomplete="new-password"
                          density="compact"
                          :disabled="isLocal('secrets')"
                          hide-details="auto"
                          label="Attendance manager password"
                          type="password"
                          variant="outlined"
                        />
                        <v-text-field
                          v-model="secretForm.admin_password"
                          autocomplete="new-password"
                          density="compact"
                          :disabled="isLocal('secrets')"
                          hint="Resets the branch's POS &quot;admin&quot; login"
                          label="POS admin password"
                          persistent-hint
                          type="password"
                          variant="outlined"
                        />
                      </div>
                    </section>
                  </div>

                  <div
                    v-show="sectionTab !== 'commands' && sectionTab !== 'users'"
                    class="bs-savebar d-flex justify-end"
                  >
                    <v-btn
                      color="primary"
                      :loading="saving"
                      text="Save"
                      type="submit"
                      variant="flat"
                    />
                  </div>
                </form>

                <div v-show="sectionTab === 'users'">
                  <div class="d-flex align-center justify-space-between mb-2">
                    <div class="bs-section-title mb-0">POS Users</div>
                    <v-btn
                      color="primary"
                      prepend-icon="mdi-plus"
                      size="small"
                      text="Add user"
                      variant="tonal"
                      @click="openUserDialog()"
                    />
                  </div>
                  <p class="text-body-2 text-medium-emphasis mb-3">
                    Roster is managed here and synced down to the till read-only. Adding a user or
                    changing a PIN sends a command the branch applies on its next sync.
                  </p>
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
                        <th>Username</th>
                        <th>Role</th>
                        <th>Akuvox ID</th>
                        <th>Active</th>
                        <th class="text-end" />
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="u in branchUsers" :key="u.id">
                        <td>{{ u.username }}</td>
                        <td>{{ u.role }}</td>
                        <td>{{ u.akuvox_user_id || '—' }}</td>
                        <td>
                          <v-chip :color="u.is_active ? 'success' : 'grey'" size="x-small" variant="tonal">
                            {{ u.is_active ? 'Active' : 'Inactive' }}
                          </v-chip>
                        </td>
                        <td class="text-end">
                          <v-icon-btn
                            icon="mdi-key"
                            size="small"
                            title="Set PIN"
                            variant="text"
                            @click="openPinDialog(u)"
                          />
                          <v-icon-btn
                            icon="mdi-pencil"
                            size="small"
                            variant="text"
                            @click="openUserDialog(u)"
                          />
                          <v-icon-btn
                            color="error"
                            icon="mdi-delete"
                            size="small"
                            variant="text"
                            @click="removeUser(u)"
                          />
                        </td>
                      </tr>
                      <tr v-if="branchUsers.length === 0">
                        <td class="text-medium-emphasis py-4" colspan="5">No POS users yet.</td>
                      </tr>
                    </tbody>
                  </v-table>
                </div>

                <div v-show="sectionTab === 'commands'">
                  <div class="d-flex align-center justify-space-between mb-2">
                    <div class="bs-section-title mb-0">Recent Commands</div>
                    <v-icon-btn
                      icon="mdi-refresh"
                      size="small"
                      variant="text"
                      @click="loadCommandHistory"
                    />
                  </div>
                  <v-table density="compact">
                    <thead>
                      <tr>
                        <th>Type</th>
                        <th>Status</th>
                        <th>Issued</th>
                        <th>Error</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="cmd in commandHistory" :key="cmd.id">
                        <td>{{ cmd.type }}</td>
                        <td>
                          <v-chip
                            :color="statusColor(cmd.status)"
                            size="small"
                          >
                            {{ cmd.status }}
                          </v-chip>
                        </td>
                        <td>{{ new Date(cmd.created_at).toLocaleString() }}</td>
                        <td>{{ cmd.error?.Valid ? cmd.error.String : '' }}</td>
                      </tr>
                      <tr v-if="commandHistory.length === 0">
                        <td class="text-medium-emphasis py-4" colspan="4">No commands issued yet.</td>
                      </tr>
                    </tbody>
                  </v-table>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </v-card>

    <!-- add / edit POS user -->
    <v-dialog v-model="userDialog" max-width="440">
      <v-card>
        <v-card-title class="text-subtitle-1">{{ editingUser ? 'Edit user' : 'Add user' }}</v-card-title>
        <v-card-text>
          <v-alert
            v-if="userFormError"
            class="mb-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ userFormError }}
          </v-alert>
          <v-text-field
            v-model="userForm.username"
            density="compact"
            label="Username"
            variant="outlined"
          />
          <v-select
            v-model="userForm.role"
            class="mt-3"
            density="compact"
            :items="['cashier', 'admin']"
            label="Role"
            variant="outlined"
          />
          <v-text-field
            v-model="userForm.akuvox_user_id"
            class="mt-3"
            density="compact"
            hide-details="auto"
            label="Akuvox UserID (optional)"
            variant="outlined"
          />
          <v-text-field
            v-if="!editingUser"
            v-model="userForm.pin"
            class="mt-3"
            density="compact"
            hint="4 to 6 digits"
            inputmode="numeric"
            label="Initial PIN"
            maxlength="6"
            persistent-hint
            variant="outlined"
          />
          <v-switch
            v-model="userForm.is_active"
            color="primary"
            hide-details
            label="Active"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="userDialog = false" />
          <v-btn
            color="primary"
            :loading="userSaving"
            text="Save"
            variant="flat"
            @click="submitUser"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- set PIN -->
    <v-dialog v-model="pinDialog" max-width="380">
      <v-card>
        <v-card-title class="text-subtitle-1">Set PIN — {{ pinTarget?.username }}</v-card-title>
        <v-card-text>
          <v-alert
            v-if="pinError"
            class="mb-3"
            density="compact"
            type="error"
            variant="tonal"
          >
            {{ pinError }}
          </v-alert>
          <v-text-field
            v-model="pinValue"
            autofocus
            density="compact"
            hint="4 to 6 digits — the branch applies it on its next sync"
            inputmode="numeric"
            label="New PIN"
            maxlength="6"
            persistent-hint
            variant="outlined"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text="Cancel" variant="text" @click="pinDialog = false" />
          <v-btn
            color="primary"
            :loading="pinSaving"
            text="Queue PIN change"
            variant="flat"
            @click="submitPin"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import PageSidebar from '@/components/PageSidebar.vue'
  import { useBranches } from '@/composables/useBranches'
  import { useBranchSettings } from '@/composables/useBranchSettings'
  import { useBranchUsers } from '@/composables/useBranchUsers'

  const route = useRoute()
  const router = useRouter()

  const { branches, fetchBranches } = useBranches()
  const {
    getBranchSettings,
    updateBranchSettings,
    listBranchCommands,
    setBranchSettingsManagedLocally,
  } = useBranchSettings()

  const listLoading = ref(false)
  const listError = ref('')
  const selectedId = ref(null)
  // Which settings section (vertical sub-tab) is showing for the active branch.
  const sectionTab = ref('general')

  const activeBranch = computed(() => branches.value.find(b => b.id === selectedId.value) ?? null)

  const branchItems = computed(() => branches.value.map(b => ({
    value: b.id,
    title: b.name,
    subtitle: b.code,
    raw: b,
  })))

  function onlineClass (branch) {
    const t = branch.last_seen_at?.Valid ? new Date(branch.last_seen_at.Time).getTime() : 0
    return Date.now() - t < 2 * 60 * 1000 ? 'is-online' : 'is-offline'
  }

  function lastSeenLabel (branch) {
    return branch.last_seen_at?.Valid
      ? new Date(branch.last_seen_at.Time).toLocaleString()
      : 'never'
  }

  function statusColor (status) {
    if (status === 'success') return 'success'
    if (status === 'failed') return 'error'
    return 'warning'
  }

  onMounted(async () => {
    listLoading.value = true
    listError.value = ''
    try {
      await fetchBranches(true)
      const wanted = Number(route.query.branch)
      selectedId.value = branches.value.find(b => b.id === wanted)?.id
        ?? branches.value[0]?.id
        ?? null
    } catch (error) {
      listError.value = error.message
    } finally {
      listLoading.value = false
    }
  })

  // ---- per-branch settings ------------------------------------------------
  const settingsForm = ref(null)
  const secretForm = ref({ akuvox_password: '', attendance_approver_password: '', admin_password: '' })
  const settingsLoading = ref(false)
  const settingsError = ref('')
  const saving = ref(false)
  const saveOutcome = ref(null)
  const socialPlatformsText = ref('[]')
  const socialHandlesText = ref('{}')
  const commandHistory = ref([])

  // Which settings groups the branch owns locally. Persisted centrally the
  // moment a toggle changes -- it's admin state, not a branch command.
  const managedLocally = ref([])
  const isLocal = group => managedLocally.value.includes(group)
  const deviceIdsSource = computed(() => (isLocal('device_ids') ? 'local' : 'central'))
  const secretsSource = computed(() => (isLocal('secrets') ? 'local' : 'central'))

  // Keys that make up each toggleable group, for stripping the command payload.
  const GROUP_KEYS = {
    device_ids: ['printer_id', 'receipt_printer', 'screen_port'],
    secrets: ['akuvox_password', 'attendance_approver_password', 'admin_password'],
  }

  const BOOL_KEYS = [
    'search_button_enabled',
    'custom_item_discounts_enabled',
    'custom_item_prices_enabled',
    'per_unit_item_prices_enabled',
    'price_change_manual_override_mode',
    'invoice_keyboard_mode',
    'client_required',
    'page_unlock_clients',
    'page_unlock_inventory',
    'page_unlock_transfers',
    'page_unlock_attendance',
    'page_unlock_salespersons',
    'receipt_enabled',
    'receipt_cutoff',
    'attendance_enabled',
    'attendance_cashier_history',
  ]

  function formFromSettings (d) {
    const form = {
      branch_name: d.branch_name,
      tax_rate: d.tax_rate,
      rounding_mode: d.rounding_mode,
      rounding_currency: d.rounding_currency,
      exchange_rate: d.exchange_rate,
      exchange_window_hours: d.exchange_window_hours,
      market_name: d.market_name,
      market_phone: d.market_phone,
      market_description: d.market_description,
      return_policy: d.return_policy,
      website: d.website,
      instagram: d.instagram,
      // Absent on branches that haven't pushed since this field was added --
      // treat only an explicit false as "off".
      search_button_enabled: d.search_button_enabled !== false,
      printer_size: d.printer_size ?? '',
      receipt_width: d.receipt_width ?? 0,
      receipt_height: d.receipt_height ?? 0,
      receipt_font: d.receipt_font ?? '',
      receipt_body_font: d.receipt_body_font ?? '',
      receipt_title_size: d.receipt_title_size ?? 0,
      receipt_body_size: d.receipt_body_size ?? 0,
      printer_id: d.printer_id ?? '',
      receipt_printer: d.receipt_printer ?? '',
      screen_port: d.screen_port ?? '',
      akuvox_ip: d.akuvox_ip ?? '',
      akuvox_username: d.akuvox_username ?? '',
      attendance_duplicate_interval_seconds: d.attendance_duplicate_interval_seconds ?? 0,
    }
    for (const k of BOOL_KEYS) {
      if (k === 'search_button_enabled') continue
      form[k] = d[k] === true
    }
    return form
  }

  watch(selectedId, id => {
    if (id == null) return
    // Keep the URL shareable / refresh-stable without stacking history entries.
    router.replace({ query: { ...route.query, branch: String(id) } })
    loadSettings(id)
  })

  async function loadSettings (id) {
    settingsLoading.value = true
    settingsError.value = ''
    saveOutcome.value = null
    sectionTab.value = 'general'
    settingsForm.value = null
    secretForm.value = { akuvox_password: '', attendance_approver_password: '', admin_password: '' }
    commandHistory.value = []
    try {
      const data = await getBranchSettings(id)
      if (data) {
        settingsForm.value = formFromSettings(data)
        managedLocally.value = Array.isArray(data.managed_locally) ? [...data.managed_locally] : []
        socialPlatformsText.value = JSON.stringify(data.social_platforms ?? [])
        socialHandlesText.value = JSON.stringify(data.social_handles ?? {})
      }
    } catch (error) {
      settingsError.value = error.message
    } finally {
      settingsLoading.value = false
    }
    loadCommandHistory()
    loadBranchUsers(id)
  }

  function reloadActive () {
    if (selectedId.value != null) loadSettings(selectedId.value)
  }

  // ---- POS users --------------------------------------------------------
  const { listBranchUsers, createBranchUser, updateBranchUser, setBranchUserPin, deleteBranchUser } = useBranchUsers()
  const branchUsers = ref([])
  const usersError = ref('')

  async function loadBranchUsers (id) {
    branchUsers.value = []
    usersError.value = ''
    try {
      branchUsers.value = await listBranchUsers(id)
    } catch (error) {
      usersError.value = error.message
    }
  }

  const userDialog = ref(false)
  const editingUser = ref(null)
  const userSaving = ref(false)
  const userFormError = ref('')
  const userForm = ref({ username: '', role: 'cashier', akuvox_user_id: '', pin: '', is_active: true })

  function openUserDialog (u = null) {
    editingUser.value = u
    userFormError.value = ''
    userForm.value = u
      ? { username: u.username, role: u.role, akuvox_user_id: u.akuvox_user_id, pin: '', is_active: u.is_active }
      : { username: '', role: 'cashier', akuvox_user_id: '', pin: '', is_active: true }
    userDialog.value = true
  }

  async function submitUser () {
    const f = userForm.value
    if (!f.username.trim()) {
      userFormError.value = 'Username is required.'
      return
    }
    if (!editingUser.value && !/^\d{4,6}$/.test(f.pin)) {
      userFormError.value = 'Initial PIN must be 4 to 6 digits.'
      return
    }
    userSaving.value = true
    userFormError.value = ''
    try {
      await (editingUser.value ? updateBranchUser(editingUser.value.id, {
        username: f.username.trim(),
        role: f.role,
        akuvox_user_id: f.akuvox_user_id,
        is_active: f.is_active,
      }) : createBranchUser(selectedId.value, {
        username: f.username.trim(),
        role: f.role,
        akuvox_user_id: f.akuvox_user_id,
        is_active: f.is_active,
        pin: f.pin,
      }));
      userDialog.value = false
      await loadBranchUsers(selectedId.value)
    } catch (error) {
      userFormError.value = error.message
    } finally {
      userSaving.value = false
    }
  }

  async function removeUser (u) {
    usersError.value = ''
    try {
      await deleteBranchUser(u.id)
      await loadBranchUsers(selectedId.value)
    } catch (error) {
      usersError.value = error.message
    }
  }

  const pinDialog = ref(false)
  const pinTarget = ref(null)
  const pinValue = ref('')
  const pinSaving = ref(false)
  const pinError = ref('')

  function openPinDialog (u) {
    pinTarget.value = u
    pinValue.value = ''
    pinError.value = ''
    pinDialog.value = true
  }

  async function submitPin () {
    if (!/^\d{4,6}$/.test(pinValue.value)) {
      pinError.value = 'PIN must be 4 to 6 digits.'
      return
    }
    pinSaving.value = true
    pinError.value = ''
    try {
      await setBranchUserPin(pinTarget.value.id, pinValue.value)
      pinDialog.value = false
      await loadCommandHistory()
    } catch (error) {
      pinError.value = error.message
    } finally {
      pinSaving.value = false
    }
  }

  async function setGroupSource (group, source) {
    const local = source === 'local'
    const next = managedLocally.value.filter(g => g !== group)
    if (local) next.push(group)
    const prev = managedLocally.value
    managedLocally.value = next
    try {
      await setBranchSettingsManagedLocally(selectedId.value, next)
    } catch (error) {
      managedLocally.value = prev
      settingsError.value = error.message
    }
  }

  async function loadCommandHistory () {
    if (selectedId.value == null) return
    try {
      commandHistory.value = await listBranchCommands(selectedId.value)
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

      const payload = {
        ...settingsForm.value,
        social_platforms: socialPlatforms,
        social_handles: socialHandles,
      }

      // Groups handed back to the branch are never pushed.
      if (isLocal('device_ids')) {
        for (const k of GROUP_KEYS.device_ids) delete payload[k]
      }
      if (!isLocal('secrets')) {
        // Write-only: only send a secret the admin actually typed.
        if (secretForm.value.akuvox_password) {
          payload.akuvox_password = secretForm.value.akuvox_password
        }
        if (secretForm.value.attendance_approver_password) {
          payload.attendance_approver_password = secretForm.value.attendance_approver_password
        }
        if (secretForm.value.admin_password) {
          payload.admin_password = secretForm.value.admin_password
        }
      }

      await updateBranchSettings(selectedId.value, payload)
      saveOutcome.value = { ok: true, message: 'Command sent -- the branch will apply it within a few seconds if online.' }
      secretForm.value = { akuvox_password: '', attendance_approver_password: '', admin_password: '' }
      await loadCommandHistory()
    } catch (error) {
      saveOutcome.value = { ok: false, message: error.message }
    } finally {
      saving.value = false
    }
  }
</script>

<style scoped>
/* Keep intrinsic height inside the layout's flex-column scroll wrapper so the
   scroll lives inside the content pane, not the whole page. */
.page-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.bs-card {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.bs-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}
.bs-content {
  flex: 1 1 auto;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
/* activeBranch wrapper: header stays put, the section pane scrolls */
.bs-detail-root {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.bs-detail {
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
}
.bs-section-tabs {
  flex: 0 0 168px;
}
.bs-section-tabs :deep(.v-tab) {
  justify-content: flex-start;
}
.bs-section-pane {
  flex: 1 1 auto;
  min-width: 0;
  overflow-y: auto;
  padding-left: 20px;
}
.bs-savebar {
  position: sticky;
  bottom: 0;
  padding: 12px 0 4px;
  margin-top: 8px;
  background: rgb(var(--v-theme-surface));
}
.bs-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  flex-shrink: 0;
}
.bs-dot.is-online {
  background: rgb(var(--v-theme-success));
}
.bs-dot.is-offline {
  background: rgba(128, 128, 128, 0.5);
}
.bs-placeholder {
  font-size: 0.875rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  height: 100%;
  color: rgba(var(--v-theme-on-surface), 0.55);
  text-align: center;
}
.bs-section {
  max-width: 720px;
  margin-bottom: 24px;
}
.bs-section-title {
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: rgba(var(--v-theme-on-surface), 0.6);
  margin-bottom: 12px;
}
.bs-group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
}
.field-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  align-items: start;
}
.field-grid > .field-full {
  grid-column: 1 / -1;
}
.bs-switch-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 0 24px;
}
</style>
