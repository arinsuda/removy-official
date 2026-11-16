<template>
  <div class="release">
    <div class="chips" role="group" :aria-label="t('anime.release.decadeLabel')">
      <button
        type="button"
        class="chip"
        :class="{ 'is-active': state.decade === null && state.year === null }"
        :aria-pressed="state.decade === null && state.year === null"
        @click="emit('update', { decade: null, year: null })"
      >
        {{ t("anime.release.any") }}
      </button>
      <button
        v-for="decade in decades"
        :key="decade"
        type="button"
        class="chip"
        :class="{ 'is-active': state.decade === decade }"
        :aria-pressed="state.decade === decade"
        @click="emit('update', { decade: state.decade === decade ? null : decade })"
      >
        {{ t("anime.release.decade", { year: decade }) }}
      </button>
    </div>
    <div class="selects">
      <label class="select">
        <span class="select__label">{{ t("anime.release.year") }}</span>
        <select class="select__input" :value="state.year ?? ''" @change="onYear">
          <option value="">{{ t("anime.release.anyYear") }}</option>
          <option v-for="year in years" :key="year" :value="year">{{ year }}</option>
        </select>
      </label>
      <!-- Season is TV only and needs a year: the API rejects it otherwise. -->
      <label v-if="state.type === 'tv'" class="select">
        <span class="select__label">{{ t("anime.release.season") }}</span>
        <select class="select__input" :value="state.season ?? ''" :disabled="state.year === null" @change="onSeason">
          <option value="">{{ t("anime.release.anySeason") }}</option>
          <option v-for="season in ANIME_SEASONS" :key="season" :value="season">{{ t(`anime.season.${season}`) }}</option>
        </select>
      </label>
    </div>
    <p v-if="state.type === 'tv' && state.year === null" class="release__hint">{{ t("anime.release.seasonHint") }}</p>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import type { AnimeSeason } from "@/types/anime"
import { ANIME_SEASONS, animeDecades, animeYears } from "./animeConfig"
import type { AnimeBrowseState } from "./animeQuery"

defineProps<{ state: AnimeBrowseState }>()
const emit = defineEmits<{ update: [patch: Partial<AnimeBrowseState>] }>()

const { t } = useI18n()
const currentYear = new Date().getFullYear()
const decades = animeDecades(currentYear)
const years = animeYears(currentYear)

function onYear(e: Event) {
  const value = (e.target as HTMLSelectElement).value
  emit("update", { year: value ? Number(value) : null })
}

function onSeason(e: Event) {
  const value = (e.target as HTMLSelectElement).value
  emit("update", { season: value ? (value as AnimeSeason) : null })
}
</script>

<style scoped>
.release {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.chips,
.selects {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.chip {
  min-height: 2rem;
  padding: 0 0.875rem;
  border-radius: 9999px;
  border: 1px solid var(--color-border-strong);
  background: transparent;
  color: var(--color-text-secondary);
  font: inherit;
  font-size: var(--text-caption);
  white-space: nowrap;
  cursor: pointer;
}
.chip:hover {
  background: var(--color-surface-3);
}
.chip.is-active {
  background: var(--color-brand);
  border-color: var(--color-brand);
  color: var(--color-on-brand);
}
.select {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.select__input {
  min-height: 2rem;
  min-width: 0;
  padding: 0 0.5rem;
  border-radius: var(--radius-btn);
  border: 1px solid var(--color-border-strong);
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  font: inherit;
  font-size: var(--text-small);
}
.select__input:disabled {
  opacity: 0.5;
}
.release__hint {
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
</style>
