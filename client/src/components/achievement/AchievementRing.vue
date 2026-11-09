<template>
  <div class="arg" role="progressbar" :aria-valuenow="pct" aria-valuemin="0" aria-valuemax="100" :aria-label="label">
    <svg viewBox="0 0 36 36" class="arg__svg" aria-hidden="true">
      <circle class="arg__track" cx="18" cy="18" r="15.915" pathLength="100" />
      <circle class="arg__fill" cx="18" cy="18" r="15.915" pathLength="100" :stroke-dasharray="`${pct} 100`" />
    </svg>
    <div class="arg__center">
      <strong class="arg__pct">{{ n(pct / 100, "percent") }}</strong>
      <span class="arg__sub">{{ sub }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"

defineProps<{ pct: number; label: string; sub: string }>()
const { n } = useI18n()
</script>

<style scoped>
.arg { position: relative; flex: none; width: 8rem; height: 8rem; }
.arg__svg { width: 100%; height: 100%; transform: rotate(-90deg); }
.arg__track, .arg__fill { fill: none; stroke-width: 3; }
.arg__track { stroke: var(--color-surface-3); }
.arg__fill { stroke: var(--color-brand); stroke-linecap: round; transition: stroke-dasharray var(--motion-wave) var(--ease-out-soft); }
.arg__center { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.125rem; }
.arg__pct { font-size: var(--text-heading); font-weight: 700; line-height: 1; color: var(--color-text-primary); }
.arg__sub { font-size: var(--text-caption); color: var(--color-text-secondary); }
@media (min-width: 768px) {
  .arg { width: 9.5rem; height: 9.5rem; }
}
@media (prefers-reduced-motion: reduce) {
  .arg__fill { transition: none; }
}
</style>
