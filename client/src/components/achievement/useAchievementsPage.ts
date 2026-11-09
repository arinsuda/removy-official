import { computed, ref, watch, type Ref } from "vue"
import { useI18n } from "vue-i18n"
import { useQuery } from "@tanstack/vue-query"
import { isAxiosError } from "axios"
import { userApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import type { UserAchievement } from "@/types/achievement"
import { DEFAULT_SORT, type CategoryKey, type SortKey, type StatusFilter } from "./config"
import { useAchievementText } from "./text"
import { buildUnits, countByStatus, filterUnits, groupByCategory, latestUnlock, matchedCount, nextUp, sortUnits, summarize } from "./units"
import { useUserAchievements } from "./useUserAchievements"

/** State of /users/:userId/achievements: profile, all achievements, filters and the open badge. */
export function useAchievementsPage(userId: Ref<number>) {
  const { locale } = useI18n()
  const { description } = useAchievementText()
  const auth = useAuthStore()
  const isOwner = computed(() => !!auth.user && auth.user.id === userId.value)

  const profile = useQuery({
    queryKey: computed(() => ["user-profile", userId.value]),
    queryFn: () => userApi.getProfile(userId.value).then((res) => res.data.user),
    enabled: computed(() => userId.value > 0),
    retry: false,
  })
  const { all, loading, failed, load } = useUserAchievements(userId)

  const notFound = computed(() => profile.isError.value && isAxiosError(profile.error.value) && profile.error.value.response?.status === 404)
  const profileFailed = computed(() => profile.isError.value && !notFound.value)

  const status = ref<StatusFilter>("all")
  const category = ref<CategoryKey | "all">("all")
  const search = ref("")
  const sort = ref<SortKey>(DEFAULT_SORT)
  const selected = ref<UserAchievement | null>(null)

  const units = computed(() => buildUnits(all.value))
  const summary = computed(() => summarize(all.value))
  const latest = computed(() => latestUnlock(all.value))
  const next = computed(() => nextUp(all.value))
  const categories = computed(() => groupByCategory(units.value).map((c) => c.key))
  /** Status counts respect category and search so the chips always match what the grid would show. */
  const query = (s: StatusFilter) => ({ status: s, category: category.value, search: search.value, describe: (a: UserAchievement) => description(a.achievement) })
  const counts = computed(() => countByStatus(filterUnits(units.value, query("all"))))
  const visible = computed(() => sortUnits(filterUnits(units.value, query(status.value)), sort.value, locale.value))
  const sections = computed(() => groupByCategory(visible.value))
  const visibleCount = computed(() => matchedCount(visible.value))
  const selectedTiers = computed(() => units.value.find((u) => u.tiers.some((t) => t.achievement_id === selected.value?.achievement_id))?.tiers ?? [])

  function clearFilters() {
    status.value = "all"
    category.value = "all"
    search.value = ""
  }

  function reload() {
    if (profile.isError.value) void profile.refetch()
    void load()
  }

  watch(userId, () => {
    clearFilters()
    selected.value = null
    void load()
  }, { immediate: true })

  return { isOwner, profile, notFound, profileFailed, all, loading, failed, status, category, search, sort, selected, summary, latest, next, categories, counts, sections, visibleCount, selectedTiers, clearFilters, reload }
}
