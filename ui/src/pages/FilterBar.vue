<script setup>
  const menu = ref(false)
  const today = new Date()
  const past = new Date()
  past.setDate(today.getDate() - 30)

  const date = ref([past, today])

  const formattedDate = computed(() => {
    if (!date.value || date.value.length === 0) return ''

    const format = d =>
      new Date(d).toLocaleDateString('en-GB')

    if (date.value.length === 1) {
      return format(date.value[0])
    }

    return `${format(date.value[0])} → ${format(date.value[1])}`
  })
</script>
<template>
  <v-sheet class="d-flex align-center pt-0" color="surface-light" style="height: 50px;" width="100%">
    <v-row class="d-flex align-center ">
      <v-spacer />
      <!-- <div class="me-2 mt-4"> -->
      <!--     Filters: -->
      <!-- </div> -->
      <v-icon class="pt-1" icon="mdi-arrow-left-thin" size="22" variant="flat" />
      <v-menu v-model="menu" :close-on-content-click="false" location="end">
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            class="ma-0 py-4 pt-5"
            color="surface-light"
            max-width="180"
            variant="flat"
          >
            {{ formattedDate }}
          </v-btn>
        </template>

        <v-card min-width="300">
          <v-date-picker v-model="date" multiple="range" />
          <v-card-actions>
            <v-spacer />
            <v-btn variant="text" @click="menu = false">
              Cancel
            </v-btn>
            <v-btn color="primary" variant="text" @click="menu = false">
              Save
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-menu>
      <v-icon class="pt-1" icon="mdi-arrow-right-thin" size="22" variant="flat" />
      <v-spacer />
      <v-select
        class="pt-6"
        density="compact"
        :items="['Branch 1', 'Branch 2', 'Branch 3', 'Branch 4']"
        label="Branch"
        max-width="150"
        variant="flat"
      />
    </v-row>
  </v-sheet>
</template>
