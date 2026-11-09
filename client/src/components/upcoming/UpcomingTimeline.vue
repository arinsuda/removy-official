<template>
  <div class="timeline" :aria-busy="loading">
    <SkeletonRegion v-if="loading" class="timeline__grid">
      <div v-for="n in 12" :key="n" class="timeline__skeleton">
        <SkeletonBlock radius="12px" aspect="2 / 3" />
        <SkeletonBlock width="80%" />
      </div>
    </SkeletonRegion>

    <AppErrorState v-else-if="error" :title="t('upcoming.grid.error')" @retry="emit('retry')" />

    <AppEmptyState v-else-if="!groups.length" :title="t('upcoming.grid.empty')" :description="t('upcoming.grid.emptyDescription')" :icon="CalendarX" />

    <template v-else>
      <section v-for="group in groups" :key="group.id" class="timeline__group" :aria-labelledby="`up-${group.id}`">
        <h2 :id="`up-${group.id}`" class="timeline__heading">
          {{ group.kind === "month" ? group.monthLabel : t(`upcoming.groups.${group.kind}`) }}
          <span class="timeline__count" aria-hidden="true">{{ group.items.length }}</span>
        </h2>
        <div class="timeline__grid">
          <HomePosterCard
            v-for="item in group.items"
            :key="item.key"
            :item="item"
            :rail="`upcoming:${group.id}`"
            :caption="caption(item.date)"
            :sizes="SIZES"
          />
        </div>
      </section>

      <div v-if="hasNext" ref="sentinel" class="timeline__more">
        <AppButton variant="outline" :loading="fetchingNext" @click="emit('load-more')">{{ t("upcoming.grid.loadMore") }}</AppButton>
        <span v-if="fetchingNext" class="sr-only" role="status">{{ t("upcoming.grid.loadingMore") }}</span>
      </div>
      <p v-else-if="capped" class="timeline__end">{{ t("upcoming.grid.capped", { n: MAX_PAGES * 20 }) }}</p>
      <p v-else class="timeline__end">{{ t("upcoming.grid.end") }}</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import { CalendarX } from "lucide-vue-next"
import { MAX_PAGES } from "@/components/browse/browseConfig"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import type { HomeMedia } from "@/components/home/homeMedia"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import { useInfiniteScroll } from "@/composables/useInfiniteScroll"
import { groupByRelease, parseLocalDate } from "./upcomingGroups"

const props = defineProps<{
  items: HomeMedia[]
  today: Date
  loading: boolean
  error: boolean
  hasNext: boolean
  fetchingNext: boolean
  /** The page cap stopped a list TMDB could still continue. */
  capped: boolean
}>()
const emit = defineEmits<{ retry: []; "load-more": [] }>()

const SIZES = "(min-width: 1280px) 190px, (min-width: 1024px) 18vw, (min-width: 768px) 23vw, (min-width: 480px) 31vw, 47vw"

const { t, locale } = useI18n()
const intlLocale = computed(() => locale.value)
const groups = computed(() => groupByRelease(props.items, props.today, intlLocale.value))
const dayFormat = computed(() => new Intl.DateTimeFormat(intlLocale.value, { weekday: "short", day: "numeric", month: "short" }))

function caption(date: string): string | undefined {
  const d = parseLocalDate(date)
  return d ? dayFormat.value.format(d) : undefined
}

const sentinel = useTemplateRef<HTMLElement>("sentinel")
useInfiniteScroll(sentinel, () => emit("load-more"), {
  disabled: computed(() => props.fetchingNext || !props.hasNext),
  rootMargin: "400px",
})
</script>

<style scoped>
.timeline {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.timeline__heading {
  position: sticky;
  top: 0;
  z-index: 5;
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  margin: 0 0 0.75rem;
  padding: 0.5rem 0;
  background: color-mix(in srgb, var(--color-surface) 92%, transparent);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--color-border);
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.timeline__count {
  font-size: var(--text-caption);
  font-weight: 500;
  color: var(--color-text-muted);
}
.timeline__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.25rem 0.75rem;
}
.timeline__skeleton {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.timeline__more {
  display: flex;
  justify-content: center;
  padding: 1rem 0;
}
.timeline__end {
  margin: 0;
  padding: 1rem 0 2rem;
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
  .timeline__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .timeline__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1.5rem 1rem;
  }
}
@media (min-width: 1024px) {
  .timeline__grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}
@media (min-width: 1280px) {
  .timeline__grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
