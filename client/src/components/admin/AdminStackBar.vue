<template>
  <figure class="stack">
    <figcaption class="stack__caption">{{ caption }}</figcaption>
    <div class="stack__bar" role="img" :aria-label="summary">
      <span
        v-for="s in segments"
        :key="s.label"
        class="stack__seg"
        :class="`stack__seg--${s.tone}`"
        :style="{ flexGrow: s.value }"
      />
    </div>
    <ul class="stack__legend">
      <li v-for="s in segments" :key="s.label">
        <span class="stack__swatch" :class="`stack__seg--${s.tone}`" aria-hidden="true" />
        {{ s.label }}
        <strong>{{ s.value.toLocaleString() }}</strong>
      </li>
    </ul>
  </figure>
</template>

<script setup lang="ts">
export interface StackSegment {
  label: string
  value: number
  tone: "success" | "danger" | "info" | "warning"
}

defineProps<{ caption: string; summary: string; segments: StackSegment[] }>()
</script>

<style scoped>
.stack {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin: 0;
}

.stack__caption {
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
}

.stack__bar {
  display: flex;
  gap: 2px;
  height: 0.75rem;
  overflow: hidden;
  border-radius: var(--radius-btn);
  background: var(--color-surface-3);
}

.stack__seg--success {
  background: var(--color-success);
}

.stack__seg--danger {
  background: var(--color-danger);
}

.stack__seg--info {
  background: var(--color-info);
}

.stack__seg--warning {
  background: var(--color-warning);
}

.stack__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1.25rem;
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
}

.stack__legend li {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.stack__legend strong {
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

.stack__swatch {
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 50%;
}
</style>
