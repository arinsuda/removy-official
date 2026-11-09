<template>
  <article class="feed-item" :data-type="item.type">
    <RouterLink :to="actorTo" class="feed-item__avatar" tabindex="-1" aria-hidden="true">
      <AppAvatar :src="item.actor.avatar_url" :name="actorName" decorative />
    </RouterLink>

    <div class="feed-item__main">
      <header class="feed-item__head">
        <div class="feed-item__headline">
          <i18n-t :keypath="`feed.types.${knownType}.sentence`" tag="p" class="feed-item__sentence">
            <template #actor>
              <RouterLink :to="actorTo" class="feed-item__link feed-item__link--actor">{{ actorName }}</RouterLink>
            </template>
            <template #title>
              <RouterLink v-if="item.media" :to="mediaTo" class="feed-item__link">{{ mediaTitle }}</RouterLink>
              <span v-else>{{ mediaTitle }}</span>
            </template>
            <template #target>
              <RouterLink v-if="item.target_user" :to="targetTo" class="feed-item__link">{{ targetName }}</RouterLink>
            </template>
            <template #achievement>
              <strong>{{ item.message }}</strong>
            </template>
          </i18n-t>
          <p class="feed-item__meta">
            <span class="feed-item__badge" :class="`feed-item__badge--${meta.tone}`">
              <component :is="meta.icon" :size="12" aria-hidden="true" />
              {{ t(meta.labelKey) }}
            </span>
            <time :datetime="item.created_at" :title="absoluteTime">{{ relativeTime }}</time>
          </p>
        </div>

        <FeedItemMenu
          v-if="canManage"
          :visibility="item.visibility"
          @update-visibility="(v) => emit('update-visibility', item.id, v)"
          @delete-activity="emit('delete-activity', item.id)"
        />
      </header>

      <div v-if="item.type === 'review_created' && item.message" class="feed-item__review">
        <p ref="excerpt" class="feed-item__excerpt" :class="{ 'is-clamped': !expanded }">{{ item.message }}</p>
        <button v-if="overflowing || expanded" type="button" class="feed-item__toggle" :aria-expanded="expanded" @click="expanded = !expanded">
          {{ expanded ? t("feed.readLess") : t("feed.readMore") }}
        </button>
      </div>

      <RouterLink v-if="item.media" :to="mediaTo" class="feed-item__media">
        <img
          v-if="item.media.poster_url && !posterFailed"
          :src="item.media.poster_url"
          alt=""
          class="feed-item__poster"
          loading="lazy"
          decoding="async"
          @error="posterFailed = true"
        />
        <span v-else class="feed-item__poster feed-item__poster--empty" aria-hidden="true"><Film :size="18" /></span>
        <span class="feed-item__media-text">
          <span class="feed-item__media-type">{{ t(`feed.mediaType.${item.media.media_type === "tv" ? "tv" : "movie"}`) }}</span>
          <span class="feed-item__media-title">{{ mediaTitle }}</span>
          <MediaRating :ratings="item.media.ratings" :vote-average="item.media.vote_average" />
        </span>
      </RouterLink>

      <RouterLink v-if="item.type === 'user_followed' && item.target_user" :to="targetTo" class="feed-item__media">
        <AppAvatar :src="item.target_user.avatar_url" :name="targetName" decorative />
        <span class="feed-item__media-text">
          <span class="feed-item__media-title">{{ targetName }}</span>
          <span class="feed-item__handle">@{{ item.target_user.username }}</span>
        </span>
      </RouterLink>

      <footer v-if="likeable" class="feed-item__footer">
        <button
          type="button"
          class="feed-item__like"
          :class="{ 'is-liked': liked }"
          :aria-pressed="liked"
          :aria-label="liked ? t('feed.unlike') : t('feed.like')"
          :disabled="liking"
          @click="toggleLike"
        >
          <Heart :size="16" :fill="liked ? 'currentColor' : 'none'" aria-hidden="true" />
        </button>
      </footer>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Film, Heart } from "lucide-vue-next"
import { reviewApi } from "@/api/api"
import MediaRating from "@/components/movie/MediaRating.vue"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import { detailRoute } from "@/components/home/homeMedia"
import { useUIStore } from "@/stores/ui"
import type { FeedItemResponse } from "@/types/feed"
import { getActivityMeta } from "@/utils/feedActivity"
import { logApiError } from "@/utils/logApiError"
import { formatRelativeTime } from "@/utils/relativeTime"
import FeedItemMenu from "./FeedItemMenu.vue"
import { LIKEABLE_TYPES, type Visibility } from "./feedConfig"

const props = defineProps<{
  item: FeedItemResponse
  canManage?: boolean
}>()

const emit = defineEmits<{
  "update-visibility": [id: number, visibility: Visibility]
  "delete-activity": [id: number]
}>()

const { t, te, locale } = useI18n()
const ui = useUIStore()

const meta = computed(() => getActivityMeta(props.item.type))
const knownType = computed(() => (te(`feed.types.${props.item.type}.sentence`) ? props.item.type : "unknown"))
const actorName = computed(() => props.item.actor.display_name || props.item.actor.username)
const targetName = computed(() => props.item.target_user?.display_name || props.item.target_user?.username || "")
const mediaTitle = computed(() => props.item.media?.title || t("feed.untitled"))
const actorTo = computed(() => ({ name: "user-profile", params: { userId: props.item.actor.id } }))
const targetTo = computed(() => ({ name: "user-profile", params: { userId: props.item.target_user?.id ?? 0 } }))
const mediaTo = computed(() =>
  detailRoute({ id: props.item.media?.id ?? 0, type: props.item.media?.media_type === "tv" ? "tv" : "movie" }),
)

const relativeTime = computed(() => formatRelativeTime(props.item.created_at, locale.value))
const absoluteTime = computed(() =>
  new Intl.DateTimeFormat(locale.value, { dateStyle: "medium", timeStyle: "short" }).format(new Date(props.item.created_at)),
)

const posterFailed = ref(false)

// Review excerpt: clamp to three lines; the toggle shows only when text really overflows.
const expanded = ref(false)
const overflowing = ref(false)
const excerpt = useTemplateRef<HTMLElement>("excerpt")
async function measure() {
  await nextTick()
  const el = excerpt.value
  overflowing.value = !!el && !expanded.value && el.scrollHeight > el.clientHeight + 1
}
onMounted(measure)
watch(() => props.item.message, measure)

// The feed API does not report whether the viewer already liked a review, so state starts unliked.
const likeable = computed(() => LIKEABLE_TYPES.includes(props.item.type) && !!props.item.review_id && !props.canManage)
const liked = ref(false)
const liking = ref(false)
async function toggleLike() {
  const reviewId = props.item.review_id
  if (!reviewId || liking.value) return
  const next = !liked.value
  liked.value = next
  liking.value = true
  try {
    await (next ? reviewApi.likeReview(reviewId) : reviewApi.unlikeReview(reviewId))
  } catch (err) {
    liked.value = !next
    logApiError("feed like failed:", err)
    ui.addToast(t("feed.likeFailed"), "error")
  } finally {
    liking.value = false
  }
}
</script>

<style scoped>
.feed-item {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  transition: border-color var(--motion-fast) ease;
}
.feed-item:hover {
  border-color: var(--color-border-strong);
}
.feed-item__avatar {
  flex-shrink: 0;
  align-self: flex-start;
}
.feed-item__main {
  min-width: 0;
  flex: 1;
}
.feed-item__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
}
.feed-item__headline {
  min-width: 0;
}
.feed-item__sentence {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.5;
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}
.feed-item__link {
  font-weight: 600;
  color: var(--color-text-primary);
  text-decoration: none;
}
.feed-item__link:hover {
  color: var(--color-brand);
}
.feed-item__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin: 0.25rem 0 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.feed-item__badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-btn);
  font-weight: 600;
  background: color-mix(in srgb, var(--badge-tone) 14%, transparent);
  color: var(--badge-tone);
  --badge-tone: var(--color-text-secondary);
}
.feed-item__badge--success { --badge-tone: var(--color-success); }
.feed-item__badge--danger { --badge-tone: var(--color-danger); }
.feed-item__badge--info { --badge-tone: var(--color-info); }
.feed-item__badge--brand { --badge-tone: var(--color-brand); }
.feed-item__badge--warning { --badge-tone: var(--color-warning); }

.feed-item__review {
  margin-top: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-left: 2px solid var(--color-border-strong);
}
.feed-item__excerpt {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.6;
  color: var(--color-text-primary);
  white-space: pre-line;
  overflow-wrap: anywhere;
}
.feed-item__excerpt.is-clamped {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.feed-item__toggle {
  margin-top: 0.25rem;
  padding: 0;
  border: 0;
  background: none;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-secondary);
  cursor: pointer;
}
.feed-item__toggle:hover {
  color: var(--color-text-primary);
}

.feed-item__media {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.75rem;
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface);
  color: inherit;
  text-decoration: none;
  transition: border-color var(--motion-fast) ease;
}
.feed-item__media:hover {
  border-color: var(--color-border-strong);
}
.feed-item__poster {
  flex-shrink: 0;
  width: 3rem;
  aspect-ratio: 2 / 3;
  border-radius: 0.375rem;
  object-fit: cover;
  background: var(--color-surface-3);
}
.feed-item__poster--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
}
.feed-item__media-text {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.125rem;
}
.feed-item__media-type {
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-muted);
}
.feed-item__media-title {
  overflow: hidden;
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-primary);
  text-overflow: ellipsis;
  white-space: nowrap;
}
.feed-item__handle {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.feed-item__footer {
  display: flex;
  margin-top: 0.5rem;
}
.feed-item__like {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 0;
  border-radius: var(--radius-btn);
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color var(--motion-fast) ease,
    background-color var(--motion-fast) ease;
}
.feed-item__like:hover:not(:disabled) {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}
.feed-item__like.is-liked {
  color: var(--color-danger);
}
.feed-item__like.is-liked svg {
  animation: feed-like-pop var(--motion-base) var(--ease-spring);
}
@keyframes feed-like-pop {
  from {
    transform: scale(0.6);
  }
}

@media (min-width: 640px) {
  .feed-item {
    gap: 1rem;
    padding: 1.25rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .feed-item__like.is-liked svg {
    animation: none;
  }
}
</style>
