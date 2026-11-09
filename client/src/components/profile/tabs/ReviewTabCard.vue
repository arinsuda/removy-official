<template>
  <article class="rc" :data-tier="tier">
    <RouterLink :to="route" class="rc__poster" :aria-label="title">
      <img v-if="review.media.poster_url" :src="review.media.poster_url" alt="" loading="lazy" decoding="async" />
      <Film v-else :size="18" aria-hidden="true" />
    </RouterLink>

    <div class="rc__main">
      <header class="rc__head">
        <div class="rc__titles">
          <RouterLink :to="route" class="rc__title">{{ title }}</RouterLink>
          <div class="rc__rating">
            <StarRatingInput :model-value="review.rating" readonly :show-label="false" :size="13" />
            <span class="rc__num">{{ review.rating.toFixed(1) }}</span>
          </div>
        </div>
        <PopcornScoreBadge :average="review.rating" :count="1" class="rc__badge" />
      </header>

      <p class="rc__meta">
        <time :datetime="review.created_at">{{ date }}</time>
        <span v-if="edited" class="rc__edited">{{ t("profile.reviews.editedAt", { date: editedDate }) }}</span>
        <span v-if="isOwner" class="rc__vis" :title="t(`visibility.${review.visibility}`)">
          <component :is="visibilityIcon" :size="12" aria-hidden="true" />{{ t(`visibility.${review.visibility}`) }}
        </span>
      </p>

      <template v-if="review.body">
        <p :id="bodyId" class="rc__body" :class="{ 'rc__body--clamp': !open && long }">{{ review.body }}</p>
        <button v-if="long" type="button" class="rc__more" :aria-expanded="open" :aria-controls="bodyId" @click="open = !open">
          {{ open ? t("profile.tabContent.reviews.readLess") : t("profile.tabContent.reviews.readMore") }}
        </button>
      </template>

      <footer class="rc__foot">
        <span class="rc__stat" :class="{ 'is-on': review.is_liked }">
          <Heart :size="14" :fill="review.is_liked ? 'currentColor' : 'none'" aria-hidden="true" />{{ review.like_count ?? 0 }}
          <span class="sr-only">{{ t("profile.tabContent.reviews.likes", { n: review.like_count ?? 0 }) }}</span>
        </span>
        <span class="rc__stat" :class="{ 'is-on': review.is_helpful_voted }">
          <ThumbsUp :size="14" :fill="review.is_helpful_voted ? 'currentColor' : 'none'" aria-hidden="true" />{{ t("profile.reviews.helpful", { count: review.helpful_count ?? 0 }) }}
        </span>
        <span v-if="isOwner" class="rc__actions">
          <TabIconButton data-action="edit" :label="`${t('profile.reviews.edit')}: ${title}`" @click="emit('edit', review)"><Pencil :size="15" aria-hidden="true" /></TabIconButton>
          <TabIconButton data-action="delete" danger :label="`${t('profile.reviews.delete')}: ${title}`" @click="emit('delete', review)"><Trash2 :size="15" aria-hidden="true" /></TabIconButton>
        </span>
      </footer>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink } from "vue-router"
import { Film, Globe, Heart, Lock, Pencil, ThumbsUp, Trash2, Users } from "lucide-vue-next"
import { detailRoute } from "@/components/home/homeMedia"
import PopcornScoreBadge from "@/components/movie/PopcornScoreBadge.vue"
import StarRatingInput from "@/components/movie/Starratinginput.vue"
import { popcornTier } from "@/components/movie/popcornTier"
import type { ReviewResponse } from "@/types"
import TabIconButton from "./TabIconButton.vue"

const CLAMP_CHARS = 220

const props = defineProps<{ review: ReviewResponse; isOwner: boolean }>()
const emit = defineEmits<{ edit: [review: ReviewResponse]; delete: [review: ReviewResponse] }>()

const { t, locale } = useI18n()
const bodyId = useId()
const open = ref(false)

const title = computed(() => props.review.media.title || `#${props.review.media.id}`)
const route = computed(() => detailRoute({ id: props.review.media.id, type: props.review.media.media_type === "tv" ? "tv" : "movie" }))
const tier = computed(() => popcornTier(props.review.rating, 1))
const long = computed(() => props.review.body.length > CLAMP_CHARS || props.review.body.split("\n").length > 4)
const visibilityIcon = computed(() => (props.review.visibility === "private" ? Lock : props.review.visibility === "followers" ? Users : Globe))
const edited = computed(() => props.review.updated_at !== props.review.created_at)

const date = computed(() => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(props.review.created_at)))
const editedDate = computed(() => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium", timeStyle: "short" }).format(new Date(props.review.updated_at)))
</script>

<style scoped>
.rc { display: flex; gap: 0.875rem; padding: 0.875rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-left: 3px solid var(--color-border-strong); border-radius: var(--radius-card); transition: box-shadow var(--motion-fast) var(--ease-water); }
/* The left rim carries the POPCORN tier of the rating, like the poster rim on home cards. */
.rc[data-tier="hot"] { border-left-color: var(--color-brand); background: linear-gradient(90deg, color-mix(in srgb, var(--color-brand) 8%, transparent), transparent 40%), var(--color-surface-2); }
.rc[data-tier="stale"] { border-left-color: var(--color-text-muted); }
.rc__poster { display: grid; flex: none; place-items: center; align-self: flex-start; width: 4rem; aspect-ratio: 2 / 3; overflow: hidden; color: var(--color-text-muted); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: 0.5rem; }
.rc__poster img { width: 100%; height: 100%; object-fit: cover; }
.rc__main { display: flex; flex: 1; flex-direction: column; gap: 0.4rem; min-width: 0; }
.rc__head { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.5rem; }
.rc__titles { display: flex; flex-direction: column; gap: 0.2rem; min-width: 0; }
.rc__title { font-size: var(--text-small); font-weight: 700; line-height: 1.3; color: var(--color-text-primary); text-decoration: none; overflow-wrap: anywhere; }
.rc__rating { display: flex; align-items: center; gap: 0.4rem; }
.rc__num { font-size: var(--text-caption); font-weight: 700; color: var(--color-rating); }
.rc__badge { flex: none; margin-top: -0.25rem; }
.rc__meta { display: flex; flex-wrap: wrap; align-items: center; gap: 0.125rem 0.625rem; margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.rc__vis { display: inline-flex; align-items: center; gap: 0.25rem; }
.rc__body { margin: 0; font-size: var(--text-small); line-height: 1.6; color: var(--color-text-secondary); white-space: pre-line; overflow-wrap: anywhere; }
.rc__body--clamp { display: -webkit-box; -webkit-line-clamp: 3; line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }
.rc__more { align-self: flex-start; padding: 0; font: inherit; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-primary); text-decoration: underline; text-underline-offset: 3px; background: none; border: 0; cursor: pointer; }
.rc__foot { display: flex; flex-wrap: wrap; align-items: center; gap: 0.25rem 0.875rem; margin-top: 0.125rem; }
.rc__stat { display: inline-flex; align-items: center; gap: 0.3rem; font-size: var(--text-caption); color: var(--color-text-muted); }
.rc__stat.is-on { color: var(--color-brand); }
.rc__actions { display: inline-flex; margin-left: auto; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
@media (min-width: 640px) { .rc__poster { width: 4.5rem; } }
@media (hover: hover) {
  .rc:hover { box-shadow: 0 0 0 1px var(--color-border-strong); }
  .rc__title:hover { color: var(--color-brand); }
}
@media (prefers-reduced-motion: reduce) { .rc { transition: none; } }
</style>
