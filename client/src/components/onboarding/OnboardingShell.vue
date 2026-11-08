<template>
  <div class="onb" :class="{ 'onb--reduced': reduced, 'onb--leaving': leaving }">
    <div class="onb__bg" aria-hidden="true">
      <span class="onb__glow onb__glow--a" />
      <span class="onb__glow onb__glow--b" />
    </div>
    <main class="onb__main">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
defineProps<{ reduced: boolean; leaving: boolean }>()
</script>

<style scoped>
.onb {
  position: relative;
  min-height: 100dvh;
  background: var(--color-surface);
  overflow-x: clip;
  transition:
    opacity var(--motion-base) var(--ease-out-soft),
    transform var(--motion-base) var(--ease-out-soft);
}

.onb--leaving {
  opacity: 0;
  transform: translateY(-12px) scale(0.99);
  pointer-events: none;
}

.onb__bg {
  position: fixed;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

/* Two soft brand glows drift slowly; only transform is animated. */
.onb__glow {
  position: absolute;
  width: 70vmax;
  aspect-ratio: 1;
  border-radius: 50%;
  background: radial-gradient(circle, color-mix(in srgb, var(--color-brand) 28%, transparent), transparent 65%);
  opacity: 0.55;
  will-change: transform;
  animation: onb-drift 26s var(--ease-water) infinite alternate;
}
.onb__glow--a {
  top: -35vmax;
  left: -25vmax;
}
.onb__glow--b {
  right: -30vmax;
  bottom: -40vmax;
  background: radial-gradient(circle, color-mix(in srgb, var(--color-info) 16%, transparent), transparent 65%);
  animation-duration: 34s;
  animation-direction: alternate-reverse;
}

@keyframes onb-drift {
  to {
    transform: translate3d(12vmax, 8vmax, 0) scale(1.12);
  }
}

.onb__main {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  min-height: 100dvh;
  width: 100%;
  max-width: 1100px;
  margin: 0 auto;
  padding: 1.5rem 1rem 0;
}

@media (min-width: 768px) {
  .onb__main {
    padding-top: 2.5rem;
  }
}

.onb--reduced,
.onb--reduced .onb__glow {
  transition-duration: 1ms;
  animation: none;
}

@media (prefers-reduced-motion: reduce) {
  .onb,
  .onb__glow {
    transition-duration: 1ms;
    animation: none;
  }
}
</style>
