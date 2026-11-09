<template>
  <div class="mg-wrap">
    <ul class="mg" :aria-label="label">
      <li v-for="(item, i) in items" :key="item.key" class="mg__cell" :class="cellState?.(item) && `is-${cellState(item)}`" :data-media-key="item.key" :style="{ '--i': Math.min(i, 12) }">
        <HomePosterCard :item="item" :rail="rail" :sizes="SIZES" :caption="caption?.(item)" />
        <div v-if="$slots.badge" class="mg__badge"><slot name="badge" :item="item" /></div>
        <div v-if="$slots.footer" class="mg__foot"><slot name="footer" :item="item" /></div>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import type { HomeMedia } from "@/components/home/homeMedia"
import { provideHomePopup } from "@/composables/useHomePopup"

defineProps<{
  items: HomeMedia[]
  /** Keeps hover-popup keys unique per tab. */
  rail: string
  label: string
  /** Replaces the release year under the title. */
  caption?: (item: HomeMedia) => string | undefined
  /** Marks a cell while it is dragged or hovered as a drop target. */
  cellState?: (item: HomeMedia) => "dragging" | "target" | undefined
}>()
defineSlots<{ footer?: (props: { item: HomeMedia }) => unknown; badge?: (props: { item: HomeMedia }) => unknown }>()

provideHomePopup()

const SIZES = "(min-width: 1024px) 15vw, (min-width: 768px) 22vw, (min-width: 480px) 30vw, 46vw"
</script>

<style scoped>
/* The profile column is narrower than the viewport, so columns follow the container. */
.mg-wrap { container-type: inline-size; }
.mg { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1.25rem 0.75rem; margin: 0; padding: 0; list-style: none; }
.mg__cell { display: flex; flex-direction: column; gap: 0.375rem; min-width: 0; animation: mg-in var(--motion-base) var(--ease-out-soft) calc(var(--i) * 30ms) both; }
.mg__cell.is-dragging { opacity: 0.5; }
.mg__cell.is-target { outline: 2px dashed var(--color-accent); outline-offset: 4px; border-radius: var(--radius-card); }
.mg :deep(.poster-info) { min-height: 0; }
.mg__badge { display: flex; min-width: 0; padding: 0 2px; }
.mg__foot { display: flex; align-items: center; justify-content: space-between; gap: 0.375rem; min-width: 0; padding: 0 2px; }
@keyframes mg-in { from { opacity: 0; transform: translateY(8px); } }
@container (min-width: 30rem) { .mg { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@container (min-width: 44rem) { .mg { grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 1.5rem 1rem; } }
@container (min-width: 60rem) { .mg { grid-template-columns: repeat(5, minmax(0, 1fr)); } }
@media (prefers-reduced-motion: reduce) { .mg__cell { animation: none; } }
</style>
