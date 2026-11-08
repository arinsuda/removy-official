import { computed, ref } from "vue"
import { useQuery } from "@tanstack/vue-query"
import { feedApi } from "@/api/endpoints/feed"
import { movieApi } from "@/api/api"
import { fromMovie, fromSeries, type HomeMedia } from "@/components/home/homeMedia"
import { filterByTier, interleave, pickHeroItems } from "@/components/home/popcornHome"
import { useFavoriteGenres } from "@/composables/useFavoriteGenres"
import { useAuthStore } from "@/stores/auth"

const RAIL_SIZE = 12
const FRIENDS_LIMIT = 8
const STALE_MS = 5 * 60 * 1000

/**
 * Data for the home page. Hero and the ป๊อปร้อน rail share four eager TMDB
 * lists; recommended, top rated and friends load once their rail nears the
 * viewport (`reveal.*`).
 */
export function useHomeData() {
  const auth = useAuthStore()
  const { hasGenres, genreQuery } = useFavoriteGenres()
  const reveal = { recommended: ref(false), friends: ref(false), topRated: ref(false) }

  const opts = { staleTime: STALE_MS }
  const popularMovies = useQuery({ queryKey: ["movies-popular-home"], queryFn: () => movieApi.getPopular(1).then((r) => r.data), ...opts })
  const nowPlaying = useQuery({ queryKey: ["movies-now-playing-home"], queryFn: () => movieApi.getNowPlaying(1).then((r) => r.data), ...opts })
  const popularTv = useQuery({ queryKey: ["tv-popular-home"], queryFn: () => movieApi.getPopularSeries(1).then((r) => r.data), ...opts })
  const airingTv = useQuery({ queryKey: ["tv-airing-home"], queryFn: () => movieApi.getNowAiringSeries(1).then((r) => r.data), ...opts })
  const topRated = useQuery({
    queryKey: ["movies-top-rated-home"],
    queryFn: () => movieApi.getTopRated(1).then((r) => r.data),
    enabled: reveal.topRated,
    ...opts,
  })
  const recommended = useQuery({
    queryKey: computed(() => ["movies-recommended", genreQuery.value]),
    queryFn: () => movieApi.getRecommended(genreQuery.value).then((r) => r.data),
    enabled: computed(() => hasGenres.value && reveal.recommended.value),
    ...opts,
  })
  const friends = useQuery({
    queryKey: ["home-friends-feed", FRIENDS_LIMIT],
    queryFn: () => feedApi.getFeed({ limit: FRIENDS_LIMIT }).then((r) => r.data),
    enabled: computed(() => auth.isLoggedIn && reveal.friends.value),
    ...opts,
  })

  const movies = (q: { data: { value?: { results?: Parameters<typeof fromMovie>[0][] } } }) =>
    (q.data.value?.results ?? []).map(fromMovie)
  const series = (q: { data: { value?: { results?: Parameters<typeof fromSeries>[0][] } } }) =>
    (q.data.value?.results ?? []).map(fromSeries)

  const popularMovieItems = computed(() => movies(popularMovies))
  const nowPlayingItems = computed(() => movies(nowPlaying))

  /** Everything the eager queries returned, movies and TV interleaved. */
  const pool = computed<HomeMedia[]>(() =>
    interleave(popularMovieItems.value, series(popularTv), nowPlayingItems.value, series(airingTv)),
  )
  const hero = computed(() => pickHeroItems(pool.value, popularMovieItems.value))

  return {
    reveal,
    hasGenres,
    pool,
    hero,
    poolLoading: computed(() => [popularMovies, nowPlaying, popularTv, airingTv].every((q) => q.isLoading.value)),
    poolError: computed(() => [popularMovies, nowPlaying, popularTv, airingTv].every((q) => q.isError.value)),
    refetchPool: () => [popularMovies, nowPlaying, popularTv, airingTv].forEach((q) => q.refetch()),
    unrated: computed(() => filterByTier(nowPlayingItems.value, "none", RAIL_SIZE)),
    nowPlayingItems: computed(() => nowPlayingItems.value.slice(0, RAIL_SIZE)),
    nowPlaying,
    topRatedItems: computed(() => movies(topRated).slice(0, RAIL_SIZE)),
    topRated,
    recommendedItems: computed(() => movies(recommended).slice(0, RAIL_SIZE)),
    recommended,
    friendItems: computed(() => friends.data.value?.items ?? []),
    friends,
  }
}
