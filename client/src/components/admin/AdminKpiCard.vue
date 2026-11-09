<template>
  <article class="kpi" :class="`kpi--${tone}`">
    <header class="kpi__head">
      <h2 class="kpi__label">{{ label }}</h2>
      <span v-if="icon" class="kpi__icon" aria-hidden="true"><component :is="icon" :size="18" /></span>
    </header>
    <p class="kpi__value"><slot name="value">{{ value }}</slot></p>
    <p v-if="$slots.sub" class="kpi__sub"><slot name="sub" /></p>
  </article>
</template>

<script setup lang="ts">
import type { Component } from "vue"

withDefaults(
  defineProps<{
    label: string
    value?: string | number
    icon?: Component
    tone?: "brand" | "success" | "info" | "warning"
  }>(),
  { tone: "brand" },
)
</script>

<style scoped>
.kpi {
  --tone: var(--color-brand);
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  min-width: 0;
  padding: 1rem 1.25rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
}

.kpi--success {
  --tone: var(--color-success);
}

.kpi--info {
  --tone: var(--color-info);
}

.kpi--warning {
  --tone: var(--color-warning);
}

.kpi__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.kpi__label {
  margin: 0;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-secondary);
}

.kpi__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--tone) 14%, transparent);
  color: var(--tone);
}

.kpi__value {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
  color: var(--color-text-primary);
}

.kpi__sub {
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
</style>
