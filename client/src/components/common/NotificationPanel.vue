<template>
  <div class="noti-panel" v-click-outside="close">
    <div class="noti-header">
      <span class="noti-header-title">การแจ้งเตือน</span>
      <button
        v-if="store.unreadCount > 0"
        class="noti-header-action"
        @click="store.markAllAsRead()"
      >
        <CheckCheck :size="13" />
        อ่านทั้งหมด
      </button>
    </div>

    <div class="noti-body" @scroll="onScroll">
      <SkeletonList v-if="store.loading" avatar :count="4" class="noti-skeleton" />

      <div v-else-if="store.error" class="noti-state noti-state--error">
        <span>{{ store.error }}</span>
        <button class="noti-retry" @click="store.refresh()">ลองใหม่</button>
      </div>

      <div v-else-if="store.notifications.length === 0" class="noti-state">
        <BellOff :size="28" />
        <span>ยังไม่มีการแจ้งเตือน</span>
      </div>

      <template v-else>
        <div
          v-for="n in store.notifications"
          :key="n.id"
          class="noti-item"
          :class="{ 'noti-item--unread': !n.is_read }"
          role="button"
          tabindex="0"
          @click="handleClick(n)"
          @keydown.enter="handleClick(n)"
          @keydown.space.prevent="handleClick(n)"
        >
          <div class="noti-item-icon" :class="`noti-item-icon--${n.category}`">
            <UserPlus v-if="n.type === 'followed_you'" :size="16" />
            <Heart
              v-else-if="
                n.type === 'review_liked' || n.type === 'following_liked_review'
              "
              :size="16"
            />
            <MessageCircle
              v-else-if="
                n.type === 'review_commented' ||
                n.type === 'following_commented'
              "
              :size="16"
            />
            <ThumbsUp
              v-else-if="
                n.type === 'review_marked_helpful' ||
                n.type === 'following_marked_helpful'
              "
              :size="16"
            />
            <Trophy v-else-if="n.type === 'achievement_unlocked'" :size="16" />
            <Clapperboard
              v-else-if="n.type === 'movie_now_playing'"
              :size="16"
            />
            <PartyPopper v-else-if="n.type === 'welcome'" :size="16" />
            <Bell v-else :size="16" />
          </div>

          <div class="noti-item-content">
            <p class="noti-item-title">{{ getNotificationTitle(n) }}</p>
            <p class="noti-item-message">{{ n.message }}</p>
            <span class="noti-item-time">{{ relativeTime(n.created_at) }}</span>
          </div>

          <span v-if="!n.is_read" class="noti-unread-dot" />

          <button
            class="noti-item-delete"
            title="ลบการแจ้งเตือนนี้"
            @click.stop="store.removeNotification(n.id)"
          >
            <X :size="13" />
          </button>
        </div>

        <div v-if="store.loadingMore" class="noti-state noti-state--inline">
          <AppSpinner :size="15" />
        </div>
        <button
          v-else-if="store.hasMore"
          class="noti-loadmore"
          @click="store.loadMore()"
        >
          โหลดเพิ่มเติม
        </button>
      </template>
    </div>

    <div class="noti-footer">
      <button class="noti-viewall" @click="goToAllNotifications">
        ดูการแจ้งเตือนทั้งหมด
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
  import { useRouter } from "vue-router"
  import {
    Bell,
    BellOff,
    CheckCheck,
    Heart,
    MessageCircle,
    Trophy,
    UserPlus,
    X,
  } from "lucide-vue-next"
  import { useNotificationStore } from "@/stores/notification"
  import type { AppNotification } from "@/types/notification"
  import { ThumbsUp, Clapperboard, PartyPopper } from "lucide-vue-next"
  import {
    getNotificationTitle,
    getNotificationActionUrl,
  } from "@/types/notification"

  const store = useNotificationStore()
  const router = useRouter()

  const emit = defineEmits(["close"])

  function close() {
    emit("close")
  }

  async function handleClick(n: AppNotification) {
    if (!n.is_read) await store.markAsRead(n.id)
    const url = getNotificationActionUrl(n)
    if (url) {
      router.push(url)
      close()
    }
  }

  function goToAllNotifications() {
    router.push("/notifications")
    close()
  }

  function onScroll(e: Event) {
    const el = e.target as HTMLElement
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 60
    if (nearBottom) store.loadMore()
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

  const vClickOutside = {
    mounted(el: HTMLElement, binding: { value: () => void }) {
      ;(el as any)._clickOutsideHandler = (e: MouseEvent) => {
        if (!el.contains(e.target as Node)) binding.value()
      }

      document.addEventListener("click", (el as any)._clickOutsideHandler)
    },

    unmounted(el: HTMLElement) {
      document.removeEventListener("click", (el as any)._clickOutsideHandler)
    },
  }
</script>

<style scoped>
  .noti-item-icon--social {
    color: #3b82f6;
  }
  .noti-item-icon--media {
    color: var(--color-brand);
  }
  .noti-item-icon--achievement {
    color: #f5c518;
  }
  .noti-item-icon--system {
    color: #10b981;
  }
  .noti-panel {
    position: absolute;
    top: calc(100% + 10px);
    right: -10px;
    width: 340px;
    max-height: 440px;
    display: flex;
    flex-direction: column;
    background: #1f1f1f;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
    z-index: 200;
    overflow: hidden;
  }

  .noti-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.85rem 1rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    flex-shrink: 0;
  }
  .noti-header-title {
    font-size: 0.9rem;
    font-weight: 700;
    color: #fff;
  }
  .noti-header-action {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    background: none;
    border: none;
    color: var(--color-brand);
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    padding: 0.2rem 0.4rem;
    border-radius: 6px;
    transition: background 0.15s;
  }
  .noti-header-action:hover {
    background: rgba(245, 67, 44, 0.1);
  }

  .noti-body {
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
  }

  .noti-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 2.5rem 1rem;
    color: #777;
    font-size: 0.82rem;
    text-align: center;
  }
  .noti-state--inline {
    padding: 0.75rem;
  }
  .noti-state--error {
    color: var(--color-brand);
  }
  .noti-retry {
    background: rgba(255, 255, 255, 0.08);
    border: none;
    color: #fff;
    font-size: 0.75rem;
    padding: 0.35rem 0.75rem;
    border-radius: 6px;
    cursor: pointer;
  }

  .spin {
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .noti-item {
    position: relative;
    display: flex;
    align-items: flex-start;
    gap: 0.65rem;
    width: 100%;
    padding: 0.75rem 2.1rem 0.75rem 1rem;
    background: none;
    border: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    cursor: pointer;
    text-align: left;
    transition: background 0.15s;
  }
  .noti-item:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  .noti-item:focus-visible {
    outline: 2px solid rgba(245, 67, 44, 0.6);
    outline-offset: -2px;
  }
  .noti-item--unread {
    background: rgba(245, 67, 44, 0.05);
  }
  .noti-item--unread:hover {
    background: rgba(245, 67, 44, 0.09);
  }

  .noti-item-icon {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.08);
    color: #bbb;
    margin-top: 1px;
  }
  .noti-item-icon--follow {
    color: #3b82f6;
  }
  .noti-item-icon--review_like {
    color: var(--color-brand);
  }
  .noti-item-icon--review_comment,
  .noti-item-icon--comment_reply {
    color: #10b981;
  }
  .noti-item-icon--achievement_unlock {
    color: #f5c518;
  }

  .noti-item-content {
    min-width: 0;
    flex: 1;
  }
  .noti-item-title {
    font-size: 0.82rem;
    font-weight: 700;
    color: #fff;
    margin: 0 0 0.15rem;
  }
  .noti-item-message {
    font-size: 0.78rem;
    color: #ccc;
    line-height: 1.4;
    margin: 0 0 0.3rem;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .noti-item-time {
    font-size: 0.7rem;
    color: #777;
  }

  .noti-unread-dot {
    position: absolute;
    top: 0.9rem;
    right: 1.9rem;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--color-brand);
  }

  .noti-item-delete {
    position: absolute;
    top: 0.6rem;
    right: 0.5rem;
    background: none;
    border: none;
    color: #666;
    cursor: pointer;
    padding: 0.3rem;
    border-radius: 4px;
    opacity: 0;
    transition:
      opacity 0.15s,
      color 0.15s,
      background 0.15s;
  }
  .noti-item:hover .noti-item-delete {
    opacity: 1;
  }
  .noti-item-delete:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .noti-loadmore {
    width: 100%;
    padding: 0.7rem;
    background: none;
    border: none;
    color: #999;
    font-size: 0.78rem;
    cursor: pointer;
    transition: background 0.15s;
  }
  .noti-loadmore:hover {
    background: rgba(255, 255, 255, 0.05);
    color: #fff;
  }

  .noti-footer {
    flex-shrink: 0;
    border-top: 1px solid rgba(255, 255, 255, 0.07);
  }
  .noti-viewall {
    width: 100%;
    padding: 0.75rem;
    background: none;
    border: none;
    color: var(--color-brand);
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s;
  }
  .noti-viewall:hover {
    background: rgba(245, 67, 44, 0.08);
  }

  @media (max-width: 576px) {
    .noti-panel {
      position: fixed;
      top: 56px;
      right: 0.5rem;
      left: 0.5rem;
      width: auto;
      max-height: calc(100vh - 80px);
    }
  }
.noti-skeleton {
  padding: 1rem;
}
</style>
