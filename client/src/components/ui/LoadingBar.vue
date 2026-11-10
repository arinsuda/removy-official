<template>
  <div class="loading-bar" role="progressbar" aria-busy="true" :aria-label="$t('common.loading')">
    <span class="loading-bar__fill" />
  </div>
</template>

<style scoped>
.loading-bar {
  width: 140px;
  height: 3px;
  overflow: hidden;
  background: var(--color-border);
  border-radius: 99px;
}

.loading-bar__fill {
  display: block;
  width: 45%;
  height: 100%;
  background: linear-gradient(90deg, var(--color-brand), var(--color-accent));
  border-radius: inherit;
  animation: loading-bar-slide 1.3s var(--ease-water) infinite;
}

@keyframes loading-bar-slide {
  from {
    transform: translateX(-110%);
  }
  to {
    transform: translateX(230%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .loading-bar__fill {
    width: 100%;
    animation: none;
  }
}
</style>
