<!--
  Category picker for the expense / loan / asset pages. Thin adapter that maps
  category rows onto the shared PageSidebar rail (see PageSidebar.vue) so every
  master/detail page shares one look.
-->
<template>
  <PageSidebar
    add-label="Add Category"
    empty-text="No categories yet."
    :groups="groupDefs"
    :items="items"
    :model-value="modelValue"
    :row-actions="true"
    :show-all="true"
    :width="240"
    @add="$emit('add')"
    @delete="$emit('delete', $event)"
    @edit="$emit('edit', $event)"
    @update:model-value="$emit('update:modelValue', $event)"
  />
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import PageSidebar from '@/components/PageSidebar.vue'
  import { categoryColor, categoryIcon } from '@/data/expenseCategoryIcons'

  const props = defineProps({
    categories: { type: Array, default: () => [] },
    // Selected category id, or null for "All".
    modelValue: { type: [Number, null], default: null },
    // Split the list into "Central" / "Branch" sub-groups by `scope`. Off for
    // taxonomies that have no scope concept (e.g. asset categories).
    groupByScope: { type: Boolean, default: true },
  })
  defineEmits(['update:modelValue', 'add', 'edit', 'delete'])

  const groupDefs = computed(() => props.groupByScope
    ? [{ key: 'central', label: 'Central' }, { key: 'branch', label: 'Branch' }]
    : [])

  const items = computed(() => props.categories
    .toSorted((a, b) => {
      if (a.is_active !== b.is_active) return a.is_active ? -1 : 1
      return a.name.localeCompare(b.name)
    })
    .map(category => ({
      value: category.id,
      title: category.name,
      subtitle: category.is_active ? '' : 'Inactive',
      inactive: !category.is_active,
      avatarColor: categoryColor(category),
      avatarIcon: categoryIcon(category),
      group: category.scope === 'branch' ? 'branch' : 'central',
      raw: category,
    })),
  )
</script>
