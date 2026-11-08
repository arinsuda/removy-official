<template>
  <article class="sc-card sc-card--story take" :class="`sc-theme-${theme}`" data-template="take">
    <span class="sc-chip">{{ t("share.myTake") }}</span>

    <div class="take__media">
      <img v-if="posterUrl" :src="posterUrl" :alt="media.title" class="sc-poster take__poster" />
      <span v-else class="sc-poster--empty take__poster" />
      <div class="take__info">
        <h1 class="sc-title" :class="titleSizeClass(media.title)">{{ media.title }}</h1>
        <p v-if="meta" class="sc-meta">{{ meta }}</p>
      </div>
    </div>

    <div v-if="stats.rating" class="take__rating" data-test="share-my-rating">
      <ShareStars :rating="stats.rating" :size="88" />
      <div class="sc-score">
        <span class="sc-score__value">{{ stats.rating.toFixed(1) }}</span>
        <span class="sc-score__scale">/5</span>
      </div>
    </div>

    <div class="sc-stat-grid">
      <div class="sc-stat">
        <p class="sc-label">{{ t("share.watched") }}</p>
        <p class="sc-stat__value" data-test="share-watch-count">{{ stats.watchCount > 0 ? `×${stats.watchCount}` : "—" }}</p>
      </div>
      <div class="sc-stat">
        <p class="sc-label">{{ t("share.lastWatched") }}</p>
        <p class="sc-stat__value sc-stat__value--sm" data-test="share-last-watched">{{ lastDate || "—" }}</p>
      </div>
    </div>

    <div v-if="score" class="take__community">
      <ShareTierBadge :tier="tier" />
      <span class="take__community-score">{{ score.value }}<small>{{ score.scale }}</small></span>
      <span class="sc-label">{{ score.source }}</span>
    </div>

    <blockquote v-if="excerpt" class="sc-quote" data-test="share-excerpt"><p class="sc-quote__text">{{ excerpt }}</p></blockquote>

    <ShareCardFooter class="take__footer" :url="shortUrl" />
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { ShareMediaContext, ShareStatsContext, ShareTheme } from "@/types/share"
import { publicExcerpt, shareGenres, shareScore, shareTier, titleSizeClass, truncateExcerpt } from "../shareCards"
import ShareCardFooter from "./ShareCardFooter.vue"
import ShareStars from "./ShareStars.vue"
import ShareTierBadge from "./ShareTierBadge.vue"
import "../shareCard.css"

const props = defineProps<{
  media: ShareMediaContext
  stats: ShareStatsContext
  posterUrl: string | null
  theme: ShareTheme
  shortUrl: string
}>()
// Sized so the quote fits 3 lines at export size (42px, ~34 chars per line).
const EXCERPT_MAX_CHARS = 90
const { t, locale } = useI18n()

const score = computed(() => shareScore(props.media))
const tier = computed(() => shareTier(props.media))
const meta = computed(() => [props.media.releaseYear, ...shareGenres(props.media, 2)].filter(Boolean).join(" · "))
// Review text only ever appears for public reviews.
const excerpt = computed(() => {
  const text = publicExcerpt(props.stats.reviewVisibility, props.stats.reviewBody)
  if (!text) return null
  return truncateExcerpt(text, EXCERPT_MAX_CHARS)
})
const lastDate = computed(() => {
  const raw = props.stats.lastWatchedOn
  if (!raw) return null
  const d = new Date(raw.length === 10 ? `${raw}T00:00:00` : raw)
  return Number.isNaN(d.getTime()) ? null : d.toLocaleDateString(locale.value, { day: "numeric", month: "short", year: "numeric" })
})
</script>

<style scoped>
.take { gap: 40px; padding: 80px; }
.take > * { flex-shrink: 0; }
.take__media { display: flex; align-items: flex-end; gap: 48px; }
.take__poster { flex: none; width: 280px; height: 420px; }
.take__info { display: flex; flex-direction: column; gap: 20px; min-width: 0; }
.take__info .sc-title { font-size: 84px; }
.take__info .sc-title.sc-title--xs { font-size: 60px; }
.take__rating { display: flex; flex-direction: column; gap: 20px; }
.take__rating .sc-score__value { font-size: 180px; }
.take__community { display: flex; align-items: center; gap: 28px; }
.take__community-score { font-size: 64px; font-weight: 800; }
.take__community-score small { color: var(--sc-muted); font-size: 40px; font-weight: 700; margin-left: 4px; }
.take__footer { margin-top: auto; }
</style>
