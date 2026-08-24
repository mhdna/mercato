<template>
  <v-navigation-drawer
    class="nav-drawer"
    density="compact"
    location="left"
    :model-value="props.mobile ? props.modelValue : true"
    :permanent="!props.mobile"
    :rail="props.rail"
    :temporary="props.mobile"
    width="240"
    @update:model-value="props.mobile && $emit('update:modelValue', $event)"
  >
    <v-list density="compact" nav>
      <template v-for="item in navItems" :key="item.title">
        <!-- Group item -->
        <v-list-group
          v-if="item.children"
          append-icon="mdi-chevron-down"
          :prepend-icon="item.icon"
          :value="item.title"
        >
          <template #activator="{ props }">
            <v-list-item v-bind="props" :title="item.title" />
          </template>

          <v-list-item
            v-for="child in item.children"
            :key="child.title"
            class="mx-4 me-0"
            :prepend-icon="child.icon"
            :title="child.title"
            :to="child.to"
            :value="child.title"
          />
        </v-list-group>

        <!-- Single item -->
        <v-list-item
          v-else
          :prepend-icon="item.icon"
          :title="item.title"
          :to="item.to"
          :value="item.title"
        />
      </template>
    </v-list>

    <template #append>
      <v-divider thickness="2" />
      <v-list class="pa-0" density="compact" nav>
        <div class="pa-2">
          <v-list-item
            v-for="item in appendItems"
            :key="item.title"
            :prepend-icon="item.icon"
            :title="item.title"
            :to="item.to"
            :value="item.title"
          />
        </div>
      </v-list>
    </template>
  </v-navigation-drawer>
</template>

<script setup>
  import { appendItems, navItems } from '@/data/navItems'

  const props = defineProps({
    modelValue: { type: Boolean, required: true },
    rail: { type: Boolean, default: false },
    mobile: { type: Boolean, default: false },
  })
  defineEmits(['update:modelValue', 'update:rail'])
</script>
