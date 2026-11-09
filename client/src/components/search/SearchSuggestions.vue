<template>
  <section class="sugg" :aria-label="t('media.search.suggestionsTitle')">
    <h2 class="sugg__title">{{ t("media.search.suggestionsTitle") }}</h2>
    <SkeletonRegion v-if="popular.isLoading.value" class="sugg__grid">
      <div v-for="n in 6" :key="n" class="sugg__skeleton">
        <SkeletonBlock radius="12px" aspect="2 / 3" />
        <SkeletonBlock width="80%" />
      </div>
    </SkeletonRegion>
    <div v-else class="sugg__grid">
      <HomePosterCard v-for="item in items" :key="item.key" :item="item" rail="search-suggestions" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useQuery } from "@tanstack/vue-query"
import { useI18n } from "vue-i18n"
import { movieApi } from "@/api/api"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import { fromMovie, fromSeries } from "@/components/home/homeMedia"
import { interleave } from "@/components/home/popcornHome"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"

const { t } = useI18n()

const popular = useQuery({
  queryKey: ["search", "suggestions"],
  queryFn: async () => {
    const [movies, series] = await Promise.allSettled([movieApi.getPopular(1), movieApi.getPopularSeries(1)])
    return {
      movies: movies.status === "fulfilled" ? movies.value.data.results.map(fromMovie) : [],
      series: series.status === "fulfilled" ? series.value.data.results.map(fromSeries) : [],
    }
  },
  staleTime: 5 * 60_000,
})

const items = computed(() => interleave(popular.data.value?.movies ?? [], popular.data.value?.series ?? []).slice(0, 12))
</script>

<style scoped>
.sugg {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 1.5rem;
}
.sugg__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.sugg__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.25rem 0.75rem;
}
.sugg__skeleton {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
@media (min-width: 640px) {
  .sugg__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .sugg__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1.5rem 1rem;
  }
}
@media (min-width: 1024px) {
  .sugg__grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
