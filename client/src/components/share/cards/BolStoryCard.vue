<template>
  <article class="sc-card sc-card--story bol" :class="`sc-theme-${theme}`" data-template="bol-story">
    <header class="bol__head">
      <span class="sc-chip">{{ typeLabel }}</span>
      <h1 class="bol__title">{{ t("library.bmol.share.cardTitle") }}</h1>
      <p v-if="username" class="sc-meta">@{{ username }}</p>
    </header>

    <section v-if="hero" class="bol__hero">
      <img v-if="posterOf(hero)" :src="posterOf(hero)!" :alt="hero.media.title" class="sc-poster bol__hero-poster" />
      <div v-else class="sc-poster--empty bol__hero-poster" />
      <div class="bol__hero-info">
        <span class="bol__crown">#1</span>
        <h2 class="bol__hero-name">{{ hero.media.title }}</h2>
        <blockquote v-if="note" class="sc-quote bol__note"><p class="sc-quote__text">{{ note }}</p></blockquote>
      </div>
    </section>

    <ol class="bol__grid">
      <li v-for="item in rest" :key="item.id" class="bol__cell">
        <img v-if="posterOf(item)" :src="posterOf(item)!" :alt="item.media.title" class="sc-poster bol__cell-poster" />
        <div v-else class="sc-poster--empty bol__cell-poster" />
        <span class="bol__rank">{{ item.rank }}</span>
      </li>
    </ol>

    <ShareCardFooter :url="shortUrl" class="bol__footer" />
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { BMOLItemResponse, MediaType } from "@/types"
import type { ShareTheme } from "@/types/share"
import { bolPosterKey, bolTemplateDef } from "../shareCards"
import ShareCardFooter from "./ShareCardFooter.vue"
import "../shareCard.css"

const props = defineProps<{
  /** Titles in rank order, already limited to the card capacity. */
  items: BMOLItemResponse[]
  mediaType: MediaType
  username: string
  theme: ShareTheme
  posters: Record<string, string | null>
  note: string | null
  shortUrl: string
}>()
const { t } = useI18n()

const hero = computed(() => props.items[0])
const rest = computed(() => props.items.slice(1, bolTemplateDef("story").capacity))
const typeLabel = computed(() => t(props.mediaType === "tv" ? "share.tvSeries" : "share.movie"))
const posterOf = (item: BMOLItemResponse) => props.posters[bolPosterKey(item)] ?? null
</script>

<style scoped>
.bol { gap: 44px; padding: 72px 72px 64px; }
.bol__head { display: flex; flex-direction: column; gap: 16px; }
.bol__title { margin: 0; font-size: 104px; font-weight: 800; letter-spacing: -0.02em; line-height: 1.05; }
.bol__hero { display: flex; gap: 48px; align-items: flex-start; }
.bol__hero-poster { flex: none; width: 380px; height: 570px; }
.bol__hero-info { display: flex; flex: 1; flex-direction: column; gap: 24px; min-width: 0; }
.bol__crown { align-self: flex-start; padding: 8px 28px; font-size: 56px; font-weight: 800; color: var(--sc-bg); background: var(--color-accent); border-radius: 999px; }
.bol__hero-name { margin: 0; font-size: 64px; font-weight: 800; line-height: 1.1; overflow-wrap: anywhere; display: -webkit-box; -webkit-box-orient: vertical; -webkit-line-clamp: 3; line-clamp: 3; overflow: hidden; }
.bol__note { padding: 24px 32px; }
.bol__note .sc-quote__text { font-size: 36px; }
.bol__grid { display: grid; flex: 1; grid-template-columns: repeat(4, 1fr); gap: 24px; align-content: start; margin: 0; padding: 0; list-style: none; }
.bol__cell { position: relative; }
.bol__cell-poster { width: 100%; aspect-ratio: 2 / 3; border-radius: 20px; }
.bol__rank { position: absolute; top: 12px; left: 12px; display: grid; place-items: center; min-width: 56px; height: 56px; padding: 0 12px; font-size: 32px; font-weight: 800; color: var(--sc-bg); background: var(--sc-fg); border-radius: 999px; }
.bol__footer { margin-top: auto; }
</style>
