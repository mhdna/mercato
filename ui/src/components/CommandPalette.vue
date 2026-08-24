<template>
  <v-command-palette
    v-model:search="search"
    height="495"
    :items="items"
    hotkey="ctrl+k"
    :no-data-text="'No commands found'"
    placeholder="Search commands..."
    @click:item="onItemClick"
  >
    <template #activator="{ props: activatorProps }">
      <v-btn icon="mdi-magnify" variant="text" v-bind="activatorProps" />
    </template>
  </v-command-palette>
</template>

<script setup>
  import { shallowRef } from 'vue'
  import { useRouter } from 'vue-router'
  import { appendItems, navItems } from '@/data/navItems'

  const router = useRouter()
  const search = shallowRef('')

  function buildItems () {
    const items = []

    for (const item of navItems) {
      if (item.children) {
        items.push({ type: 'subheader', title: item.title })
        for (const child of item.children) {
          items.push({
            title: child.title,
            subtitle: item.title,
            prependIcon: child.icon,
            value: child.to,
          })
        }
      } else {
        items.push({
          title: item.title,
          prependIcon: item.icon,
          value: item.to,
        })
      }
    }

    items.push({ type: 'divider' })
    items.push({ type: 'subheader', title: 'More' })
    for (const item of appendItems) {
      items.push({
        title: item.title,
        prependIcon: item.icon,
        value: item.to,
      })
    }

    return items
  }

  const items = buildItems()

  function onItemClick (item) {
    if (item.value) {
      router.push(item.value)
    }
  }
</script>
