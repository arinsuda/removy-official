<template>
  <div class="feed-skeleton">
    <SkeletonBlock circle width="40px" height="40px" />
    <div class="feed-skeleton__body">
      <SkeletonBlock width="70%" height="14px" />
      <SkeletonBlock width="30%" height="12px" />
      <SkeletonBlock height="64px" radius="12px" />
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
</script>

<style scoped>
.feed-skeleton {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.feed-skeleton__body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.625rem;
}
</style>
