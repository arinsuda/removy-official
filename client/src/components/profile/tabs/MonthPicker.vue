<template>
  <div ref="root" class="mp" @keydown.escape="closeAndFocus">
    <button
      ref="trigger"
      type="button"
      class="mp__btn"
      :class="{ 'is-active': !!model }"
      :aria-expanded="open"
      aria-haspopup="dialog"
      :aria-label="t('profile.reviews.month')"
      @click="toggle"
    >
      <Calendar :size="14" aria-hidden="true" />
      <span class="mp__label">{{ displayLabel }}</span>
      <ChevronDown :size="14" class="mp__chev" :class="{ 'is-open': open }" aria-hidden="true" />
    </button>
    <TabIconButton v-if="model" :label="t('profile.reviews.clearDate')" @click="select('')">
      <X :size="14" aria-hidden="true" />
    </TabIconButton>

    <div v-if="open" ref="pop" class="mp__pop" :class="{ 'mp__pop--end': alignEnd }" role="dialog" :aria-label="t('profile.reviews.month')">
      <div class="mp__head">
        <button
          type="button"
          class="mp__nav"
          :disabled="year <= floorYear"
          :aria-label="t('common.previous')"
          @click="year--"
        >
          <ChevronLeft :size="15" aria-hidden="true" />
        </button>
        <span class="mp__year">{{ formatYear(year) }}</span>
        <button
          type="button"
          class="mp__nav"
          :disabled="year >= currentYear"
          :aria-label="t('common.next')"
          @click="year++"
        >
          <ChevronRight :size="15" aria-hidden="true" />
        </button>
      </div>

      <div class="mp__grid">
        <button
          v-for="m in MONTHS"
          :key="m"
          type="button"
          class="mp__cell"
          :class="{ 'is-active': model === keyFor(year, m) }"
          :aria-pressed="model === keyFor(year, m)"
          :disabled="isFuture(year, m)"
          @click="select(keyFor(year, m))"
        >
          {{ formatMonth(year, m) }}
        </button>
      </div>

      <button type="button" class="mp__all" :class="{ 'is-active': !model }" @click="select('')">
        {{ t("profile.tabContent.reviews.allMonths") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Calendar, ChevronDown, ChevronLeft, ChevronRight, X } from "lucide-vue-next"
import { useClickOutside } from "@/composables/useClickOutside"
import TabIconButton from "./TabIconButton.vue"

const MIN_YEAR = 1970
/** Minimum gap (px) kept between the popup and the viewport edge. */
const VIEWPORT_EDGE = 8
const MONTHS = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12] as const

const props = withDefaults(defineProps<{ minYear?: number }>(), {
  minYear: MIN_YEAR,
})
const model = defineModel<string>({ default: "" })
const emit = defineEmits<{ change: [value: string] }>()

const { t, locale } = useI18n()
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const pop = ref<HTMLElement | null>(null)
const alignEnd = ref(false)
const open = ref(false)

const now = new Date()
const currentYear = now.getFullYear()
const currentMonth = now.getMonth() + 1
const floorYear = computed(() => Math.min(props.minYear || MIN_YEAR, currentYear))
const year = ref(currentYear)

const displayLabel = computed(() => {
  if (!model.value) return t("profile.tabContent.reviews.allMonths")
  const [y, m] = model.value.split("-").map(Number)
  if (!y || !m) return t("profile.tabContent.reviews.allMonths")
  return new Intl.DateTimeFormat(locale.value, { month: "short", year: "numeric" }).format(new Date(y, m - 1, 1))
})

const formatYear = (y: number) =>
  new Intl.DateTimeFormat(locale.value, { year: "numeric" }).format(new Date(y, 0, 1))

const formatMonth = (y: number, m: number) =>
  new Intl.DateTimeFormat(locale.value, { month: "short" }).format(new Date(y, m - 1, 1))

const keyFor = (y: number, m: number) => `${y}-${String(m).padStart(2, "0")}`
const isFuture = (y: number, m: number) => y > currentYear || (y === currentYear && m > currentMonth)

// Open toward the side with room so the popup never runs off a narrow screen.
function updatePlacement() {
  const btn = trigger.value
  const width = pop.value?.offsetWidth ?? 0
  if (!btn || !width) return
  const rect = btn.getBoundingClientRect()
  alignEnd.value = rect.left + width > window.innerWidth - VIEWPORT_EDGE && rect.right - width >= VIEWPORT_EDGE
}

async function toggle() {
  if (open.value) return close()
  if (model.value) {
    const parsedYear = Number(model.value.split("-")[0])
    if (parsedYear) year.value = parsedYear
  }
  alignEnd.value = false
  open.value = true
  await nextTick()
  updatePlacement()
  const target = pop.value?.querySelector<HTMLElement>(".mp__cell.is-active") ?? pop.value?.querySelector<HTMLElement>(".mp__cell:not(:disabled)")
  target?.focus()
}

function close() {
  open.value = false
}

function closeAndFocus() {
  if (!open.value) return
  close()
  trigger.value?.focus()
}

function select(next: string) {
  model.value = next
  closeAndFocus()
  emit("change", next)
}

useClickOutside(root, close)
</script>

<style scoped>
.mp { position: relative; display: inline-flex; align-items: center; gap: 0.25rem; }
.mp__btn { display: inline-flex; align-items: center; gap: 0.45rem; min-height: 2.25rem; padding: 0 0.75rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); white-space: nowrap; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); cursor: pointer; transition: border-color var(--motion-fast) var(--ease-water), color var(--motion-fast) var(--ease-water); }
.mp__btn.is-active { color: var(--color-text-primary); border-color: var(--color-border-strong); }
.mp__chev { color: var(--color-text-muted); transition: transform var(--motion-fast) var(--ease-water); }
.mp__chev.is-open { transform: rotate(180deg); }
.mp__pop { position: absolute; inset-block-start: calc(100% + 0.375rem); inset-inline-start: 0; z-index: var(--z-popover); display: flex; flex-direction: column; gap: 0.625rem; width: 15.5rem; max-width: calc(100vw - 1rem); max-height: 70vh; max-height: 70dvh; overflow-y: auto; padding: 0.75rem; background: var(--color-surface-2); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card); box-shadow: var(--shadow-card); }
.mp__pop--end { inset-inline-start: auto; inset-inline-end: 0; }
.mp__head { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
.mp__year { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.mp__nav { display: grid; place-items: center; width: 2rem; height: 2rem; color: var(--color-text-secondary); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-btn); cursor: pointer; }
.mp__nav:disabled { opacity: 0.35; cursor: not-allowed; }
.mp__grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.375rem; }
.mp__cell { min-height: 2.125rem; padding: 0 0.375rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); background: var(--color-surface-3); border: 1px solid transparent; border-radius: var(--radius-btn); cursor: pointer; transition: background-color var(--motion-fast) var(--ease-water), color var(--motion-fast) var(--ease-water); }
.mp__cell.is-active { font-weight: 600; color: var(--color-on-brand); background: var(--color-brand); }
.mp__cell:disabled { opacity: 0.35; cursor: not-allowed; }
.mp__all { min-height: 2rem; padding: 0 0.625rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); background: none; border: 1px solid var(--color-border); border-radius: var(--radius-btn); cursor: pointer; }
.mp__all.is-active { color: var(--color-text-primary); border-color: var(--color-border-strong); background: var(--color-surface-3); }
@media (hover: hover) {
  .mp__btn:hover, .mp__nav:not(:disabled):hover, .mp__all:hover { border-color: var(--color-border-strong); color: var(--color-text-primary); }
  .mp__cell:not(.is-active):not(:disabled):hover { color: var(--color-text-primary); border-color: var(--color-border-strong); }
}
@media (prefers-reduced-motion: reduce) {
  .mp__btn, .mp__chev, .mp__cell { transition: none; }
}
</style>
