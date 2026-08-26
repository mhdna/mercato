<template>
  <v-dialog v-model="dialog" max-height="780" max-width="920">
    <template #default>
      <v-card class="px-4">
        <v-card-title>
          Create a New Product
          <v-spacer />
        </v-card-title>
        <NewProductForm @created="onProductCreated" />

        <v-card-actions>
          <v-spacer />

          <v-btn text="Close Dialog" @click="dialog = false" />
        </v-card-actions>
      </v-card>
    </template>
  </v-dialog>

  <div class="d-flex justify-space-between">
    <v-btn text="Add a New Product" @click="dialog = true" />
    <div>
      <v-icon-btn icon="mdi-format-list-bulleted" @click="galleryView = false" />
    </div>
  </div>

  <ProductsTable v-if="!galleryView" ref="productsTableRef" />
</template>

<script setup>
  import { ref } from 'vue'
  import NewProductForm from '@/components/Forms/NewProductForm.vue'
  import ProductsTable from '@/components/ProductsTable.vue'

  const dialog = ref(false)
  const galleryView = ref(false)
  const productsTableRef = ref(null)

  function onProductCreated () {
    dialog.value = false
    productsTableRef.value?.reload()
  }
</script>
