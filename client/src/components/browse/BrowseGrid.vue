<template>
  <section class="browse-grid-wrap" :aria-busy="loading">
    <p class="browse-count" role="status" aria-live="polite">
      <template v-if="!loading && !error">{{ t("browse.grid.count", { n: items.length }) }}</template>
    </p>

    <SkeletonRegion v-if="loading" class="browse-grid">
      <div v-for="n in 18" :key="n" class="browse-grid__skeleton">
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
      <div class="browse-grid">
        <HomePosterCard v-for="item in items" :key="item.key" :item="item" rail="browse" :sizes="SIZES" />
      </div>

      <div v-if="hasNext" ref="sentinel" class="browse-more">
        <AppButton variant="outline" :loading="fetchingNext" @click="emit('load-more')">{{ t("browse.grid.loadMore") }}</AppButton>
        <span v-if="fetchingNext" class="sr-only" role="status">{{ t("browse.grid.loadingMore") }}</span>
      </div>
      <p v-else-if="capped" class="browse-end">{{ t("browse.grid.capped", { n: MAX_PAGES * 20 }) }}</p>
      <p v-else-if="items.length" class="browse-end">{{ t("browse.grid.end") }}</p>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import type { HomeMedia } from "@/components/home/homeMedia"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import { useInfiniteScroll } from "@/composables/useInfiniteScroll"
import { MAX_PAGES } from "./browseConfig"

const props = defineProps<{
  items: HomeMedia[]
  loading: boolean
  error: boolean
  hasNext: boolean
  fetchingNext: boolean
  /** The page cap stopped a list TMDB could still continue. */
  capped: boolean
  /** Auto-load on scroll. Off for client-side sorts so cards do not jump under the reader. */
  infinite: boolean
}>()
const emit = defineEmits<{ retry: []; reset: []; "load-more": [] }>()

const SIZES = "(min-width: 1280px) 190px, (min-width: 1024px) 18vw, (min-width: 768px) 23vw, (min-width: 480px) 31vw, 47vw"

const { t } = useI18n()
const sentinel = useTemplateRef<HTMLElement>("sentinel")
useInfiniteScroll(sentinel, () => emit("load-more"), {
  disabled: computed(() => !props.infinite || props.fetchingNext || !props.hasNext),
  rootMargin: "400px",
})
</script>

<style scoped>
.browse-count {
  min-height: 1.25rem;
  margin: 0 0 0.75rem;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.browse-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.25rem 0.75rem;
}
.browse-grid__skeleton {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.browse-more {
  display: flex;
  justify-content: center;
  padding: 2rem 0;
}
.browse-end {
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
  .browse-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .browse-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1.5rem 1rem;
  }
}
@media (min-width: 1024px) {
  .browse-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}
@media (min-width: 1280px) {
  .browse-grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
