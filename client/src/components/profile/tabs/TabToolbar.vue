<template>
  <div class="tb">
    <div class="tb__row">
      <p class="tb__count" role="status" aria-live="polite">{{ countLabel }}</p>
      <div class="tb__tools">
        <label v-if="searchable" class="tb__search">
          <Search :size="14" aria-hidden="true" />
          <input v-model="query" type="search" class="tb__input" :placeholder="t('profile.tabContent.common.search')" :aria-label="t('profile.tabContent.common.search')" />
        </label>
        <label v-if="sorts.length" class="tb__sort">
          <span class="sr-only">{{ t("profile.tabContent.common.sort") }}</span>
          <select v-model="sort" class="tb__select" :aria-label="t('profile.tabContent.common.sort')">
            <option v-for="s in sorts" :key="s.key" :value="s.key">{{ s.label }}</option>
          </select>
        </label>
        <slot name="extra" />
        <div v-if="viewToggle" class="tb__view" role="group" :aria-label="t('profile.tabContent.common.view')">
          <button v-for="v in VIEWS" :key="v.key" type="button" class="tb__view-btn" :class="{ 'is-active': view === v.key }" :aria-pressed="view === v.key" :aria-label="viewLabels?.[v.key] ?? t(v.label)" @click="view = v.key">
            <component :is="v.icon" :size="15" aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>
    <div v-if="filters.length" class="tb__chips" role="group" :aria-label="t('profile.tabContent.common.filter')">
      <button v-for="f in filters" :key="f.key" type="button" class="tb__chip" :class="{ 'is-active': filter === f.key }" :aria-pressed="filter === f.key" @click="filter = f.key">
        {{ f.label }}<span v-if="f.count !== undefined" class="tb__chip-n">{{ f.count }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { LayoutGrid, List, Search } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import type { ViewMode } from "./useViewMode"

export interface ToolbarOption {
  key: string
  label: string
  count?: number
}

withDefaults(
  defineProps<{
    countLabel: string
    filters?: ToolbarOption[]
    sorts?: ToolbarOption[]
    searchable?: boolean
    viewToggle?: boolean
    /** Accessible names for the grid/list buttons when a tab calls them something else. */
    viewLabels?: Partial<Record<ViewMode, string>>
  }>(),
  { filters: () => [], sorts: () => [], searchable: true, viewToggle: false },
)

const query = defineModel<string>("query", { default: "" })
const sort = defineModel<string>("sort", { default: "" })
const filter = defineModel<string>("filter", { default: "all" })
const view = defineModel<ViewMode>("view", { default: "grid" })

const { t } = useI18n()
const VIEWS = [
  { key: "grid" as const, icon: LayoutGrid, label: "profile.tabContent.common.viewGrid" },
  { key: "list" as const, icon: List, label: "profile.tabContent.common.viewList" },
]
</script>

<style scoped>
.tb { display: flex; flex-direction: column; gap: 0.75rem; margin-bottom: 1.25rem; }
.tb__row { display: flex; flex-direction: column; gap: 0.625rem; }
.tb__count { margin: 0; font-size: var(--text-small); font-weight: 600; color: var(--color-text-secondary); }
.tb__tools { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
.tb__search { display: flex; flex: 1 1 11rem; align-items: center; gap: 0.5rem; min-height: 2.25rem; padding: 0 0.75rem; color: var(--color-text-muted); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); }
.tb__search:focus-within { border-color: var(--color-border-strong); }
.tb__input { flex: 1; min-width: 0; font: inherit; font-size: var(--text-small); color: var(--color-text-primary); background: none; border: 0; outline: 0; }
.tb__input::placeholder { color: var(--color-text-muted); }
.tb__select { min-height: 2.25rem; padding: 0 0.75rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); cursor: pointer; }
.tb__view { display: inline-flex; padding: 2px; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); }
.tb__view-btn { display: grid; place-items: center; width: 2rem; height: 2rem; color: var(--color-text-muted); background: none; border: 0; border-radius: var(--radius-btn); cursor: pointer; transition: color var(--motion-fast) var(--ease-water), background-color var(--motion-fast) var(--ease-water); }
.tb__view-btn.is-active { color: var(--color-text-primary); background: var(--color-surface-3); }
.tb__chips { display: flex; gap: 0.5rem; overflow-x: auto; padding-bottom: 2px; scrollbar-width: none; }
.tb__chips::-webkit-scrollbar { display: none; }
.tb__chip { display: inline-flex; flex: none; align-items: center; gap: 0.4rem; min-height: 2rem; padding: 0 0.875rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); white-space: nowrap; background: none; border: 1px solid var(--color-border-strong); border-radius: 9999px; cursor: pointer; transition: background-color var(--motion-fast) var(--ease-water), color var(--motion-fast) var(--ease-water); }
.tb__chip.is-active { color: var(--color-on-brand); background: var(--color-brand); border-color: var(--color-brand); }
.tb__chip-n { font-size: 0.6875rem; opacity: 0.8; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
@media (hover: hover) {
  .tb__chip:not(.is-active):hover, .tb__view-btn:not(.is-active):hover { background: var(--color-surface-3); }
}
@media (min-width: 768px) {
  .tb__row { flex-direction: row; align-items: center; justify-content: space-between; }
  .tb__tools { flex-wrap: nowrap; }
  .tb__search { flex: 0 1 14rem; }
}
@media (prefers-reduced-motion: reduce) {
  .tb__chip, .tb__view-btn { transition: none; }
}
</style>
