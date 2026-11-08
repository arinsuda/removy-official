<template>
  <article class="sc-card sc-card--square square" :class="`sc-theme-${theme}`" data-template="square">
    <img v-if="posterUrl" :src="posterUrl" :alt="media.title" class="square__art" />
    <div v-else class="square__art sc-bleed--empty" />
    <div class="square__body">
      <span class="sc-chip">{{ typeLabel }}</span>
      <h1 class="sc-title" :class="titleSizeClass(media.title)">{{ media.title }}</h1>
      <p v-if="meta" class="sc-meta">{{ meta }}</p>
      <div class="square__score">
        <ShareTierBadge :tier="tier" />
        <div v-if="score" class="sc-score" data-test="share-score">
          <span class="sc-score__value square__value">{{ score.value }}</span>
          <span class="sc-score__scale">{{ score.scale }}</span>
        </div>
        <span v-if="score" class="sc-label">{{ score.source }}</span>
      </div>
      <ShareCardFooter :url="shortUrl" />
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { ShareMediaContext, ShareTheme } from "@/types/share"
import { shareGenres, shareScore, shareTier, titleSizeClass } from "../shareCards"
import ShareCardFooter from "./ShareCardFooter.vue"
import ShareTierBadge from "./ShareTierBadge.vue"
import "../shareCard.css"

const props = defineProps<{ media: ShareMediaContext; posterUrl: string | null; theme: ShareTheme; shortUrl: string }>()
const { t } = useI18n()

const typeLabel = computed(() => t(props.media.mediaType === "movie" ? "share.movie" : "share.tvSeries"))
const score = computed(() => shareScore(props.media))
const tier = computed(() => shareTier(props.media))
const meta = computed(() => [props.media.releaseYear, ...shareGenres(props.media, 2)].filter(Boolean).join(" · "))
</script>

<style scoped>
.square { flex-direction: row; }
.square__art { flex: none; width: 450px; height: 100%; object-fit: cover; }
.square__body { display: flex; flex: 1; flex-direction: column; gap: 28px; min-width: 0; padding: 72px 64px 64px; }
.square__body .sc-title { font-size: 84px; }
.square__body .sc-title.sc-title--xs { font-size: 60px; }
.square__body .sc-meta { font-size: 38px; }
.square__score { display: flex; flex-direction: column; gap: 16px; margin-top: auto; padding-bottom: 8px; }
.square__value { font-size: 200px; }
.square__body .sc-footer { flex-direction: column; align-items: flex-start; gap: 8px; padding-top: 28px; }
.square__body .sc-wordmark { font-size: 40px; }
.square__body .sc-footer__url { font-size: 28px; }
</style>
