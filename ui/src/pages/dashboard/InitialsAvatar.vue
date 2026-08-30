<template>
  <div class="d-flex align-center" style="min-width: 0">
    <v-avatar
      class="flex-shrink-0 text-caption font-weight-bold"
      :color="color"
      size="26"
    >
      {{ initials || '·' }}
    </v-avatar>
    <span v-if="showName" class="ms-2 text-body-2 text-truncate">{{ name || '—' }}</span>
  </div>
</template>

<script setup>
  import { computed } from 'vue'

  const props = defineProps({
    name: {
      type: String,
      default: '',
    },
    showName: {
      type: Boolean,
      default: true,
    },
  })

  const initials = computed(() => {
    const parts = (props.name || '').trim().split(/\s+/).filter(Boolean)
    if (parts.length === 0) return ''
    if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
    return (parts[0][0] + parts.at(-1)[0]).toUpperCase()
  })

  // Deterministic muted colour per name so avatars are distinguishable
  // without being loud.
  const palette = [
    'indigo-lighten-1', 'teal-lighten-1', 'deep-purple-lighten-1',
    'blue-grey-lighten-1', 'cyan-darken-1', 'green-lighten-1',
    'orange-lighten-1', 'pink-lighten-1',
  ]
  const color = computed(() => {
    const key = props.name || ''
    let hash = 0
    for (let i = 0; i < key.length; i++) {
      hash = Math.trunc(hash * 31 + key.codePointAt(i))
    }
    return key ? palette[Math.abs(hash) % palette.length] : 'grey-lighten-1'
  })
</script>
