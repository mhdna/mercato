<template>
  <v-command-palette
    v-model:search="search"
    height="75vh"
    hotkey="cmd+k"
    :items="items"
    location="center center"
    :no-data-text="'No commands found'"
    offset-top="0"
    placeholder="Search commands..."
    @click:item="onItemClick"
    @update:model-value="onToggle"
  >
    <template #activator="{ props: activatorProps }">
      <v-btn icon="mdi-magnify" variant="text" v-bind="activatorProps" />
    </template>
  </v-command-palette>
</template>

<script setup>
  import { nextTick, shallowRef } from 'vue'
  import { useRouter } from 'vue-router'
  import { appendItems, navItems } from '@/data/navItems'

  const router = useRouter()
  const search = shallowRef('')

  function onToggle (value) {
    if (!value) {
      // VCommandPalette restores focus to the activator button on close,
      // which leaves it with a lingering focus ring. Drop that focus.
      nextTick(() => {
        requestAnimationFrame(() => {
          const el = document.activeElement
          if (el instanceof HTMLElement) {
            el.blur()
          }
        })
      })
    }
  }

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

    items.push({ type: 'divider' }, { type: 'subheader', title: 'More' })
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
