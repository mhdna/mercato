<template>
  <v-card class="pa-0" flat rounded="0" width="100%">
    <!-- color="surface-light" -->
    <template #prepend>
      <v-icon
        class="me-4"
        :color="checking ? 'red lighten-2' : 'indigo-lighten-2'"
        :icon="props.icon"
        size="37"
        @click="takePulse"
      />
    </template>

    <template #title>
      <div class="text-body-small text-uppercase">
        <!-- text-grey -->
        {{ props.title }}
      </div>

      <strong v-if="avg"> {{ props.unit }}</strong>
      <span class="text-headline-small font-weight-black" v-text="avg || '—'" />
    </template>

    <template #append>
      <v-btn
        class="align-self-start"
        icon="mdi-arrow-right-thick"
        size="34"
        variant="text"
      />
    </template>

    <!-- <v-sheet color="transparent" class="ma-0"> -->
    <!--     <v-sparkline height="40" :key="String(avg)" :line-width="1" :model-value="heartbeats" :smooth="0" -->
    <!--         auto-draw></v-sparkline> -->
    <!-- color="grey" :gradient="['#f72047', '#ffd200', '#1feaea']" class="ma-0 my-0" stroke-linecap="round" -->
    <!-- </v-sheet> -->
  </v-card>
</template>
<script setup>
import { computed, ref } from "vue";

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  icon: {
    type: String,
    required: true,
  },
  unit: {
    type: String,
    required: true,
  },
});

const exhale = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const checking = ref(false);
const heartbeats = ref([]);
const avg = computed(() => {
  const sum = heartbeats.value.reduce((acc, cur) => acc + cur, 0);
  const length = heartbeats.value.length;
  if (!sum && !length) return 0;
  return Math.ceil(sum / length);
});

function heartbeat() {
  return Math.ceil(Math.random() * (1_111_120 - 80) + 80);
}
async function takePulse(inhale = true) {
  checking.value = true;
  inhale && (await exhale(100_000));
  heartbeats.value = Array.from({ length: 20 }, heartbeat);
  checking.value = false;
}
takePulse(false);
</script>
