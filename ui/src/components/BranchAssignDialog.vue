<template>
  <v-dialog max-width="480" :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title class="d-flex align-center ga-2">
        <v-icon icon="mdi-store-outline" />
        <span>{{ title }}</span>
      </v-card-title>
      <v-divider />
      <v-card-text>
        <p class="text-medium-emphasis text-body-2 mb-3">
          A branch can only use one {{ kind }} at a time. Ticking a branch that
          already belongs to another {{ kind }} is rejected — untick it there first.
        </p>

        <div v-if="loading" class="py-4 text-center">
          <v-progress-circular indeterminate size="24" />
        </div>

        <template v-else>
          <div v-if="branches.length === 0" class="text-medium-emphasis py-2">
            No branches yet.
          </div>
          <v-list v-else class="py-0" density="compact">
            <v-list-item
              v-for="b in branches"
              :key="b.id"
              @click="toggle(b.id)"
            >
              <template #prepend>
                <v-checkbox-btn
                  :model-value="selected.includes(b.id)"
                  @click.stop="toggle(b.id)"
                />
              </template>
              <v-list-item-title>{{ b.name }}</v-list-item-title>
              <v-list-item-subtitle>{{ b.code }}</v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </template>

        <v-alert
          v-if="error"
          class="mt-3"
          density="compact"
          type="error"
          variant="tonal"
        >
          {{ error }}
        </v-alert>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn text="Cancel" variant="text" @click="$emit('update:modelValue', false)" />
        <v-btn
          color="primary"
          :loading="saving"
          text="Save"
          variant="flat"
          @click="save"
        />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { useBranches } from '@/composables/useBranches'

  const props = defineProps({
    modelValue: { type: Boolean, default: false },
    title: { type: String, default: 'Enabled Branches' },
    kind: { type: String, default: 'price list' },
    // async (branchIds) => any ; throws Error on conflict
    saveFn: { type: Function, required: true },
    // async () => number[] ; currently-assigned branch ids
    loadFn: { type: Function, required: true },
  })
  const emit = defineEmits(['update:modelValue', 'saved'])

  const { branches, fetchBranches } = useBranches()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const selected = ref([])

  function toggle (id) {
    const i = selected.value.indexOf(id)
    if (i === -1) {
      selected.value.push(id)
    } else {
      selected.value.splice(i, 1)
    }
  }

  watch(() => props.modelValue, async open => {
    if (!open) return
    error.value = ''
    loading.value = true
    try {
      await fetchBranches()
      selected.value = [...await props.loadFn()]
    } catch (error_) {
      error.value = error_.message
    } finally {
      loading.value = false
    }
  })

  async function save () {
    saving.value = true
    error.value = ''
    try {
      await props.saveFn([...selected.value])
      emit('saved')
      emit('update:modelValue', false)
    } catch (error_) {
      error.value = error_.message
    } finally {
      saving.value = false
    }
  }
</script>
