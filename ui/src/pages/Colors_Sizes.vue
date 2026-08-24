<template>
  <v-card class="mx-4 my-2 px-4" max-height="585" style="overflow-y: auto;" title="Colors">
    <v-row>
      <v-col cols="3">
        <v-color-picker class="ma-2" swatches-max-height="450px" />
        <v-text-field label="Name" prepend-icon="mdi-text" type="name" />
        <v-btn color="green">Add Color</v-btn>
      </v-col>
      <v-col cols="9">
        <div class="py-2 mx-4" style="display:flex; gap:10px; flex-wrap:wrap;">
          <v-card
            v-for="c in colors"
            :key="c.hex"
            class="d-flex align-center justify-center"
            :color="c.hex"
            height="40"
            width="86"
          >
            {{ c.name }}
          </v-card>
        </div>
      </v-col>
    </v-row>
  </v-card>
  <v-card class="ma-2 my-4 pb-6" title="Sizes">
    <v-sheet class="mx-auto d-flex">
      <div class="d-flex justify-middle align-center mx-15 font-weight-bold">
        Shoes:
      </div>
      <v-slide-group show-arrows>
        <v-slide-group-item v-for="n in 25" :key="n" v-slot="{ isSelected, toggle }">
          <v-btn class="ma-2" :color="isSelected ? 'primary' : undefined" rounded @click="toggle">
            {{ n }}
          </v-btn>
        </v-slide-group-item>
      </v-slide-group>
    </v-sheet>
    <v-sheet class="mx-auto d-flex">
      <div class="d-flex justify-middle align-center mx-15 font-weight-bold">
        Pants:
      </div>
      <v-slide-group show-arrows>
        <v-slide-group-item v-for="n in 25" :key="n" v-slot="{ isSelected, toggle }">
          <v-btn class="ma-2" :color="isSelected ? 'primary' : undefined" rounded @click="toggle">
            {{ n }}
          </v-btn>
        </v-slide-group-item>
      </v-slide-group>
    </v-sheet>
  </v-card>
</template>

<script setup>
  function hslToHex (h, s, l) {
    s /= 100
    l /= 100

    const k = n => (n + h / 30) % 12
    const a = s * Math.min(l, 1 - l)
    const f = n =>
      l - a * Math.max(-1, Math.min(k(n) - 3, Math.min(9 - k(n), 1)))

    const toHex = x =>
      Math.round(255 * x).toString(16).padStart(2, '0')

    return `#${toHex(f(0))}${toHex(f(8))}${toHex(f(4))}`
  }

  const colors = [
    { name: 'Black', hex: '#000000' },
    { name: 'Charcoal', hex: '#36454F' },
    { name: 'Jet Black', hex: '#0A0A0A' },
    { name: 'Graphite', hex: '#383838' },
    { name: 'Ash Gray', hex: '#B2BEB5' },
    { name: 'Light Gray', hex: '#D3D3D3' },
    { name: 'Silver', hex: '#C0C0C0' },
    { name: 'White', hex: '#FFFFFF' },
    { name: 'Ivory', hex: '#FFFFF0' },
    { name: 'Cream', hex: '#FFFDD0' },

    { name: 'Beige', hex: '#F5F5DC' },
    { name: 'Sand', hex: '#C2B280' },
    { name: 'Khaki', hex: '#C3B091' },
    { name: 'Camel', hex: '#C19A6B' },
    { name: 'Tan', hex: '#D2B48C' },
    { name: 'Mocha', hex: '#967969' },
    { name: 'Chocolate', hex: '#7B3F00' },
    { name: 'Coffee', hex: '#6F4E37' },
    { name: 'Espresso', hex: '#4B3621' },
    { name: 'Chestnut', hex: '#954535' },

    { name: 'Navy', hex: '#000080' },
    { name: 'Midnight Blue', hex: '#191970' },
    { name: 'Denim', hex: '#1560BD' },
    { name: 'Sky Blue', hex: '#87CEEB' },
    { name: 'Powder Blue', hex: '#B0E0E6' },
    { name: 'Baby Blue', hex: '#89CFF0' },
    { name: 'Steel Blue', hex: '#4682B4' },
    { name: 'Slate Blue', hex: '#6A5ACD' },
    { name: 'Royal Blue', hex: '#4169E1' },
    { name: 'Cobalt', hex: '#0047AB' },

    { name: 'Forest Green', hex: '#228B22' },
    { name: 'Olive', hex: '#808000' },
    { name: 'Army Green', hex: '#4B5320' },
    { name: 'Sage', hex: '#BCB88A' },
    { name: 'Mint', hex: '#98FF98' },
    { name: 'Seafoam', hex: '#9FE2BF' },
    { name: 'Emerald', hex: '#50C878' },
    { name: 'Hunter Green', hex: '#355E3B' },
    { name: 'Pine', hex: '#01796F' },
    { name: 'Moss', hex: '#8A9A5B' },

    { name: 'Red', hex: '#FF0000' },
    { name: 'Crimson', hex: '#DC143C' },
    { name: 'Burgundy', hex: '#800020' },
    { name: 'Maroon', hex: '#800000' },
    { name: 'Wine', hex: '#722F37' },
    { name: 'Rose', hex: '#FF007F' },
    { name: 'Blush', hex: '#DE5D83' },
    { name: 'Coral', hex: '#FF7F50' },
    { name: 'Salmon', hex: '#FA8072' },
    { name: 'Brick', hex: '#CB4154' },

    { name: 'Orange', hex: '#FFA500' },
    { name: 'Burnt Orange', hex: '#CC5500' },
    { name: 'Rust', hex: '#B7410E' },
    { name: 'Terracotta', hex: '#E2725B' },
    { name: 'Peach', hex: '#FFE5B4' },
    { name: 'Apricot', hex: '#FBCEB1' },
    { name: 'Amber', hex: '#FFBF00' },
    { name: 'Mustard', hex: '#FFDB58' },
    { name: 'Goldenrod', hex: '#DAA520' },
    { name: 'Honey', hex: '#FFC30B' },

    { name: 'Purple', hex: '#800080' },
    { name: 'Lavender', hex: '#E6E6FA' },
    { name: 'Lilac', hex: '#C8A2C8' },
    { name: 'Violet', hex: '#8F00FF' },
    { name: 'Plum', hex: '#8E4585' },
    { name: 'Mauve', hex: '#E0B0FF' },
    { name: 'Orchid', hex: '#DA70D6' },
    { name: 'Amethyst', hex: '#9966CC' },
    { name: 'Mulberry', hex: '#70193D' },
    { name: 'Eggplant', hex: '#614051' },

    { name: 'Pink', hex: '#FFC0CB' },
    { name: 'Hot Pink', hex: '#FF69B4' },
    { name: 'Fuchsia', hex: '#FF00FF' },
    { name: 'Magenta', hex: '#FF0090' },
    { name: 'Bubblegum', hex: '#FFC1CC' },
    { name: 'Dusty Rose', hex: '#C08081' },
    { name: 'Cotton Candy', hex: '#FFBCD9' },
    { name: 'Rosewood', hex: '#65000B' },
    { name: 'Cherry', hex: '#DE3163' },
    { name: 'Raspberry', hex: '#E30B5C' },

    { name: 'Gold', hex: '#FFD700' },
    { name: 'Champagne', hex: '#F7E7CE' },
    { name: 'Bronze', hex: '#CD7F32' },
    { name: 'Copper', hex: '#B87333' },
    { name: 'Pearl', hex: '#FDEEF4' },
    { name: 'Stone', hex: '#928E85' },
    { name: 'Clay', hex: '#B66A50' },
    { name: 'Smoke', hex: '#738276' },
    { name: 'Ice', hex: '#E0F7FA' },
    { name: 'Ocean', hex: '#4F42B5' },
  ]
</script>
