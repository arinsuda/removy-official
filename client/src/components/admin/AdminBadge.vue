<template>
  <span class="admin-badge" :class="`admin-badge--${tone}`"><slot /></span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ tone?: "neutral" | "brand" | "success" | "danger" | "warning" | "info" }>(), {
  tone: "neutral",
})
</script>

<style scoped>
.admin-badge {
  --tone: var(--color-text-secondary);
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-btn);
  background: color-mix(in srgb, var(--tone) 14%, transparent);
  font-size: var(--text-caption);
  font-weight: 700;
  line-height: 1.5;
  white-space: nowrap;
  color: var(--tone);
}

.admin-badge--brand {
  --tone: var(--color-brand);
}

.admin-badge--success {
  --tone: var(--color-success);
}

.admin-badge--danger {
  --tone: var(--color-danger);
}

.admin-badge--warning {
  --tone: var(--color-warning);
}

.admin-badge--info {
  --tone: var(--color-info);
}
</style>
