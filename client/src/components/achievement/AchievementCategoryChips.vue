<template>
  <div class="cc" role="group" :aria-label="t('achievements.toolbar.categories')">
    <button type="button" class="cc__chip" :class="{ 'is-active': model === 'all' }" :aria-pressed="model === 'all'" @click="model = 'all'">
      <LayoutGrid :size="14" aria-hidden="true" />{{ t("achievements.categories.all") }}
    </button>
    <button v-for="c in categories" :key="c" type="button" class="cc__chip" :class="{ 'is-active': model === c }" :aria-pressed="model === c" @click="model = c">
      <component :is="CATEGORIES[c].icon" :size="14" aria-hidden="true" />{{ t(`achievements.categories.${c}`) }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { LayoutGrid } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import { CATEGORIES, type CategoryKey } from "./config"

defineProps<{ categories: CategoryKey[] }>()
const model = defineModel<CategoryKey | "all">({ default: "all" })
const { t } = useI18n()
</script>

<style scoped>
.cc { display: flex; gap: 0.5rem; overflow-x: auto; padding-bottom: 2px; scrollbar-width: none; }
.cc::-webkit-scrollbar { display: none; }
.cc__chip { display: inline-flex; flex: none; align-items: center; gap: 0.4rem; min-height: 2.25rem; padding: 0 0.875rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); white-space: nowrap; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); cursor: pointer; transition: background-color var(--motion-fast) var(--ease-water), color var(--motion-fast) var(--ease-water); }
.cc__chip.is-active { color: var(--color-text-primary); background: var(--color-surface-3); border-color: var(--color-text-secondary); }
@media (hover: hover) {
  .cc__chip:not(.is-active):hover { background: var(--color-surface-3); }
}
@media (prefers-reduced-motion: reduce) {
  .cc__chip { transition: none; }
}
</style>
