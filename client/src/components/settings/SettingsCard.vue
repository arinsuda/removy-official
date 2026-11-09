<template>
  <section :id="id" class="card" :aria-labelledby="`${id}-title`">
    <h2 :id="`${id}-title`" class="card__title">{{ title }}</h2>
    <p v-if="description" class="card__desc">{{ description }}</p>
    <div class="card__body"><slot /></div>
  </section>
</template>

<script setup lang="ts">
defineProps<{ id: string; title: string; description?: string }>()
</script>

<style scoped>
.card { scroll-margin-top: 6rem; padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.card__title { font-size: 1.0625rem; font-weight: 700; color: var(--color-text-primary); }
.card__desc { margin-top: 0.25rem; font-size: 0.8125rem; color: var(--color-text-muted); }
.card__body { display: flex; flex-direction: column; gap: 1rem; margin-top: 1rem; }
</style>
