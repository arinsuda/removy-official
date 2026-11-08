<template>
  <section class="dist" :aria-labelledby="titleId">
    <h2 :id="titleId" class="dist__title">{{ t("media.detail.distribution.title") }}</h2>
    <p v-if="result.total === 0" class="dist__empty">{{ t("media.detail.distribution.empty") }}</p>
    <template v-else>
      <ol class="dist__rows">
        <li
          v-for="b in rows"
          :key="b.value"
          class="dist__row"
          :aria-label="t('media.detail.distribution.row', { value: b.value.toFixed(1), count: b.count }, b.count)"
        >
          <span class="dist__label" aria-hidden="true">{{ b.value.toFixed(1) }}</span>
          <span class="dist__track" aria-hidden="true">
            <span class="dist__fill" :style="{ width: `${b.pct}%` }" />
          </span>
          <span class="dist__count" aria-hidden="true">{{ b.count }}</span>
        </li>
      </ol>
      <p class="dist__based" data-test="dist-based">
        {{ t("media.detail.distribution.basedOn", { count: result.total }, result.total) }}
      </p>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import type { ReviewResponse } from "@/types/review"
import { ratingDistribution } from "./ratingDistribution"

const props = defineProps<{ reviews: Pick<ReviewResponse, "rating" | "visibility">[] }>()
const { t } = useI18n()
const titleId = useId()
const result = computed(() => ratingDistribution(props.reviews))
/** Highest rating first. */
const rows = computed(() => [...result.value.buckets].reverse())
</script>

<style scoped>
.dist { padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.dist__title { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.dist__empty, .dist__based { margin-top: 0.75rem; font-size: var(--text-caption); color: var(--color-text-muted); }
.dist__rows { display: flex; flex-direction: column; gap: 0.3rem; margin-top: 0.75rem; padding: 0; list-style: none; }
.dist__row { display: grid; grid-template-columns: 2rem minmax(0, 1fr) 1.75rem; align-items: center; gap: 0.5rem; font-size: var(--text-caption); font-variant-numeric: tabular-nums; }
.dist__label { color: var(--color-text-secondary); text-align: right; }
.dist__count { color: var(--color-text-muted); }
.dist__track { height: 0.5rem; overflow: hidden; background: var(--color-surface-3); border-radius: var(--radius-btn); }
.dist__fill { display: block; height: 100%; background: linear-gradient(90deg, var(--color-brand), var(--color-accent)); border-radius: inherit; transition: width var(--motion-base) var(--ease-out-soft); }
@media (prefers-reduced-motion: reduce) {
  .dist__fill { transition: none; }
}
</style>
