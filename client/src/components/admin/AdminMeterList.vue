<template>
  <figure class="meters">
    <figcaption class="meters__caption">{{ caption }}</figcaption>
    <ul class="meters__list">
      <li v-for="item in items" :key="item.label" class="meters__row">
        <span class="meters__label">{{ item.label }}</span>
        <span class="meters__value">{{ item.value.toLocaleString() }}</span>
        <span class="meters__track" role="meter" :aria-label="item.label" :aria-valuenow="item.value" aria-valuemin="0" :aria-valuemax="max">
          <span class="meters__fill" :style="{ width: `${max ? (item.value / max) * 100 : 0}%` }" />
        </span>
      </li>
    </ul>
  </figure>
</template>

<script setup lang="ts">
import { computed } from "vue"

const props = defineProps<{ caption: string; items: { label: string; value: number }[] }>()
const max = computed(() => Math.max(0, ...props.items.map((i) => i.value)))
</script>

<style scoped>
.meters {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin: 0;
}

.meters__caption {
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
}

.meters__list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.meters__row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 0.25rem 0.75rem;
  font-size: var(--text-caption);
}

.meters__label {
  color: var(--color-text-secondary);
}

.meters__value {
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-primary);
}

.meters__track {
  grid-column: 1 / -1;
  height: 0.5rem;
  overflow: hidden;
  border-radius: var(--radius-btn);
  background: var(--color-surface-3);
}

.meters__fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--color-info);
  transition: width var(--motion-base) var(--ease-out-soft);
}

@media (prefers-reduced-motion: reduce) {
  .meters__fill {
    transition: none;
  }
}
</style>
