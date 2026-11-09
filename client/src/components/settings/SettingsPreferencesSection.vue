<template>
  <SettingsCard id="preferences" :title="t('settings.preferences.title')" :description="t('settings.preferences.desc')">
    <div class="chips" role="group" :aria-label="t('settings.preferences.title')">
      <button
        v-for="id in KNOWN_GENRE_IDS"
        :key="id"
        type="button"
        class="chip"
        :class="{ 'is-on': model.includes(id) }"
        :aria-pressed="model.includes(id)"
        @click="toggle(id)"
      >
        {{ t(`onboarding.genres.${id}`) }}
      </button>
    </div>
  </SettingsCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { KNOWN_GENRE_IDS } from "@/components/onboarding/onboardingConfig"
import SettingsCard from "./SettingsCard.vue"

const model = defineModel<number[]>({ required: true })
const { t } = useI18n()

// Keep the same order as the list so the dirty check can compare arrays.
function toggle(id: number) {
  const next = model.value.includes(id) ? model.value.filter(g => g !== id) : [...model.value, id]
  model.value = KNOWN_GENRE_IDS.filter(g => next.includes(g)).concat(next.filter(g => !KNOWN_GENRE_IDS.includes(g)))
}
</script>

<style scoped>
.chips { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.chip { padding: 0.375rem 0.875rem; font-size: 0.8125rem; color: var(--color-text-secondary); cursor: pointer; background: none; border: 1px solid var(--color-border-strong); border-radius: 999px; }
.chip:hover { color: var(--color-text-primary); }
.chip.is-on { color: var(--color-surface); background: var(--color-text-primary); border-color: var(--color-text-primary); }
</style>
