<template>
  <div
    class="media-rating inline-flex"
    :class="variant === 'detail' ? 'flex-col items-start gap-2' : 'items-center gap-1.5'"
    :data-variant="variant"
  >
    <div class="flex items-center" :class="variant === 'detail' ? 'gap-3' : ''">
      <PopcornScoreBadge :average="popcorn.average" :count="popcorn.count" :size="variant" />
      <span v-if="variant === 'detail'" data-test="review-count" class="text-caption text-text-muted">
        {{ t("media.rating.outOfReviews", { count: popcorn.count }, popcorn.count) }}
      </span>
    </div>

    <ul
      v-if="secondary.length"
      class="m-0 flex list-none flex-wrap items-center p-0 text-text-muted"
      :class="variant === 'detail' ? 'gap-x-3 gap-y-1 text-caption' : 'gap-x-2 text-[10px] leading-none'"
    >
      <li
        v-for="s in secondary"
        :key="s.key"
        :data-source="s.key"
        :aria-label="s.aria"
        class="inline-flex items-baseline gap-1 whitespace-nowrap"
      >
        <span aria-hidden="true" class="font-semibold uppercase tracking-wide">{{ s.name }}</span>
        <span aria-hidden="true">{{ s.value }}</span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { MediaRatings, RatingSource } from "@/types/movie"
import PopcornScoreBadge from "./PopcornScoreBadge.vue"

const props = withDefaults(
  defineProps<{
    ratings?: MediaRatings | null
    /** Legacy TMDB 0-10 score, used when the API did not send `ratings.tmdb`. */
    voteAverage?: number | null
    variant?: "card" | "detail"
  }>(),
  { ratings: null, voteAverage: null, variant: "card" },
)

const { t } = useI18n()

const popcorn = computed<RatingSource>(
  () => props.ratings?.popcorn ?? { average: 0, count: 0, available: false, scale: 5 },
)

const tmdbScore = computed(() => {
  const src = props.ratings?.tmdb
  if (src) return src.available && src.average ? src.average : null
  return props.voteAverage && props.voteAverage > 0 ? props.voteAverage : null
})

interface Secondary {
  key: "tmdb" | "imdb" | "rt"
  name: string
  value: string
  aria: string
}

const secondary = computed<Secondary[]>(() => {
  const out: Secondary[] = []
  if (tmdbScore.value !== null) {
    const v = tmdbScore.value.toFixed(1)
    out.push({
      key: "tmdb",
      name: "TMDB",
      value: props.variant === "detail" ? `${v}/10` : v,
      aria: t("media.rating.tmdbScore", { score: v }),
    })
  }
  if (props.variant === "detail") {
    const imdb = props.ratings?.imdb
    if (imdb?.available && imdb.average) {
      const v = imdb.average.toFixed(1)
      out.push({ key: "imdb", name: "IMDb", value: `${v}/10`, aria: t("media.rating.imdbScore", { score: v }) })
    }
    const rt = props.ratings?.rotten_tomatoes
    if (rt?.available && rt.average !== null) {
      const v = String(Math.round(rt.average))
      out.push({ key: "rt", name: "RT", value: `${v}%`, aria: t("media.rating.rtScore", { score: v }) })
    }
  }
  return out
})
</script>
