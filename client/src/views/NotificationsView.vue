<template>
  <div class="notifications-page">
    <div class="notifications-page-header">
      <h1 class="notifications-page-title">การแจ้งเตือน</h1>
      <button
        v-if="notiStore.unreadCount > 0"
        class="notifications-page-markall"
        @click="handleMarkAllRead"
      >
        <CheckCheck :size="15" />
        อ่านทั้งหมด
      </button>
    </div>

    <div class="notifications-page-tabs">
      <button
        v-for="tab in NOTIFICATION_CATEGORY_TABS"
        :key="tab.label"
        class="notifications-page-tab"
        :class="{
          'notifications-page-tab--active': activeCategory === tab.value,
        }"
        @click="selectCategory(tab.value)"
      >
        {{ tab.label }}
      </button>
    </div>

    <div class="notifications-page-body">
      <SkeletonList v-if="loading" avatar :count="7" />

      <AppErrorState v-else-if="error" :title="error" @retry="fetchList(1)" />

      <AppEmptyState
        v-else-if="items.length === 0"
        :icon="BellOff"
        title="ยังไม่มีการแจ้งเตือนในหมวดนี้"
      />

      <template v-else>
        <div
          v-for="n in items"
          :key="n.id"
          class="notifications-page-item"
          :class="{ 'notifications-page-item--unread': !n.is_read }"
          role="button"
          tabindex="0"
          @click="handleClick(n)"
          @keydown.enter="handleClick(n)"
          @keydown.space.prevent="handleClick(n)"
        >
          <div
            class="notifications-page-item-icon"
            :class="`notifications-page-item-icon--${n.category}`"
          >
            <UserPlus v-if="n.type === 'followed_you' || n.type === 'follow_requested'" :size="18" />
            <UserCheck v-else-if="n.type === 'follow_accepted'" :size="18" />
            <Heart
              v-else-if="
                n.type === 'review_liked' || n.type === 'following_liked_review'
              "
              :size="18"
            />
            <MessageCircle
              v-else-if="
                n.type === 'review_commented' ||
                n.type === 'following_commented'
              "
              :size="18"
            />
            <ThumbsUp
              v-else-if="
                n.type === 'review_marked_helpful' ||
                n.type === 'following_marked_helpful'
              "
              :size="18"
            />
            <Trophy v-else-if="n.type === 'achievement_unlocked'" :size="18" />
            <Clapperboard
              v-else-if="n.type === 'movie_now_playing'"
              :size="18"
            />
            <PartyPopper v-else-if="n.type === 'welcome'" :size="18" />
            <Bell v-else :size="18" />
          </div>

          <div class="notifications-page-item-content">
            <p class="notifications-page-item-title">
              {{ getNotificationTitle(n) }}
            </p>
            <p class="notifications-page-item-message">{{ n.message }}</p>
            <span class="notifications-page-item-time">{{
              relativeTime(n.created_at)
            }}</span>
          </div>

          <span v-if="!n.is_read" class="notifications-page-unread-dot" />
        </div>

        <div
          v-if="loadingMore"
          class="notifications-page-state notifications-page-state--inline"
        >
          <AppSpinner :size="18" />
        </div>
        <button
          v-else-if="hasMore"
          class="notifications-page-loadmore"
          @click="loadMore"
        >
          โหลดเพิ่มเติม
        </button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
  import { ref, computed, watch } from "vue"
  import { useRoute, useRouter } from "vue-router"
  import {
    Bell,
    BellOff,
    CheckCheck,
    Clapperboard,
    Heart,
    MessageCircle,
    PartyPopper,
    ThumbsUp,
    Trophy,
    UserPlus,
    UserCheck,
  } from "lucide-vue-next"
  import { notiApi } from "@/api/endpoints/notification"
  import { useNotificationStore } from "@/stores/notification"
  import {
    getNotificationActionUrl,
    getNotificationTitle,
    NOTIFICATION_CATEGORY_TABS,
    type AppNotification,
    type NotificationCategory,
    type NotificationPaginationMeta,
  } from "@/types/notification"

  const DEFAULT_PAGINATION: NotificationPaginationMeta = {
    page: 1,
    limit: 20,
    total: 0,
    total_pages: 0,
  }

  const route = useRoute()
  const router = useRouter()
  const notiStore = useNotificationStore()

  const items = ref<AppNotification[]>([])
  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref<string | null>(null)
  const pagination = ref<NotificationPaginationMeta>({ ...DEFAULT_PAGINATION })

  const hasMore = computed(
    () => pagination.value.page < pagination.value.total_pages,
  )

  const activeCategory = computed<NotificationCategory | null>(() => {
    const raw = route.query.category
    return (typeof raw === "string" ? raw : null) as NotificationCategory | null
  })

  function selectCategory(category: NotificationCategory | null) {
    if (category === activeCategory.value) return
    router.push({
      path: "/notifications",
      query: category ? { category } : {},
    })
  }

  async function fetchList(page = 1) {
    if (loading.value || loadingMore.value) return

    if (page === 1) loading.value = true
    else loadingMore.value = true
    error.value = null

    try {
      const res = await notiApi.list({
        page,
        limit: pagination.value.limit,
        category: activeCategory.value ?? undefined,
      })

      const next = res.data.notifications ?? []
      items.value = page === 1 ? next : [...items.value, ...next]

      pagination.value = {
        ...DEFAULT_PAGINATION,
        ...res.data.pagination,
        page,
        limit: pagination.value.limit,
      }
    } catch (err) {
      console.error("fetchList (notifications page) failed:", err)
      error.value = "โหลดการแจ้งเตือนไม่สำเร็จ"
    } finally {
      loading.value = false
      loadingMore.value = false
    }
  }

  function loadMore() {
    if (!hasMore.value || loadingMore.value) return
    fetchList(pagination.value.page + 1)
  }

  async function handleClick(n: AppNotification) {
    if (!n.is_read) {
      n.is_read = true
      const ok = await notiStore.markAsReadRemote(n.id)
      if (!ok) n.is_read = false
    }

    const url = getNotificationActionUrl(n)
    if (url) router.push(url)
  }

  async function handleMarkAllRead() {
    await notiStore.markAllAsRead()
    items.value.forEach(n => (n.is_read = true))
  }

  function relativeTime(iso: string): string {
    const diffMs = Date.now() - new Date(iso).getTime()
    const sec = Math.floor(diffMs / 1000)
    if (sec < 60) return "เมื่อสักครู่"
    const min = Math.floor(sec / 60)
    if (min < 60) return `${min} นาทีที่แล้ว`
    const hr = Math.floor(min / 60)
    if (hr < 24) return `${hr} ชั่วโมงที่แล้ว`
    const day = Math.floor(hr / 24)
    if (day < 7) return `${day} วันที่แล้ว`
    return new Date(iso).toLocaleDateString("th-TH", {
      day: "numeric",
      month: "short",
      year:
        new Date(iso).getFullYear() !== new Date().getFullYear()
          ? "numeric"
          : undefined,
    })
  }

  watch(
    activeCategory,
    () => {
      pagination.value = { ...DEFAULT_PAGINATION }
      items.value = []
      fetchList(1)
    },
    { immediate: true },
  )
</script>

<style scoped>
  .notifications-page {
    max-width: 720px;
    margin: 0 auto;
    padding: 2rem 1.25rem 4rem;
    color: #fff;
  }

  .notifications-page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.25rem;
  }

  .notifications-page-title {
    font-size: 1.5rem;
    font-weight: 800;
    margin: 0;
  }

  .notifications-page-markall {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    background: none;
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: var(--color-brand);
    font-size: 0.82rem;
    font-weight: 600;
    cursor: pointer;
    padding: 0.4rem 0.75rem;
    border-radius: 8px;
    transition: background 0.15s;
  }
  .notifications-page-markall:hover {
    background: rgba(245, 67, 44, 0.1);
  }

  .notifications-page-tabs {
    display: flex;
    gap: 0.5rem;
    overflow-x: auto;
    padding-bottom: 0.75rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    scrollbar-width: none;
  }
  .notifications-page-tabs::-webkit-scrollbar {
    display: none;
  }

  .notifications-page-tab {
    flex-shrink: 0;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid transparent;
    color: #aaa;
    font-size: 0.85rem;
    font-weight: 600;
    padding: 0.5rem 1rem;
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.15s;
    white-space: nowrap;
  }
  .notifications-page-tab:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.09);
  }
  .notifications-page-tab--active {
    background: var(--color-brand);
    color: #fff;
  }

  .notifications-page-body {
    display: flex;
    flex-direction: column;
  }

  .notifications-page-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.6rem;
    padding: 4rem 1rem;
    color: #777;
    font-size: 0.88rem;
    text-align: center;
  }
  .notifications-page-state--inline {
    padding: 1rem;
  }
  .spin {
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .notifications-page-item {
    position: relative;
    display: flex;
    align-items: flex-start;
    gap: 0.85rem;
    padding: 1rem;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.06);
    background: #1a1a1a;
    margin-bottom: 0.6rem;
    cursor: pointer;
    transition: background 0.15s;
  }
  .notifications-page-item:hover {
    background: #222;
  }
  .notifications-page-item:focus-visible {
    outline: 2px solid rgba(245, 67, 44, 0.6);
    outline-offset: -2px;
  }
  .notifications-page-item--unread {
    background: rgba(245, 67, 44, 0.06);
    border-color: rgba(245, 67, 44, 0.2);
  }
  .notifications-page-item--unread:hover {
    background: rgba(245, 67, 44, 0.1);
  }

  .notifications-page-item-icon {
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.08);
    color: #bbb;
  }
  .notifications-page-item-icon--social {
    color: #3b82f6;
  }
  .notifications-page-item-icon--media {
    color: var(--color-brand);
  }
  .notifications-page-item-icon--achievement {
    color: #f5c518;
  }
  .notifications-page-item-icon--system {
    color: #10b981;
  }

  .notifications-page-item-content {
    min-width: 0;
    flex: 1;
  }
  .notifications-page-item-title {
    font-size: 0.9rem;
    font-weight: 700;
    color: #fff;
    margin: 0 0 0.2rem;
  }
  .notifications-page-item-message {
    font-size: 0.84rem;
    color: #ccc;
    line-height: 1.45;
    margin: 0 0 0.35rem;
  }
  .notifications-page-item-time {
    font-size: 0.75rem;
    color: #777;
  }

  .notifications-page-unread-dot {
    position: absolute;
    top: 1.1rem;
    right: 1rem;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--color-brand);
  }

  .notifications-page-loadmore {
    width: 100%;
    padding: 0.85rem;
    background: rgba(255, 255, 255, 0.04);
    border: none;
    border-radius: 10px;
    color: #999;
    font-size: 0.82rem;
    cursor: pointer;
    transition: background 0.15s;
  }
  .notifications-page-loadmore:hover {
    background: rgba(255, 255, 255, 0.08);
    color: #fff;
  }
</style>
