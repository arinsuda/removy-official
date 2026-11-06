<template>
  <SkeletonRegion class="skeleton-detail">
    <SkeletonBlock class="skeleton-detail__poster" aspect="2 / 3" radius="16px" />
    <div class="skeleton-detail__info">
      <SkeletonBlock width="65%" height="40px" radius="10px" />
      <SkeletonBlock width="35%" />
      <div class="skeleton-detail__chips">
        <SkeletonBlock v-for="n in 3" :key="n" width="84px" height="30px" radius="999px" />
      </div>
      <SkeletonBlock v-for="n in 3" :key="`line-${n}`" :width="n === 3 ? '60%' : '100%'" />
      <div class="skeleton-detail__chips">
        <SkeletonBlock v-for="n in 2" :key="n" width="136px" height="44px" radius="999px" />
      </div>
    </div>
  </SkeletonRegion>
</template>

<script setup lang="ts">
import SkeletonBlock from "./SkeletonBlock.vue"
import SkeletonRegion from "./SkeletonRegion.vue"
</script>

<style scoped>
.skeleton-detail {
  display: grid;
  grid-template-columns: minmax(180px, 300px) minmax(0, 1fr);
  gap: 2.5rem;
  max-width: 1200px;
  margin: 0 auto;
  padding: 6rem 1.5rem 3rem;
}

.skeleton-detail__info {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.skeleton-detail__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (max-width: 720px) {
  .skeleton-detail {
    grid-template-columns: minmax(0, 1fr);
    gap: 1.5rem;
    padding-top: 5rem;
  }

  .skeleton-detail__poster {
    width: 55%;
    margin: 0 auto;
  }
}
</style>
