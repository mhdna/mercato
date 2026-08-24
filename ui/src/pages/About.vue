<template>
  <div>
    <h1>Health Check Status</h1>
    <p v-if="error">{{ error }}</p>
    <p v-else-if="loading">Loading...</p>
    <p v-else>{{ healthData }}</p>
  </div>
</template>

<script>
  export default {
    data () {
      return {
        healthData: null,
        loading: true,
        error: null,
      }
    },
    mounted () {
      this.fetchHealthCheck()
    },
    methods: {
      async fetchHealthCheck () {
        try {
          const response = await fetch('http://localhost:4000/v1/healthcheck')
          if (!response.ok) {
            throw new Error('Network response was not ok')
          }
          this.healthData = await response.json()
        } catch (error) {
          this.error = 'Failed to fetch health check: ' + error.message
        } finally {
          this.loading = false
        }
      },
    },
  }
</script>

<style scoped>
/* Add your styles here if needed */
</style>
