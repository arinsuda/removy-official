<template>
  <RouterLink :to="to" class="lc">
    <div class="lc__cover" aria-hidden="true">
      <template v-if="posters.length">
        <img v-for="(p, i) in posters" :key="`${p}-${i}`" :src="posterUrl(p)" alt="" class="lc__poster" loading="lazy" decoding="async" />
      </template>
      <component :is="icon" v-else :size="40" class="lc__empty-icon" />
    </div>
    <div class="lc__body">
      <h3 class="lc__name"><component :is="icon" :size="16" aria-hidden="true" />{{ t(`library.lists.names.${listKey}`) }}</h3>
      <p class="lc__desc">{{ t(`library.lists.descriptions.${listKey}`) }}</p>
      <p class="lc__meta">
        <span class="lc__count">{{ t("library.lists.count", { n: count }, count) }}</span>
        <span class="lc__date">{{ updatedAt ? t("library.lists.updated", { date: formatDate(updatedAt) }) : t("library.lists.noActivity") }}</span>
      </p>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { RouterLink, type RouteLocationRaw } from "vue-router"
import { useI18n } from "vue-i18n"
import { getTmdbImageUrl } from "@/utils/image"
import type { LibraryListKey } from "@/types"
import { COLLAGE_POSTER_SIZE, LIST_ICONS } from "../config"

const props = defineProps<{
  listKey: LibraryListKey
  to: RouteLocationRaw
  count: number
  posters: string[]
  updatedAt: string | null
}>()

const { t, locale } = useI18n()
const icon = computed(() => LIST_ICONS[props.listKey])
const posterUrl = (path: string) => getTmdbImageUrl(path, COLLAGE_POSTER_SIZE) ?? ""
const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso))
</script>

<style scoped>
.lc { display: flex; flex-direction: column; min-width: 0; overflow: hidden; color: inherit; text-decoration: none; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); transition: border-color var(--motion-fast) var(--ease-water), transform var(--motion-base) var(--ease-out-soft); }
.lc__cover { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); aspect-ratio: 2 / 1; background: var(--color-surface-3); }
.lc__poster { width: 100%; height: 100%; object-fit: cover; }
.lc__empty-icon { grid-column: 1 / -1; align-self: center; justify-self: center; color: var(--color-text-muted); }
.lc__body { display: flex; flex-direction: column; gap: 0.25rem; padding: 1rem; }
.lc__name { display: flex; align-items: center; gap: 0.5rem; margin: 0; font-size: var(--text-body); font-weight: 700; color: var(--color-text-primary); }
.lc__desc { margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.lc__meta { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 0.25rem 0.75rem; margin: 0.5rem 0 0; font-size: var(--text-caption); color: var(--color-text-secondary); }
.lc__count { font-weight: 600; color: var(--color-text-primary); }
@media (hover: hover) {
  .lc:hover { border-color: var(--color-border-strong); transform: translateY(-2px); }
}
@media (prefers-reduced-motion: reduce) {
  .lc { transition: none; }
  .lc:hover { transform: none; }
}
</style>
