<template>
  <div class="carousel">
    <div
      ref="track"
      class="carousel__track"
      role="region"
      tabindex="0"
      aria-roledescription="carousel"
      :aria-label="t('share.previewLabel')"
      @scroll.passive="onScroll"
      @keydown="onKeydown"
    >
      <div
        v-for="(def, i) in templates"
        :key="def.key"
        class="carousel__slide"
        role="group"
        aria-roledescription="slide"
        :aria-label="`${t(def.labelKey)} (${i + 1} / ${templates.length})`"
        :aria-hidden="def.key !== model"
        :style="{ minHeight: `${stageHeight}px` }"
      >
        <div
          class="carousel__frame"
          :class="{ 'carousel__frame--transparent': def.transparent }"
          :style="{ width: `${def.width * scaleFor(def)}px`, height: `${def.height * scaleFor(def)}px` }"
        >
          <div class="carousel__scaler" :style="{ width: `${def.width}px`, height: `${def.height}px`, transform: `scale(${scaleFor(def)})` }">
            <ShareCardRenderer :ref="(el) => setRenderer(def.key, el)" v-bind="cardProps" :template="def.key" />
          </div>
        </div>
      </div>
    </div>
    <p class="carousel__live" aria-live="polite">{{ activeLabel }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import type { ShareMediaContext, ShareReviewContext, ShareStatsContext, ShareTemplateKey, ShareTheme } from "@/types/share"
import ShareCardRenderer from "./ShareCardRenderer.vue"
import type { ShareTemplateDef } from "./shareCards"

const props = defineProps<{
  templates: ShareTemplateDef[]
  media: ShareMediaContext
  posterUrl: string | null
  avatarUrl: string | null
  theme: ShareTheme
  shortUrl: string
  stats: ShareStatsContext | null
  review: ShareReviewContext | null
  /** Tallest a preview may be, in px. */
  maxHeight: number
}>()
const model = defineModel<ShareTemplateKey>({ required: true })
const { t } = useI18n()

const track = useTemplateRef<HTMLElement>("track")
const trackWidth = ref(320)
const renderers = new Map<ShareTemplateKey, InstanceType<typeof ShareCardRenderer> | null>()
let resizeObserver: ResizeObserver | null = null

const cardProps = computed(() => ({
  media: props.media,
  posterUrl: props.posterUrl,
  avatarUrl: props.avatarUrl,
  theme: props.theme,
  shortUrl: props.shortUrl,
  stats: props.stats,
  review: props.review,
}))

const activeIndex = computed(() => Math.max(0, props.templates.findIndex((d) => d.key === model.value)))
const activeLabel = computed(() => {
  const def = props.templates[activeIndex.value]
  return def ? `${t(def.labelKey)} (${activeIndex.value + 1} / ${props.templates.length})` : ""
})

function scaleFor(def: ShareTemplateDef): number {
  return Math.min(trackWidth.value / def.width, props.maxHeight / def.height)
}

const stageHeight = computed(() => Math.max(0, ...props.templates.map((d) => d.height * scaleFor(d))))

function setRenderer(key: ShareTemplateKey, el: unknown) {
  renderers.set(key, el as InstanceType<typeof ShareCardRenderer> | null)
}

/** The full-size DOM node of a template, for export. */
function getNode(key: ShareTemplateKey): HTMLElement | null {
  return renderers.get(key)?.getNode() ?? null
}
defineExpose({ getNode })

const reducedMotion = () => window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? false

function scrollToIndex(i: number) {
  const el = track.value
  if (!el) return
  el.scrollTo({ left: i * el.clientWidth, behavior: reducedMotion() ? "auto" : "smooth" })
}

let raf = 0
function onScroll() {
  cancelAnimationFrame(raf)
  raf = requestAnimationFrame(() => {
    const el = track.value
    if (!el || !el.clientWidth) return
    const def = props.templates[Math.round(el.scrollLeft / el.clientWidth)]
    if (def && def.key !== model.value) model.value = def.key
  })
}

function onKeydown(e: KeyboardEvent) {
  const last = props.templates.length - 1
  let next = activeIndex.value
  if (e.key === "ArrowRight") next = Math.min(last, next + 1)
  else if (e.key === "ArrowLeft") next = Math.max(0, next - 1)
  else if (e.key === "Home") next = 0
  else if (e.key === "End") next = last
  else return
  e.preventDefault()
  model.value = props.templates[next]!.key
}

// Tabs and keys drive the model; keep the scroll position in step with it.
watch(model, async () => {
  await nextTick()
  const el = track.value
  if (el && Math.round(el.scrollLeft / (el.clientWidth || 1)) !== activeIndex.value) scrollToIndex(activeIndex.value)
})

onMounted(() => {
  const el = track.value
  if (!el) return
  trackWidth.value = el.clientWidth || trackWidth.value
  if (typeof ResizeObserver === "undefined") return
  resizeObserver = new ResizeObserver(() => {
    trackWidth.value = el.clientWidth || trackWidth.value
    el.scrollLeft = activeIndex.value * el.clientWidth
  })
  resizeObserver.observe(el)
})
onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  cancelAnimationFrame(raf)
})
</script>

<style scoped>
.carousel__track {
  display: flex;
  overflow-x: auto;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
  border-radius: var(--radius-card);
  outline-offset: 2px;
}
.carousel__track::-webkit-scrollbar { display: none; }
.carousel__slide {
  display: flex;
  flex: 0 0 100%;
  align-items: center;
  justify-content: center;
  scroll-snap-align: center;
  scroll-snap-stop: always;
}
.carousel__frame {
  position: relative;
  overflow: hidden;
  border-radius: 12px;
  background: var(--color-surface);
  pointer-events: none;
}
/* Checkerboard shows the sticker is transparent. */
.carousel__frame--transparent {
  background:
    repeating-conic-gradient(var(--color-surface-3) 0% 25%, var(--color-surface-2) 0% 50%) 0 0 / 24px 24px;
}
.carousel__scaler { transform-origin: top left; }
.carousel__live {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
}
</style>
