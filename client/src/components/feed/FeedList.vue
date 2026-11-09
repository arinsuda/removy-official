<template>
  <div class="feed-list">
    <Transition name="feed-fade" mode="out-in">
      <SkeletonRegion v-if="loading && items.length === 0" key="loading">
        <FeedItemSkeleton v-for="n in 5" :key="n" />
      </SkeletonRegion>

      <AppErrorState
        v-else-if="error && items.length === 0"
        key="error"
        :title="t('feed.error')"
        @retry="emit('retry')"
      />

      <AppEmptyState
        v-else-if="isEmpty"
        key="empty"
        :icon="Clapperboard"
        :title="filtered ? t('feed.empty.filtered') : t('feed.empty.title')"
        :description="filtered ? undefined : t('feed.empty.description')"
      >
        <template v-if="!filtered" #action>
          <div class="feed-list__empty-actions">
            <AppButton :to="{ name: 'search-results', query: { q: '' } }" size="sm">
              <UserPlus :size="14" aria-hidden="true" />
              {{ t("feed.empty.findPeople") }}
            </AppButton>
            <AppButton :to="{ name: 'movies' }" variant="outline" size="sm">{{ t("feed.empty.browse") }}</AppButton>
          </div>
        </template>
      </AppEmptyState>

      <div v-else key="content">
        <TransitionGroup name="feed-item" tag="div" class="feed-list__items">
          <FeedItem
            v-for="item in items"
            :key="item.id"
            :item="item"
            :can-manage="item.actor.id === currentUserId"
            @update-visibility="(id, val) => emit('update-visibility', id, val)"
            @delete-activity="(id) => emit('delete-activity', id)"
          />
        </TransitionGroup>

        <div ref="sentinel" class="feed-list__sentinel" />

        <p v-if="loadingMore" class="feed-list__note" role="status">{{ t("feed.loadingMore") }}</p>
        <p v-else-if="!hasMore && items.length > 0" class="feed-list__note">{{ t("feed.end") }}</p>

        <div v-if="error" class="feed-list__inline-error">
          <p role="alert">{{ t("feed.error") }}</p>
          <AppButton variant="outline" size="sm" @click="emit('retry')">{{ t("common.retry") }}</AppButton>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Clapperboard, UserPlus } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import { useInfiniteScroll } from "@/composables/useInfiniteScroll"
import type { FeedItemResponse } from "@/types/feed"
import FeedItem from "./FeedItem.vue"
import FeedItemSkeleton from "./FeedItemSkeleton.vue"
import type { Visibility } from "./feedConfig"

const props = defineProps<{
  items: FeedItemResponse[]
  loading: boolean
  loadingMore: boolean
  error: string | null
  hasMore: boolean
  isEmpty: boolean
  /** A category filter is active, so the empty state is about the filter, not follows. */
  filtered?: boolean
  currentUserId?: number
}>()

const emit = defineEmits<{
  "load-more": []
  retry: []
  "update-visibility": [id: number, visibility: Visibility]
  "delete-activity": [id: number]
}>()

const { t } = useI18n()
const sentinel = ref<HTMLElement | null>(null)

useInfiniteScroll(sentinel, () => emit("load-more"), {
  disabled: computed(() => props.loading || props.loadingMore || !props.hasMore),
})
</script>

<style scoped>
.feed-list__items {
  position: relative;
}
.feed-list__empty-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.5rem;
}
.feed-list__sentinel {
  height: 1px;
}
.feed-list__note {
  padding: 1rem 0;
  text-align: center;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.feed-list__inline-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0;
  font-size: var(--text-caption);
  color: var(--color-danger);
}

.feed-item-enter-active,
.feed-item-leave-active {
  transition:
    opacity var(--motion-base) var(--ease-out-soft),
    transform var(--motion-base) var(--ease-out-soft);
}
.feed-item-enter-from {
  opacity: 0;
  transform: translateY(12px);
}
.feed-item-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}
.feed-item-leave-active {
  position: absolute;
  width: 100%;
}
.feed-item-move {
  transition: transform var(--motion-base) var(--ease-out-soft);
}

.feed-fade-enter-active,
.feed-fade-leave-active {
  transition: opacity var(--motion-fast) ease-out;
}
.feed-fade-enter-from,
.feed-fade-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .feed-item-enter-from,
  .feed-item-leave-to {
    transform: none;
  }
  .feed-item-move {
    transition: none;
  }
}
</style>
