<template>
  <div>
    <v-img
      class="mx-auto my-6"
      max-width="228"
      src="https://cdn.vuetifyjs.com/docs/images/logos/vuetify-logo-v3-slim-text-light.svg"
    />

    <v-card class="mx-auto pa-12 pb-8" elevation="8" max-width="448" rounded="lg">
      <div class="text-subtitle-1 text-medium-emphasis">Account</div>

      <v-text-field
        v-model="username"
        density="compact"
        placeholder="Username"
        prepend-inner-icon="mdi-account-outline"
        variant="outlined"
        @keyup.enter="submit"
      />

      <div class="text-subtitle-1 text-medium-emphasis d-flex align-center justify-space-between">
        Password

        <!-- <a -->
        <!--   class="text-caption text-decoration-none text-blue" -->
        <!--   href="#" -->
        <!--   rel="noopener noreferrer" -->
        <!--   target="_blank" -->
        <!-- > -->
        <!--   Forgot login password?</a> -->
      </div>

      <v-text-field
        v-model="password"
        :append-inner-icon="visible ? 'mdi-eye-off' : 'mdi-eye'"
        density="compact"
        placeholder="Enter your password"
        prepend-inner-icon="mdi-lock-outline"
        :type="visible ? 'text' : 'password'"
        variant="outlined"
        @click:append-inner="visible = !visible"
        @keyup.enter="submit"
      />

      <v-alert v-if="error" class="mb-4" type="error" variant="tonal">
        {{ error }}
      </v-alert>

      <v-card class="mb-12" color="surface-variant" variant="tonal">
        <v-card-text class="text-medium-emphasis text-caption">
          Warning: After 3 consecutive failed login attempts, you account will be temporarily locked for three hours. If
          you must login now, you can also click "Forgot login password?" below to reset the login password.
        </v-card-text>
      </v-card>

      <v-btn
        block
        class="mb-8"
        color="blue"
        :loading="submitting"
        size="large"
        variant="tonal"
        @click="submit"
      >
        Log In
      </v-btn>

      <v-card-text class="text-center">
        <a class="text-blue text-decoration-none" href="#" rel="noopener noreferrer" target="_blank">
          Sign up now <v-icon icon="mdi-chevron-right" />
        </a>
      </v-card-text>
    </v-card>
  </div>
</template>
<script setup lang="ts">
  import { ref } from 'vue'
  import { useRouter } from 'vue-router'
  import { useAuthStore } from '@/stores/auth'

  const router = useRouter()
  const authStore = useAuthStore()

  const visible = ref(false)
  const username = ref('')
  const password = ref('')
  const submitting = ref(false)
  const error = ref('')

  async function submit () {
    error.value = ''
    submitting.value = true
    try {
      await authStore.login(username.value, password.value)
      router.push('/')
    } catch (error_) {
      error.value = error_.message
    } finally {
      submitting.value = false
    }
  }
</script>
