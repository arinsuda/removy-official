<template>
  <SkeletonList v-if="loading" :count="3" />
  <p v-else-if="error" class="toggles__error" role="alert">{{ t("settings.activity.loadFailed") }}</p>
  <ul v-else-if="settings" class="toggles">
    <li v-for="opt in ACTIVITY_OPTIONS" :key="opt.key" class="toggles__row">
      <component :is="opt.icon" class="toggles__icon" :size="16" aria-hidden="true" />
      <span :id="`${uid}-${opt.key}`" class="toggles__label">{{ t(`settings.activity.options.${opt.key}`) }}</span>
      <button
        type="button"
        role="switch"
        class="switch"
        :class="{ 'is-on': settings[opt.key] }"
        :aria-checked="settings[opt.key]"
        :aria-labelledby="`${uid}-${opt.key}`"
        @click="toggle(opt.key)"
      >
        <span class="switch__knob" />
      </button>
    </li>
  </ul>
</template>

<script setup lang="ts">
import { onMounted, useId } from "vue"
import { useI18n } from "vue-i18n"
import { storeToRefs } from "pinia"
import { useFeedStore } from "@/stores/feed"
import { useToast } from "@/composables/useToast"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import type { ActivitySettingsResponse } from "@/types/feed"
import { ACTIVITY_OPTIONS } from "./activityOptions"

const { t } = useI18n()
const toast = useToast()
const uid = useId()
const store = useFeedStore()
const { settings, settingsLoading: loading, settingsError: error } = storeToRefs(store)

async function toggle(key: keyof ActivitySettingsResponse) {
  if (!settings.value) return
  const ok = await store.updateSettings({ [key]: !settings.value[key] })
  if (!ok) toast.error(t("settings.activity.saveFailed"))
}

onMounted(() => {
  if (!settings.value) store.fetchSettings()
})
</script>

<style scoped>
.toggles { display: flex; flex-direction: column; }
.toggles__row { display: flex; align-items: center; gap: 0.75rem; padding: 0.625rem 0; border-bottom: 1px solid var(--color-border); }
.toggles__row:last-child { border-bottom: 0; }
.toggles__icon { flex: none; color: var(--color-text-muted); }
.toggles__label { flex: 1; font-size: 0.875rem; color: var(--color-text-primary); }
.toggles__error { font-size: 0.8125rem; color: var(--color-danger); }
.switch { position: relative; flex: none; width: 2.75rem; height: 1.5rem; cursor: pointer; background: var(--color-surface-3); border: 1px solid var(--color-border-strong); border-radius: 999px; transition: background-color var(--motion-base, 0.2s); }
.switch.is-on { background: var(--color-brand); border-color: var(--color-brand); }
.switch__knob { position: absolute; top: 0.125rem; left: 0.125rem; width: 1.125rem; height: 1.125rem; background: var(--color-text-primary); border-radius: 50%; transition: transform var(--motion-base, 0.2s); }
.switch.is-on .switch__knob { transform: translateX(1.25rem); }
@media (prefers-reduced-motion: reduce) { .switch, .switch__knob { transition: none; } }
</style>
