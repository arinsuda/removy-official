<template>
  <HomeRail
    :title="t('home.friends.title')"
    :subtitle="t('home.friends.subtitle')"
    :see-all-to="{ name: 'feed' }"
    :loading="loading"
    :error="error"
    :empty="!items.length"
    :empty-title="t('home.friends.emptyTitle')"
    :empty-description="t('home.friends.emptyDescription')"
    wide
    :skeleton-count="4"
    @retry="emit('retry')"
    @visible="emit('visible')"
  >
    <template #empty>
      <AppButton :to="{ name: 'search-results', query: { q: '' } }" variant="outline" size="sm">
        <UserPlus :size="14" aria-hidden="true" />
        {{ t("home.friends.findPeople") }}
      </AppButton>
    </template>
    <article v-for="item in items" :key="item.id" class="friend-card">
      <RouterLink :to="{ name: 'user-profile', params: { userId: item.actor.id } }" class="friend-card__actor">
        <AppAvatar
          :src="item.actor.avatar_url"
          :name="item.actor.display_name || item.actor.username"
          size="sm"
          decorative
        />
        <span class="friend-card__name">{{ item.actor.display_name || item.actor.username }}</span>
      </RouterLink>
      <p class="friend-card__action">
        <component :is="meta(item).icon" :size="13" aria-hidden="true" />
        {{ t(meta(item).labelKey) }}
      </p>
      <RouterLink v-if="item.media" :to="mediaTo(item.media)" class="friend-card__media">
        <img
          v-if="item.media.poster_url"
          :src="item.media.poster_url"
          :alt="item.media.title ?? ''"
          loading="lazy"
          decoding="async"
        />
        <span class="friend-card__title">{{ item.media.title }}</span>
      </RouterLink>
      <time :datetime="item.created_at" class="friend-card__time">{{ formatRelativeTime(item.created_at, locale) }}</time>
    </article>
  </HomeRail>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { UserPlus } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import AppButton from "@/components/ui/AppButton.vue"
import type { FeedItemResponse, FeedMediaSummary } from "@/types/feed"
import { getActivityMeta } from "@/utils/feedActivity"
import { formatRelativeTime } from "@/utils/relativeTime"
import { detailRoute } from "./homeMedia"
import HomeRail from "./HomeRail.vue"

defineProps<{ items: FeedItemResponse[]; loading?: boolean; error?: boolean }>()
const emit = defineEmits<{ retry: []; visible: [] }>()
const { t, locale } = useI18n()

const meta = (item: FeedItemResponse) => getActivityMeta(item.type)
const mediaTo = (m: FeedMediaSummary) => detailRoute({ id: m.id, type: m.media_type === "tv" ? "tv" : "movie" })
</script>

<style scoped>
.friend-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.friend-card__actor {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  color: var(--color-text-primary);
  text-decoration: none;
}
.friend-card__name {
  overflow: hidden;
  font-size: var(--text-small);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.friend-card__action {
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
}
.friend-card__media {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 8px;
  min-width: 0;
  color: var(--color-text-primary);
  text-decoration: none;
}
.friend-card__media img {
  width: 36px;
  aspect-ratio: 2 / 3;
  flex-shrink: 0;
  object-fit: cover;
  border-radius: 4px;
}
.friend-card__title {
  display: -webkit-box;
  overflow: hidden;
  font-size: var(--text-small);
  font-weight: 600;
  line-height: 1.3;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.friend-card__time {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
</style>
