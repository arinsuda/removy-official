<template>
  <Transition name="fade-up">
    <div v-if="dirty" class="bar" role="region" :aria-label="t('settings.saveBar.label')">
      <p class="bar__text">{{ error ? t("settings.saveBar.failed") : t("settings.saveBar.unsaved") }}</p>
      <div class="bar__actions">
        <AppButton variant="ghost" :disabled="saving" @click="emit('discard')">{{ t("settings.saveBar.discard") }}</AppButton>
        <AppButton variant="primary" :loading="saving" @click="emit('save')">{{ t("settings.saveBar.save") }}</AppButton>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"

defineProps<{ dirty: boolean; saving: boolean; error?: string }>()
const emit = defineEmits<{ save: []; discard: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.bar { position: sticky; bottom: calc(0.75rem + env(safe-area-inset-bottom)); z-index: var(--z-sticky); display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.5rem; padding: 0.75rem 1rem; background: var(--color-surface-3); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card); box-shadow: var(--shadow-raised); }
.bar__text { font-size: 0.875rem; color: var(--color-text-primary); }
.bar__actions { display: flex; gap: 0.5rem; margin-left: auto; }
</style>
