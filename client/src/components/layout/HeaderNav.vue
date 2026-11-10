<template>
  <nav ref="nav" class="hn" :aria-label="t('navigation.header.mainNav')">
    <RouterLink
      v-for="item in NAV_ITEMS"
      :key="item.key"
      :to="item.to"
      class="hn-link"
      :class="{ 'hn-link--active': item.isActive(route) }"
      :aria-current="item.isActive(route) ? 'page' : undefined"
    >
      {{ t(item.labelKey) }}
    </RouterLink>
    <span class="hn-indicator" :class="{ 'hn-indicator--ready': ready }" :style="indicatorStyle" aria-hidden="true" />
  </nav>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { useRoute } from "vue-router"
import { useI18n } from "vue-i18n"
import { NAV_ITEMS } from "./navItems"

const { t, locale } = useI18n()
const route = useRoute()
const nav = ref<HTMLElement | null>(null)
const x = ref(0)
const w = ref(0)
const ready = ref(false)
let observer: ResizeObserver | null = null

const indicatorStyle = computed(() => ({
  width: `${w.value}px`,
  transform: `translateX(${x.value}px)`,
  opacity: w.value > 0 ? 1 : 0,
}))

function measure() {
  const active = nav.value?.querySelector<HTMLElement>(".hn-link--active")
  if (!active) {
    w.value = 0
    return
  }
  // Inset the bar to the text, not the padded hit area.
  const inset = 12
  x.value = active.offsetLeft + inset
  w.value = Math.max(active.offsetWidth - inset * 2, 0)
}

watch(
  () => [route.fullPath, locale.value],
  async () => {
    await nextTick()
    measure()
  },
)

onMounted(async () => {
  await nextTick()
  measure()
  // Enable the slide only after the first placement so it does not sweep in from 0.
  requestAnimationFrame?.(() => (ready.value = true))
  if (typeof ResizeObserver !== "undefined" && nav.value) {
    observer = new ResizeObserver(measure)
    observer.observe(nav.value)
  }
})
onBeforeUnmount(() => observer?.disconnect())
</script>

<style scoped>
.hn {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.125rem;
}
.hn-link {
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-decoration: none;
  white-space: nowrap;
  border-radius: 0.5rem;
  transition: color var(--motion-fast), background var(--motion-fast);
}
.hn-link--active {
  color: var(--color-text-primary);
  font-weight: 600;
}
@media (hover: hover) {
  .hn-link:hover {
    color: var(--color-text-primary);
    background: color-mix(in srgb, var(--color-text-primary) 7%, transparent);
  }
}
.hn-indicator {
  position: absolute;
  left: 0;
  bottom: -0.5rem;
  height: 2px;
  border-radius: 2px;
  background: var(--color-brand);
  pointer-events: none;
}
.hn-indicator--ready {
  transition: transform var(--motion-base) var(--ease-out-soft), width var(--motion-base) var(--ease-out-soft), opacity var(--motion-fast);
}
@media (max-width: 1279.98px) {
  .hn-link {
    padding: 0.5rem 0.5rem;
    font-size: 0.8125rem;
  }
}
@media (prefers-reduced-motion: reduce) {
  .hn-indicator--ready {
    transition: none;
  }
}
</style>
