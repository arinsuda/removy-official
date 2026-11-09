<template>
  <main class="noti-page">
    <header class="noti-page__header">
      <div>
        <h1 class="noti-page__title">{{ t("notifications.title") }}</h1>
        <p class="noti-page__subtitle">{{ t("notifications.subtitle") }}</p>
      </div>
      <AppButton v-if="notiStore.unreadCount > 0" variant="outline" size="sm" @click="list.markAllRead">
        <CheckCheck :size="14" aria-hidden="true" />
        {{ t("notifications.markAllAsRead") }}
      </AppButton>
    </header>

    <NotificationFilters v-model:category="categoryKey" v-model:unread-only="unreadOnly" />

    <div class="noti-page__body" :aria-busy="list.loading.value">
      <SkeletonList v-if="list.loading.value" avatar :count="7" />

      <AppErrorState v-else-if="list.error.value && !list.items.value.length" :title="t('notifications.error')" @retry="list.reload" />

      <AppEmptyState
        v-else-if="!list.items.value.length"
        :icon="BellOff"
        :title="emptyTitle"
        :description="t('notifications.emptyStates.description')"
      />

      <template v-else>
        <NotificationGroups
          :items="list.items.value"
          :fresh-ids="list.freshIds.value"
          @open="open"
          @remove="list.remove"
          @responded="list.markRead"
        />

        <div v-if="list.loadingMore.value" class="noti-page__more" role="status">
          <AppSpinner :size="18" />
          <span class="sr-only">{{ t("notifications.loadingMore") }}</span>
        </div>
        <div v-else-if="list.hasMore.value" class="noti-page__more">
          <AppButton variant="outline" @click="list.loadMore">{{ t("notifications.loadMore") }}</AppButton>
        </div>
      </template>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { useRoute, useRouter } from "vue-router"
import { BellOff, CheckCheck } from "lucide-vue-next"
import NotificationFilters from "@/components/notifications/NotificationFilters.vue"
import NotificationGroups from "@/components/notifications/NotificationGroups.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useNotificationList } from "@/composables/useNotificationList"
import { useNotificationStore } from "@/stores/notification"
import {
  getNotificationActionUrl,
  NOTIFICATION_CATEGORIES,
  type AppNotification,
  type NotificationCategory,
} from "@/types/notification"

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const notiStore = useNotificationStore()

// Filters live in the URL (?category=social&unread=1) so they survive reloads and back/forward.
const category = computed<NotificationCategory | null>(() => {
  const raw = route.query.category
  return NOTIFICATION_CATEGORIES.find((c) => c === raw) ?? null
})
const unreadOnly = computed({
  get: () => route.query.unread === "1",
  set: (on: boolean) => setQuery({ unread: on ? "1" : undefined }),
})
const categoryKey = computed({
  get: () => category.value ?? "all",
  set: (key: string) => setQuery({ category: key === "all" ? undefined : key }),
})

function setQuery(patch: Record<string, string | undefined>) {
  const query = { ...route.query, ...patch }
  for (const key of Object.keys(query)) if (query[key] === undefined) delete query[key]
  router.push({ path: "/notifications", query })
}

const list = useNotificationList(category, unreadOnly)

const emptyTitle = computed(() =>
  unreadOnly.value
    ? t("notifications.emptyStates.unread")
    : t(category.value ? "notifications.emptyStates.category" : "notifications.emptyStates.all"),
)

async function open(n: AppNotification) {
  await list.markRead(n)
  const url = getNotificationActionUrl(n)
  if (url) router.push(url)
}
</script>

<style scoped>
.noti-page {
  max-width: 45rem;
  margin: 0 auto;
  padding: 1.25rem 1rem 4rem;
  color: var(--color-text-primary);
}
.noti-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}
.noti-page__title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.2;
}
.noti-page__subtitle {
  margin: 0.25rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
.noti-page__more {
  display: flex;
  justify-content: center;
  padding: 1.25rem 0;
}

@media (min-width: 768px) {
  .noti-page {
    padding: 2rem 1.5rem 5rem;
  }
  .noti-page__title {
    font-size: var(--text-page);
  }
}
</style>
