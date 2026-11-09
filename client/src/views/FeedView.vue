<template>
  <main class="feed-page">
    <header class="feed-header">
      <div class="feed-header__text">
        <h1 class="feed-title">{{ t("feed.title") }}</h1>
        <p class="feed-subtitle">{{ t("feed.subtitle") }}</p>
      </div>

      <div class="feed-header__actions">
        <AppButton
          variant="outline"
          size="sm"
          icon-only
          :aria-label="t('feed.refresh')"
          :disabled="feedStore.loading"
          @click="feedStore.refresh"
        >
          <RefreshCw :size="16" :class="{ 'is-spinning': feedStore.loading }" aria-hidden="true" />
        </AppButton>

        <div ref="settingsRef" class="feed-settings">
          <AppButton
            variant="outline"
            size="sm"
            icon-only
            :aria-label="t('feed.settings.open')"
            :aria-expanded="isSettingsOpen"
            @click="isSettingsOpen = !isSettingsOpen"
          >
            <Settings :size="16" aria-hidden="true" />
          </AppButton>
          <Transition name="feed-pop">
            <div v-if="isSettingsOpen" class="feed-settings__panel">
              <FeedSettingsPanel />
            </div>
          </Transition>
        </div>
      </div>
    </header>

    <FeedFilters :model-value="feedStore.selectedCategory" @update:model-value="feedStore.setCategory" />

    <Transition name="feed-pill">
      <div v-if="feedStore.newItemsCount > 0" class="feed-new">
        <AppButton size="sm" role="status" @click="feedStore.showNewItems()">
          <ArrowUp :size="14" aria-hidden="true" />
          {{ t("feed.newItems", { n: feedStore.newItemsCount }, feedStore.newItemsCount) }}
        </AppButton>
      </div>
    </Transition>

    <FeedList
      :items="feedStore.items"
      :loading="feedStore.loading"
      :loading-more="feedStore.loadingMore"
      :error="feedStore.error"
      :has-more="feedStore.hasMore"
      :is-empty="feedStore.isEmpty"
      :filtered="feedStore.selectedCategory !== ''"
      :current-user-id="currentUserId"
      @load-more="feedStore.loadMore"
      @retry="feedStore.refresh"
      @update-visibility="changeVisibility"
      @delete-activity="removeActivity"
    />
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { ArrowUp, RefreshCw, Settings } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import FeedFilters from "@/components/feed/FeedFilters.vue"
import FeedList from "@/components/feed/FeedList.vue"
import FeedSettingsPanel from "@/components/feed/FeedSettingsPanel.vue"
import { useClickOutside } from "@/composables/useClickOutside"
import { useAuthStore } from "@/stores/auth"
import { useFeedStore } from "@/stores/feed"
import { useUIStore } from "@/stores/ui"
import type { Visibility } from "@/components/feed/feedConfig"

const { t } = useI18n()
const feedStore = useFeedStore()
const authStore = useAuthStore()
const ui = useUIStore()
const currentUserId = computed(() => authStore.user?.id)

// The store rolls back optimistic updates and rethrows, so surface the failure here.
async function changeVisibility(id: number, visibility: Visibility) {
  try {
    await feedStore.updateVisibility(id, visibility)
  } catch {
    ui.addToast(t("feed.manage.visibilityFailed"), "error")
  }
}

async function removeActivity(id: number) {
  try {
    await feedStore.deleteActivity(id)
  } catch {
    ui.addToast(t("feed.manage.deleteFailed"), "error")
  }
}

const isSettingsOpen = ref(false)
const settingsRef = ref<HTMLElement | null>(null)
useClickOutside(settingsRef, () => (isSettingsOpen.value = false))

onMounted(() => {
  feedStore.bindSocket()
  if (feedStore.items.length === 0) feedStore.fetchFeed(1)
})

onUnmounted(() => {
  feedStore.unbindSocket()
})
</script>

<style scoped>
.feed-page {
  max-width: 42rem;
  margin: 0 auto;
  padding: 1.25rem 1rem 4rem;
  color: var(--color-text-primary);
}
.feed-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}
.feed-title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.2;
}
.feed-subtitle {
  margin: 0.25rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
.feed-header__actions {
  display: flex;
  flex-shrink: 0;
  gap: 0.5rem;
}
.feed-settings {
  position: relative;
}
.feed-settings__panel {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: var(--z-popover);
  margin-top: 0.5rem;
  max-width: calc(100vw - 2rem);
}
.is-spinning {
  animation: feed-spin 0.8s linear infinite;
}
.feed-new {
  position: sticky;
  top: 4.5rem;
  z-index: 5;
  display: flex;
  justify-content: center;
  margin: 0.75rem 0;
}

.feed-pop-enter-active,
.feed-pill-enter-active {
  transition:
    opacity var(--motion-base) var(--ease-out-soft),
    transform var(--motion-base) var(--ease-spring);
}
.feed-pop-leave-active,
.feed-pill-leave-active {
  transition:
    opacity var(--motion-fast) ease,
    transform var(--motion-fast) ease;
}
.feed-pop-enter-from,
.feed-pop-leave-to,
.feed-pill-enter-from,
.feed-pill-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

@keyframes feed-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (min-width: 768px) {
  .feed-page {
    padding: 2rem 1.5rem 5rem;
  }
  .feed-title {
    font-size: var(--text-page);
  }
}

@media (prefers-reduced-motion: reduce) {
  .is-spinning {
    animation: none;
  }
  .feed-pop-enter-from,
  .feed-pop-leave-to,
  .feed-pill-enter-from,
  .feed-pill-leave-to {
    transform: none;
  }
}
</style>
