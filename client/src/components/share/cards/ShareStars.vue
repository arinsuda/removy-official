<template>
  <span class="sc-stars" role="img" :aria-label="`${rating} / 5`" :style="{ '--sc-star': `${size}px` }">
    <span v-for="n in 5" :key="n" class="sc-star">
      <svg class="sc-star__bg" viewBox="0 0 24 24" aria-hidden="true"><path :d="PATH" fill="currentColor" /></svg>
      <span class="sc-star__fill" :style="{ width: `${fillFor(n)}%` }">
        <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="PATH" fill="currentColor" /></svg>
      </span>
    </span>
  </span>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{ rating: number; size?: number }>(), { size: 96 })

const PATH = "M12 2.5l2.9 6.3 6.9.8-5.1 4.7 1.4 6.8L12 17.7 5.9 21.1l1.4-6.8L2.2 9.6l6.9-.8z"

/** 0, 50 or 100 for each star; ratings use half steps. */
function fillFor(n: number): number {
  const diff = props.rating - (n - 1)
  return diff >= 1 ? 100 : diff >= 0.5 ? 50 : 0
}
</script>

<style scoped>
.sc-stars {
  display: inline-flex;
  gap: calc(var(--sc-star) * 0.08);
}
.sc-star {
  position: relative;
  display: block;
  width: var(--sc-star);
  height: var(--sc-star);
}
.sc-star svg {
  display: block;
  width: var(--sc-star);
  height: var(--sc-star);
}
/* currentColor (not var() on fill) so html-to-image keeps the colour in the export. */
.sc-star__bg {
  color: var(--sc-star-empty);
}
.sc-star__fill {
  position: absolute;
  inset: 0 auto 0 0;
  overflow: hidden;
  color: var(--sc-star-fill);
}
</style>
