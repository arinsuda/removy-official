<template>
  <SettingsCard id="privacy" :title="t('settings.privacy.title')">
    <div class="row">
      <div>
        <p :id="labelId" class="row__label">{{ t("settings.privacy.private") }}</p>
        <p :id="descId" class="row__desc">{{ t("settings.privacy.privateDesc") }}</p>
      </div>
      <button
        type="button"
        role="switch"
        class="switch"
        :class="{ 'is-on': model }"
        :aria-checked="model"
        :aria-labelledby="labelId"
        :aria-describedby="descId"
        @click="model = !model"
      >
        <span class="switch__knob" />
      </button>
    </div>
    <p class="hidden-title">{{ t("settings.privacy.hiddenTitle") }}</p>
    <ul class="hidden-list">
      <li v-for="key in hidden" :key="key">{{ t(`settings.privacy.hidden.${key}`) }}</li>
    </ul>
    <p class="note">{{ t("settings.privacy.visibleNote") }}</p>
  </SettingsCard>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import SettingsCard from "./SettingsCard.vue"

const model = defineModel<boolean>({ required: true })
const { t } = useI18n()
const labelId = useId()
const descId = useId()
const hidden = ["library", "taste", "bmol", "achievements", "lists"] as const
</script>

<style scoped>
.row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
.row__label { font-size: 0.9375rem; font-weight: 600; color: var(--color-text-primary); }
.row__desc, .note { font-size: var(--text-caption); color: var(--color-text-muted); }
.switch { position: relative; flex: none; width: 2.75rem; height: 1.5rem; cursor: pointer; background: var(--color-surface-3); border: 1px solid var(--color-border-strong); border-radius: 999px; transition: background-color var(--motion-base, 0.2s); }
.switch.is-on { background: var(--color-brand); border-color: var(--color-brand); }
.switch__knob { position: absolute; top: 0.125rem; left: 0.125rem; width: 1.125rem; height: 1.125rem; background: var(--color-text-primary); border-radius: 50%; transition: transform var(--motion-base, 0.2s); }
.switch.is-on .switch__knob { transform: translateX(1.25rem); }
.hidden-title { margin-top: 1rem; font-size: 0.8125rem; font-weight: 600; color: var(--color-text-secondary); }
.hidden-list { margin: 0.375rem 0 0.75rem; padding-left: 1.125rem; font-size: 0.8125rem; color: var(--color-text-secondary); list-style: disc; }
@media (prefers-reduced-motion: reduce) { .switch, .switch__knob { transition: none; } }
</style>
