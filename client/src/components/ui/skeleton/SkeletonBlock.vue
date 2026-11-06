<template>
  <span
    class="skeleton-block"
    :class="{ 'skeleton-block--circle': circle }"
    :style="{ width, height: aspect ? 'auto' : height, aspectRatio: aspect, borderRadius: circle ? undefined : radius }"
    aria-hidden="true"
  />
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{ width?: string; height?: string; radius?: string; circle?: boolean; aspect?: string }>(),
  { width: "100%", height: "14px", radius: "8px", circle: false },
)
</script>

<style scoped>
/* background-attachment: fixed makes every block share one sweep across the page. */
.skeleton-block {
  display: block;
  flex-shrink: 0;
  max-width: 100%;
  background-color: #1f1f1f;
  background-image: linear-gradient(100deg, transparent 35%, rgb(255 255 255 / 7%) 50%, transparent 65%);
  background-attachment: fixed;
  background-size: 200vw 100%;
  animation: skeleton-sweep 1.8s linear infinite;
}

.skeleton-block--circle {
  border-radius: 50%;
}

@keyframes skeleton-sweep {
  from {
    background-position: 200vw 0;
  }
  to {
    background-position: -200vw 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton-block {
    animation: none;
    background-image: none;
  }
}
</style>
