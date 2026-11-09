<template>
  <figure class="bar-chart">
    <figcaption class="bar-chart__caption">{{ caption }}</figcaption>
    <p class="sr-only">{{ summary }}</p>
    <table class="sr-only">
      <thead>
        <tr>
          <th scope="col">{{ labelHeader }}</th>
          <th scope="col">{{ valueHeader }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in points" :key="p.label">
          <th scope="row">{{ p.label }}</th>
          <td>{{ p.value }}</td>
        </tr>
      </tbody>
    </table>

    <div class="bar-chart__plot" aria-hidden="true">
      <div class="bar-chart__axis">
        <span v-for="tick in ticks" :key="tick">{{ tick }}</span>
      </div>
      <ol class="bar-chart__bars">
        <li v-for="p in points" :key="p.label" class="bar-chart__col" :class="{ 'is-peak': p.value === max && max > 0 }">
          <span class="bar-chart__value">{{ p.value }}</span>
          <span class="bar-chart__bar" :style="{ height: `${(p.value / scaleMax) * 100}%` }" />
          <span class="bar-chart__label">{{ p.label }}</span>
        </li>
      </ol>
    </div>
  </figure>
</template>

<script setup lang="ts">
import { computed } from "vue"

export interface BarPoint {
  label: string
  value: number
}

const props = defineProps<{
  points: BarPoint[]
  caption: string
  summary: string
  labelHeader: string
  valueHeader: string
}>()

const max = computed(() => Math.max(0, ...props.points.map((p) => p.value)))
// Round the scale up so the top gridline is a clean number.
const scaleMax = computed(() => {
  if (max.value <= 4) return Math.max(max.value, 1)
  const step = 10 ** Math.floor(Math.log10(max.value))
  return Math.ceil(max.value / step) * step
})
const ticks = computed(() => [scaleMax.value, Math.round(scaleMax.value / 2), 0])
</script>

<style scoped>
.bar-chart {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin: 0;
}

.bar-chart__caption {
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
}

.bar-chart__plot {
  display: flex;
  gap: 0.5rem;
  height: 15rem;
}

.bar-chart__axis {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding-bottom: 1.5rem;
  font-size: var(--text-caption);
  font-variant-numeric: tabular-nums;
  text-align: right;
  color: var(--color-text-muted);
}

.bar-chart__bars {
  display: flex;
  flex: 1;
  gap: 0.25rem;
  min-width: 0;
  margin: 0;
  padding: 0;
  list-style: none;
  background-image: linear-gradient(var(--color-border) 1px, transparent 1px);
  background-size: 100% calc(50% - 0.75rem);
  background-repeat: repeat-y;
  background-position: 0 0;
}

.bar-chart__col {
  position: relative;
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 0.25rem;
  min-width: 0;
  padding-bottom: 1.5rem;
}

.bar-chart__bar {
  width: 100%;
  max-width: 2.5rem;
  min-height: 2px;
  background: color-mix(in srgb, var(--color-brand) 55%, transparent);
  border-radius: 0.25rem 0.25rem 0 0;
  transition: height var(--motion-base) var(--ease-out-soft);
}

.is-peak .bar-chart__bar {
  background: var(--color-brand);
}

.bar-chart__value {
  font-size: var(--text-caption);
  font-variant-numeric: tabular-nums;
  color: var(--color-text-secondary);
}

.bar-chart__label {
  position: absolute;
  bottom: 0;
  font-size: var(--text-caption);
  white-space: nowrap;
  color: var(--color-text-muted);
}

@media (max-width: 639px) {
  /* Thin out labels on phones: every second column keeps its label. */
  .bar-chart__col:nth-child(even) .bar-chart__label {
    visibility: hidden;
  }
}

@media (prefers-reduced-motion: reduce) {
  .bar-chart__bar {
    transition: none;
  }
}
</style>
