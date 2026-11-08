<template>
  <article class="sc-card sc-card--story" :class="`sc-theme-${theme}`" data-template="poster">
    <img v-if="posterUrl" :src="posterUrl" :alt="media.title" class="sc-bleed" />
    <div v-else class="sc-bleed sc-bleed--empty" />
    <div class="sc-scrim" />
    <div class="story__top">
      <span class="sc-chip">{{ typeLabel }}</span>
    </div>
    <div class="story__bottom">
      <ShareTierBadge :tier="tier" />
      <div v-if="score" class="sc-score" data-test="share-score">
        <span class="sc-score__value">{{ score.value }}</span>
        <span class="sc-score__scale">{{ score.scale }}</span>
        <span class="sc-score__source">{{ score.source }}</span>
      </div>
      <h1 class="sc-title" :class="titleSizeClass(media.title)">{{ media.title }}</h1>
      <p v-if="meta" class="sc-meta">{{ meta }}</p>
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
.story__top { position: relative; padding: 96px 96px 0; }
.story__bottom { position: relative; display: flex; flex-direction: column; gap: 40px; margin-top: auto; padding: 0 96px 96px; }
</style>
