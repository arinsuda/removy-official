<template>
  <section ref="root" class="rail" :class="{ 'rail--wide': wide }" :aria-labelledby="titleId">
    <header class="rail__header">
      <div class="rail__heading">
        <h2 :id="titleId" class="rail__title">{{ title }}</h2>
        <p v-if="subtitle" class="rail__subtitle">{{ subtitle }}</p>
      </div>
      <RouterLink v-if="seeAllTo" :to="seeAllTo" class="rail__see-all link-nudge">
        <span>{{ t("home.rail.seeAll") }}</span>
        <ChevronRight :size="16" aria-hidden="true" />
      </RouterLink>
    </header>

    <slot name="toolbar" />

    <div v-if="loading" class="rail__track" aria-hidden="true">
      <div v-for="i in skeletonCount" :key="i" class="rail__cell">
        <SkeletonBlock aspect="2 / 3" radius="12px" />
      </div>
    </div>
    <AppErrorState v-else-if="error" :title="t('home.rail.error')" @retry="emit('retry')" />
    <AppEmptyState v-else-if="empty" :title="emptyTitle ?? t('home.rail.empty')" :description="emptyDescription">
      <template v-if="$slots.empty" #action><slot name="empty" /></template>
    </AppEmptyState>
    <div v-else class="rail__viewport">
      <AppButton
        v-show="canPrev"
        variant="outline"
        size="sm"
        icon-only
        class="rail__arrow rail__arrow--prev"
        :aria-label="t('home.rail.prev')"
        @click="page(-1)"
      >
        <ChevronLeft :size="18" aria-hidden="true" />
      </AppButton>
      <div ref="track" class="rail__track" @scroll.passive="measure">
        <slot />
      </div>
      <AppButton
        v-show="canNext"
        variant="outline"
        size="sm"
        icon-only
        class="rail__arrow rail__arrow--next"
        :aria-label="t('home.rail.next')"
        @click="page(1)"
      >
        <ChevronRight :size="18" aria-hidden="true" />
      </AppButton>
    </div>
  </section>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, useId, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { ChevronLeft, ChevronRight } from "lucide-vue-next"
import type { RouteLocationRaw } from "vue-router"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import { useInView } from "@/composables/useInView"

withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    seeAllTo?: RouteLocationRaw
    loading?: boolean
    error?: boolean
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
    skeletonCount?: number
    /** Fewer, wider cells for text-heavy cards. */
    wide?: boolean
  }>(),
  { skeletonCount: 8 },
)
const emit = defineEmits<{
  retry: []
  /** Fires once, when the rail first gets near the viewport (lazy data loading). */
  visible: []
}>()

const { t } = useI18n()
const titleId = useId()
const root = useTemplateRef<HTMLElement>("root")
const track = useTemplateRef<HTMLElement>("track")

const seen = useInView(root)
watch(seen, (v) => v && emit("visible"), { immediate: true })

const canPrev = ref(false)
const canNext = ref(false)
let resizeObserver: ResizeObserver | null = null

function measure() {
  const el = track.value
  if (!el) return
  canPrev.value = el.scrollLeft > 4
  canNext.value = el.scrollLeft + el.clientWidth < el.scrollWidth - 4
}

function page(dir: 1 | -1) {
  const el = track.value
  if (!el) return
  const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches
  el.scrollBy({ left: dir * el.clientWidth * 0.85, behavior: reduce ? "auto" : "smooth" })
}

// The track mounts only in the ready state, so (re)attach whenever it appears.
watch(track, async (el) => {
  resizeObserver?.disconnect()
  if (!el) return
  await nextTick()
  measure()
  if (typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(measure)
    resizeObserver.observe(el)
  }
})
onMounted(measure)
onBeforeUnmount(() => resizeObserver?.disconnect())
</script>

<style scoped>
.rail {
  --rail-cols: 2.3;
  --rail-gap: 12px;
  min-width: 0;
}
.rail__header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 1rem;
}
.rail__title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  line-height: 1.2;
  color: var(--color-text-primary);
}
.rail__subtitle {
  display: none;
  margin: 2px 0 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.rail__see-all {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 4px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: color 0.2s;
}
.rail__see-all:hover {
  color: var(--color-brand);
}
.rail__viewport {
  position: relative;
}
.rail__track {
  display: flex;
  gap: var(--rail-gap);
  overflow-x: auto;
  overscroll-behavior-x: contain;
  scroll-snap-type: x proximity;
  scrollbar-width: none;
  /* Room for the hover lift and focus ring that overflow would clip. */
  padding: 6px 2px 8px;
  margin: -6px -2px -8px;
}
.rail__track::-webkit-scrollbar {
  display: none;
}
.rail__track > :deep(*),
.rail__cell {
  flex: 0 0 calc((100% - var(--rail-gap) * (var(--rail-cols) - 1)) / var(--rail-cols));
  min-width: 0;
  scroll-snap-align: start;
}
.rail__arrow {
  position: absolute;
  top: 38%;
  z-index: 5;
  display: none;
  background: var(--color-surface-2);
  transform: translateY(-50%);
}
.rail__arrow--prev {
  left: -14px;
}
.rail__arrow--next {
  right: -14px;
}
@media (min-width: 480px) {
  .rail {
    --rail-cols: 2.6;
  }
  .rail--wide {
    --rail-cols: 1.8;
  }
}
@media (min-width: 640px) {
  .rail {
    --rail-cols: 3.5;
  }
  .rail--wide {
    --rail-cols: 2.5;
  }
  .rail__title {
    font-size: 1.25rem;
  }
  .rail__subtitle {
    display: block;
  }
}
@media (min-width: 768px) {
  .rail {
    --rail-cols: 4.5;
    --rail-gap: 14px;
  }
  .rail--wide {
    --rail-cols: 3.2;
  }
}
@media (min-width: 1024px) {
  .rail {
    --rail-cols: 6;
    --rail-gap: 16px;
  }
  .rail--wide {
    --rail-cols: 4;
  }
  .rail__track {
    scroll-snap-type: x mandatory;
  }
  .rail__arrow {
    display: inline-flex;
  }
}
</style>
