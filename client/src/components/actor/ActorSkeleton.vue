<template>
  <SkeletonRegion class="actor-skeleton">
    <div class="actor-skeleton__hero">
      <SkeletonBlock width="9.5rem" aspect="2 / 3" radius="12px" />
      <div class="actor-skeleton__text">
        <SkeletonBlock width="55%" height="32px" />
        <SkeletonBlock width="75%" height="20px" />
        <SkeletonBlock height="96px" />
      </div>
    </div>
    <div class="actor-skeleton__grid">
      <SkeletonBlock v-for="n in 12" :key="n" aspect="2 / 3" radius="12px" />
    </div>
  </SkeletonRegion>
</template>

<script setup lang="ts">
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonRegion from "@/components/ui/skeleton/SkeletonRegion.vue"
</script>

<style scoped>
.actor-skeleton__hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  margin-bottom: 2rem;
}
.actor-skeleton__text {
  display: flex;
  width: 100%;
  flex-direction: column;
  gap: 0.75rem;
}
.actor-skeleton__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

@media (min-width: 768px) {
  .actor-skeleton__hero {
    flex-direction: row;
    align-items: flex-start;
  }
  .actor-skeleton__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
@media (min-width: 1280px) {
  .actor-skeleton__grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
