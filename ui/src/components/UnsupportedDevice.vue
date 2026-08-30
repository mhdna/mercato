<template>
  <v-main class="unsupported-device">
    <section class="unsupported-device__content text-center">
      <v-img
        alt="GB Cloud"
        class="unsupported-device__logo mx-auto"
        src="/app-icon.png"
      />

      <h1 class="text-h4 font-weight-bold mt-6">Desktop only</h1>
      <p class="text-body-1 text-medium-emphasis mt-3 mb-0">
        Please continue on a desktop computer.
      </p>
    </section>
  </v-main>
</template>

<style scoped>
.unsupported-device {
  align-items: center;
  background:
    radial-gradient(circle at top, rgba(var(--v-theme-primary), 0.14), transparent 45%),
    rgb(var(--v-theme-background));
  display: flex;
  justify-content: center;
  min-height: 100dvh;
  padding: 24px;
}

.unsupported-device__content {
  max-width: 360px;
  width: 100%;
}

.unsupported-device__logo {
  height: 112px;
  width: 112px;
}
</style>
