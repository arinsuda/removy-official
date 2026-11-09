<template>
  <section class="lr" :aria-labelledby="titleId">
    <header class="lr__head">
      <h3 :id="titleId" class="lr__title">{{ title }}</h3>
      <RouterLink :to="seeAllTo" class="lr__all">{{ seeAllLabel }}<ChevronRight :size="14" aria-hidden="true" /></RouterLink>
    </header>
    <p v-if="!items.length" class="lr__empty">{{ emptyLabel }}</p>
    <div v-else class="lr__grid"><MediaGrid :items="items" :rail="rail" :label="title" /></div>
  </section>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { RouterLink, type RouteLocationRaw } from "vue-router"
import { ChevronRight } from "lucide-vue-next"
import MediaGrid from "@/components/profile/tabs/MediaGrid.vue"
import type { HomeMedia } from "@/components/home/homeMedia"

defineProps<{
  title: string
  seeAllLabel: string
  emptyLabel: string
  seeAllTo: RouteLocationRaw
  items: HomeMedia[]
  rail: string
}>()
const titleId = useId()
</script>

<style scoped>
.lr__head { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; margin-bottom: 0.75rem; }
.lr__title { margin: 0; font-size: var(--text-title); font-weight: 700; color: var(--color-text-primary); }
.lr__all { display: inline-flex; align-items: center; gap: 0.125rem; min-height: 2.75rem; font-size: var(--text-caption); color: var(--color-text-secondary); text-decoration: none; }
.lr__empty { margin: 0; padding: 1.5rem 1rem; font-size: var(--text-small); text-align: center; color: var(--color-text-muted); background: var(--color-surface-2); border: 1px dashed var(--color-border-strong); border-radius: var(--radius-card); }
@media (hover: hover) {
  .lr__all:hover { color: var(--color-text-primary); }
}
/* The grid shows 2 columns on phones and 5 on wide screens; keep rows full by trimming the extra card. */
.lr__grid { container-type: inline-size; }
@container (min-width: 60rem) {
  .lr__grid :deep(.mg__cell:nth-child(n + 6)) { display: none; }
}
@container (min-width: 44rem) and (max-width: 59.99rem) {
  .lr__grid :deep(.mg__cell:nth-child(n + 5)) { display: none; }
}
@container (min-width: 30rem) and (max-width: 43.99rem) {
  .lr__grid :deep(.mg__cell:nth-child(n + 4)) { display: none; }
}
</style>
