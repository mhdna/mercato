<script setup lang="ts">
  import ClientRankTag from '@/components/ClientRankTag.vue'
  import Heatmap from '@/components/Stats/Heatmap.vue'
  import PieChart from '@/components/Stats/PieChart.vue'

  const clientRank = ref('Gold Tier')

</script>

<template>
  <v-sheet class="pa-0" color="surface-light" max-height="200">
    <v-row class="pa-8">
      <v-col md="2">
        <InitialsAvatar initials="AM" />
      </v-col>
      <v-col class="pt-4">
        <div class="d-flex">
          <div class="text-headline-large pb-2">
            Aisha Miliki
          </div>
          <ClientRankTag :rank="clientRank" />
        </div>
        <div class="d-flex">
          <div class="text-title-medium text-grey" style="width: 130px;">
            ID:
          </div>
          <div>
            dfkj3-fjie2-jkfdi2-klio9
          </div>
        </div>
        <div class="d-flex">
          <div class="text-title-medium text-grey" style="width: 130px;">
            Phone:
          </div>
          <div>
            772837421
          </div>
        </div>
        <div class="d-flex">
          <div class="text-title-medium text-grey" style="width: 130px;">
            First Purchase:
          </div>
          <div>
            13-04-2026
          </div>
        </div>
      </v-col>
      <v-col class="pt-4" md="6">
        <v-row>
          <v-col>
            <SparklineCard icon="mdi-currency-usd" title="Money Spent" unit="$" />
            <SparklineCard icon="mdi-cart-check" title="Items Bought" unit="" />
          </v-col>
          <v-col>
            <SparklineCard icon="mdi-circle" title="Points Remaining" unit="" />
            <SparklineCard icon="mdi-check-circle" title="Points Redeemed" unit="" />
          </v-col>
        </v-row>
      </v-col>
    </v-row>
  </v-sheet>
  <v-row>
    <v-col md="7">
      <v-card class="pa-4 mb-2" flat max-height="180" max-width="800">
        <v-card-title class="card-title">Activity Over the Year</v-card-title>
        <Heatmap />
      </v-card>
      <ServerSideTable api-u-r-l="http://localhost:4123/invoices" root-key="invoices" />
    </v-col>
    <v-col md="5">
      <v-card class="pa-4 mb-2" flat max-height="300" max-width="800">
        <v-card-title class="card-title">Most Bought Genre</v-card-title>
        <PieChart />
      </v-card>
      <v-card class="card-title" flat max-height="300" max-width="800">
        <v-card-title class="px-2 py-2 font-weight-bold">Most Bought Categories</v-card-title>
        <PieChart />
      </v-card>
    </v-col>
  </v-row>
</template>
