<template>
  <SkeletonRegion class="skeleton-grid" :style="{ '--min': minWidth }">
    <div v-for="n in count" :key="n" class="skeleton-grid__item">
      <SkeletonBlock radius="12px" :aspect="aspect" />
      <template v-if="caption">
        <SkeletonBlock width="80%" />
        <SkeletonBlock width="45%" height="12px" />
      </template>
    </div>
  </SkeletonRegion>
</template>

<script setup lang="ts">
import SkeletonBlock from "./SkeletonBlock.vue"
import SkeletonRegion from "./SkeletonRegion.vue"

withDefaults(
  defineProps<{ count?: number; minWidth?: string; aspect?: string; caption?: boolean }>(),
  { count: 12, minWidth: "160px", aspect: "2 / 3", caption: true },
)
</script>

<style scoped>
.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(var(--min), 100%), 1fr));
  gap: 1.5rem 1rem;
}

.skeleton-grid__item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
