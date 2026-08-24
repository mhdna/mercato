<!-- TODO properize this -->
<template>
  <div class="text-center d-flex justify-space-between align-center">
    <v-menu v-model="menu" :close-on-content-click="false">
      <template #activator="{ props }">
        <!-- <v-tooltip text="Notifications" max-width="300" location="bottom"> -->
        <!--   <template v-slot:activator="{ props }"> -->
        <v-badge v-bind="props" class="me-1" color="red-darken-4" content="9+">
          <v-icon v-bind="on">mdi-bell</v-icon>
        </v-badge>
        <!--   </template> -->
        <!-- </v-tooltip> -->
      </template>

      <v-card class="overflow-y-auto" max-height="520" max-width="400">
        <!-- <v-list> -->
        <!--   <v-list-item -->
        <!--     prepend-avatar="https://cdn.vuetifyjs.com/images/john.jpg" -->
        <!--     subtitle="Founder of Vuetify" -->
        <!--     title="John Leider" -->
        <!--   > -->
        <!--     <template v-slot:append> -->
        <!--       <v-btn -->
        <!--         :class="fav ? 'text-red' : ''" -->
        <!--         icon="mdi-heart" -->
        <!--         variant="text" -->
        <!--         @click="fav = !fav" -->
        <!--       ></v-btn> -->
        <!--     </template> -->
        <!--   </v-list-item> -->
        <!-- </v-list> -->

        <v-divider />

        <v-list item-props :items="items" lines="three">
          <template #subtitle="{ subtitle }">
            <div v-html="subtitle" />
          </template>
        </v-list>

        <!-- <v-card-actions> -->
        <!--   <v-spacer></v-spacer> -->

        <!--   <v-btn variant="text" @click="menu = false"> Cancel </v-btn> -->
        <!--   <v-btn color="primary" variant="text" @click="menu = false"> -->
        <!--     Save -->
        <!--   </v-btn> -->
        <!-- </v-card-actions> -->
      </v-card>
    </v-menu>
  </div>
</template>
<script setup>
  import { ref } from 'vue'

  const fav = ref(true)
  const menu = ref(false)
  const message = ref(false)
  const hints = ref(true)
  const stock = product => Math.floor(Math.random() * 25)

  const items = [
    { type: 'subheader', title: 'Low Stock Alerts' },
    {
      prependIcon: 'mdi-tshirt-crew',
      title: 'Lady T-Shirt',
      subtitle: `Code: 8000000135892 &mdash; Only ${stock()} items left &mdash; $22.00`,
    },
    { type: 'divider', inset: true },
    {
      prependIcon: 'mdi-bag-personal',
      title: 'Travel Bag',
      subtitle: `Code: 8000000246713 &mdash; Only ${stock()} items left &mdash; $90.00`,
    },
    { type: 'divider', inset: true },
    {
      prependIcon: 'mdi-hanger',
      title: 'Kids Boy Pants',
      subtitle: `Code: 8000000357424 &mdash; Only ${stock()} items left &mdash; $49.99`,
    },
    { type: 'divider', inset: true },
    {
      prependIcon: 'mdi-hanger',
      title: 'Baby Shorts',
      subtitle: `Code: 8000000468135 &mdash; Only ${stock()} items left &mdash; $39.99`,
    },
    { type: 'divider', inset: true },
    {
      prependIcon: 'mdi-hanger',
      title: 'Kids Sweatshirt',
      subtitle: `Code: 8000000579246 &mdash; Only ${stock()} items left &mdash; $39.00`,
    },
    { type: 'divider', inset: true },
    {
      prependIcon: 'mdi-shoe-sneaker',
      title: 'Men Sneakers',
      subtitle: `Code: 8000000680157 &mdash; Only ${stock()} items left &mdash; $35.00`,
    },
  ]
</script>
