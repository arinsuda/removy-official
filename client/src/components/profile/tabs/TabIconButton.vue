<template>
  <button type="button" class="tib" :class="{ 'tib--danger': danger }" :aria-label="label" :title="label">
    <slot />
  </button>
</template>

<script setup lang="ts">
defineProps<{ label: string; danger?: boolean }>()
</script>

<style scoped>
.tib { display: inline-grid; place-items: center; width: 2rem; height: 2rem; color: var(--color-text-muted); background: none; border: 0; border-radius: var(--radius-btn); cursor: pointer; transition: color var(--motion-fast) var(--ease-water), background-color var(--motion-fast) var(--ease-water); }
@media (hover: hover) {
  .tib:hover { color: var(--color-text-primary); background: var(--color-surface-3); }
  .tib--danger:hover { color: var(--color-danger); }
}
@media (prefers-reduced-motion: reduce) { .tib { transition: none; } }
</style>
