<template>
  <div class="app-tabs" role="tablist" :aria-label="label" @keydown="onKeydown">
    <button
      v-for="tab in tabs"
      :id="tabId(tab.key)"
      :key="tab.key"
      ref="buttons"
      type="button"
      role="tab"
      class="app-tabs__tab"
      :class="{ 'is-active': model === tab.key }"
      :aria-selected="model === tab.key"
      :aria-controls="controls ? panelId(tab.key) : undefined"
      :tabindex="model === tab.key ? 0 : -1"
      :disabled="tab.disabled"
      @click="model = tab.key"
    >
      <slot name="tab" :tab="tab">{{ tab.label }}</slot>
    </button>
  </div>
</template>

<script setup lang="ts" generic="K extends string = string">
import { nextTick, useId, useTemplateRef } from "vue"

interface AppTab<T extends string> {
  key: T
  label: string
  disabled?: boolean
}

const props = withDefaults(
  defineProps<{
    tabs: AppTab<K>[]
    label?: string
    /** Set false when the tabs filter one region and have no dedicated panels. */
    controls?: boolean
  }>(),
  { label: undefined, controls: true },
)

const model = defineModel<K>({ required: true })

// Panels should use id `${id}-panel-${key}` (see panelId) and aria-labelledby the tab id.
const baseId = useId()
const tabId = (key: K) => `${baseId}-tab-${key}`
const panelId = (key: K) => `${baseId}-panel-${key}`
defineExpose({ tabId, panelId })

const buttons = useTemplateRef<HTMLButtonElement[]>("buttons")

async function select(index: number) {
  const tab = props.tabs[index]
  if (!tab) return
  model.value = tab.key
  await nextTick()
  buttons.value?.[index]?.focus()
}

function onKeydown(e: KeyboardEvent) {
  const enabled = props.tabs.map((t, i) => (t.disabled ? -1 : i)).filter((i) => i >= 0)
  if (!enabled.length) return
  const current = Math.max(0, props.tabs.findIndex((t) => t.key === model.value))
  const pos = enabled.indexOf(current)
  let next: number | undefined
  if (e.key === "ArrowRight" || e.key === "ArrowDown") next = enabled[(pos + 1) % enabled.length]
  else if (e.key === "ArrowLeft" || e.key === "ArrowUp") next = enabled[(pos - 1 + enabled.length) % enabled.length]
  else if (e.key === "Home") next = enabled[0]
  else if (e.key === "End") next = enabled[enabled.length - 1]
  if (next === undefined) return
  e.preventDefault()
  void select(next)
}
</script>

<style scoped>
.app-tabs {
  display: flex;
  gap: 0.375rem;
  max-width: 100%;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  scrollbar-width: none;
}
.app-tabs::-webkit-scrollbar {
  display: none;
}
.app-tabs__tab {
  font-family: var(--font-sans);
  flex: none;
  font-size: var(--text-caption);
  font-weight: 500;
  color: var(--color-text-secondary);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  padding: 0.375rem 0.875rem;
  border-radius: var(--radius-full);
  white-space: nowrap;
  cursor: pointer;
  transition: color var(--motion-fast) var(--ease-water), border-color var(--motion-fast) var(--ease-water), background-color var(--motion-fast) var(--ease-water);
}
.app-tabs__tab:hover:not(:disabled) {
  color: var(--color-text-primary);
  border-color: var(--color-border-strong);
}
.app-tabs__tab.is-active {
  color: var(--color-surface);
  background: var(--color-text-primary);
  border-color: var(--color-text-primary);
}
.app-tabs__tab.is-active:hover:not(:disabled) {
  color: var(--color-surface);
  border-color: var(--color-text-primary);
}
.app-tabs__tab:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
