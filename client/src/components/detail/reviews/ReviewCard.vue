<template>
  <article class="card">
    <header class="card__head">
      <button type="button" class="card__user" @click="emit('profile', review.user.id)">
        <AppAvatar :src="review.user.avatar_url ?? undefined" :name="review.user.display_name || review.user.username" size="md" />
        <span class="card__who">
          <span class="card__name">{{ review.user.display_name || review.user.username }}</span>
          <span class="card__date">{{ date }}</span>
        </span>
      </button>
      <span v-if="followStatus" class="card__follow" :class="`card__follow--${followStatus}`">
        {{ t(`reviews.follow.${followStatus}`) }}
      </span>
      <span class="card__score"><Star :size="14" fill="currentColor" aria-hidden="true" />{{ review.rating }}/5</span>
    </header>

    <p v-if="review.visibility !== 'public'" class="card__vis">
      <component :is="review.visibility === 'private' ? Lock : Users" :size="12" aria-hidden="true" />
      {{ t(`visibility.${review.visibility}`) }}
    </p>

    <p class="card__body">{{ review.body }}</p>

    <footer class="card__actions">
      <button type="button" class="card__btn" :class="{ 'card__btn--on': review.is_liked }" :aria-pressed="review.is_liked" :disabled="likePending" @click="emit('like', review)">
        <Heart :size="16" :fill="review.is_liked ? 'currentColor' : 'none'" aria-hidden="true" />
        <span>{{ review.like_count > 0 ? review.like_count : t("reviews.actions.like") }}</span>
      </button>
      <button type="button" class="card__btn" :class="{ 'card__btn--on': review.is_helpful_voted }" :aria-pressed="review.is_helpful_voted" :disabled="helpfulPending" @click="emit('helpful', review)">
        <ThumbsUp :size="16" :fill="review.is_helpful_voted ? 'currentColor' : 'none'" aria-hidden="true" />
        <span>{{ t("reviews.actions.helpful") }}{{ review.helpful_count > 0 ? ` (${review.helpful_count})` : "" }}</span>
      </button>
      <span v-if="review.comment_count > 0" class="card__comments">
        <MessageCircle :size="14" aria-hidden="true" />
        {{ t("reviews.actions.comments", { count: review.comment_count }, review.comment_count) }}
      </span>
      <button v-if="review.visibility === 'public'" type="button" class="card__btn" @click="emit('share', review)">
        <Share2 :size="16" aria-hidden="true" />
        <span>{{ t("share.share") }}</span>
      </button>
    </footer>
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { Heart, Lock, MessageCircle, Share2, Star, ThumbsUp, Users } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import type { ReviewResponse } from "@/types/review"

const props = defineProps<{
  review: ReviewResponse
  /** null hides the badge (guests and your own reviews). */
  followStatus: "none" | "pending" | "accepted" | null
  likePending?: boolean
  helpfulPending?: boolean
}>()
const emit = defineEmits<{
  like: [review: ReviewResponse]
  helpful: [review: ReviewResponse]
  share: [review: ReviewResponse]
  profile: [userId: number]
}>()

const { t, locale } = useI18n()
const date = computed(() => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(props.review.created_at)))
</script>

<style scoped>
.card { display: flex; flex-direction: column; gap: 0.6rem; padding: 1rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.card__head { display: flex; align-items: center; gap: 0.5rem; }
.card__user { display: flex; flex: 1; align-items: center; gap: 0.6rem; min-width: 0; padding: 0; text-align: left; color: inherit; background: none; border: 0; cursor: pointer; }
.card__who { display: flex; flex-direction: column; min-width: 0; }
.card__name { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); overflow-wrap: anywhere; }
.card__user:hover .card__name { color: var(--color-brand); text-decoration: underline; }
.card__date { font-size: var(--text-caption); color: var(--color-text-muted); }
.card__follow { padding: 0.1rem 0.5rem; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-muted); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-btn); white-space: nowrap; }
.card__follow--accepted { color: var(--color-brand); border-color: color-mix(in srgb, var(--color-brand) 40%, transparent); }
.card__score { display: inline-flex; align-items: center; gap: 0.25rem; font-size: var(--text-small); font-weight: 700; color: var(--color-rating); white-space: nowrap; }
.card__vis { display: inline-flex; align-items: center; gap: 0.3rem; align-self: flex-start; font-size: var(--text-caption); color: var(--color-text-muted); }
.card__body { font-size: var(--text-small); line-height: 1.6; color: var(--color-text-secondary); white-space: pre-line; overflow-wrap: anywhere; }
.card__actions { display: flex; flex-wrap: wrap; align-items: center; gap: 0.25rem 0.5rem; }
.card__btn { display: inline-flex; align-items: center; gap: 0.35rem; min-height: 36px; padding: 0 0.6rem; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-secondary); background: none; border: 0; border-radius: var(--radius-btn); cursor: pointer; transition: color var(--motion-fast) var(--ease-water), background-color var(--motion-fast) var(--ease-water); }
.card__btn--on { color: var(--color-brand); }
.card__btn:disabled { opacity: 0.6; cursor: progress; }
.card__comments { display: inline-flex; align-items: center; gap: 0.3rem; font-size: var(--text-caption); color: var(--color-text-muted); }
@media (hover: hover) {
  .card__btn:hover:not(:disabled) { color: var(--color-text-primary); background: var(--color-surface-3); }
}
@media (prefers-reduced-motion: reduce) {
  .card__btn { transition: none; }
}
</style>
