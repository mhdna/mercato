<template>
  <v-dialog v-model="dialog" max-width="640">
    <template #activator="{ props: activatorProps }">
      <v-btn
        v-bind="activatorProps"
        prepend-icon="mdi-palette"
        text="Colors & Sizes"
        variant="tonal"
      />
    </template>

    <v-card class="px-2">
      <v-card-title>Manage Colors & Sizes</v-card-title>

      <v-tabs v-model="tab" color="primary">
        <v-tab value="color">Colors</v-tab>
        <v-tab value="size">Sizes</v-tab>
      </v-tabs>

      <v-card-text>
        <v-window v-model="tab">
          <v-window-item value="color">
            <form @submit.prevent="submitColor">
              <v-row align="center">
                <v-col cols="6">
                  <v-text-field
                    v-model="colorName.value.value"
                    density="compact"
                    :error-messages="colorName.errorMessage.value"
                    label="Color Name"
                  />
                </v-col>
                <v-col cols="6">
                  <v-text-field
                    v-model="colorHex.value.value"
                    density="compact"
                    :error-messages="colorHex.errorMessage.value"
                    label="Hex Value"
                    placeholder="#FF0000"
                  >
                    <template #append-inner>
                      <div
                        class="rounded"
                        :style="{ backgroundColor: isValidHex(colorHex.value.value) ? colorHex.value.value : 'transparent', width: '20px', height: '20px', border: '1px solid #999' }"
                      />
                    </template>
                  </v-text-field>
                </v-col>
              </v-row>

              <v-alert v-if="colorError" class="mb-4" type="error" variant="tonal">{{ colorError }}</v-alert>

              <v-btn color="primary" :loading="colorSubmitting" text="Add Color" type="submit" />
            </form>

            <v-divider class="my-4" />

            <div class="d-flex flex-wrap" style="gap: 8px;">
              <v-chip
                v-for="c in colors"
                :key="c.id"
                :color="c.hexValue"
                label
                variant="flat"
              >
                {{ c.name }}
              </v-chip>
            </div>
          </v-window-item>

          <v-window-item value="size">
            <form @submit.prevent="submitSize">
              <v-row>
                <v-col cols="4">
                  <v-text-field
                    v-model="sizeName.value.value"
                    density="compact"
                    :error-messages="sizeName.errorMessage.value"
                    label="Size Name"
                    placeholder="XL"
                  />
                </v-col>
                <v-col cols="4">
                  <v-text-field
                    v-model="sizeType.value.value"
                    density="compact"
                    :error-messages="sizeType.errorMessage.value"
                    label="Type"
                    placeholder="Shoes"
                  />
                </v-col>
                <v-col cols="4">
                  <v-text-field
                    v-model="sizeOrder.value.value"
                    density="compact"
                    :error-messages="sizeOrder.errorMessage.value"
                    label="Order"
                    placeholder="1"
                  />
                </v-col>
              </v-row>

              <v-alert v-if="sizeError" class="mb-4" type="error" variant="tonal">{{ sizeError }}</v-alert>

              <v-btn color="primary" :loading="sizeSubmitting" text="Add Size" type="submit" />
            </form>

            <v-divider class="my-4" />

            <div class="d-flex flex-wrap" style="gap: 8px;">
              <v-chip
                v-for="s in sizes"
                :key="s.id"
                label
              >
                {{ s.type }}: {{ s.name }}
              </v-chip>
            </div>
          </v-window-item>
        </v-window>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn text="Close" @click="dialog = false" />
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
  import { useField, useForm } from 'vee-validate'
  import { ref } from 'vue'
  import { useColorsAndSizes } from '@/composables/useColorsAndSizes'

  const { colors, sizes, fetchColors, fetchSizes, createColor, createSize } = useColorsAndSizes()

  const dialog = ref(false)
  const tab = ref('color')

  fetchColors()
  fetchSizes()

  function isValidHex (value) {
    return /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(value ?? '')
  }

  const { handleSubmit: handleColorSubmit, handleReset: resetColorForm } = useForm({
    validationSchema: {
      colorName (value) {
        if (value?.length >= 2) return true
        return 'Name needs to be at least 2 characters.'
      },
      colorHex (value) {
        if (isValidHex(value)) return true
        return 'Enter a valid hex color, e.g. #FF0000.'
      },
    },
  })
  const colorName = useField('colorName')
  const colorHex = useField('colorHex')
  const colorSubmitting = ref(false)
  const colorError = ref('')

  const submitColor = handleColorSubmit(async values => {
    colorSubmitting.value = true
    colorError.value = ''
    try {
      await createColor({ name: values.colorName, hexValue: values.colorHex })
      resetColorForm()
    } catch (error) {
      colorError.value = error.message
    } finally {
      colorSubmitting.value = false
    }
  })

  const { handleSubmit: handleSizeSubmit, handleReset: resetSizeForm } = useForm({
    validationSchema: {
      sizeName (value) {
        if (value?.length >= 1) return true
        return 'Name is required.'
      },
      sizeType (value) {
        if (value?.length >= 1) return true
        return 'Type is required.'
      },
      sizeOrder (value) {
        if (value?.length >= 1) return true
        return 'Order is required.'
      },
    },
  })
  const sizeName = useField('sizeName')
  const sizeType = useField('sizeType')
  const sizeOrder = useField('sizeOrder')
  const sizeSubmitting = ref(false)
  const sizeError = ref('')

  const submitSize = handleSizeSubmit(async values => {
    sizeSubmitting.value = true
    sizeError.value = ''
    try {
      await createSize({ name: values.sizeName, type: values.sizeType, order: values.sizeOrder })
      resetSizeForm()
    } catch (error) {
      sizeError.value = error.message
    } finally {
      sizeSubmitting.value = false
    }
  })
</script>
