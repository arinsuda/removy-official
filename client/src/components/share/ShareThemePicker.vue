<template>
  <div class="picker" role="radiogroup" :aria-label="t('share.theme.label')" @keydown="onKeydown">
    <button
      v-for="theme in SHARE_THEMES"
      :key="theme"
      ref="buttons"
      type="button"
      role="radio"
      class="picker__btn"
      :class="`picker__btn--${theme}`"
      :aria-checked="model === theme"
      :tabindex="model === theme ? 0 : -1"
      :title="t(`share.theme.${theme}`)"
      @click="model = theme"
    >
      <span class="sr-only">{{ t(`share.theme.${theme}`) }}</span>
    </button>
    <span class="picker__name" aria-hidden="true">{{ t(`share.theme.${model}`) }}</span>
  </div>
</template>

<script setup lang="ts">
import { nextTick, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import type { ShareTheme } from "@/types/share"
import { SHARE_THEMES } from "./shareCards"

const model = defineModel<ShareTheme>({ required: true })
const { t } = useI18n()
const buttons = useTemplateRef<HTMLButtonElement[]>("buttons")

async function onKeydown(e: KeyboardEvent) {
  const i = SHARE_THEMES.indexOf(model.value)
  const step = e.key === "ArrowRight" || e.key === "ArrowDown" ? 1 : e.key === "ArrowLeft" || e.key === "ArrowUp" ? -1 : 0
  if (!step) return
  e.preventDefault()
  const next = (i + step + SHARE_THEMES.length) % SHARE_THEMES.length
  model.value = SHARE_THEMES[next]!
  await nextTick()
  buttons.value?.[next]?.focus()
}
</script>

<style scoped>
.picker { display: flex; align-items: center; gap: 12px; }
.picker__btn {
  width: 32px; height: 32px; border-radius: 50%; border: 2px solid var(--color-border-strong); cursor: pointer;
}
.picker__btn--dark { background: #0b0b0c; } /* hardcode-ok: swatch previews the fixed share-card theme background */
.picker__btn--light { background: #f6f3ee; } /* hardcode-ok: swatch previews the fixed share-card theme background */
.picker__btn--brand { background: var(--color-brand); }
.picker__btn[aria-checked="true"] { outline: 2px solid var(--color-text-primary); outline-offset: 2px; }
.picker__name { color: var(--color-text-secondary); font-size: 13px; font-weight: 600; }
</style>
