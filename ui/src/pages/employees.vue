<script setup lang="ts">
  import EmployeesPerformanceTable from '@/components/Tables/EmployeesPerformanceTable.vue'

</script>

<template>
  <div class="d-flex flex-wrap text-center">
    <EmployeesPerformanceTable />
    <!-- <v-card v-for="n in 6" to="/employee_page" class="ma-4 pa-4" rounded="xl">
      <EmployeeCard name="Sara" title="Salesman" />
    </v-card> -->
  </div>
</template>
