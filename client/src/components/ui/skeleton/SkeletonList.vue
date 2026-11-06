<template>
  <SkeletonRegion class="skeleton-list">
    <div v-for="n in count" :key="n" class="skeleton-list__row">
      <SkeletonBlock v-if="avatar" circle width="40px" height="40px" />
      <div class="skeleton-list__text">
        <SkeletonBlock width="40%" />
        <SkeletonBlock :width="n % 2 ? '90%' : '70%'" height="12px" />
      </div>
    </div>
  </SkeletonRegion>
</template>

<script setup lang="ts">
import SkeletonBlock from "./SkeletonBlock.vue"
import SkeletonRegion from "./SkeletonRegion.vue"

withDefaults(defineProps<{ count?: number; avatar?: boolean }>(), { count: 4, avatar: false })
</script>

<style scoped>
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 0.5rem 0;
}

.skeleton-list__row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.skeleton-list__text {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}
</style>
