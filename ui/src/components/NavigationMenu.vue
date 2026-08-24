<!-- TODO disgusting code -->
<template>
  <v-list-group v-if="item.subItems" :id="item.id" :fluid="isRail">
    <!-- :ripple="false" -->
    <template #activator="{ props }">
      <v-list-item
        v-bind="props"
        :prepend-icon="item.icon"
        :title="item.text"
      />
    </template>
    <NavigationMenu
      v-for="(subItem, i) in item.subItems"
      :id="subItem.id"
      :key="i"
      :item="subItem"
    />
    <!-- :ripple="false" -->
  </v-list-group>

  <!-- TODO: color="primary" -->
  <v-list-item
    v-else
    :id="item.id"
    :prepend-icon="item.icon"
    :title="item.text"
    :to="item.to"
    :value="item"
  >
    <!-- :ripple="false" -->
    <!-- TODO change these -->
    <template v-if="item.text === 'Requests'" #append>
      <v-badge color="red-darken-4" content="6" inline />
    </template>
    <template v-if="item.text === 'Live Chat'" #append>
      <v-badge color="green-darken-4" content="6" inline />
    </template>
  </v-list-item>
</template>

<script setup>
  const props = defineProps({
    item: String,
  // isRail: Boolean,
  })
</script>
