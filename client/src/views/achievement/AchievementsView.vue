<template>
  <div class="av">
    <AchievementSkeleton v-if="page.loading.value || (page.profile.isPending.value && !page.failed.value)" />

    <AppEmptyState v-else-if="page.notFound.value" :icon="UserX" :title="t('profile.states.notFoundTitle')" :description="t('profile.states.notFoundBody')">
      <template #action>
        <AppButton variant="outline" size="sm" :to="{ name: 'home' }">{{ t('navigation.home') }}</AppButton>
      </template>
    </AppEmptyState>

    <PrivateSectionLock v-else-if="page.locked.value" :user-id="userId" @reload="page.reload()" />

    <AppErrorState v-else-if="page.failed.value || page.profileFailed.value" :title="t('achievements.states.errorTitle')" :description="t('achievements.states.errorBody')" @retry="page.reload()" />

    <div v-else class="av__layout">
      <AchievementHero :user-id="userId" :profile="page.profile.data.value ?? null" :is-owner="page.isOwner.value" :summary="page.summary.value" :latest="page.latest.value" :next="page.next.value" @open="page.selected.value = $event" />

      <TabEmptyState v-if="page.all.value.length === 0" :icon="Trophy" :is-owner="page.isOwner.value" :owner-title="t('achievements.empty.ownerTitle')" :owner-body="t('achievements.empty.ownerBody')" :visitor-title="t('achievements.empty.visitorTitle')" :visitor-body="t('achievements.empty.visitorBody')" :cta-label="t('achievements.cta.browse')" />

      <template v-else>
        <div class="av__tools">
          <TabToolbar v-model:query="page.search.value" v-model:sort="page.sort.value" v-model:filter="statusModel" :count-label="t('achievements.toolbar.count', { shown: page.visibleCount.value, total: page.all.value.length })" :filters="statusFilters" :sorts="sorts" />
          <AchievementCategoryChips v-model="page.category.value" :categories="page.categories.value" />
        </div>

        <TabEmptyState v-if="page.sections.value.length === 0" :icon="Trophy" :is-owner="page.isOwner.value" filtered @clear="page.clearFilters()" />

        <div v-else class="av__sections">
          <AchievementSection v-for="s in page.sections.value" :key="s.key" :section="s" @open="page.selected.value = $event" />
        </div>
      </template>
    </div>

    <AchievementDetailModal :achievement="page.selected.value" :tiers="page.selectedTiers.value" :user-id="userId" :is-owner="page.isOwner.value" @close="page.selected.value = null" @select="page.selected.value = $event" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { useRoute } from "vue-router"
import { Trophy, UserX } from "lucide-vue-next"
import PrivateSectionLock from "@/components/profile/PrivateSectionLock.vue"
import AchievementCategoryChips from "@/components/achievement/AchievementCategoryChips.vue"
import AchievementDetailModal from "@/components/achievement/AchievementDetailModal.vue"
import AchievementHero from "@/components/achievement/AchievementHero.vue"
import AchievementSection from "@/components/achievement/AchievementSection.vue"
import AchievementSkeleton from "@/components/achievement/AchievementSkeleton.vue"
import { SORT_KEYS, STATUS_FILTERS, type StatusFilter } from "@/components/achievement/config"
import { useAchievementsPage } from "@/components/achievement/useAchievementsPage"
import TabEmptyState from "@/components/profile/tabs/TabEmptyState.vue"
import TabToolbar from "@/components/profile/tabs/TabToolbar.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"

const { t } = useI18n()
const route = useRoute()
const userId = computed(() => Number(route.params.userId))
const page = useAchievementsPage(userId)

const statusModel = computed<string>({
  get: () => page.status.value,
  set: (v) => (page.status.value = v as StatusFilter),
})
const statusFilters = computed(() => STATUS_FILTERS.map((key) => ({ key, label: t(`achievements.filters.${key}`), count: page.counts.value[key] })))
const sorts = computed(() => SORT_KEYS.map((key) => ({ key, label: t(`achievements.sorts.${key}`) })))
</script>

<style scoped>
.av { width: 100%; min-height: 100vh; min-height: 100dvh; color: var(--color-text-primary); background: var(--color-surface); }
.av > :not(.av__layout) { max-width: 75rem; margin-inline: auto; padding-inline: 1rem; }
.av__layout { display: flex; flex-direction: column; gap: 1.25rem; max-width: 75rem; margin: 0 auto; padding: 0 1rem 5rem; }
.av__tools { display: flex; flex-direction: column; gap: 0.75rem; }
.av__sections { display: flex; flex-direction: column; gap: 2rem; }
@media (min-width: 768px) {
  .av__layout, .av > :not(.av__layout) { padding-inline: 1.5rem; }
}
</style>
