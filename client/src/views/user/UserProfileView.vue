<template>
  <div class="profile">
    <SkeletonProfile v-if="profile.isPending.value" />

    <AppEmptyState
      v-else-if="errorKind === 'notFound'"
      :icon="UserX"
      :title="t('profile.states.notFoundTitle')"
      :description="t('profile.states.notFoundBody')"
    />
    <AppErrorState
      v-else-if="errorKind === 'error' || !user"
      :title="t('profile.states.errorTitle')"
      :description="t('profile.states.errorBody')"
      @retry="profile.refetch()"
    />

    <div v-else class="profile__layout">
      <ProfileHero
        :user="user"
        :is-owner="isOwner"
        :can-follow="!isOwner"
        :is-guest="!auth.isLoggedIn"
        :follow-state="followState"
        :busy="follow.isPending.value || unfollow.isPending.value"
        @follow="handleFollow"
        @unfollow="showUnfollowConfirm = true"
        @cancel="unfollow.mutate()"
        @open-list="openList"
      />

      <ProfileLockedState
        v-if="isLocked"
        :can-follow="!isOwner"
        :is-guest="!auth.isLoggedIn"
        :follow-state="followState"
        :busy="follow.isPending.value || unfollow.isPending.value"
        @follow="handleFollow"
        @cancel="unfollow.mutate()"
      />

      <div v-else class="profile__grid">
        <main class="profile__main">
          <div class="profile__tabs">
            <AppTabs v-model="activeTab" :tabs="tabs" :label="t('profile.title')" :controls="false" />
          </div>
          <section role="tabpanel" class="profile__panel">
            <component :is="activeComponent" :key="`${activeTab}-${userId}`" v-bind="activeProps" />
          </section>
        </main>
        <div class="profile__rail">
          <ProfileTastePanel
            :user-id="userId"
            :taste="taste"
            :reviews="reviews.data.value ?? []"
            :stats="stats.data.value"
            :level="user.level ?? 1"
          />
          <ProfileBmolHighlight :user-id="userId" />
        </div>
      </div>
    </div>

    <FollowListModal v-if="user" v-model="listOpen" :user-id="userId" :initial-tab="listTab" :is-owner="isOwner" />

    <ConfirmModal
      v-model="showUnfollowConfirm"
      list-type="unfollow"
      :item-name="user?.display_name || user?.username"
      :confirm-disabled="unfollow.isPending.value"
      @confirm="confirmUnfollow"
      @cancel="showUnfollowConfirm = false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, type Component } from "vue"
import { useRoute, useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { UserX } from "lucide-vue-next"
import { useAuthStore } from "@/stores/auth"
import { useAuthGate } from "@/composables/useAuthGate"
import { useUserProfile } from "@/composables/useUserProfile"
import { useFollowActions } from "@/composables/useFollow"
import type { ListType } from "@/types"
import SkeletonProfile from "@/components/ui/skeleton/SkeletonProfile.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import ProfileHero from "@/components/profile/ProfileHero.vue"
import ProfileLockedState from "@/components/profile/ProfileLockedState.vue"
import ProfileTastePanel from "@/components/profile/ProfileTastePanel.vue"
import ProfileBmolHighlight from "@/components/profile/ProfileBmolHighlight.vue"
import FollowListModal from "@/components/profile/FollowListModal.vue"
import UserReviews from "@/components/profile/UserReviews.vue"
import UserWatchlist from "@/components/profile/UserWatchlist.vue"
import UserWatched from "@/components/profile/UserWatched.vue"
import UserLikes from "@/components/profile/UserLikes.vue"
import UserAchievements from "@/components/profile/UserAchievements.vue"
import UserBmol from "@/components/profile/UserBmol.vue"

type TabKey = "reviews" | "watchlist" | "likes" | "watched" | "bol" | "achievements"
const TAB_KEYS: TabKey[] = ["reviews", "watchlist", "likes", "watched", "bol", "achievements"]

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { requireAuth } = useAuthGate()
const { t } = useI18n()

const userIdParam = computed(() => Number(route.params.userId))
const { userId, isOwner, profile, followStatus, isLocked, stats, reviews, taste, errorKind } = useUserProfile(userIdParam)
const { follow, unfollow } = useFollowActions(userId)

function handleFollow() {
  if (!requireAuth("follow")) return
  follow.mutate()
}

const user = computed(() => profile.data.value ?? null)
const followState = computed(() => followStatus.data.value?.follow_status)

// ?tab=<key> <-> active tab; replace so tab switches don't fill history.
function tabFromRoute(): TabKey {
  const q = route.query.tab
  return typeof q === "string" && (TAB_KEYS as string[]).includes(q) ? (q as TabKey) : "reviews"
}
const activeTab = ref<TabKey>(tabFromRoute())
watch(() => route.query.tab, () => (activeTab.value = tabFromRoute()))
watch(activeTab, key => {
  if (route.query.tab !== key) router.replace({ query: { ...route.query, tab: key } })
})

const tabs = computed(() => {
  const s = stats.data.value
  const count = (n?: number) => (n === undefined ? "" : ` (${n})`)
  return [
    { key: "reviews" as const, label: t("profile.tabs.reviews") + count(s?.review_count) },
    { key: "watchlist" as const, label: t("profile.tabs.watchlist") + count(s?.watchlist_count) },
    { key: "likes" as const, label: t("profile.tabs.likes") + count(s?.like_count) },
    { key: "watched" as const, label: t("profile.tabs.watched") + count(s?.watched_count) },
    { key: "bol" as const, label: t("profile.tabs.bol") },
    { key: "achievements" as const, label: t("profile.tabs.achievements") },
  ]
})

const components: Record<TabKey, Component> = {
  reviews: UserReviews,
  watchlist: UserWatchlist,
  likes: UserLikes,
  watched: UserWatched,
  bol: UserBmol,
  achievements: UserAchievements,
}
const listTypes: Partial<Record<TabKey, ListType>> = { watchlist: "watchlist", likes: "likes", watched: "watched" }
const activeComponent = computed(() => components[activeTab.value])
const activeProps = computed(() => ({
  userId: userId.value,
  listType: listTypes[activeTab.value],
  // Only the reviews tab declares createdAt; other tabs would leak it as a DOM attribute.
  ...(activeTab.value === "reviews" ? { createdAt: user.value?.created_at } : {}),
}))

const listOpen = ref(false)
const listTab = ref<"followers" | "following">("followers")
function openList(kind: "followers" | "following") {
  listTab.value = kind
  listOpen.value = true
}

const showUnfollowConfirm = ref(false)
function confirmUnfollow() {
  unfollow.mutate(undefined, { onSuccess: () => (showUnfollowConfirm.value = false) })
}
</script>

<style scoped>
.profile { width: 100%; min-height: 100vh; min-height: 100dvh; color: var(--color-text-primary); background: var(--color-surface); }
.profile__layout { max-width: 75rem; margin: 0 auto; padding: 0 1rem 5rem; }
.profile__grid { display: grid; grid-template-columns: minmax(0, 1fr); gap: 1.5rem; }
.profile__rail { display: flex; flex-direction: column; gap: 1rem; }
.profile__tabs { overflow-x: auto; scrollbar-width: none; padding-bottom: 0.25rem; }
.profile__tabs :deep(.app-tabs__tab) { flex: none; white-space: nowrap; }
.profile__tabs::-webkit-scrollbar { display: none; }
.profile__panel { padding-top: 1.25rem; }
@media (min-width: 768px) {
  .profile__layout { padding: 0 1.5rem 5rem; }
}
@media (min-width: 1024px) {
  .profile__grid { grid-template-columns: minmax(0, 8fr) minmax(0, 4fr); align-items: start; }
  .profile__rail { position: sticky; top: 5rem; order: 2; }
}
</style>
