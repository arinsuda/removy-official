<template>
  <div class="ds">
    <section v-if="isEmpty" class="ds__onboarding">
      <Clapperboard :size="36" class="ds__onboarding-icon" aria-hidden="true" />
      <h3 class="ds__onboarding-title">{{ t("library.dashboard.onboarding.title") }}</h3>
      <p class="ds__onboarding-body">{{ t("library.dashboard.onboarding.body") }}</p>
      <AppButton v-if="isOwner" :to="{ name: 'movies' }">{{ t("library.dashboard.onboarding.cta") }}</AppButton>
    </section>

    <template v-else>
      <ul class="ds__stats" :aria-label="t('library.dashboard.stats.watched')">
        <li v-for="s in stats" :key="s.key" class="ds__stat">
          <component :is="s.to ? RouterLink : 'div'" v-bind="s.to ? { to: s.to } : {}" class="ds__stat-body">
            <span class="ds__stat-icon"><component :is="s.icon" :size="18" aria-hidden="true" /></span>
            <span class="ds__stat-label">{{ t(`library.dashboard.stats.${s.key}`) }}</span>
            <span class="ds__stat-value">{{ s.value }}</span>
            <span v-if="s.note" class="ds__stat-note">{{ s.note }}</span>
          </component>
        </li>
      </ul>

      <div v-if="isOwner" class="ds__actions" role="group" :aria-label="t('library.dashboard.actions.title')">
        <AppButton variant="outline" size="sm" :to="{ name: 'movies' }"><Film :size="14" aria-hidden="true" />{{ t("library.dashboard.actions.browseMovies") }}</AppButton>
        <AppButton variant="outline" size="sm" :to="{ name: 'tv' }"><Tv :size="14" aria-hidden="true" />{{ t("library.dashboard.actions.browseSeries") }}</AppButton>
        <AppButton variant="outline" size="sm" data-action="quick-add" @click="emit('add')"><Plus :size="14" aria-hidden="true" />{{ t("library.dashboard.actions.add") }}</AppButton>
        <AppButton variant="outline" size="sm" :to="tabTo('bmol')"><Trophy :size="14" aria-hidden="true" />{{ t("library.dashboard.actions.bmol") }}</AppButton>
      </div>

      <LibraryRail
        rail="library-up-next"
        :title="t('library.dashboard.upNext.title')"
        :see-all-label="t('library.dashboard.upNext.seeAll')"
        :empty-label="t('library.dashboard.upNext.empty')"
        :see-all-to="listTo('watchlist')"
        :items="upNext"
      />
      <LibraryRail
        rail="library-recent"
        :title="t('library.dashboard.recent.title')"
        :see-all-label="t('library.dashboard.recent.seeAll')"
        :empty-label="t('library.dashboard.recent.empty')"
        :see-all-to="listTo('watched')"
        :items="recent"
      />

      <div class="ds__charts">
        <section class="ds__card" aria-labelledby="ds-genres">
          <h3 id="ds-genres" class="ds__card-title">{{ t("library.dashboard.genres.title") }}</h3>
          <p v-if="!genres.length" class="ds__muted">{{ t("library.dashboard.genres.empty") }}</p>
          <ol v-else class="ds__genres">
            <li v-for="(g, i) in genres" :key="g.name" class="ds__genre">
              <span class="ds__rank">{{ fmt(i + 1) }}</span>
              <span class="ds__genre-name">{{ g.name }}</span>
              <span class="ds__genre-count">{{ t("library.dashboard.genres.count", { n: fmt(g.count) }) }}</span>
            </li>
          </ol>
        </section>

        <section class="ds__card" aria-labelledby="ds-ratings">
          <h3 id="ds-ratings" class="ds__card-title">{{ t("library.dashboard.ratings.title") }}</h3>
          <p v-if="!reviews.length" class="ds__muted">{{ t("library.dashboard.ratings.empty") }}</p>
          <ul v-else class="ds__bars">
            <li v-for="b in buckets" :key="b.rating" class="ds__bar" :aria-label="t('library.dashboard.ratings.bar', { rating: fmt(b.rating), count: fmt(b.count) })">
              <span class="ds__bar-label" aria-hidden="true">{{ fmt(b.rating) }}</span>
              <span class="ds__bar-track" aria-hidden="true"><span class="ds__bar-fill" :style="{ width: `${b.percent}%` }" /></span>
              <span class="ds__bar-count" aria-hidden="true">{{ fmt(b.count) }}</span>
            </li>
          </ul>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink, type RouteLocationRaw } from "vue-router"
import { Bookmark, Clapperboard, Eye, Film, Heart, Plus, Star, Trophy, Tv, Award } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import type { LibraryListKey, LibraryTab, ReviewResponse } from "@/types"
import { RAIL_SIZE, RATING_MAX, TOP_GENRES } from "../config"
import { averageRating, newestFirst, ratingBuckets, reviewMonthDiff, topGenres, type LibraryEntry } from "../libraryStats"
import LibraryRail from "./LibraryRail.vue"

const props = defineProps<{
  isOwner: boolean
  watchlist: LibraryEntry[]
  watched: LibraryEntry[]
  likes: LibraryEntry[]
  reviews: ReviewResponse[]
  listTo: (key: LibraryListKey) => RouteLocationRaw
  tabTo: (tab: LibraryTab) => RouteLocationRaw
}>()
const emit = defineEmits<{ add: [] }>()

const { t, n } = useI18n()
const fmt = (value: number) => n(value, { maximumFractionDigits: 1 })

const isEmpty = computed(() => !props.watchlist.length && !props.watched.length && !props.likes.length && !props.reviews.length)

const upNext = computed(() => newestFirst(props.watchlist).slice(0, RAIL_SIZE).map((e) => e.media))
const recent = computed(() => newestFirst(props.watched).slice(0, RAIL_SIZE).map((e) => e.media))
const genres = computed(() => topGenres(props.watched, TOP_GENRES))
const buckets = computed(() => ratingBuckets(props.reviews))

const trend = computed(() => {
  const diff = reviewMonthDiff(props.reviews)
  if (!props.reviews.length) return undefined
  if (diff > 0) return t("library.dashboard.trend.more", { count: fmt(diff) })
  if (diff < 0) return t("library.dashboard.trend.less", { count: fmt(Math.abs(diff)) })
  return t("library.dashboard.trend.same")
})

const stats = computed(() => {
  const avg = averageRating(props.reviews)
  return [
    { key: "watched", icon: Eye, value: fmt(props.watched.length), to: props.listTo("watched") },
    { key: "watchlist", icon: Bookmark, value: fmt(props.watchlist.length), to: props.listTo("watchlist") },
    { key: "likes", icon: Heart, value: fmt(props.likes.length), to: props.listTo("likes") },
    { key: "reviews", icon: Star, value: fmt(props.reviews.length), note: trend.value, to: props.listTo("reviews") },
    { key: "average", icon: Award, value: avg ? t("library.dashboard.averageOf", { value: fmt(avg), max: fmt(RATING_MAX) }) : t("library.dashboard.noData") },
    { key: "topGenre", icon: Trophy, value: genres.value[0]?.name ?? t("library.dashboard.noData") },
  ] as { key: string; icon: typeof Eye; value: string; note?: string; to?: RouteLocationRaw }[]
})
</script>

<style scoped>
.ds { display: flex; flex-direction: column; gap: 2rem; }
.ds__stats { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.75rem; margin: 0; padding: 0; list-style: none; }
.ds__stat-body { display: flex; flex-direction: column; gap: 0.25rem; height: 100%; padding: 1rem; color: inherit; text-decoration: none; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); transition: border-color var(--motion-fast) var(--ease-water); }
.ds__stat-icon { display: grid; place-items: center; width: 2rem; height: 2rem; margin-bottom: 0.25rem; color: var(--color-text-secondary); background: var(--color-surface-3); border-radius: var(--radius-btn); }
.ds__stat-label { font-size: var(--text-caption); color: var(--color-text-muted); }
.ds__stat-value { overflow: hidden; font-size: var(--text-heading); font-weight: 700; line-height: 1.2; color: var(--color-text-primary); text-overflow: ellipsis; }
.ds__stat-note { font-size: var(--text-caption); color: var(--color-text-secondary); }
.ds__actions { display: flex; flex-wrap: wrap; gap: 0.5rem; }
.ds__charts { display: grid; grid-template-columns: minmax(0, 1fr); gap: 1rem; }
.ds__card { padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.ds__card-title { margin: 0 0 1rem; font-size: var(--text-body); font-weight: 700; color: var(--color-text-primary); }
.ds__muted { margin: 0; font-size: var(--text-small); color: var(--color-text-muted); }
.ds__genres, .ds__bars { display: flex; flex-direction: column; gap: 0.625rem; margin: 0; padding: 0; list-style: none; }
.ds__genre { display: flex; align-items: center; gap: 0.75rem; font-size: var(--text-small); }
.ds__rank { display: grid; flex: none; place-items: center; width: 1.5rem; height: 1.5rem; font-size: var(--text-caption); font-weight: 700; color: var(--color-text-secondary); background: var(--color-surface-3); border-radius: var(--radius-btn); }
.ds__genre-name { flex: 1; min-width: 0; overflow: hidden; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.ds__genre-count { flex: none; font-size: var(--text-caption); color: var(--color-text-muted); }
.ds__bar { display: flex; align-items: center; gap: 0.75rem; font-size: var(--text-caption); color: var(--color-text-secondary); }
.ds__bar-label { flex: none; width: 2rem; text-align: end; }
.ds__bar-track { flex: 1; height: 0.5rem; overflow: hidden; background: var(--color-surface-3); border-radius: var(--radius-btn); }
.ds__bar-fill { display: block; height: 100%; background: var(--color-rating); border-radius: var(--radius-btn); transition: width var(--motion-base) var(--ease-out-soft); }
.ds__bar-count { flex: none; width: 2rem; color: var(--color-text-muted); }
.ds__onboarding { display: flex; flex-direction: column; align-items: center; gap: 0.75rem; padding: 3rem 1.5rem; text-align: center; background: var(--color-surface-2); border: 1px dashed var(--color-border-strong); border-radius: var(--radius-card); }
.ds__onboarding-icon { color: var(--color-text-muted); }
.ds__onboarding-title { margin: 0; font-size: var(--text-title); font-weight: 700; color: var(--color-text-primary); }
.ds__onboarding-body { max-width: 28rem; margin: 0; font-size: var(--text-small); color: var(--color-text-secondary); }
@media (hover: hover) {
  a.ds__stat-body:hover { border-color: var(--color-border-strong); }
}
@media (min-width: 640px) {
  .ds__stats { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}
@media (min-width: 1024px) {
  .ds__stats { grid-template-columns: repeat(6, minmax(0, 1fr)); }
  .ds__charts { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (prefers-reduced-motion: reduce) {
  .ds__stat-body, .ds__bar-fill { transition: none; }
}
</style>
