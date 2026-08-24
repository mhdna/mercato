<template>
  <v-card class="mx-auto" max-width="500">
    <v-sheet class="pa-4" color="surface-variant">
      <v-text-field
        v-model="search"
        clear-icon="mdi-close-circle-outline"
        clearable
        flat
        hide-details
        label="Search Company Directory"
        variant="solo"
      />

      <v-checkbox-btn
        v-model="caseSensitive"
        label="Case sensitive search"
      />
    </v-sheet>

    <v-treeview
      v-model:opened="open"
      :custom-filter="filter"
      item-value="id"
      :items="items"
      open-on-click
      :search="search"
    >
      <template #prepend="{ item }">
        <v-icon
          v-if="item.children"
          :icon="`mdi-${item.id === 1 ? 'home-variant' : 'folder-network'}`"
        />
      </template>
    </v-treeview>
  </v-card>
</template>

<script setup>
  import { shallowRef } from 'vue'

  const items = [
    {
      id: 1,
      title: 'Vuetify Human Resources',
      children: [
        {
          id: 2,
          title: 'Core team',
          children: [
            {
              id: 201,
              title: 'John',
            },
            {
              id: 202,
              title: 'Kael',
            },
            {
              id: 203,
              title: 'Nekosaur',
            },
            {
              id: 204,
              title: 'Jacek',
            },
            {
              id: 205,
              title: 'Andrew',
            },
          ],
        },
        {
          id: 3,
          title: 'Administrators',
          children: [
            {
              id: 301,
              title: 'Blaine',
            },
            {
              id: 302,
              title: 'Yuchao',
            },
          ],
        },
        {
          id: 4,
          title: 'Contributors',
          children: [
            {
              id: 401,
              title: 'Phlow',
            },
            {
              id: 402,
              title: 'Brandon',
            },
            {
              id: 403,
              title: 'Sean',
            },
          ],
        },
      ],
    },
  ]

  const open = shallowRef([1, 2])
  const search = shallowRef(null)
  const caseSensitive = shallowRef(false)

  function filter (value, search, item) {
    return caseSensitive.value
      ? value.includes(search)
      : value.toLowerCase().includes(search.toLowerCase())
  }
</script>
