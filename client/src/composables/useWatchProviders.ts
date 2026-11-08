import { computed, ref, type Ref } from "vue"
import { useQuery } from "@tanstack/vue-query"
import { movieApi } from "@/api/endpoints/movie"

export const PROVIDER_REGIONS = ["TH", "US", "JP", "KR", "GB", "SG"] as const
const REGION_KEY = "popcorn-watch-region"
const DEFAULT_REGION = "TH"

function readRegion(): string {
  try {
    const v = localStorage.getItem(REGION_KEY)
    if (v && (PROVIDER_REGIONS as readonly string[]).includes(v)) return v
  } catch {
    // storage unavailable
  }
  return DEFAULT_REGION
}

// Shared so the choice survives title changes within a session.
const region = ref(readRegion())

export function setWatchRegion(next: string) {
  region.value = next
  try {
    localStorage.setItem(REGION_KEY, next)
  } catch {
    // storage unavailable
  }
}

export function useWatchProviders(type: "movie" | "tv", id: Ref<number>) {
  const query = useQuery({
    queryKey: ["watch-providers", type, id, region],
    queryFn: () => movieApi.getWatchProviders(type, id.value, region.value).then(r => r.data.providers),
    enabled: computed(() => id.value > 0),
    staleTime: 60 * 60 * 1000,
  })
  return { region, query }
}
