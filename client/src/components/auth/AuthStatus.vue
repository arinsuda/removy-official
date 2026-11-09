<template>
  <div class="auth-status form-item" :class="`auth-status--${tone}`">
    <span v-if="icon" class="auth-status__icon" aria-hidden="true"><component :is="icon" :size="28" /></span>
    <slot />
  </div>
</template>

<script setup lang="ts">
import type { Component } from "vue"

withDefaults(defineProps<{ icon?: Component; tone?: "info" | "success" | "danger" }>(), { tone: "info" })
</script>

<style scoped>
.auth-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  text-align: center;
}

.auth-status--info {
  --status-color: var(--color-brand);
}

.auth-status--success {
  --status-color: var(--color-success);
}

.auth-status--danger {
  --status-color: var(--color-danger);
}

.auth-status__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  border-radius: 50%;
  background: color-mix(in srgb, var(--status-color) 14%, transparent);
  color: var(--status-color);
}

.auth-status :slotted(p) {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.6;
  color: var(--color-text-secondary);
}

.auth-status :slotted(strong) {
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.auth-status :slotted(.auth-status__note) {
  padding: 0.625rem 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  font-size: var(--text-caption);
}

.auth-status :slotted(.auth-link) {
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.auth-status :slotted(.auth-status__actions) {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
}
</style>
