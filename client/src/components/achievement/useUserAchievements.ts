import { ref, toValue, type MaybeRefOrGetter } from "vue"
import { achievementApi } from "@/api/endpoints/achievement"
import type { UserAchievement } from "@/types/achievement"
import { MAX_PAGES, PAGE_LIMIT } from "./config"
import { sortAchievements } from "./achievements"

/** Loads every achievement of a user (all pages), newest unlocks first. */
export function useUserAchievements(userId: MaybeRefOrGetter<number>) {
  const all = ref<UserAchievement[]>([])
  const loading = ref(false)
  const failed = ref(false)

  async function load() {
    loading.value = true
    failed.value = false
    try {
      const out: UserAchievement[] = []
      for (let page = 1; page <= MAX_PAGES; page++) {
        const res = await achievementApi.listByUser(toValue(userId), { page, limit: PAGE_LIMIT })
        out.push(...res.data)
        if (page >= (res.pagination?.total_pages ?? 1)) break
      }
      all.value = sortAchievements(out)
    } catch (err) {
      console.error("Fetch achievements failed:", err)
      failed.value = true
    } finally {
      loading.value = false
    }
  }

  return { all, loading, failed, load }
}
