<template>
  <div class="home">
    <HomeHero :items="heroItems" />

    <div class="home__content">
      <HomeGenreChips />

      <HomeRail
        v-if="hasGenres"
        rail="recommended"
        :title="t('home.recommended.title')"
        :subtitle="t('home.recommended.subtitle')"
        :loading="recommended.isLoading.value"
        :error="recommended.isError.value"
        :empty="!recommendedItems.length"
        @retry="recommended.refetch()"
      >
        <HomePosterCard v-for="item in recommendedItems" :key="item.key" :item="item" rail="recommended" />
      </HomeRail>

      <HomeRail
        v-for="rail in rails"
        :key="rail.id"
        :title="t(`home.${rail.id}.title`)"
        :subtitle="t(`home.${rail.id}.subtitle`)"
        :see-all-to="{ name: 'movies' }"
        :loading="rail.query.isLoading.value"
        :error="rail.query.isError.value"
        :empty="!rail.items.value.length"
        @retry="rail.query.refetch()"
      >
        <HomePosterCard v-for="item in rail.items.value" :key="item.key" :item="item" :rail="rail.id" />
      </HomeRail>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useQuery } from "@tanstack/vue-query"
import { useI18n } from "vue-i18n"
import { movieApi } from "@/api/api"
import HomeGenreChips from "@/components/home/HomeGenreChips.vue"
import HomeHero from "@/components/home/HomeHero.vue"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import HomeRail from "@/components/home/HomeRail.vue"
import { fromMovie, type HomeMedia } from "@/components/home/homeMedia"
import { useFavoriteGenres } from "@/composables/useFavoriteGenres"
import { provideHomePopup } from "@/composables/useHomePopup"
import type { Movie, PaginatedResult } from "@/types"

const { t } = useI18n()
provideHomePopup()

const RAIL_SIZE = 10
const { hasGenres, genreQuery } = useFavoriteGenres()

function useMovieRail(id: string, key: string, fetcher: () => Promise<{ data: PaginatedResult<Movie> }>) {
  const query = useQuery({ queryKey: [key], queryFn: () => fetcher().then((r) => r.data) })
  const items = computed<HomeMedia[]>(() => (query.data.value?.results ?? []).slice(0, RAIL_SIZE).map(fromMovie))
  return { id, query, items }
}

const recommended = useQuery({
  queryKey: computed(() => ["movies-recommended", genreQuery.value]),
  queryFn: () => movieApi.getRecommended(genreQuery.value).then((r) => r.data),
  enabled: hasGenres,
})
const recommendedItems = computed<HomeMedia[]>(() =>
  (recommended.data.value?.results ?? []).slice(0, RAIL_SIZE).map(fromMovie),
)

const popular = useMovieRail("popular", "movies-popular-home", () => movieApi.getPopular(1))
const nowPlaying = useMovieRail("nowPlaying", "movies-now-playing-home", () => movieApi.getNowPlaying(1))
const topRated = useMovieRail("topRated", "movies-top-rated-home", () => movieApi.getTopRated(1))
const rails = [popular, nowPlaying, topRated]

const heroItems = computed(() => {
  const source = recommendedItems.value.length
    ? recommendedItems.value
    : popular.items.value.length
      ? popular.items.value
      : nowPlaying.items.value
  return source.slice(0, 5)
})
</script>

<style scoped>
.home {
  min-height: 100vh;
  background: var(--color-surface);
  color: var(--color-text-primary);
}
.home__content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1.25rem 1rem 3rem;
}
@media (min-width: 768px) {
  .home__content {
    gap: 3rem;
    padding: 1.75rem 1.5rem 4rem;
  }
}
</style>
