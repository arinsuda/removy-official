<template>
  <article class="sc-card sc-card--square bol" :class="`sc-theme-${theme}`" data-template="bol-square">
    <header class="bol__head">
      <div class="bol__titles">
        <span class="sc-chip">{{ typeLabel }}</span>
        <h1 class="bol__title">{{ t("library.bmol.share.cardTitle") }}</h1>
      </div>
      <p v-if="username" class="sc-meta bol__user">@{{ username }}</p>
    </header>

    <ol class="bol__row">
      <li v-for="item in items" :key="item.id" class="bol__cell">
        <div class="bol__art">
          <img v-if="posterOf(item)" :src="posterOf(item)!" :alt="item.media.title" class="sc-poster bol__poster" />
          <div v-else class="sc-poster--empty bol__poster" />
          <span class="bol__rank">{{ item.rank }}</span>
        </div>
        <p class="bol__name">{{ item.media.title }}</p>
      </li>
    </ol>

    <ShareCardFooter :url="shortUrl" />
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { BMOLItemResponse, MediaType } from "@/types"
import type { ShareTheme } from "@/types/share"
import { bolPosterKey } from "../shareCards"
import ShareCardFooter from "./ShareCardFooter.vue"
import "../shareCard.css"

const props = defineProps<{
  /** The top three titles in rank order. */
  items: BMOLItemResponse[]
  mediaType: MediaType
  username: string
  theme: ShareTheme
  posters: Record<string, string | null>
  shortUrl: string
}>()
const { t } = useI18n()

const typeLabel = computed(() => t(props.mediaType === "tv" ? "share.tvSeries" : "share.movie"))
const posterOf = (item: BMOLItemResponse) => props.posters[bolPosterKey(item)] ?? null
</script>

<style scoped>
.bol { gap: 36px; padding: 64px 72px 56px; }
.bol__head { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; }
.bol__titles { display: flex; flex-direction: column; gap: 16px; }
.bol__title { margin: 0; font-size: 84px; font-weight: 800; letter-spacing: -0.02em; line-height: 1.05; }
.bol__user { font-size: 36px; }
.bol__row { display: grid; flex: 1; grid-template-columns: repeat(3, 1fr); gap: 36px; margin: 0; padding: 0; list-style: none; min-height: 0; }
.bol__cell { display: flex; flex-direction: column; gap: 16px; min-width: 0; }
.bol__art { position: relative; }
.bol__poster { width: 100%; aspect-ratio: 2 / 3; }
.bol__rank { position: absolute; top: 16px; left: 16px; display: grid; place-items: center; min-width: 72px; height: 72px; padding: 0 16px; font-size: 44px; font-weight: 800; color: var(--sc-bg); background: var(--color-accent); border-radius: 999px; }
.bol__name { margin: 0; font-size: 32px; font-weight: 700; display: -webkit-box; -webkit-box-orient: vertical; -webkit-line-clamp: 2; line-clamp: 2; overflow: hidden; overflow-wrap: anywhere; }
</style>
