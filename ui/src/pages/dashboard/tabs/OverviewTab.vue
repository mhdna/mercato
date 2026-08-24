<template>
  <div class="d-flex flex-column flex-grow-1" style="min-height: 0">
    <div class="d-flex">
      <SparklineCard icon="mdi-cart-check" title="Items Sold" unit="" />
      <SparklineCard
        icon="mdi-card-account-details"
        title="New Clients"
        unit=""
      />
      <SparklineCard icon="mdi-currency-usd" title="Revenue" unit="$" />
      <SparklineCard icon="mdi-currency-usd-off" title="Expenses" unit="$" />
      <SparklineCard icon="mdi-receipt-text-check" title="Invoices" unit="" />
    </div>
    <ChartsRow />
    <TablesRow />
  </div>
</template>

<script setup>
import SparklineCard from "@/components/Cards/SparklineCard.vue";
import ChartsRow from "../ChartsRow.vue";
import TablesRow from "../TablesRow.vue";
</script>
