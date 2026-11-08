<template>
  <PosterStoryCard v-if="template === 'poster'" ref="card" v-bind="common" />
  <TakeCard v-else-if="template === 'take' && stats" ref="card" v-bind="common" :stats="stats" />
  <StickerCard v-else-if="template === 'sticker'" ref="card" :media="media" :stats="stats" :poster-url="posterUrl" :theme="theme" />
  <SquareCard v-else-if="template === 'square'" ref="card" v-bind="common" />
  <ReviewShareTemplate
    v-else-if="template === 'review' && review"
    ref="card"
    :media="media"
    :review="review"
    :poster-blob-url="posterUrl"
    :avatar-blob-url="avatarUrl ?? null"
  />
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from "vue"
import type { ShareMediaContext, ShareReviewContext, ShareStatsContext, ShareTemplateKey, ShareTheme } from "@/types/share"
import PosterStoryCard from "./cards/PosterStoryCard.vue"
import SquareCard from "./cards/SquareCard.vue"
import StickerCard from "./cards/StickerCard.vue"
import TakeCard from "./cards/TakeCard.vue"
import ReviewShareTemplate from "./ReviewShareTemplate.vue"

const props = defineProps<{
  template: ShareTemplateKey
  media: ShareMediaContext
  posterUrl: string | null
  avatarUrl?: string | null
  theme: ShareTheme
  shortUrl: string
  stats?: ShareStatsContext | null
  review?: ShareReviewContext | null
}>()

const common = computed(() => ({
  media: props.media,
  posterUrl: props.posterUrl,
  theme: props.theme,
  shortUrl: props.shortUrl,
}))

const card = useTemplateRef<{ $el: HTMLElement }>("card")
/** The card's root element, the node that gets rasterised. */
defineExpose({ getNode: () => card.value?.$el ?? null })
</script>
