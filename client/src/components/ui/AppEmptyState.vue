<template>
  <div class="app-state">
    <component :is="icon" v-if="icon" :size="36" class="app-state__icon" aria-hidden="true" />
    <p class="app-state__title">{{ title }}</p>
    <p v-if="description" class="app-state__desc">{{ description }}</p>
    <div v-if="$slots.action" class="app-state__action"><slot name="action" /></div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from "vue"

defineProps<{
  title: string
  description?: string
  /** Icon component (e.g. a lucide icon). */
  icon?: Component
}>()
</script>

<style scoped>
.app-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 4rem 1rem;
  text-align: center;
}
.app-state__icon {
  color: var(--color-text-muted);
}
.app-state__title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}
.app-state__desc {
  max-width: 28rem;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
.app-state__action {
  margin-top: 0.5rem;
}
</style>
