import { computed, type Ref } from "vue"
import { useQuery } from "@tanstack/vue-query"
import { mangaApi } from "@/api/api"
import { MANGA_DETAIL_STALE_MS, MANGA_SIMILAR_SIZE } from "@/components/manga/mangaConfig"

/** Detail and similar rail for one title. The id comes from the route. */
export function useMangaDetail(id: Ref<number>, includeAdult: Ref<boolean>) {
  const detail = useQuery({
    queryKey: computed(() => ["manga", "detail", id.value]),
    queryFn: async () => (await mangaApi.detail(id.value)).data.manga,
    staleTime: MANGA_DETAIL_STALE_MS,
  })

  const similar = useQuery({
    queryKey: computed(() => ["manga", "similar", id.value, includeAdult.value]),
    queryFn: async () => (await mangaApi.similar(id.value, includeAdult.value)).data.results.slice(0, MANGA_SIMILAR_SIZE),
    staleTime: MANGA_DETAIL_STALE_MS,
  })

  return { detail, similar }
}
