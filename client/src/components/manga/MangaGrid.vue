<template>
  <section class="manga-grid-wrap" :aria-busy="loading">
    <p class="manga-grid-count" role="status" aria-live="polite">
      <template v-if="!loading && !error">{{ t("browse.grid.count", { n: items.length }) }}</template>
    </p>

    <SkeletonRegion v-if="loading" class="manga-grid">
      <div v-for="n in SKELETON_COUNT" :key="n" class="manga-grid__skeleton">
        <SkeletonBlock radius="12px" aspect="2 / 3" />
        <SkeletonBlock width="80%" />
      </div>
    </SkeletonRegion>

    <AppErrorState v-else-if="error" :title="t('browse.grid.error')" @retry="emit('retry')" />

    <AppEmptyState v-else-if="!items.length && !hasNext" :title="t('browse.grid.empty')" :description="t('browse.grid.emptyDescription')">
      <template #action>
        <AppButton variant="outline" size="sm" @click="emit('reset')">{{ t("browse.grid.reset") }}</AppButton>
      </template>
    </AppEmptyState>

    <template v-else>
      <div class="manga-grid">
        <MangaCard v-for="item in items" :key="item.id" :item="item" />
      </div>

      <div v-if="hasNext" ref="sentinel" class="manga-grid__more">
        <AppButton variant="outline" :loading="fetchingNext" @click="emit('load-more')">{{ t("browse.grid.loadMore") }}</AppButton>
        <span v-if="fetchingNext" class="sr-only" role="status">{{ t("browse.grid.loadingMore") }}</span>
      </div>
      <p v-else-if="capped" class="manga-grid__end">{{ t("manga.capped") }}</p>
      <p v-else-if="items.length" class="manga-grid__end">{{ t("browse.grid.end") }}</p>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import MangaCard from "./MangaCard.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import { useInfiniteScroll } from "@/composables/useInfiniteScroll"
import type { MangaItem } from "@/types/manga"

const props = defineProps<{
  items: MangaItem[]
  loading: boolean
  error: boolean
  hasNext: boolean
  fetchingNext: boolean
  /** The page cap stopped a list the API could still continue. */
  capped: boolean
}>()
const emit = defineEmits<{ retry: []; reset: []; "load-more": [] }>()

const SKELETON_COUNT = 18

const { t } = useI18n()
const sentinel = useTemplateRef<HTMLElement>("sentinel")
useInfiniteScroll(sentinel, () => emit("load-more"), {
  disabled: computed(() => props.fetchingNext || !props.hasNext),
  rootMargin: "400px",
})
</script>

<style scoped>
.manga-grid-count {
  min-height: 1.25rem;
  margin: 0 0 0.75rem;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.manga-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.25rem 0.75rem;
}
.manga-grid__skeleton {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.manga-grid__more {
  display: flex;
  justify-content: center;
  padding: 2rem 0;
}
.manga-grid__end {
  margin: 0;
  padding: 2rem 0;
  text-align: center;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
@media (min-width: 480px) {
  .manga-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .manga-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1.5rem 1rem;
  }
}
@media (min-width: 1024px) {
  .manga-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}
@media (min-width: 1280px) {
  .manga-grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
