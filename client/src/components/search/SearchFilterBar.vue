<template>
  <div class="sfb">
    <AppTabs :model-value="state.tab" :tabs="tabs" :label="t('media.search.tabs.label')" :controls="false" @update:model-value="emit('update', { tab: $event })">
      <template #tab="{ tab }">
        {{ tab.label }}
        <span class="sfb__count">{{ counts[tab.key] }}</span>
      </template>
    </AppTabs>

    <div v-if="state.tab !== 'people'" class="sfb__row">
      <label class="sfb__field">
        <span class="sfb__label">{{ t("media.search.filters.year") }}</span>
        <select class="sfb__select" :value="state.year ?? ''" @change="onYear">
          <option value="">{{ t("media.search.filters.anyYear") }}</option>
          <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
        </select>
      </label>
      <label class="sfb__field">
        <span class="sfb__label">{{ t("media.search.filters.genre") }}</span>
        <select class="sfb__select" :value="state.genre ?? ''" @change="onGenre">
          <option value="">{{ t("media.search.filters.anyGenre") }}</option>
          <option v-for="g in genres" :key="g.id" :value="g.id">{{ g.name }}</option>
        </select>
      </label>
      <label class="sfb__field">
        <span class="sfb__label">{{ t("browse.sort.label") }}</span>
        <select class="sfb__select" :value="state.sort" @change="onSort">
          <option v-for="s in BROWSE_SORTS" :key="s" :value="s">{{ t(`browse.sort.${s}`) }}</option>
        </select>
      </label>
      <AppButton v-if="state.year || state.genre" variant="ghost" size="sm" @click="emit('update', { year: null, genre: null })">
        {{ t("media.search.filters.clear") }}
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { BROWSE_SORTS, type BrowseSort } from "@/components/browse/browseSort"
import AppButton from "@/components/ui/AppButton.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import { SEARCH_TABS, type SearchState, type SearchTab } from "./searchQuery"

defineProps<{
  state: SearchState
  counts: Record<SearchTab, number>
  years: number[]
  genres: { id: number; name: string }[]
}>()
const emit = defineEmits<{ update: [patch: Partial<SearchState>] }>()

const { t } = useI18n()
const tabs = computed(() => SEARCH_TABS.map((key) => ({ key, label: t(`media.search.tabs.${key}`) })))

const value = (e: Event) => (e.target as HTMLSelectElement).value
const onYear = (e: Event) => emit("update", { year: Number(value(e)) || null })
const onGenre = (e: Event) => emit("update", { genre: Number(value(e)) || null })
const onSort = (e: Event) => emit("update", { sort: value(e) as BrowseSort })
</script>

<style scoped>
.sfb {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.sfb__count {
  margin-inline-start: 0.375rem;
  font-size: var(--text-caption);
  opacity: 0.7;
  font-variant-numeric: tabular-nums;
}
.sfb__row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
}
.sfb__field {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.sfb__select {
  min-height: 2rem;
  max-width: 11rem;
  padding: 0 0.5rem;
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-btn);
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  font: inherit;
  font-size: var(--text-small);
}
@media (max-width: 479.98px) {
  .sfb__field {
    flex: 1 1 100%;
    justify-content: space-between;
  }
  .sfb__select {
    max-width: none;
    flex: 1;
  }
}
</style>
