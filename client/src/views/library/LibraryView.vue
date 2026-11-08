<template>
  <div class="library-root">
    <div class="grain" aria-hidden="true" />

    <!-- Loading State -->
    <SkeletonProfile v-if="loading" />

    <!-- Main Layout -->
    <div v-else class="library-layout">
      <!-- Back to Profile Link -->
      <router-link :to="`/users/${userId}`" class="back-link">
        <ChevronLeft :size="16" />
        <span>{{ $t("navigation.profile") }}</span>
      </router-link>

      <!-- Profile Header Summary Card -->
      <header class="profile-summary-card">
        <div class="user-info-section">
          <div class="avatar-wrapper">
            <div class="avatar-ring">
              <img
                v-if="profileUser?.avatar_url"
                :src="profileUser.avatar_url"
                :alt="profileUser.username"
                class="avatar-img"
              />
              <div v-else class="avatar-fallback">
                <UserIcon :size="28" />
              </div>
            </div>
            <span class="level-chip">Lv.{{ profileUser?.level ?? 1 }}</span>
          </div>
          <div class="user-meta">
            <h1 class="user-name">{{ profileUser?.display_name || profileUser?.username }}</h1>
            <p class="user-tag">@{{ profileUser?.username }}</p>
            <p class="user-bio">{{ profileUser?.bio || $t("profile.noBio") }}</p>
          </div>
        </div>
        <div class="user-counts-section">
          <div class="count-box">
            <span class="count-val">{{ profileUser?.review_count ?? 0 }}</span>
            <span class="count-lbl">{{ $t("library.tabs.reviews") }}</span>
          </div>
          <div class="count-box">
            <span class="count-val">{{ profileUser?.follower_count ?? 0 }}</span>
            <span class="count-lbl">{{ $t("profile.stats.followers") }}</span>
          </div>
          <div class="count-box">
            <span class="count-val">{{ profileUser?.following_count ?? 0 }}</span>
            <span class="count-lbl">{{ $t("profile.stats.following") }}</span>
          </div>
        </div>
      </header>

      <!-- Navigation Tabs -->
      <nav class="library-tabs">
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === 'dashboard' }"
          @click="activeTab = 'dashboard'"
        >
          <Flame :size="16" />
          <span>{{ $t("library.dashboard.title") }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === 'bmol' }"
          @click="activeTab = 'bmol'"
        >
          <Trophy :size="16" />
          <span>{{ $t("library.bmol.title") }}</span>
          <span class="tab-count">{{ bmolItems.length }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === 'watchlist' }"
          @click="activeTab = 'watchlist'"
        >
          <Bookmark :size="16" />
          <span>{{ $t("library.tabs.watchlist") }}</span>
          <span class="tab-count">{{ watchlistItems.length }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === 'watched' }"
          @click="activeTab = 'watched'"
        >
          <Eye :size="16" />
          <span>{{ $t("library.tabs.watched") }}</span>
          <span class="tab-count">{{ watchedItems.length }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === 'likes' }"
          @click="activeTab = 'likes'"
        >
          <Heart :size="16" />
          <span>{{ $t("library.tabs.likes") }}</span>
          <span class="tab-count">{{ likedItems.length }}</span>
        </button>
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': activeTab === 'reviews' }"
          @click="activeTab = 'reviews'"
        >
          <Star :size="16" />
          <span>{{ $t("library.tabs.reviews") }}</span>
          <span class="tab-count">{{ userReviews.length }}</span>
        </button>
      </nav>

      <!-- TAB CONTENT SUB-COMPONENTS -->
      <DashboardSection
        v-if="activeTab === 'dashboard'"
        :watched-items="watchedItems"
        :user-reviews="userReviews"
        :liked-items="likedItems"
      />

      <BmolSection
        v-else-if="activeTab === 'bmol'"
        v-model:bmolItems="bmolItems"
        :is-owner="isOwner"
        :user-id="userId"
      />

      <ListsSection
        v-else-if="isLibraryListTab(activeTab)"
        :active-tab="activeTab"
        :user-id="userId"
        :is-owner="isOwner"
        :watchlist-items="watchlistItems"
        :watched-items="watchedItems"
        :liked-items="likedItems"
        :user-reviews="userReviews"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonProfile from "@/components/ui/skeleton/SkeletonProfile.vue"
import { ref, computed, onMounted, watch } from "vue"
import { useRoute } from "vue-router"
import { useAuthStore } from "@/stores/auth"
import { libraryApi, reviewApi, userApi, bmolApi } from "@/api/api"
import type { UserProfile } from "@/types/user"
import type {
  LibraryItemResponse,
  ReviewResponse,
  BMOLItemResponse,
  LibraryTab,
  LibraryListTab
} from "@/types"
import {
  ChevronLeft,
  User as UserIcon,
  Flame,
  Trophy,
  Bookmark,
  Eye,
  Heart,
  Star
} from "lucide-vue-next"

// Sub-components
import DashboardSection from "./components/DashboardSection.vue"
import ListsSection from "./components/ListsSection.vue"
import BmolSection from "./components/bmol/BmolSection.vue"

const route = useRoute()
const auth = useAuthStore()

const userId = computed(() => {
  const raw = route.params.userId
  return Number(Array.isArray(raw) ? raw[0] : raw)
})

const loading = ref(true)
const activeTab = ref<LibraryTab>("dashboard")

const profileUser = ref<UserProfile | null>(null)
const watchlistItems = ref<LibraryItemResponse[]>([])
const watchedItems = ref<LibraryItemResponse[]>([])
const likedItems = ref<LibraryItemResponse[]>([])
const userReviews = ref<ReviewResponse[]>([])
const bmolItems = ref<BMOLItemResponse[]>([])

const isOwner = computed(() => auth.user?.id === userId.value)

// Type guard to safely narrow activeTab to LibraryListTab for ListsSection
function isLibraryListTab(tab: LibraryTab): tab is LibraryListTab {
  return (
    tab === "watchlist" ||
    tab === "watched" ||
    tab === "likes" ||
    tab === "reviews"
  )
}

// Load everything on mount
const loadData = async () => {
  try {
    loading.value = true
    const id = userId.value

    // Fetch profile info
    const profileRes = await userApi.getProfile(id)
    profileUser.value = profileRes.data.user

    // Fetch lists in parallel
    const [watchlistRes, watchedRes, likedRes, reviewsRes, bmolRes] = await Promise.all([
      libraryApi.getVisibleUserLibrary(id, { list_type: "watchlist" }),
      libraryApi.getVisibleUserLibrary(id, { list_type: "watched" }),
      libraryApi.getVisibleUserLibrary(id, { list_type: "likes" }),
      reviewApi.getUserReviews(id),
      bmolApi.getUserBMOL(id)
    ])

    watchlistItems.value = watchlistRes.data.items
    watchedItems.value = watchedRes.data.items
    likedItems.value = likedRes.data.items
    userReviews.value = reviewsRes.data.reviews
    bmolItems.value = bmolRes.data.items
  } catch (err) {
    console.error("Failed to load library data:", err)
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
watch(userId, () => {
  activeTab.value = "dashboard"
  loadData()
})
</script>

<style scoped>
.library-root {
  font-family: var(--font-sans);
  background-color: var(--color-surface);
  color: #f3f4f6;
  min-height: 100vh;
  padding: 2.5rem 2rem;
  position: relative;
  overflow: hidden;
}

.grain {
  position: fixed;
  inset: 0;
  background-image: radial-gradient(rgba(255, 255, 255, 0.015) 1px, transparent 0);
  background-size: 24px 24px;
  pointer-events: none;
  z-index: 1;
}

.library-layout {
  position: relative;
  z-index: 2;
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

/* Back Link */
.back-link {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: #9ca3af;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
  transition: color 0.2s;
  align-self: flex-start;
}

.back-link:hover {
  color: #e1251b;
}

/* Summary Card Header */
.profile-summary-card {
  background: rgba(25, 25, 28, 0.65);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

@media (min-width: 768px) {
  .profile-summary-card {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
}

.user-info-section {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.avatar-wrapper {
  position: relative;
}

.avatar-ring {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: linear-gradient(135deg, #e1251b, #8a1612);
  padding: 2px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.avatar-fallback {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: #27272a;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
}

.level-chip {
  position: absolute;
  bottom: -4px;
  right: -4px;
  background: #e1251b;
  color: #ffffff;
  font-size: 0.675rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 8px;
  border: 2px solid #161618;
}

.user-meta {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 1.5rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
  letter-spacing: -0.01em;
}

.user-tag {
  font-size: 0.875rem;
  color: #9ca3af;
  margin: 0.125rem 0 0.5rem 0;
}

.user-bio {
  font-size: 0.875rem;
  color: #d1d5db;
  margin: 0;
  max-width: 500px;
  line-height: 1.5;
}

.user-counts-section {
  display: flex;
  gap: 2rem;
}

.count-box {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.count-val {
  font-size: 1.5rem;
  font-weight: 800;
  color: #ffffff;
}

.count-lbl {
  font-size: 0.75rem;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-top: 0.25rem;
}

/* Tabs */
.library-tabs {
  display: flex;
  gap: 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 0.5rem;
  overflow-x: auto;
}

.tab-btn {
  background: transparent;
  border: none;
  color: #9ca3af;
  padding: 0.75rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 10px;
  transition: background-color 0.2s, color 0.2s;
  white-space: nowrap;
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.03);
  color: #ffffff;
}

.tab-btn--active {
  background: rgba(225, 37, 27, 0.1);
  color: #e1251b;
}

.tab-count {
  font-size: 0.75rem;
  background: rgba(255, 255, 255, 0.08);
  color: #d1d5db;
  padding: 2px 6px;
  border-radius: 6px;
  margin-left: 0.25rem;
}

.tab-btn--active .tab-count {
  background: rgba(225, 37, 27, 0.2);
  color: #e1251b;
}
</style>
