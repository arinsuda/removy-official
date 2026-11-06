<template>
  <SkeletonRegion class="skeleton-profile">
    <div class="skeleton-profile__hero">
      <SkeletonBlock circle width="96px" height="96px" />
      <div class="skeleton-profile__identity">
        <SkeletonBlock width="220px" height="28px" />
        <SkeletonBlock width="130px" />
        <SkeletonBlock width="320px" height="12px" />
      </div>
    </div>
    <div class="skeleton-profile__tabs">
      <SkeletonBlock v-for="n in 4" :key="n" width="88px" height="34px" radius="999px" />
    </div>
    <SkeletonGrid :count="6" />
  </SkeletonRegion>
</template>

<script setup lang="ts">
import SkeletonBlock from "./SkeletonBlock.vue"
import SkeletonGrid from "./SkeletonGrid.vue"
import SkeletonRegion from "./SkeletonRegion.vue"
</script>

<style scoped>
.skeleton-profile {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1.5rem;
}

.skeleton-profile__hero {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 1.25rem;
}

.skeleton-profile__identity {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}

.skeleton-profile__tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
