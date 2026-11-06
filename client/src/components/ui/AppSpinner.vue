<template>
  <span
    class="app-spinner"
    :style="{ width: `${size}px`, height: `${size}px` }"
    aria-hidden="true"
  />
</template>

<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 18 })
</script>

<style scoped>
/* currentColor lets the spinner match any button or surface it sits on. */
.app-spinner {
  display: inline-block;
  flex-shrink: 0;
  border: 2px solid color-mix(in srgb, currentcolor 25%, transparent);
  border-top-color: currentcolor;
  border-radius: 50%;
  animation: app-spinner-turn 0.7s linear infinite;
}

@keyframes app-spinner-turn {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .app-spinner {
    animation-duration: 1.6s;
  }
}
</style>
