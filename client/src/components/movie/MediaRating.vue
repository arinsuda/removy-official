<template>
  <div
    class="media-rating @container inline-flex"
    :class="variant === 'detail' ? 'w-full' : 'items-center gap-1.5'"
    :data-variant="variant"
  >
    <template v-if="variant === 'detail'">
      <div class="flex w-full flex-col items-stretch gap-3 min-[360px]:flex-row min-[360px]:items-center">
        <div class="flex flex-1 flex-col items-center gap-1 text-center">
          <PopcornScoreBadge :average="popcorn.average" :count="popcorn.count" size="detail" />
          <span data-test="review-count" class="text-caption text-text-muted">
            {{ t("media.rating.outOfReviews", { count: popcorn.count }, popcorn.count) }}
          </span>
        </div>
        <ul
          class="m-0 flex list-none flex-col justify-center gap-2 border-t border-border p-0 pt-3 min-[360px]:border-l min-[360px]:border-t-0 min-[360px]:pl-4 min-[360px]:pt-0"
        >
          <li
            v-for="s in secondary"
            :key="s.key"
            :data-source="s.key"
            :aria-label="s.aria"
            class="flex items-center gap-2 whitespace-nowrap text-caption"
          >
            <RatingSourceMark :source="s.key" class="w-[44px] justify-start" />
            <SkeletonBlock
              v-if="!s.value && externalLoading && s.key !== 'tmdb'"
              :data-test="`${s.key}-loading`"
              width="36px"
              height="12px"
              radius="4px"
            />
            <span v-else aria-hidden="true" class="tabular-nums" :class="s.value ? 'font-semibold text-text-primary' : 'text-text-muted'">{{ s.value ?? "—" }}</span>
          </li>
        </ul>
      </div>
    </template>
    <template v-else>
      <PopcornScoreBadge :average="popcorn.average" :count="popcorn.count" size="card" />
      <ul
        v-if="secondary[0]?.value"
        class="m-0 flex list-none flex-wrap items-center gap-x-2 p-0 text-[10px] leading-none text-text-muted"
      >
        <li :data-source="secondary[0].key" :aria-label="secondary[0].aria" class="inline-flex items-baseline gap-1 whitespace-nowrap">
          <span aria-hidden="true" class="font-semibold uppercase tracking-wide">{{ t("media.rating.sources.tmdb") }}</span>
          <span aria-hidden="true">{{ secondary[0].short }}</span>
        </li>
      </ul>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { MediaRatings, RatingSource } from "@/types/movie"
import PopcornScoreBadge from "./PopcornScoreBadge.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import RatingSourceMark from "./RatingSourceMark.vue"

const props = withDefaults(
  defineProps<{
    ratings?: MediaRatings | null
    /** Legacy TMDB 0-10 score, used when the API did not send `ratings.tmdb`. */
    voteAverage?: number | null
    variant?: "card" | "detail"
    /** IMDb / Rotten Tomatoes are still loading: show a skeleton in empty slots. */
    externalLoading?: boolean
  }>(),
  { ratings: null, voteAverage: null, variant: "card", externalLoading: false },
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

interface Row {
  key: "tmdb" | "imdb" | "rt"
  /** Display text, or null when the source has no score (rendered as an em dash). */
  value: string | null
  short: string
  aria: string
}

const secondary = computed<Row[]>(() => {
  const tmdb = tmdbScore.value?.toFixed(1)
  const imdbSrc = props.ratings?.imdb
  const imdb = imdbSrc?.available && imdbSrc.average ? imdbSrc.average.toFixed(1) : undefined
  const rtSrc = props.ratings?.rotten_tomatoes
  const rt = rtSrc?.available && rtSrc.average !== null ? String(Math.round(rtSrc.average)) : undefined
  return [
    { key: "tmdb", value: tmdb ? `${tmdb}/10` : null, short: tmdb ?? "", aria: tmdb ? t("media.rating.tmdbScore", { score: tmdb }) : t("media.rating.unavailable", { source: "TMDB" }) },
    { key: "imdb", value: imdb ? `${imdb}/10` : null, short: imdb ?? "", aria: imdb ? t("media.rating.imdbScore", { score: imdb }) : t("media.rating.unavailable", { source: "IMDb" }) },
    { key: "rt", value: rt ? `${rt}%` : null, short: rt ?? "", aria: rt ? t("media.rating.rtScore", { score: rt }) : t("media.rating.unavailable", { source: t("media.rating.sources.rottenTomatoes") }) },
  ]
})
</script>
