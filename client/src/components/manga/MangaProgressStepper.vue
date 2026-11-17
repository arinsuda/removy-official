<template>
  <div class="manga-progress-stepper">
    <span class="manga-progress-stepper__label">{{ label }}</span>
    <div class="manga-progress-stepper__controls">
      <AppButton
        variant="outline"
        size="sm"
        icon-only
        :aria-label="decreaseLabel"
        :disabled="!canDecrement"
        @click="emit('step', -1)"
      >
        <Minus :size="16" aria-hidden="true" />
      </AppButton>
      <output class="manga-progress-stepper__value" aria-live="polite">{{ n(value) }}</output>
      <AppButton
        variant="outline"
        size="sm"
        icon-only
        :aria-label="increaseLabel"
        :disabled="!canIncrement"
        @click="emit('step', 1)"
      >
        <Plus :size="16" aria-hidden="true" />
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Minus, Plus } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"

defineProps<{
  label: string
  value: number
  decreaseLabel: string
  increaseLabel: string
  canDecrement: boolean
  canIncrement: boolean
}>()

const emit = defineEmits<{ step: [delta: number] }>()
const { n } = useI18n()
</script>

<style scoped>
.manga-progress-stepper {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  min-width: 0;
}
.manga-progress-stepper__label {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.manga-progress-stepper__controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.manga-progress-stepper__value {
  min-width: 3.5rem;
  text-align: center;
  font-size: var(--text-body);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-primary);
}
</style>
