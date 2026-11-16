<template>
  <main class="anime-page">
    <h1 class="anime-title">{{ t("anime.title") }}</h1>
    <AnimeFilterBar :state="state" :genres="data.genres.value" @update="update" />
    <AnimeReleaseFilter :state="state" @update="update" />
    <BrowseTierChips v-model="tier" :counts="data.tierCounts.value" :searching="false" />
    <BrowseGrid
      :items="items"
      :loading="data.isLoading.value"
      :error="data.isError.value"
      :has-next="data.hasNextPage.value"
      :fetching-next="data.isFetchingNextPage.value"
      :capped="data.capped.value"
      :infinite="true"
      @retry="data.refetch()"
      @reset="reset"
      @load-more="data.fetchNextPage()"
    />
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import AnimeFilterBar from "@/components/anime/AnimeFilterBar.vue"
import AnimeReleaseFilter from "@/components/anime/AnimeReleaseFilter.vue"
import { DEFAULT_ANIME_STATE } from "@/components/anime/animeQuery"
import BrowseGrid from "@/components/browse/BrowseGrid.vue"
import BrowseTierChips from "@/components/browse/BrowseTierChips.vue"
import type { BrowseTierFilter } from "@/components/browse/browseQuery"
import { filterTier } from "@/components/browse/browseSort"
import { useAnimeBrowse } from "@/composables/useAnimeBrowse"
import { useAnimeRouteState } from "@/composables/useAnimeRouteState"
import { provideHomePopup } from "@/composables/useHomePopup"

const { t } = useI18n()
provideHomePopup()
const { state, update } = useAnimeRouteState()
const data = useAnimeBrowse(state)

/** Popcorn tier is a client-side filter over the loaded titles, so it stays out of the URL. */
const tier = ref<BrowseTierFilter>("all")
const items = computed(() => filterTier(data.loaded.value, tier.value))

/** Clears every filter but keeps the content type. */
function reset() {
  tier.value = "all"
  void update({ ...DEFAULT_ANIME_STATE, type: state.value.type })
}
</script>

<style scoped>
.anime-page {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  width: 100%;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}
.anime-title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
@media (min-width: 768px) {
  .anime-page {
    padding: 2rem 1.5rem 4rem;
  }
}
</style>
