<template>
  <button
    ref="trigger"
    type="button"
    class="row-menu__trigger"
    aria-haspopup="menu"
    :aria-expanded="open"
    :aria-label="label"
    @click="toggle"
    @keydown.down.prevent="openMenu"
  >
    <MoreHorizontal :size="18" aria-hidden="true" />
  </button>
  <Teleport to="body">
    <div v-if="open" ref="menu" class="row-menu" role="menu" :aria-label="label" :style="position" @keydown="onKeydown">
      <button
        v-for="item in items"
        :key="item.key"
        type="button"
        role="menuitem"
        class="row-menu__item"
        :class="item.tone && `row-menu__item--${item.tone}`"
        :disabled="item.disabled"
        :title="item.disabled ? item.hint : undefined"
        @click="choose(item.key)"
      >
        <component :is="item.icon" v-if="item.icon" :size="16" aria-hidden="true" />
        {{ item.label }}
      </button>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, type Component } from "vue"
import { MoreHorizontal } from "lucide-vue-next"

export interface RowMenuItem {
  key: string
  label: string
  icon?: Component
  tone?: "danger" | "warning"
  disabled?: boolean
  /** Why the item is disabled; shown as a tooltip. */
  hint?: string
}

defineProps<{ label: string; items: RowMenuItem[] }>()
const emit = defineEmits<{ select: [key: string] }>()

const open = ref(false)
const trigger = ref<HTMLButtonElement | null>(null)
const menu = ref<HTMLElement | null>(null)
const position = ref<Record<string, string>>({})

const enabledItems = () => Array.from(menu.value?.querySelectorAll<HTMLButtonElement>("button:not(:disabled)") ?? [])

function place() {
  const t = trigger.value?.getBoundingClientRect()
  if (!t) return
  const height = menu.value?.offsetHeight ?? 0
  const below = t.bottom + 4
  const top = below + height > window.innerHeight && t.top - 4 - height > 0 ? t.top - 4 - height : below
  position.value = { top: `${top}px`, right: `${Math.max(8, window.innerWidth - t.right)}px` }
}

async function openMenu() {
  open.value = true
  place()
  await nextTick()
  place()
  enabledItems()[0]?.focus()
  document.addEventListener("pointerdown", onOutside, true)
  window.addEventListener("resize", dismiss)
  window.addEventListener("scroll", place, true)
}

const dismiss = () => close()

function close(refocus = false) {
  if (!open.value) return
  open.value = false
  document.removeEventListener("pointerdown", onOutside, true)
  window.removeEventListener("resize", dismiss)
  window.removeEventListener("scroll", place, true)
  if (refocus) void nextTick(() => trigger.value?.focus())
}

function toggle() {
  if (open.value) close()
  else void openMenu()
}

function onOutside(e: Event) {
  const target = e.target as Node
  if (!menu.value?.contains(target) && !trigger.value?.contains(target)) close()
}

function choose(key: string) {
  close(true)
  emit("select", key)
}

function onKeydown(e: KeyboardEvent) {
  const els = enabledItems()
  const i = els.indexOf(document.activeElement as HTMLButtonElement)
  if (e.key === "Escape") {
    e.preventDefault()
    e.stopPropagation()
    close(true)
  } else if (e.key === "ArrowDown") {
    e.preventDefault()
    els[(i + 1) % els.length]?.focus()
  } else if (e.key === "ArrowUp") {
    e.preventDefault()
    els[(i - 1 + els.length) % els.length]?.focus()
  } else if (e.key === "Home") {
    e.preventDefault()
    els[0]?.focus()
  } else if (e.key === "End") {
    e.preventDefault()
    els[els.length - 1]?.focus()
  } else if (e.key === "Tab") {
    close()
  }
}

onBeforeUnmount(() => close())
</script>

<style scoped>
.row-menu__trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0;
  background: none;
  border: 1px solid transparent;
  border-radius: var(--radius-btn);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition:
    background var(--motion-fast) var(--ease-out-soft),
    color var(--motion-fast) var(--ease-out-soft);
}

.row-menu__trigger[aria-expanded="true"] {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}

.row-menu {
  position: fixed;
  z-index: var(--z-popover);
  min-width: 13rem;
  padding: 0.25rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  box-shadow: var(--shadow-popover);
}

.row-menu__item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  width: 100%;
  padding: 0.625rem 0.75rem;
  background: none;
  border: 0;
  border-radius: 0.5rem;
  font-size: var(--text-small);
  text-align: left;
  color: var(--color-text-primary);
  cursor: pointer;
}

.row-menu__item--warning {
  color: var(--color-warning);
}

.row-menu__item--danger {
  color: var(--color-danger);
}

.row-menu__item:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.row-menu__item:focus-visible,
.row-menu__item:hover:not(:disabled) {
  background: var(--color-surface-3);
}
</style>
