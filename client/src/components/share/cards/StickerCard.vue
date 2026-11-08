<template>
  <article class="sc-card sc-card--square sc-card--transparent" :class="`sc-theme-${theme}`" data-template="sticker">
    <div class="sticker">
      <div class="sticker__media">
        <img v-if="posterUrl" :src="posterUrl" :alt="media.title" class="sc-poster sticker__poster" />
        <span v-else class="sc-poster--empty sticker__poster" />
        <div class="sticker__info">
          <span class="sc-chip sticker__chip">{{ typeLabel }}</span>
          <h1 class="sc-title sticker__title" :class="titleSizeClass(media.title)">{{ media.title }}</h1>
          <p v-if="media.releaseYear" class="sc-meta">{{ media.releaseYear }}</p>
        </div>
      </div>
      <div v-if="display" class="sticker__score" data-test="share-score">
        <p class="sc-label">{{ display.label }}</p>
        <ShareStars v-if="display.stars" :rating="display.stars" :size="84" />
        <div class="sc-score">
          <span class="sc-score__value sticker__value">{{ display.value }}</span>
          <span class="sc-score__scale">{{ display.scale }}</span>
        </div>
        <ShareTierBadge :tier="tier" />
      </div>
      <span class="sc-wordmark sticker__brand"><span class="sc-wordmark__pop">POP</span><span>CORN</span></span>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { ShareMediaContext, ShareStatsContext, ShareTheme } from "@/types/share"
import { shareScore, shareTier, titleSizeClass } from "../shareCards"
import ShareStars from "./ShareStars.vue"
import ShareTierBadge from "./ShareTierBadge.vue"
import "../shareCard.css"

const props = defineProps<{
  media: ShareMediaContext
  stats?: ShareStatsContext | null
  posterUrl: string | null
  theme: ShareTheme
}>()
const { t } = useI18n()

const typeLabel = computed(() => t(props.media.mediaType === "movie" ? "share.movie" : "share.tvSeries"))
const tier = computed(() => shareTier(props.media))
// The owner's own rating wins; otherwise the community score.
const display = computed(() => {
  const mine = props.stats?.rating
  if (mine) return { label: t("share.myRating"), stars: mine, value: mine.toFixed(1), scale: "/5" }
  const s = shareScore(props.media)
  return s ? { label: s.source, stars: 0, value: s.value, scale: s.scale } : null
})
</script>

<style scoped>
.sticker {
  display: flex; flex-direction: column; align-items: stretch; gap: 36px; width: 840px; padding: 56px;
  border-radius: 80px; background: var(--sc-bg); border: 6px solid var(--sc-line);
  box-shadow: 0 32px 80px rgba(0, 0, 0, 0.35);
}
.sticker__media { display: flex; align-items: flex-end; gap: 40px; }
.sticker__poster { flex: none; width: 220px; height: 330px; border-radius: 28px; }
.sticker__info { display: flex; flex-direction: column; gap: 20px; min-width: 0; }
.sticker__chip { padding: 8px 24px; font-size: 28px; }
.sticker__title { font-size: 64px; -webkit-line-clamp: 4; }
.sticker__score { display: flex; flex-direction: column; gap: 16px; padding-top: 32px; border-top: 3px solid var(--sc-line); }
.sticker__value { font-size: 140px; }
.sticker__brand { align-self: center; font-size: 40px; }
</style>
