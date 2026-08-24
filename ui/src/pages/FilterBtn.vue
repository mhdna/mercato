<template>
    <v-btn class="mx-2 mt-1" rounded="xl" prepend-icon="mdi-filter" @click="dialog = !dialog">
        Filters
    </v-btn>
    <v-dialog v-model="dialog" max-width="400">
        <v-card max-width="400">
            <v-date-picker></v-date-picker>

            Filter
        </v-card>
    </v-dialog>
</template>

<script setup>
import { ref } from 'vue';
const dialog = ref(false)
</script>