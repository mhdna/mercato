<template>
  <v-container>
    <v-row>
      <v-btn>Import Excel Sheet</v-btn>
      <v-checkbox v-model="multiLine" label="Multi-Line Mode" />
    </v-row>
    <v-row v-if="!multiLine">
      <v-text-field
        hide-details="auto"
        label="Enter your barcode"
      />
    </v-row>
    <v-row>
      <v-col v-if="multiLine" cols="4">
        <v-textarea label="Barcode,Quantity" variant="outlined" />
      </v-col>
      <v-col>
        <server-side-table />
      </v-col>
    </v-row>
    <v-card>
      <v-card-title>Total Items to Print: 0</v-card-title>
      Code 1 Code 2
    </v-card>
  </v-container>
</template>

<script setup>
  import ServerSideTable from '@/components/Tables/ServerSideTable.vue'

  const multiLine = ref(false)
</script>
