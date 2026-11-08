<template>
  <span
    class="popcorn-score-badge relative inline-flex flex-col items-center"
    :class="size === 'detail' ? 'gap-1' : ''"
    :data-tier="tier"
    :data-size="size"
  >
    <span
      role="img"
      :aria-label="tier === 'none' ? t('media.rating.tier.none') : t('media.rating.popcornScore', { score: display })"
      class="relative inline-block shrink-0"
      :class="size === 'detail' ? 'h-[72px] w-16' : 'h-[34px] w-[30px]'"
    >
      <svg viewBox="0 0 48 56" class="absolute inset-0 h-full w-full overflow-visible" aria-hidden="true" focusable="false">
        <!-- Kernels -->
        <g v-if="tier !== 'none'" :class="kernelClass">
          <circle cx="13" cy="19" r="8" />
          <circle cx="24" cy="14" r="9" />
          <circle cx="35" cy="19" r="8" />
          <circle cx="18" cy="10" r="6" />
          <circle cx="30" cy="10" r="6" />
          <template v-if="tier === 'hot'">
            <circle cx="8" cy="22" r="5" />
            <circle cx="40" cy="22" r="5" />
            <circle cx="24" cy="5" r="5" />
          </template>
        </g>
        <!-- Hot accent glow behind the bucket rim -->
        <ellipse v-if="tier === 'hot'" cx="24" cy="24" rx="21" ry="4" class="fill-brand/40" />
        <!-- Bucket body -->
        <path
          d="M5 24 H43 L38 54 H10 Z"
          :class="tier === 'none' ? 'fill-none stroke-text-muted' : 'fill-brand stroke-transparent'"
          stroke-width="1.5"
          stroke-linejoin="round"
          :stroke-dasharray="tier === 'none' ? '3 2' : undefined"
        />
        <g v-if="tier !== 'none'" class="fill-text-primary/25">
          <path d="M14 24 L13 54 H17 L18 24 Z" />
          <path d="M30 24 L31 54 H35 L34 24 Z" />
        </g>
        <!-- Label band that carries the number -->
        <rect x="9" y="33" width="30" height="14" rx="3" class="fill-surface" :class="tier === 'hot' ? 'stroke-accent' : 'stroke-transparent'" stroke-width="1" />
      </svg>
      <span
        data-test="score"
        class="absolute left-0 right-0 text-center font-bold leading-none tabular-nums text-text-primary"
        :class="size === 'detail' ? 'top-[71%] -translate-y-1/2 text-[17px]' : 'top-[71%] -translate-y-1/2 text-[10px]'"
        aria-hidden="true"
      >{{ display }}</span>
    </span>
    <span
      v-if="size === 'detail'"
      data-test="tier-label"
      class="text-caption font-semibold leading-none"
      :class="labelClass"
    >{{ t(`media.rating.tier.${tier}`) }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { popcornTier, roundScore } from "./popcornTier"

const props = withDefaults(
  defineProps<{
    average?: number | null
    count?: number
    size?: "card" | "detail"
  }>(),
  { average: 0, count: 0, size: "card" },
)

const { t } = useI18n()

const display = computed(() => roundScore(props.average).toFixed(1))
const tier = computed(() => popcornTier(props.average, props.count))

const kernelClass = computed(() => {
  if (tier.value === "stale") return "fill-text-muted"
  if (tier.value === "hot") return "fill-accent"
  return "fill-rating"
})

const labelClass = computed(() => {
  if (tier.value === "hot") return "text-accent"
  if (tier.value === "fresh") return "text-rating"
  return "text-text-muted"
})
</script>
