<template>
  <Teleport to="body">
    <Transition name="app-modal">
      <div
        v-if="modelValue"
        class="app-modal"
        :style="{ zIndex: `calc(var(--z-modal-base) + ${modalZOffset})` }"
        @mousedown="onBackdropDown"
        @mouseup="onBackdropUp"
      >
        <div
          ref="panel"
          v-bind="$attrs"
          class="app-modal__panel"
          :class="`app-modal__panel--${size}`"
          :role="role"
          aria-modal="true"
          :aria-label="ariaLabel"
          :aria-labelledby="ariaLabel ? undefined : hasTitleSlot ? titleId : ($attrs['aria-labelledby'] as string | undefined)"
          tabindex="-1"
          @keydown="onKeydown"
        >
          <header v-if="hasTitleSlot || showClose" class="app-modal__header">
            <div v-if="hasTitleSlot" :id="titleId" class="app-modal__title">
              <slot name="header">
                <h2 v-if="title">{{ title }}</h2>
              </slot>
            </div>
            <AppButton
              v-if="showClose"
              variant="ghost"
              size="sm"
              icon-only
              :disabled="dismissDisabled"
              :aria-label="t('common.close')"
              @click="dismiss('close-button')"
            >
              <X :size="18" aria-hidden="true" />
            </AppButton>
          </header>
          <div class="app-modal__body"><slot /></div>
          <footer v-if="$slots.footer" class="app-modal__footer"><slot name="footer" /></footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script lang="ts">
export type AppModalCloseReason = "escape" | "backdrop" | "close-button"

// Shared by every open modal: the stack decides z-order and which modal owns Escape.
const openModals: symbol[] = []
let nextZOrder = 0
const Z_MODAL_STEP = 10
</script>

<script setup lang="ts">
import AppButton from "@/components/ui/AppButton.vue"
import { lockScroll, unlockScroll } from "@/composables/scrollLock"
import { useDialogFocus } from "@/composables/useDialogFocus"
import { X } from "lucide-vue-next"
import { computed, onBeforeUnmount, ref, toRef, useAttrs, useId, useSlots, watch } from "vue"
import { useI18n } from "vue-i18n"

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    ariaLabel?: string
    size?: "sm" | "md" | "lg"
    role?: "dialog" | "alertdialog"
    closeOnBackdrop?: boolean
    /**
     * When embedding `@vuepic/vue-datepicker` inside the modal, its menu keeps
     * focus on the input inside the panel. Bind `:close-on-escape="!menuOpen"`
     * driven by the picker's `@open` / `@closed` events so Escape closes the
     * menu first instead of dismissing the modal.
     */
    closeOnEscape?: boolean
    /** Blocks every user dismissal (e.g. while a request is saving). */
    dismissDisabled?: boolean
    /** CSS selector or element to focus first instead of the first focusable. */
    initialFocus?: string | HTMLElement
    showClose?: boolean
  }>(),
  { size: "md", role: "dialog", closeOnBackdrop: true, closeOnEscape: true, showClose: true },
)

const emit = defineEmits<{
  "update:modelValue": [value: boolean]
  close: [reason: AppModalCloseReason]
}>()

const { t } = useI18n()
const slots = useSlots()
const attrs = useAttrs()
const titleId = useId()
const panel = ref<HTMLElement | null>(null)
const hasTitleSlot = computed(() => Boolean(props.title || slots.header))
const token = Symbol("app-modal")
const modalZOffset = ref(0)

function isTop(): boolean {
  const len = openModals.length
  return len === 0 || openModals[len - 1] === token
}

const { onKeydown } = useDialogFocus(panel, toRef(props, "modelValue"), toRef(props, "initialFocus"), isTop)

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      if (
        import.meta.env.DEV &&
        !props.title &&
        !props.ariaLabel &&
        !slots.header &&
        !attrs["aria-labelledby"]
      ) {
        console.warn("[AppModal] dialogs need an accessible name via title, ariaLabel, or the header slot.")
      }
      if (!openModals.includes(token)) {
        openModals.push(token)
        modalZOffset.value = nextZOrder++ * Z_MODAL_STEP
        lockScroll()
        document.addEventListener("keydown", onDocumentKeydown)
      }
    } else {
      release()
    }
  },
  { immediate: true },
)

function dismiss(reason: AppModalCloseReason) {
  if (!props.modelValue || props.dismissDisabled) return
  emit("close", reason)
  emit("update:modelValue", false)
}

function isInsideTeleportedMenu(target: Node | null): boolean {
  return target instanceof Element && Boolean(target.closest(".dp--menu-wrapper, [role='menu'], [role='listbox']"))
}

function onDocumentKeydown(event: KeyboardEvent) {
  if (event.defaultPrevented || !isTop()) return
  const target = event.target as Node | null
  if (event.key === "Tab") {
    if (!panel.value?.contains(target) && !isInsideTeleportedMenu(target)) {
      onKeydown(event)
    }
    return
  }
  if (event.key !== "Escape" || !props.closeOnEscape) return
  // Ignore Escape from elsewhere, such as a menu teleported to body.
  if (target !== document.body && !panel.value?.contains(target)) return
  event.preventDefault()
  dismiss("escape")
}

function release() {
  const index = openModals.indexOf(token)
  if (index === -1) return
  openModals.splice(index, 1)
  if (openModals.length === 0) nextZOrder = 0
  unlockScroll()
  document.removeEventListener("keydown", onDocumentKeydown)
}

onBeforeUnmount(release)

// A drag that starts inside the panel and ends on the backdrop must not close.
let pressedBackdrop = false

function onBackdropDown(event: MouseEvent) {
  if (event.button !== 0 || event.target !== event.currentTarget) {
    pressedBackdrop = false
    return
  }
  event.preventDefault()
  pressedBackdrop = true
}

function onBackdropUp(event: MouseEvent) {
  const clicked = pressedBackdrop && event.button === 0 && event.target === event.currentTarget
  pressedBackdrop = false
  if (clicked && props.closeOnBackdrop) dismiss("backdrop")
}
</script>

<style scoped>
.app-modal {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal-base);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: color-mix(in srgb, black 72%, transparent);
  backdrop-filter: blur(6px);
}

.app-modal__panel {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-height: calc(100dvh - 32px);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  color: var(--color-text-primary);
  outline: none;
}

.app-modal__panel--sm {
  max-width: 400px;
}

.app-modal__panel--md {
  max-width: 520px;
}

.app-modal__panel--lg {
  max-width: 760px;
}

.app-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 16px 0 20px;
}

.app-modal__header:not(:has(.app-modal__title)) {
  justify-content: flex-end;
}

.app-modal__title {
  flex: 1;
  min-width: 0;
  font-size: 18px;
  font-weight: 800;
}

.app-modal__body {
  flex: 1;
  min-height: 0;
  padding: 16px 20px;
  overflow-y: auto;
}

.app-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 0 20px 16px;
}

.app-modal-enter-active,
.app-modal-leave-active {
  transition: opacity var(--motion-fast) var(--ease-out-soft);
}

.app-modal-enter-active .app-modal__panel,
.app-modal-leave-active .app-modal__panel {
  transition: transform var(--motion-fast) var(--ease-out-soft);
}

.app-modal-enter-from,
.app-modal-leave-to {
  opacity: 0;
}

.app-modal-enter-from .app-modal__panel,
.app-modal-leave-to .app-modal__panel {
  transform: translateY(8px);
}

@media (prefers-reduced-motion: reduce) {
  .app-modal-enter-active,
  .app-modal-leave-active,
  .app-modal-enter-active .app-modal__panel,
  .app-modal-leave-active .app-modal__panel {
    transition: none;
  }
}
</style>
