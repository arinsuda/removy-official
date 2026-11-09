<template>
  <div class="ua">
    <SkeletonGrid v-if="loading" :count="8" min-width="140px" aspect="1 / 1" />

    <AppErrorState v-else-if="loadError" :title="t('profile.states.sectionError')" @retry="load" />

    <AppEmptyState v-else-if="all.length === 0" :icon="Trophy" :title="t('profile.tabContent.achievements.emptyTitle')" :description="t('profile.tabContent.achievements.emptyBody')" />

    <template v-else>
      <header class="ua__head">
        <div class="ua__counter">
          <p class="ua__count" role="status">{{ t("profile.tabContent.achievements.counter", { unlocked, total: all.length }) }}</p>
          <span class="ua__bar" aria-hidden="true"><span class="ua__fill" :style="{ width: `${(unlocked / all.length) * 100}%` }" /></span>
        </div>
        <RouterLink :to="{ name: 'user-achievements', params: { userId } }" class="ua__all">{{ t("profile.tabContent.achievements.viewAll") }}</RouterLink>
      </header>

      <div class="ua__chips" role="group" :aria-label="t('profile.tabContent.common.filter')">
        <button v-for="f in FILTERS" :key="f" type="button" class="ua__chip" :class="{ 'is-active': filter === f }" :aria-pressed="filter === f" @click="filter = f">
          {{ t(`profile.tabContent.achievements.filters.${f}`) }}
        </button>
      </div>

      <ul class="ua__grid" :aria-label="t('profile.tabs.achievements')">
        <li v-for="(a, i) in shown" :key="a.achievement_id" :style="{ '--i': Math.min(i, 12) }">
          <AchievementBadge :achievement="a" @open="selected = $event" />
        </li>
      </ul>
    </template>

    <AppModal :model-value="!!selected" size="sm" :title="selected?.achievement.name" @close="selected = null">
      <div v-if="selected" class="ua__detail">
        <span class="ua__detail-icon" :class="{ 'is-on': selected.is_unlocked }">
          <component :is="getAchievementIcon(selected.achievement.action_type)" :size="34" :stroke-width="1.6" aria-hidden="true" />
        </span>
        <p class="ua__detail-desc">{{ selected.achievement.description }}</p>
        <p v-if="selected.is_unlocked && selected.unlocked_at" class="ua__detail-state">{{ t("profile.tabContent.achievements.unlockedOn", { date: formatDate(selected.unlocked_at) }) }}</p>
        <template v-else-if="!selected.is_unlocked">
          <span class="ua__detail-bar" role="progressbar" :aria-valuenow="progressOf(selected)" aria-valuemin="0" aria-valuemax="100" :aria-label="t('profile.tabContent.achievements.progressLabel')">
            <span :style="{ width: `${progressOf(selected)}%` }" />
          </span>
          <p class="ua__detail-state">{{ t("profile.tabContent.achievements.progress", { current: Math.min(selected.current_count, selected.achievement.target_count), target: selected.achievement.target_count }) }}</p>
          <p class="ua__detail-hint">{{ t("profile.tabContent.achievements.lockedHint") }}</p>
        </template>
        <p v-if="selected.achievement.exp_reward" class="ua__detail-xp">{{ t("profile.tabContent.achievements.xp", { n: selected.achievement.exp_reward }) }}</p>
      </div>
      <template #footer>
        <AppButton variant="outline" @click="selected = null">{{ t("profile.tabContent.achievements.close") }}</AppButton>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink } from "vue-router"
import { Trophy } from "lucide-vue-next"
import { achievementApi } from "@/api/endpoints/achievement"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppModal from "@/components/ui/AppModal.vue"
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"
import type { UserAchievement } from "@/types/achievement"
import { getAchievementIcon } from "@/utils/achievementIcons"
import AchievementBadge from "./tabs/AchievementBadge.vue"
import { countUnlocked, filterAchievements, progressOf, sortAchievements, type AchievementFilter } from "./tabs/achievements"

const props = defineProps<{ userId: number }>()

const { t, locale } = useI18n()
const FILTERS: AchievementFilter[] = ["all", "unlocked", "locked"]
const PAGE_LIMIT = 100 // server maximum
const MAX_PAGES = 5

const all = ref<UserAchievement[]>([])
const loading = ref(false)
const loadError = ref(false)
const filter = ref<AchievementFilter>("all")
const selected = ref<UserAchievement | null>(null)

const unlocked = computed(() => countUnlocked(all.value))
const shown = computed(() => filterAchievements(all.value, filter.value))
const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "long" }).format(new Date(iso))

async function load() {
  try {
    loading.value = true
    loadError.value = false
    const out: UserAchievement[] = []
    for (let page = 1; page <= MAX_PAGES; page++) {
      const res = await achievementApi.listByUser(props.userId, { page, limit: PAGE_LIMIT })
      out.push(...res.data)
      if (page >= (res.pagination?.total_pages ?? 1)) break
    }
    all.value = sortAchievements(out)
  } catch (err) {
    console.error("Fetch achievements failed:", err)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.ua__head { display: flex; align-items: flex-end; justify-content: space-between; gap: 1rem; margin-bottom: 1rem; }
.ua__counter { display: flex; flex: 1; flex-direction: column; gap: 0.5rem; max-width: 24rem; }
.ua__count { margin: 0; font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.ua__bar { display: block; height: 0.375rem; overflow: hidden; background: var(--color-surface-3); border-radius: 9999px; }
.ua__fill { display: block; height: 100%; background: var(--color-brand); border-radius: inherit; transition: width var(--motion-base) var(--ease-out-soft); }
.ua__all { flex: none; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-secondary); text-decoration: underline; text-underline-offset: 3px; }
.ua__chips { display: flex; gap: 0.5rem; margin-bottom: 1.25rem; overflow-x: auto; scrollbar-width: none; }
.ua__chip { flex: none; min-height: 2rem; padding: 0 0.875rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); white-space: nowrap; background: none; border: 1px solid var(--color-border-strong); border-radius: 9999px; cursor: pointer; }
.ua__chip.is-active { color: var(--color-on-brand); background: var(--color-brand); border-color: var(--color-brand); }
.ua__grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.75rem; margin: 0; padding: 0; list-style: none; }
.ua__grid > li { display: flex; flex-direction: column; animation: ua-in var(--motion-base) var(--ease-out-soft) calc(var(--i) * 30ms) both; }
.ua__grid > li > * { flex: 1; }
@keyframes ua-in { from { opacity: 0; transform: translateY(8px); } }
.ua__detail { display: flex; flex-direction: column; align-items: center; gap: 0.75rem; text-align: center; }
.ua__detail-icon { display: grid; place-items: center; width: 5rem; height: 5rem; color: var(--color-text-muted); background: var(--color-surface-3); border-radius: 50%; filter: grayscale(1); }
.ua__detail-icon.is-on { color: var(--color-accent); background: var(--color-brand); filter: none; box-shadow: 0 0 28px color-mix(in srgb, var(--color-brand) 50%, transparent); }
.ua__detail-desc { margin: 0; font-size: var(--text-small); line-height: 1.6; color: var(--color-text-secondary); }
.ua__detail-state { margin: 0; font-size: var(--text-small); font-weight: 600; color: var(--color-text-primary); }
.ua__detail-hint { margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.ua__detail-xp { margin: 0; padding: 0.125rem 0.75rem; font-size: var(--text-caption); font-weight: 700; color: var(--color-accent); border: 1px solid var(--color-border-strong); border-radius: 9999px; }
.ua__detail-bar { display: block; width: 100%; height: 0.4rem; overflow: hidden; background: var(--color-surface-3); border-radius: 9999px; }
.ua__detail-bar > span { display: block; height: 100%; background: var(--color-text-secondary); border-radius: inherit; }
@media (min-width: 480px) { .ua__grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (min-width: 768px) { .ua__grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
@media (prefers-reduced-motion: reduce) { .ua__grid > li { animation: none; } .ua__fill { transition: none; } }
</style>
