<template>
  <div class="mfb">
    <AppTabs
      :model-value="state.country"
      :tabs="countryTabs"
      :label="t('manga.country.label')"
      :controls="false"
      @update:model-value="emit('update', { country: $event })"
    />
    <AppTabs
      :model-value="state.sort"
      :tabs="sortTabs"
      :label="t('manga.sort.label')"
      :controls="false"
      @update:model-value="emit('update', { sort: $event })"
    />
    <label v-if="adultAllowed" class="adult-toggle">
      <input
        type="checkbox"
        class="adult-toggle__input"
        :checked="state.adult && acknowledged"
        @change="onAdult"
      />
      <span>{{ t("manga.adult.toggle") }}</span>
    </label>
    <div class="mfb__row">
      <div class="chips" role="group" :aria-label="t('manga.format.label')">
        <button
          v-for="format in MANGA_FORMATS"
          :key="format"
          type="button"
          class="chip"
          :class="{ 'is-active': state.format === format }"
          :aria-pressed="state.format === format"
          @click="emit('update', { format })"
        >
          {{ t(`manga.format.${format}`) }}
        </button>
      </div>
      <div class="selects">
        <label class="select">
          <span class="select__label">{{ t("manga.genre.label") }}</span>
          <select class="select__input" :value="state.genre ?? ''" @change="onGenre">
            <option value="">{{ t("manga.genre.all") }}</option>
            <option v-for="genre in genres" :key="genre" :value="genre">{{ genre }}</option>
          </select>
        </label>
        <label class="select">
          <span class="select__label">{{ t("manga.status.label") }}</span>
          <select class="select__input" :value="state.status" @change="onStatus">
            <option v-for="status in MANGA_STATUSES" :key="status" :value="status">{{ t(`manga.status.${status}`) }}</option>
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
import type { MangaStatus } from "@/types/manga"
import { MANGA_COUNTRIES, MANGA_FORMATS, MANGA_SORTS, MANGA_STATUSES } from "./mangaConfig"
import type { MangaBrowseState } from "./mangaQuery"

const props = defineProps<{
  state: MangaBrowseState
  genres: string[]
  /** The toggle exists only for signed-in viewers who are 18+. */
  adultAllowed: boolean
  /** Turning adult titles on for an unacknowledged viewer goes through the confirm modal first. */
  acknowledged: boolean
}>()
const emit = defineEmits<{ update: [patch: Partial<MangaBrowseState>]; "request-adult": [] }>()

const { t } = useI18n()

const countryTabs = computed(() => MANGA_COUNTRIES.map((key) => ({ key, label: t(`manga.country.${key}`) })))
const sortTabs = computed(() => MANGA_SORTS.map((key) => ({ key, label: t(`manga.sort.${key}`) })))

function onGenre(e: Event) {
  const value = (e.target as HTMLSelectElement).value
  emit("update", { genre: value || null })
}

function onAdult(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.checked && !props.acknowledged) {
    // Keep the box unchecked until the modal confirms; the state change is what sends include_adult.
    input.checked = false
    emit("request-adult")
    return
  }
  emit("update", { adult: input.checked })
}

function onStatus(e: Event) {
  emit("update", { status: (e.target as HTMLSelectElement).value as MangaStatus })
}
</script>

<style scoped>
.mfb {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.mfb__row,
.selects {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.adult-toggle {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  gap: 0.5rem;
  min-height: 2rem;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  cursor: pointer;
}
.adult-toggle__input {
  width: 1rem;
  height: 1rem;
  accent-color: var(--color-brand);
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.chip {
  min-height: 2rem;
  padding: 0 0.875rem;
  border-radius: var(--radius-full);
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
  .mfb__row {
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
