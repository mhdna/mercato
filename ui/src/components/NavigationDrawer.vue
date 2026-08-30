<template>
  <v-navigation-drawer
    class="nav-drawer"
    density="compact"
    location="left"
    :model-value="props.mobile ? props.modelValue : true"
    :permanent="!props.mobile"
    :rail="props.rail"
    :temporary="props.mobile"
    width="260"
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
          <template #activator="{ props: activatorProps }">
            <v-list-item v-bind="activatorProps" :title="item.title" />
          </template>

          <v-list-item
            v-for="child in item.children"
            :key="child.title"
            class="ms-1 me-0"
            :prepend-icon="child.icon"
            :title="child.title"
            :to="child.to"
            :value="child.to"
          >
            <template v-if="child.append === 'status'" #append>
              <span
                aria-label="Stock health status is healthy"
                class="nav-status d-flex align-center ga-2"
                role="img"
              >
                <span class="nav-status-dot" />
              </span>
            </template>
          </v-list-item>
        </v-list-group>

        <!-- Single item -->
        <v-list-item
          v-else
          :color="item.color"
          :prepend-icon="item.icon"
          :title="item.title"
          :to="item.to"
          :value="item.title"
        >
          <template v-if="item.append === 'status'" #append>
            <span
              aria-label="Stock health status is healthy"
              class="nav-status d-flex align-center ga-2"
              role="img"
            >
              <span class="nav-status-dot" />
            </span>
          </template>
          <template v-else-if="item.append === 'alert-count'" #append>
            <v-badge
              color="error"
              :content="item.appendCount"
              inline
            />
          </template>
        </v-list-item>
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

<style scoped>
.nav-drawer {
  --v-list-prepend-gap: 20px;
}

.nav-drawer :deep(.v-list-group__items .v-list-item) {
  padding-inline-start: 16px !important;
}

.nav-status-dot {
  width: 18px;
  height: 18px;
  display: block;
  border-radius: 50%;
  background-color: rgb(var(--v-theme-success));
}

.nav-status {
  min-width: 24px;
  justify-content: center;
  margin-inline-end: 8px;
}
</style>
