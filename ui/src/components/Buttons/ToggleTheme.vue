<template>
  <!-- TODO auto theme and list chosen -->
  <v-btn icon @click="theme.toggle()">
    <!-- v-tooltip:bottom="'Toggle Theme'" -->
    <!-- <v-icon :style="{ -->
    <!--   color: theme.global.current.value.dark ? 'darkblue' : 'yellow', -->
    <!-- }"> -->
    <v-icon icon="mdi-theme-light-dark">
      <!-- {{ -->
      <!--   theme.global.current.value.dark -->
      <!--     ? "mdi-weather-night" -->
      <!--     : "mdi-white-balance-sunny" -->
      <!-- }} -->
    </v-icon>
  </v-btn>
</template>

<script lang="ts" setup>
  import { useTheme } from 'vuetify'

  const theme = useTheme()
</script>
