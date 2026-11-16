<template>
  <div class="filter-bar">
    <AppTabs
      :model-value="state.type"
      :tabs="typeTabs"
      :label="t('anime.type.label')"
      :controls="false"
      @update:model-value="emit('update', { type: $event })"
    />
    <AppTabs
      :model-value="state.sort"
      :tabs="sortTabs"
      :label="t('anime.sort.label')"
      :controls="false"
      @update:model-value="emit('update', { sort: $event })"
    />
    <div class="filter-bar__row">
      <div class="chips" role="group" :aria-label="t('anime.keyword.label')">
        <button
          type="button"
          class="chip"
          :class="{ 'is-active': !state.keyword }"
          :aria-pressed="!state.keyword"
          @click="emit('update', { keyword: null })"
        >
          {{ t("anime.keyword.all") }}
        </button>
        <button
          v-for="keyword in ANIME_KEYWORDS"
          :key="keyword"
          type="button"
          class="chip"
          :class="{ 'is-active': state.keyword === keyword }"
          :aria-pressed="state.keyword === keyword"
          @click="emit('update', { keyword: state.keyword === keyword ? null : keyword })"
        >
          {{ t(`anime.keyword.${keyword}`) }}
        </button>
      </div>
      <div class="selects">
        <label class="select">
          <span class="select__label">{{ t("anime.genre.label") }}</span>
          <select class="select__input" :value="state.genre ?? ''" @change="onGenre">
            <option value="">{{ t("anime.genre.all") }}</option>
            <option v-for="g in genres" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </label>
        <label class="select">
          <span class="select__label">{{ t("anime.country.label") }}</span>
          <select class="select__input" :value="state.country" @change="onCountry">
            <option v-for="c in ANIME_COUNTRIES" :key="c" :value="c">{{ t(`anime.country.${c}`) }}</option>
          </select>
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import AppTabs from "@/components/ui/AppTabs.vue"
import type { AnimeCountry } from "@/types/anime"
import type { Genre } from "@/types"
import { ANIME_COUNTRIES, ANIME_KEYWORDS, ANIME_SORTS, ANIME_TYPES } from "./animeConfig"
import type { AnimeBrowseState } from "./animeQuery"

const props = defineProps<{ state: AnimeBrowseState; genres: Genre[] }>()
const emit = defineEmits<{ update: [patch: Partial<AnimeBrowseState>] }>()

const { t } = useI18n()

const typeTabs = computed(() => ANIME_TYPES.map((key) => ({ key, label: t(`anime.type.${key}`) })))
const sortTabs = computed(() =>
  ANIME_SORTS[props.state.type].map((key) => ({ key, label: t(`anime.sort.${key}`) })),
)

function onGenre(e: Event) {
  const value = (e.target as HTMLSelectElement).value
  emit("update", { genre: value ? Number(value) : null })
}

function onCountry(e: Event) {
  emit("update", { country: (e.target as HTMLSelectElement).value as AnimeCountry })
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.filter-bar__row,
.selects {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.chips {
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
@media (min-width: 768px) {
  .filter-bar__row {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }
  .selects {
    flex-direction: row;
    gap: 1rem;
  }
}
</style>
