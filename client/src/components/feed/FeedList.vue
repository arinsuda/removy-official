<template>
  <div>
    <Transition name="fade-fast" mode="out-in">
      <!-- 1. Loading Skeletons -->
      <SkeletonRegion v-if="loading && items.length === 0" key="loading">
        <FeedItemSkeleton v-for="n in 5" :key="n" />
      </SkeletonRegion>

      <!-- 2. Error Message -->
      <div
        v-else-if="error && items.length === 0"
        key="error"
        class="flex flex-col items-center gap-3 py-16 text-center"
      >
        <p class="text-sm text-white/50">{{ error }}</p>
        <button
          type="button"
          class="rounded-md bg-white/10 px-4 py-2 text-sm text-white hover:bg-white/20"
          @click="emit('retry')"
        >
          ลองอีกครั้ง
        </button>
      </div>

      <!-- 3. Empty State -->
      <div
        v-else-if="isEmpty"
        key="empty"
        class="flex flex-col items-center gap-2 py-20 text-center text-white/40"
      >
        <p class="text-sm">ยังไม่มีความเคลื่อนไหวในฟีด</p>
        <p class="text-xs">ติดตามเพื่อน ๆ เพิ่ม เพื่อดูกิจกรรมของพวกเขาที่นี่</p>
      </div>

      <!-- 4. Content List -->
      <div v-else key="content">
        <TransitionGroup name="feed-list" tag="div" class="relative">
          <FeedItem
            v-for="item in items"
            :key="item.id"
            :item="item"
            :can-manage="item.actor.id === currentUserId"
            @update-visibility="
              (id, val) => emit('update-visibility', id, val)
            "
            @delete-activity="(id) => emit('delete-activity', id)"
          />
        </TransitionGroup>

        <div ref="sentinel" class="h-1" />

        <div v-if="loadingMore" class="flex justify-center py-4">
          <span class="text-xs text-white/40">กำลังโหลดเพิ่มเติม...</span>
        </div>
        <p
          v-else-if="!hasMore && items.length > 0"
          class="py-4 text-center text-xs text-white/30"
        >
          ดูครบทุกกิจกรรมแล้ว
        </p>

        <div
          v-if="error"
          class="flex flex-col items-center gap-2 py-4 text-center"
        >
          <p class="text-xs text-red-400">{{ error }}</p>
          <button
            type="button"
            class="text-xs text-white/60 underline"
            @click="emit('retry')"
          >
            ลองอีกครั้ง
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
import { computed, ref } from "vue";
import type { FeedItemResponse } from "@/types/feed";
import { useInfiniteScroll } from "@/composables/useInfiniteScroll";
import FeedItem from "./FeedItem.vue";
import FeedItemSkeleton from "./FeedItemSkeleton.vue";

const props = defineProps<{
  items: FeedItemResponse[];
  loading: boolean;
  loadingMore: boolean;
  error: string | null;
  hasMore: boolean;
  isEmpty: boolean;
  currentUserId?: number;
}>();

const emit = defineEmits<{
  "load-more": [];
  retry: [];
  "update-visibility": [
    id: number,
    visibility: "default" | "public" | "followers" | "private",
  ];
  "delete-activity": [id: number];
}>();

const sentinel = ref<HTMLElement | null>(null);

useInfiniteScroll(sentinel, () => emit("load-more"), {
  disabled: computed(
    () => props.loading || props.loadingMore || !props.hasMore
  ),
});
</script>

<style scoped>
.feed-list-enter-active,
.feed-list-leave-active {
  transition: all 0.45s cubic-bezier(0.16, 1, 0.3, 1);
}
.feed-list-enter-from {
  opacity: 0;
  transform: translateY(16px) scale(0.98);
}
.feed-list-leave-to {
  opacity: 0;
  transform: translateY(-16px) scale(0.98);
}
.feed-list-leave-active {
  position: absolute;
  width: 100%;
}
.feed-list-move {
  transition: transform 0.45s cubic-bezier(0.16, 1, 0.3, 1);
}

.fade-fast-enter-active,
.fade-fast-leave-active {
  transition: opacity 0.2s ease-out;
}
.fade-fast-enter-from,
.fade-fast-leave-to {
  opacity: 0;
}
</style>
