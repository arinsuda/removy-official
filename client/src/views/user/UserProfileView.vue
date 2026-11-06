<template>
  <div class="profile-root">
    <div class="grain" aria-hidden="true" />

    <SkeletonProfile v-if="loading" />

    <div v-else-if="!user" class="profile-empty">
      <Monitor class="empty-icon" />
      <p class="empty-title">User not found</p>
      <p class="empty-sub">Please try logging in again.</p>
    </div>

    <div v-else class="profile-layout">
      <!-- ── Hero Section ── -->
      <header class="profile-hero">
        <div class="hero-backdrop" aria-hidden="true" />
        <div class="hero-inner">
          <!-- Avatar -->
          <div class="avatar-shell">
            <div class="avatar-ring">
              <div class="avatar-inner">
                <Avatar
                  v-if="user?.avatar_url"
                  :image="user.avatar_url"
                  shape="circle"
                  class="avatar-img"
                />
                <div v-else class="avatar-fallback">
                  <UserIcon :size="32" />
                </div>
              </div>
            </div>
            <span class="level-chip">Lv.{{ user?.level ?? 1 }}</span>
          </div>

          <!-- Identity -->
          <div class="hero-identity">
            <h1 class="display-name">
              {{ user?.display_name || user?.username }}
            </h1>
            <p class="username">@{{ user?.username }}</p>
            <p v-if="user?.bio" class="bio">{{ user.bio }}</p>
            <p v-else class="bio bio--empty">No bio yet.</p>

            <!-- Inline stats row (Instagram-style) -->
            <div class="hero-stats">
              <div class="hstat">
                <span class="hstat-val">{{ user?.review_count ?? 0 }}</span>
                <span class="hstat-lbl">Reviews</span>
              </div>
              <div class="hstat-sep" />
              <div class="hstat">
                <span class="hstat-val">{{ user?.follower_count ?? 0 }}</span>
                <span class="hstat-lbl">Followers</span>
              </div>
              <div class="hstat-sep" />
              <div class="hstat">
                <span class="hstat-val">{{ user?.following_count ?? 0 }}</span>
                <span class="hstat-lbl">Following</span>
              </div>
              <div class="hstat-sep" />
              <div class="hstat">
                <span class="hstat-val">{{ joinYear }}</span>
                <span class="hstat-lbl">Joined</span>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="hero-actions">
            <button
              v-if="notMe"
              class="btn"
              :class="followButtonClass"
              :disabled="followLoading || isPendingFollow"
              @click="handleFollowButtonClick"
            >
              <component :is="followButtonIcon" :size="14" />
              <span>{{ followButtonLabel }}</span>
            </button>

            <button v-else class="btn btn-ghost" @click="showEdit = true">
              <Settings :size="14" />
              <span>Edit Profile</span>
            </button>
          </div>
        </div>
      </header>

      <div class="divider" />

      <!-- ── Mobile Dedicated Navigation Bar (< 960px) ── -->
      <div class="mobile-nav-section">
        <div class="mobile-nav-bar" role="tablist">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="mobile-nav-tab"
            :class="{ 'mobile-nav-tab--active': activeTab === tab.key }"
            role="tab"
            :aria-selected="activeTab === tab.key"
            @click="activeTab = tab.key"
          >
            <component :is="tab.icon" :size="15" />
            <span class="mobile-nav-label">{{ tab.label }}</span>
            <span v-if="tab.count !== undefined" class="mobile-nav-count">{{ tab.count }}</span>
          </button>
        </div>
      </div>

      <!-- ── Main Content Grid ── -->
      <div class="content-grid">
        <!-- Left: Desktop Nav Sidebar (PC >= 960px) -->
        <aside class="desktop-nav-aside">
          <p class="nav-eyebrow">Menu</p>
          <nav>
            <button
              v-for="tab in tabs"
              :key="tab.key"
              class="nav-item"
              :class="{ 'nav-item--active': activeTab === tab.key }"
              @click="activeTab = tab.key"
            >
              <component :is="tab.icon" :size="14" />
              <span>{{ tab.label }}</span>
              <span v-if="tab.count !== undefined" class="nav-count">{{
                tab.count
              }}</span>
            </button>
          </nav>

          <!-- Badges Section in Sidebar -->
          <div class="sidebar-badges">
            <p class="nav-eyebrow" style="margin-top: 24px;">Badges</p>
            <slot name="badges">
              <p class="badges-empty">No badges unlocked yet.</p>
            </slot>
          </div>
        </aside>

        <!-- Center: Tab Content Container -->
        <main class="content-main">
          <Transition name="fade-up" mode="out-in">
            <component
              :is="activeComponent"
              v-bind="activeProps"
              :key="activeComponentKey"
            />
          </Transition>
        </main>
      </div>
    </div>

    <!-- Edit Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showEdit"
          class="modal-backdrop"
          @click.self="showEdit = false"
        >
          <div class="modal-panel">
            <EditProfile :user="user" @close="handleEditClose" />
          </div>
        </div>
      </Transition>
    </Teleport>

    <ConfirmModal
      v-model="showUnfollowConfirm"
      list-type="unfollow"
      :item-name="user?.display_name || user?.username"
      :confirm-disabled="followLoading"
      @confirm="confirmUnfollow"
      @cancel="showUnfollowConfirm = false"
    />
  </div>
</template>

<script setup lang="ts">
import SkeletonProfile from "@/components/ui/skeleton/SkeletonProfile.vue"
  import { followApi, userApi } from "@/api/api"
  import { useAuthStore } from "@/stores/auth"
  import type { UserProfile } from "@/types/user"
  import type { ListType } from "@/types"
  import { computed, ref, watch, onBeforeUnmount } from "vue"
  import { useRoute } from "vue-router"

  import Avatar from "primevue/avatar"

  import {
    Star,
    Bookmark,
    Heart,
    Trophy,
    User as UserIcon,
    Settings,
    UserPlus,
    UserCheck,
    Clock,
    Monitor,
    TvMinimalPlay,
    Award,
  } from "lucide-vue-next"

  import ProfileInfo from "@/components/profile/ProfileInfo.vue"
  import UserReviews from "@/components/profile/UserReviews.vue"
  import UserWatchlist from "@/components/profile/UserWatchlist.vue"
  import UserWatched from "@/components/profile/UserWatched.vue"
  import UserLikes from "@/components/profile/UserLikes.vue"
  import UserAchievements from "@/components/profile/UserAchievements.vue"
  import UserBmol from "@/components/profile/UserBmol.vue"
  import EditProfile from "@/components/profile/EditProfile.vue"
  import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"

  const route = useRoute()
  const auth = useAuthStore()

  const user = ref<UserProfile | null>(null)
  const loading = ref(true)
  const activeTab = ref("reviews")
  const showEdit = ref(false)
  const isFollowing = ref(false)
  const isPendingFollow = ref(false)
  const followLoading = ref(false)
  const showUnfollowConfirm = ref(false)

  const userIdParam = computed(() => Number(route.params.userId))
  const notMe = computed(() => auth.user?.id !== userIdParam.value)

  const joinYear = computed(() => {
    if (!user.value?.created_at) return new Date().getFullYear()
    return new Date(user.value.created_at).getFullYear()
  })

  // Synchronize active tab from route query (e.g. ?tab=bol)
  watch(
    () => route.query.tab,
    (newTab) => {
      if (newTab && typeof newTab === "string") {
        const validKeys = ["reviews", "watchlist", "likes", "watched", "bol", "achievements"]
        if (validKeys.includes(newTab)) {
          activeTab.value = newTab
        }
      }
    },
    { immediate: true }
  )

  const tabs = computed(() => [
    {
      key: "reviews",
      label: "Reviews",
      icon: Star,
      count: (user.value as any)?.review_count,
    },
    {
      key: "watchlist",
      label: "Watchlist",
      icon: Bookmark,
      count: (user.value as any)?.watchlist_count,
    },
    {
      key: "likes",
      label: "Likes",
      icon: Heart,
      count: (user.value as any)?.like_count,
    },
    {
      key: "watched",
      label: "Watched",
      icon: TvMinimalPlay,
      count: (user.value as any)?.watched_count,
    },
    {
      key: "bol",
      label: "BOL",
      icon: Trophy,
      count: (user.value as any)?.bmol_count,
    },
    {
      key: "achievements",
      label: "Achievements",
      icon: Award,
    },
  ])

  const activeComponentMap: Record<string, any> = {
    reviews: UserReviews,
    watchlist: UserWatchlist,
    likes: UserLikes,
    watched: UserWatched,
    bol: UserBmol,
    achievements: UserAchievements,
  }

  const activeComponent = computed(
    () => activeComponentMap[activeTab.value] || UserReviews
  )

  const activeProps = computed(() => {
    const key = activeTab.value
    if (key === "reviews") return { userId: userIdParam.value }
    if (key === "watchlist")
      return { userId: userIdParam.value, listType: "watchlist" as ListType }
    if (key === "likes")
      return { userId: userIdParam.value, listType: "likes" as ListType }
    if (key === "watched")
      return { userId: userIdParam.value, listType: "watched" as ListType }
    if (key === "bol") return { userId: userIdParam.value }
    if (key === "achievements") return { userId: userIdParam.value }
    return {}
  })

  const activeComponentKey = computed(() => `${activeTab.value}-${userIdParam.value}`)

  const followButtonLabel = computed(() => {
    if (isFollowing.value) return "Following"
    if (isPendingFollow.value) return "Requested"
    return "Follow"
  })

  const followButtonClass = computed(() => {
    if (isFollowing.value) return "btn-following"
    if (isPendingFollow.value) return "btn-pending"
    return "btn-primary"
  })

  const followButtonIcon = computed(() => {
    if (isFollowing.value) return UserCheck
    if (isPendingFollow.value) return Clock
    return UserPlus
  })

  async function fetchProfile() {
    loading.value = true
    try {
      const targetId = userIdParam.value
      const res = await userApi.getProfile(targetId)
      user.value = (res.data as any).user || res.data
      if (notMe.value && auth.isLoggedIn) {
        const fRes = await followApi.getFollowStatus(targetId)
        isFollowing.value = Boolean(fRes.data.is_following)
        isPendingFollow.value = fRes.data.follow_status === "pending"
      }
    } catch (e) {
      user.value = null
    } finally {
      loading.value = false
    }
  }

  function handleFollowButtonClick() {
    if (isFollowing.value) {
      showUnfollowConfirm.value = true
    } else {
      handleFollow()
    }
  }

  async function handleFollow() {
    if (followLoading.value) return
    followLoading.value = true
    try {
      const res = await followApi.follow(userIdParam.value)
      const st = (res.data as any).status
      if (st === "following" || st === "accepted") {
        isFollowing.value = true
        isPendingFollow.value = false
        if (user.value) user.value.follower_count++
      } else if (st === "pending") {
        isFollowing.value = false
        isPendingFollow.value = true
      }
    } catch (e) {
      console.error(e)
    } finally {
      followLoading.value = false
    }
  }

  async function confirmUnfollow() {
    if (followLoading.value) return
    followLoading.value = true
    try {
      await followApi.unfollow(userIdParam.value)
      isFollowing.value = false
      isPendingFollow.value = false
      if (user.value) user.value.follower_count = Math.max(0, user.value.follower_count - 1)
      showUnfollowConfirm.value = false
    } catch (e) {
      console.error(e)
    } finally {
      followLoading.value = false
    }
  }

  function handleEditClose(updatedUser?: UserProfile) {
    showEdit.value = false
    if (updatedUser) {
      user.value = updatedUser
    }
  }

  watch(userIdParam, () => fetchProfile(), { immediate: true })
</script>

<style scoped>
  .profile-root {
    --c-bg: var(--color-surface);
    --c-surface: var(--color-surface-2);
    --c-card: var(--color-surface-2);
    --c-border: var(--color-border);
    --c-border-h: var(--color-border-strong);
    --c-red: var(--color-brand);
    --c-red-dim: color-mix(in srgb, var(--color-brand) 8%, transparent);
    --c-text: var(--color-text-primary);
    --c-sub: var(--color-text-secondary);
    --c-muted: var(--color-text-muted);
    --font-display: var(--font-sans);
    --font-ui: var(--font-sans);
    --radius: 10px;
    --ease: cubic-bezier(0.16, 1, 0.3, 1);
    --profile-gutter: 24px;

    position: relative;
    min-height: 100vh;
    width: 100%;
    overflow-x: hidden;
    background: var(--c-bg);
    color: var(--c-text);
    font-family: var(--font-sans);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  /* Grain overlay */
  .grain {
    position: fixed;
    inset: 0;
    pointer-events: none;
    z-index: 0;
    opacity: 0.025;
    background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
    background-size: 200px 200px;
  }

  /* ─────────── Empty ─────────── */
  .profile-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-height: 60vh;
  }
  .empty-icon {
    width: 28px;
    height: 28px;
    color: var(--c-muted);
  }
  .empty-title {
    font-size: 0.95rem;
    font-weight: 500;
    margin: 0;
  }
  .empty-sub {
    font-size: 0.8rem;
    color: var(--c-sub);
    margin: 0;
  }

  /* ─────────── Layout ─────────── */
  .profile-layout {
    position: relative;
    z-index: 1;
    max-width: 1100px;
    margin: 0 auto;
    padding: 0 var(--profile-gutter) 80px;
    animation: fadeIn 0.5s var(--ease) both;
  }
  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* ─────────── Hero ─────────── */
  .profile-hero {
    position: relative;
    padding: 56px 0 40px;
    margin-bottom: 0;
  }

  /* Subtle red-tinted backdrop at top */
  .hero-backdrop {
    position: absolute;
    top: 0;
    left: calc(-1 * var(--profile-gutter));
    right: calc(-1 * var(--profile-gutter));
    height: 220px;
    background: linear-gradient(
      180deg,
      rgba(225, 37, 27, 0.06) 0%,
      transparent 100%
    );
    pointer-events: none;
  }

  .hero-inner {
    position: relative;
    display: flex;
    align-items: flex-end;
    gap: 32px;
  }

  /* ── Avatar ── */
  .avatar-shell {
    position: relative;
    flex-shrink: 0;
  }

  .avatar-ring {
    width: 120px;
    height: 120px;
    border-radius: 50%;
    padding: 3px;
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.18),
      rgba(255, 255, 255, 0.04)
    );
  }

  .avatar-inner {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    overflow: hidden;
    background: var(--c-card);
    border: 2px solid var(--c-bg);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .avatar-img {
    width: 100% !important;
    height: 100% !important;
  }
  .avatar-img :deep(img) {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
    display: block;
  }

  .avatar-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--c-muted);
  }

  .level-chip {
    position: absolute;
    bottom: -2px;
    left: 50%;
    transform: translateX(-50%);
    background: var(--c-surface);
    border: 1px solid var(--c-border-h);
    color: #fff;
    font-size: 0.55rem;
    font-weight: 700;
    letter-spacing: 0.06em;
    padding: 2px 8px;
    border-radius: 20px;
    white-space: nowrap;
    text-transform: uppercase;
  }

  /* ── Identity ── */
  .hero-identity {
    flex: 1;
    min-width: 0;
    padding-bottom: 6px;
  }

  .display-name {
    font-family: var(--font-sans);
    font-size: 2rem;
    font-weight: 700;
    color: #fff;
    margin: 0 0 4px;
    line-height: 1.1;
    letter-spacing: -0.03em;
  }

  .username {
    font-size: 0.85rem;
    color: var(--c-sub);
    margin: 0 0 10px;
    letter-spacing: -0.01em;
  }

  .bio {
    font-size: 0.88rem;
    color: rgba(240, 240, 240, 0.72);
    line-height: 1.65;
    margin: 0 0 18px;
    max-width: 520px;
    font-weight: 400;
  }
  .bio--empty {
    color: var(--c-muted);
    font-style: italic;
  }

  /* ── Inline hero stats ── */
  .hero-stats {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .hstat {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .hstat-val {
    font-family: var(--font-sans);
    font-size: 1.1rem;
    font-weight: 700;
    color: #fff;
    letter-spacing: -0.025em;
    line-height: 1;
  }
  .hstat-lbl {
    font-size: 0.62rem;
    font-weight: 500;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--c-sub);
  }
  .hstat-sep {
    width: 1px;
    height: 30px;
    background: var(--c-border);
    flex-shrink: 0;
  }

  /* ── Buttons ── */
  .hero-actions {
    display: flex;
    gap: 8px;
    align-self: flex-end;
    padding-bottom: 6px;
    flex-shrink: 0;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-sans);
    font-size: 0.78rem;
    font-weight: 500;
    padding: 9px 18px;
    border-radius: 8px;
    border: none;
    cursor: pointer;
    transition: all 0.2s var(--ease);
    white-space: nowrap;
    touch-action: manipulation;
  }
  .btn-primary {
    background: var(--c-red);
    color: #fff;
  }
  .btn-primary:hover {
    background: #ff3b30;
    transform: translateY(-1px);
    box-shadow: 0 6px 16px rgba(225, 37, 27, 0.3);
  }
  .btn-ghost {
    background: transparent;
    color: var(--c-text);
    border: 1px solid var(--c-border-h);
  }
  .btn-ghost:hover {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.2);
  }

  /* ─────────── Divider ─────────── */
  .divider {
    height: 1px;
    background: var(--c-border);
    margin-bottom: 24px;
  }

  /* ─────────── Mobile Navigation Section (< 960px) ─────────── */
  .mobile-nav-section {
    display: none;
  }

  /* ─────────── Main Content Grid (PC >= 960px) ─────────── */
  .content-grid {
    display: grid;
    grid-template-columns: 188px 1fr;
    gap: 0;
    min-height: 480px;
    width: 100%;
  }

  /* Desktop Nav Sidebar */
  .desktop-nav-aside {
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--c-border);
    padding-right: 20px;
    padding-top: 4px;
  }

  .nav-eyebrow {
    font-size: 0.58rem;
    font-weight: 700;
    letter-spacing: 0.09em;
    text-transform: uppercase;
    color: var(--c-muted);
    margin: 0 0 12px 8px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 9px;
    width: 100%;
    padding: 9px 10px;
    border-radius: 7px;
    border: none;
    background: none;
    font-family: var(--font-sans);
    font-size: 0.8rem;
    color: var(--c-sub);
    cursor: pointer;
    transition: all 0.15s var(--ease);
    margin-bottom: 2px;
    text-align: left;
    touch-action: manipulation;
  }
  .nav-item:hover {
    color: var(--c-text);
    background: rgba(255, 255, 255, 0.03);
  }
  .nav-item--active {
    color: #fff;
    background: rgba(255, 255, 255, 0.06);
    font-weight: 500;
  }
  .nav-count {
    margin-left: auto;
    font-size: 0.58rem;
    font-weight: 700;
    background: var(--c-red);
    color: #fff;
    padding: 1px 5px;
    border-radius: 4px;
  }

  /* Main Content Container */
  .content-main {
    padding: 0 28px;
    min-width: 0;
    width: 100%;
    box-sizing: border-box;
  }

  /* Sidebar Badges */
  .sidebar-badges {
    margin-top: 24px;
    padding-top: 16px;
    border-top: 1px dashed var(--c-border);
  }
  .badges-empty {
    font-size: 0.78rem;
    color: var(--c-muted);
    margin: 0 0 0 8px;
  }

  /* ─────────── Transitions ─────────── */
  .fade-up-enter-active,
  .fade-up-leave-active {
    transition: all 0.25s var(--ease);
  }
  .fade-up-enter-from {
    opacity: 0;
    transform: translateY(8px);
  }
  .fade-up-leave-to {
    opacity: 0;
    transform: translateY(-4px);
  }

  /* Modal */
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.72);
    backdrop-filter: blur(6px);
    z-index: 999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    touch-action: manipulation;
  }
  .modal-panel {
    background: var(--c-surface);
    border: 1px solid var(--c-border-h);
    border-radius: 14px;
    width: 100%;
    max-width: 480px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.6);
    touch-action: manipulation;
  }
  .modal-enter-active,
  .modal-leave-active {
    transition: all 0.3s var(--ease);
  }
  .modal-enter-from,
  .modal-leave-to {
    opacity: 0;
    transform: scale(0.96);
  }
  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  .btn-following {
    background: rgba(255, 255, 255, 0.06);
    color: #fff;
    border: 1px solid rgba(255, 255, 255, 0.16);
  }

  .btn-following:hover:not(:disabled) {
    background: rgba(225, 37, 27, 0.12);
    color: #ff6b6b;
    border-color: rgba(225, 37, 27, 0.32);
  }

  .btn-pending {
    background: rgba(255, 184, 0, 0.1);
    color: #ffb800;
    border: 1px solid rgba(255, 184, 0, 0.22);
  }

  /* ─────────── Responsive Mobile & Tablet (< 960px) ─────────── */
  @media (max-width: 959px) {
    .desktop-nav-aside {
      display: none !important;
    }

    .mobile-nav-section {
      display: block;
      width: 100%;
      margin-bottom: 20px;
      box-sizing: border-box;
    }

    .mobile-nav-bar {
      display: flex;
      gap: 8px;
      overflow-x: auto;
      -webkit-overflow-scrolling: touch;
      padding: 6px;
      background: #141418;
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 12px;
      scrollbar-width: none;
      touch-action: manipulation;
    }

    .mobile-nav-bar::-webkit-scrollbar {
      display: none;
    }

    .mobile-nav-tab {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 10px 14px;
      border-radius: 8px;
      background: transparent;
      border: 1px solid transparent;
      color: var(--c-sub);
      font-size: 0.82rem;
      font-weight: 500;
      white-space: nowrap;
      cursor: pointer;
      flex-shrink: 0;
      transition: all 0.2s ease;
      touch-action: manipulation;
    }

    .mobile-nav-tab--active {
      background: rgba(225, 37, 27, 0.15);
      color: #ffffff;
      border-color: rgba(225, 37, 27, 0.35);
      font-weight: 600;
      box-shadow: 0 2px 8px rgba(225, 37, 27, 0.2);
    }

    .mobile-nav-tab--active svg {
      color: var(--c-red);
    }

    .mobile-nav-count {
      font-size: 0.62rem;
      font-weight: 700;
      background: var(--c-red);
      color: #ffffff;
      padding: 1px 6px;
      border-radius: 10px;
    }

    .content-grid {
      grid-template-columns: 1fr !important;
      width: 100% !important;
      max-width: 100% !important;
      min-width: 0 !important;
    }

    .content-main {
      padding: 0 !important;
      min-width: 0 !important;
      width: 100% !important;
      max-width: 100% !important;
      box-sizing: border-box !important;
      overflow-x: hidden !important;
    }
  }

  @media (max-width: 720px) {
    .profile-hero {
      padding: 40px 0 32px;
    }
    .hero-inner {
      flex-direction: column;
      align-items: flex-start;
      gap: 20px;
    }
    .hero-actions {
      width: 100%;
      align-self: auto;
      padding-bottom: 0;
    }
    .hero-actions .btn {
      flex: 1;
      justify-content: center;
    }
    .display-name {
      font-size: 1.65rem;
    }
    .avatar-ring {
      width: 96px;
      height: 96px;
    }
  }

  @media (max-width: 480px) {
    .profile-root {
      --profile-gutter: 16px;
    }
    .profile-layout {
      padding-bottom: 60px;
    }
    .display-name {
      font-size: 1.4rem;
    }
    .avatar-ring {
      width: 84px;
      height: 84px;
    }
    .hero-stats {
      gap: 10px 14px;
      flex-wrap: wrap;
    }
    .hstat-val {
      font-size: 0.95rem;
    }
  }

  @media (max-width: 360px) {
    .profile-root {
      --profile-gutter: 12px;
    }
    .display-name {
      font-size: 1.25rem;
    }
    .avatar-ring {
      width: 72px;
      height: 72px;
    }
  }
</style>
