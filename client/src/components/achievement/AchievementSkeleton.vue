<template>
  <div class="asx">
    <SkeletonBlock height="13rem" radius="var(--radius-card)" />
    <div class="asx__spot">
      <SkeletonBlock height="4.5rem" radius="var(--radius-card)" />
      <SkeletonBlock height="4.5rem" radius="var(--radius-card)" />
    </div>
    <SkeletonBlock height="2.25rem" radius="var(--radius-btn)" />
    <SkeletonGrid :count="SKELETON_BADGES" min-width="9rem" aspect="1 / 1" :caption="false" />
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"

const SKELETON_BADGES = 8
</script>

<style scoped>
.asx { display: flex; flex-direction: column; gap: 1.25rem; padding-top: 1rem; }
.asx__spot { display: grid; gap: 0.75rem; }
@media (min-width: 640px) {
  .asx__spot { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
</style>
