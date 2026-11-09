<template>
  <div class="filter-bar">
    <AppTabs
      v-if="!searching"
      :model-value="state.tab"
      :tabs="tabs"
      :label="t('browse.tabs.label')"
      :controls="false"
      class="filter-bar__tabs"
      @update:model-value="emit('update', { tab: $event })"
    />
    <div class="filter-bar__row">
      <div v-if="!searching" class="genres" role="group" :aria-label="t('browse.genre.label')">
        <button
          type="button"
          class="chip"
          :class="{ 'is-active': !state.genre }"
          :aria-pressed="!state.genre"
          @click="emit('update', { genre: null })"
        >
          {{ t("browse.genre.all") }}
        </button>
        <button
          v-for="g in genres"
          :key="g.id"
          type="button"
          class="chip"
          :class="{ 'is-active': state.genre === g.id }"
          :aria-pressed="state.genre === g.id"
          @click="emit('update', { genre: g.id })"
        >
          {{ g.name }}
        </button>
      </div>
      <label class="sort">
        <span class="sort__label">{{ t("browse.sort.label") }}</span>
        <select class="sort__select" :value="state.sort" @change="onSort">
          <option v-for="s in BROWSE_SORTS" :key="s" :value="s">{{ t(`browse.sort.${s}`) }}</option>
        </select>
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import AppTabs from "@/components/ui/AppTabs.vue"
import type { HomeMediaType } from "@/components/home/homeMedia"
import type { Genre } from "@/types"
import { BROWSE_TABS } from "./browseConfig"
import type { BrowseState } from "./browseQuery"
import { BROWSE_SORTS, type BrowseSort } from "./browseSort"

const props = defineProps<{ state: BrowseState; type: HomeMediaType; genres: Genre[]; searching: boolean }>()
const emit = defineEmits<{ update: [patch: Partial<BrowseState>] }>()

const { t } = useI18n()

const tabs = computed(() =>
  BROWSE_TABS.map((key) => ({
    key,
    label: key === "now" ? t(`browse.tabs.now.${props.type}`) : t(`browse.tabs.${key}`),
  })),
)

function onSort(e: Event) {
  emit("update", { sort: (e.target as HTMLSelectElement).value as BrowseSort })
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.filter-bar__row {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.genres {
  display: flex;
  gap: 0.5rem;
  min-width: 0;
  overflow-x: auto;
  padding-bottom: 2px;
  scrollbar-width: none;
}
.genres::-webkit-scrollbar {
  display: none;
}
.chip {
  flex: none;
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
.sort {
  display: flex;
  flex: none;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.sort__select {
  min-height: 2rem;
  padding: 0 0.5rem;
  border-radius: var(--radius-btn);
  border: 1px solid var(--color-border-strong);
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  font: inherit;
  font-size: var(--text-small);
}

/* Phones keep the bar visible while scrolling through the grid. */
@media (max-width: 767.98px) {
  .filter-bar {
    position: sticky;
    top: 0;
    z-index: var(--z-sticky);
    margin-inline: -1rem;
    padding: 0.5rem 1rem;
    background: color-mix(in srgb, var(--color-surface) 92%, transparent);
    backdrop-filter: blur(8px);
    border-bottom: 1px solid var(--color-border);
  }
}
@media (min-width: 768px) {
  .filter-bar__row {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }
}
</style>
