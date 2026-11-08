<template>
  <section class="meter" :data-tier="tier" :aria-labelledby="titleId">
    <div class="meter__head">
      <div>
        <h2 :id="titleId" class="meter__eyebrow">POPCORN</h2>
        <p class="meter__score">
          <span class="meter__num" data-test="meter-score">{{ tier === "none" ? "—" : score.toFixed(1) }}</span>
          <span class="meter__max" aria-hidden="true">/ 5</span>
        </p>
        <p class="meter__tier" data-test="meter-tier">{{ t(`media.rating.tier.${tier}`) }}</p>
      </div>
      <button
        type="button"
        class="meter__like"
        :class="{ 'meter__like--on': liked }"
        :aria-pressed="liked"
        :aria-label="t(liked ? 'media.detail.actions.unlike' : 'media.detail.actions.like')"
        :disabled="likePending"
        @click="emit('toggle-like')"
      >
        <Heart :size="20" :fill="liked ? 'currentColor' : 'none'" aria-hidden="true" />
        <span v-if="likeCount > 0" class="meter__like-count">{{ likeCount.toLocaleString() }}</span>
      </button>
    </div>

    <div
      class="meter__kernels"
      role="img"
      :aria-label="tier === 'none' ? t('media.rating.tier.none') : t('media.rating.popcornScore', { score: score.toFixed(1) })"
    >
      <span v-for="(fill, i) in fills" :key="i" class="meter__kernel" :style="{ '--i': i }">
        <svg viewBox="0 0 24 24" class="meter__shape meter__shape--base" aria-hidden="true"><path :d="KERNEL" /></svg>
        <svg
          v-if="fill > 0"
          viewBox="0 0 24 24"
          class="meter__shape meter__shape--fill"
          :style="{ clipPath: `inset(0 ${(1 - fill) * 100}% 0 0)` }"
          aria-hidden="true"
        >
          <path :d="KERNEL" />
        </svg>
      </span>
    </div>

    <p class="meter__count" data-test="meter-count">
      {{ count > 0 ? t("media.rating.reviewCount", { count }, count) : t("media.detail.meter.noReviews") }}
    </p>

    <ul v-if="secondary.length" class="meter__sources">
      <li v-for="s in secondary" :key="s.key" :data-source="s.key" :aria-label="s.aria">
        <RatingSourceMark :source="s.key" />
        <span aria-hidden="true">{{ s.text }}</span>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import { Heart } from "lucide-vue-next"
import RatingSourceMark from "@/components/movie/RatingSourceMark.vue"
import { popcornTier, roundScore } from "@/components/movie/popcornTier"
import type { MediaRatings, RatingSource } from "@/types/movie"
import { kernelFills } from "./popcornMeter"

const props = withDefaults(
  defineProps<{
    /** POPCORN average (0-5) and review count, from live stats. */
    average: number
    count: number
    ratings?: MediaRatings | null
    voteAverage?: number
    liked?: boolean
    likeCount?: number
    likePending?: boolean
  }>(),
  { ratings: null, voteAverage: 0, liked: false, likeCount: 0, likePending: false },
)
const emit = defineEmits<{ "toggle-like": [] }>()

const { t } = useI18n()
const titleId = useId()
const KERNEL =
  "M12 2.5c2.4 0 4 1.5 4.6 3.3 2.2.2 4 2 4 4.3 0 1.3-.6 2.4-1.5 3.1.3.5.4 1 .4 1.6 0 2.4-2.200 3.800-4.500 3.600-.7 1.600-2.100 2.600-3 2.600s-2.300-1-3-2.600c-2.300.2-4.500-1.200-4.500-3.600 0-.6.1-1.100.4-1.600-.9-.7-1.500-1.800-1.500-3.100 0-2.300 1.800-4.100 4-4.300C8 4 9.600 2.500 12 2.500z"

const score = computed(() => roundScore(props.average))
const tier = computed(() => popcornTier(props.average, props.count))
const fills = computed(() => (tier.value === "none" ? kernelFills(0) : kernelFills(props.average)))

const secondary = computed(() => {
  const r = props.ratings
  const rows: { key: "tmdb" | "imdb" | "rt"; text: string; aria: string }[] = []
  const val = (src: RatingSource | undefined) => (src?.available && src.average ? src.average : null)
  const tmdb = val(r?.tmdb) ?? (props.voteAverage > 0 ? props.voteAverage : null)
  if (tmdb) rows.push({ key: "tmdb", text: `${tmdb.toFixed(1)}/10`, aria: t("media.rating.tmdbScore", { score: tmdb.toFixed(1) }) })
  const imdb = val(r?.imdb)
  if (imdb) rows.push({ key: "imdb", text: `${imdb.toFixed(1)}/10`, aria: t("media.rating.imdbScore", { score: imdb.toFixed(1) }) })
  const rt = val(r?.rotten_tomatoes)
  if (rt) rows.push({ key: "rt", text: `${Math.round(rt)}%`, aria: t("media.rating.rtScore", { score: Math.round(rt) }) })
  return rows
})
</script>

<style scoped>
.meter {
  --tier-color: var(--color-text-muted);
  padding: 1.25rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
}
.meter[data-tier="fresh"] { --tier-color: var(--color-rating); }
.meter[data-tier="hot"] {
  --tier-color: var(--color-accent);
  border-color: color-mix(in srgb, var(--color-brand) 55%, transparent);
  box-shadow: 0 0 32px color-mix(in srgb, var(--color-brand) 18%, transparent);
}
.meter__head { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.meter__eyebrow { font-size: var(--text-caption); font-weight: 800; letter-spacing: 0.14em; color: var(--color-brand); }
.meter__score { display: flex; align-items: baseline; gap: 0.35rem; line-height: 1; margin-top: 0.35rem; }
.meter__num { font-size: var(--text-display); font-weight: 800; color: var(--tier-color); font-variant-numeric: tabular-nums; }
.meter__max { font-size: var(--text-small); color: var(--color-text-muted); }
.meter__tier { margin-top: 0.35rem; font-size: var(--text-small); font-weight: 600; color: var(--tier-color); }
.meter__like {
  display: inline-flex; align-items: center; gap: 0.4rem;
  min-width: 44px; min-height: 44px; padding: 0 0.75rem; justify-content: center;
  color: var(--color-text-secondary);
  background: var(--color-surface-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-btn);
  cursor: pointer;
  transition: color var(--motion-fast) var(--ease-water), border-color var(--motion-fast) var(--ease-water), transform var(--motion-fast) var(--ease-out-soft);
}
.meter__like--on { color: var(--color-brand); border-color: var(--color-brand); }
.meter__like:disabled { opacity: 0.6; cursor: progress; }
.meter__like:active:not(:disabled) { transform: scale(0.94); }
.meter__like-count { font-size: var(--text-caption); font-weight: 700; font-variant-numeric: tabular-nums; }
.meter__kernels { display: grid; grid-template-columns: repeat(10, minmax(0, 1fr)); gap: 0.15rem; margin-top: 1rem; }
.meter__kernel { position: relative; aspect-ratio: 1; animation: kernel-pop 520ms var(--ease-out-soft) both; animation-delay: calc(var(--i) * 40ms); }
.meter__shape { position: absolute; inset: 0; width: 100%; height: 100%; }
.meter__shape--base { fill: var(--color-surface-3); }
.meter__shape--fill { fill: var(--tier-color); }
.meter__count { margin-top: 0.75rem; font-size: var(--text-small); color: var(--color-text-secondary); }
.meter__sources {
  display: flex; flex-wrap: wrap; gap: 0.5rem 1rem; margin-top: 0.75rem; padding: 0.75rem 0 0;
  list-style: none; border-top: 1px solid var(--color-border);
  font-size: var(--text-caption); font-weight: 600; color: var(--color-text-primary);
}
.meter__sources li { display: inline-flex; align-items: center; gap: 0.4rem; }

@keyframes kernel-pop {
  from { opacity: 0; transform: translateY(6px) scale(0.7); }
  to { opacity: 1; transform: none; }
}
@media (prefers-reduced-motion: reduce) {
  .meter__kernel { animation: none; }
  .meter__like { transition: none; }
}
</style>
