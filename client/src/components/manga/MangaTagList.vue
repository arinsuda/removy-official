<template>
  <ul class="manga-tags" :aria-label="label">
    <li v-for="tag in tags" :key="tag" class="manga-tags__tag">{{ tag }}</li>
  </ul>
</template>

<script setup lang="ts">
defineProps<{ label: string; tags: string[] }>()
</script>

<style scoped>
.manga-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.manga-tags__tag {
  padding: 0.25rem 0.75rem;
  border-radius: var(--radius-full);
  border: 1px solid var(--color-border-strong);
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
}
</style>
