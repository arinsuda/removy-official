<template>
  <svg
    class="popcorn-rating-icon"
    viewBox="0 0 24 24"
    :width="size"
    :height="size"
  >
    <defs>
      <linearGradient :id="gradId" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stop-color="#FF512F" /> <!-- hardcode-ok: brand gradient -->
        <stop offset="100%" stop-color="#DD2476" /> <!-- hardcode-ok: brand gradient -->
      </linearGradient>
    </defs>
    <!-- Branded Ticket Notch SVG -->
    <path
      :fill="`url(#${gradId})`"
      d="M20 2H4c-1.1 0-1.99.9-1.99 2v4c1.1 0 2 .9 2 2s-.9 2-2 2v4c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2v-4c-1.1 0-2-.9-2-2s.9-2 2-2V4c0-1.1-.9-2-2-2zm-9 12V6l5 4-5 4z"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from "vue";

withDefaults(
  defineProps<{
    size?: number | string;
  }>(),
  {
    size: 24,
  }
);

// Generate unique ID per instance to ensure gradient safety in DOM
const uid = Math.random().toString(36).substring(2, 9);
const gradId = computed(() => `popcorn-grad-${uid}`);
</script>

<style scoped>
.popcorn-rating-icon {
  display: inline-block;
  vertical-align: middle;
}
</style>
