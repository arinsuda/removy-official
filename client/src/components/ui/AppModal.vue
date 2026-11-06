<template>
  <Teleport to="body">
    <Transition name="app-modal">
      <div
        v-if="modelValue"
        class="app-modal"
        @mousedown="onBackdropDown"
        @mouseup="onBackdropUp"
      >
        <div
          ref="panel"
          class="app-modal__panel"
          :class="`app-modal__panel--${size}`"
          :role="role"
          aria-modal="true"
          :aria-label="ariaLabel"
          :aria-labelledby="ariaLabel ? undefined : titleId"
          tabindex="-1"
          @keydown="onKeydown"
        >
          <header v-if="title || $slots.header || showClose" class="app-modal__header">
            <div :id="titleId" class="app-modal__title">
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
// Shared by every open modal: the stack decides which one owns Escape, and the
// counter keeps the page locked until the last modal closes.
const openModals: symbol[] = []
let scrollLocks = 0
let previousOverflow = ""

function lockScroll() {
  if (scrollLocks++ === 0) {
    previousOverflow = document.body.style.overflow
    document.body.style.overflow = "hidden"
  }
}

function unlockScroll() {
  if (scrollLocks > 0 && --scrollLocks === 0) document.body.style.overflow = previousOverflow
}
</script>

<script setup lang="ts">
import AppButton from "@/components/ui/AppButton.vue"
import { useDialogFocus } from "@/composables/useDialogFocus"
import { X } from "lucide-vue-next"
import { computed, onBeforeUnmount, ref, toRef, useId, watch } from "vue"
import { useI18n } from "vue-i18n"

export type AppModalCloseReason = "escape" | "backdrop" | "close-button"

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    ariaLabel?: string
    size?: "sm" | "md" | "lg"
    role?: "dialog" | "alertdialog"
    closeOnBackdrop?: boolean
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
const titleId = useId()
const panel = ref<HTMLElement | null>(null)
const { onKeydown } = useDialogFocus(panel, toRef(props, "modelValue"), toRef(props, "initialFocus"))

function dismiss(reason: AppModalCloseReason) {
  if (props.dismissDisabled) return
  emit("close", reason)
  emit("update:modelValue", false)
}

const token = Symbol("app-modal")
const isTop = computed(() => openModals[openModals.length - 1] === token)

function onDocumentKeydown(event: KeyboardEvent) {
  if (event.key !== "Escape" || !props.closeOnEscape || !isTop.value) return
  // Ignore Escape from elsewhere, such as a date menu teleported to body.
  const target = event.target as Node | null
  if (target !== document.body && !panel.value?.contains(target)) return
  event.preventDefault()
  dismiss("escape")
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      openModals.push(token)
      lockScroll()
      document.addEventListener("keydown", onDocumentKeydown)
    } else {
      release()
    }
  },
  { immediate: true },
)

function release() {
  const index = openModals.indexOf(token)
  if (index === -1) return
  openModals.splice(index, 1)
  unlockScroll()
  document.removeEventListener("keydown", onDocumentKeydown)
}

onBeforeUnmount(release)

// A drag that starts inside the panel and ends on the backdrop must not close.
let pressedBackdrop = false

function onBackdropDown(event: MouseEvent) {
  pressedBackdrop = event.target === event.currentTarget
}

function onBackdropUp(event: MouseEvent) {
  const clicked = pressedBackdrop && event.target === event.currentTarget
  pressedBackdrop = false
  if (clicked && props.closeOnBackdrop) dismiss("backdrop")
}
</script>

<style scoped>
.app-modal {
  position: fixed;
  inset: 0;
  z-index: 9999;
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
