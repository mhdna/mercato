<template>
  <div class="text-center">
    <v-menu>
      <template v-slot:activator="{ props }">
        <v-btn
          rounded="lg"
          variant="outlined"
          class="me-4 mt-2"
          v-bind="props"
          style="border-color: rgb(var(--v-theme-surface-light));"
        >
          All Branches
        </v-btn>
      </template>

      <v-list>
        <v-list-item
          v-for="(item, index) in items"
          :key="index"
          :value="index"
        >
          <v-list-item-title>{{ item }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </div>
</template>

<script setup>
        const items=['All Branches', 'Centro', 'Beirut', 'Hamra', 'Sour']
</script>