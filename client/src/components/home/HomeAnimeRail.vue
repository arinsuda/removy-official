<template>
  <HomeRail
    :title="t('anime.home.title')"
    :subtitle="t('anime.home.subtitle')"
    :see-all-to="{ name: 'anime' }"
    :loading="loading"
    :error="failed"
    :empty="!items.length"
    @retry="refetch"
    @visible="visible = true"
  >
    <HomePosterCard v-for="item in items" :key="item.key" :item="item" rail="anime" />
  </HomeRail>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { useQuery } from "@tanstack/vue-query"
import { animeApi } from "@/api/api"
import { ANIME_DEFAULT_SORT, ANIME_RAIL_SIZE, ANIME_RAIL_STALE_MS } from "@/components/anime/animeConfig"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import HomeRail from "@/components/home/HomeRail.vue"
import { fromMovie, fromSeries } from "@/components/home/homeMedia"
import { interleave } from "@/components/home/popcornHome"

const { t } = useI18n()
/** Loads once the rail nears the viewport, like the other lazy home rails. */
const visible = ref(false)
const query = { sort: ANIME_DEFAULT_SORT }

const tv = useQuery({
  queryKey: ["anime", "home", "tv"],
  queryFn: () => animeApi.list("tv", query).then((r) => r.data.results.map(fromSeries)),
  enabled: visible,
  staleTime: ANIME_RAIL_STALE_MS,
})
const movie = useQuery({
  queryKey: ["anime", "home", "movie"],
  queryFn: () => animeApi.list("movie", query).then((r) => r.data.results.map(fromMovie)),
  enabled: visible,
  staleTime: ANIME_RAIL_STALE_MS,
})

/** Series and films alternate, so one type does not fill the whole rail. */
const items = computed(() => interleave(tv.data.value ?? [], movie.data.value ?? []).slice(0, ANIME_RAIL_SIZE))
const loading = computed(() => !visible.value || (tv.isLoading.value && movie.isLoading.value))
const failed = computed(() => tv.isError.value && movie.isError.value)

function refetch() {
  void tv.refetch()
  void movie.refetch()
}
</script>
