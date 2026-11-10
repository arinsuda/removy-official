<template>
  <div v-if="variant === 'segmented'" class="lt-seg" role="group" :aria-label="t('navigation.header.language')">
    <button
      v-for="opt in OPTIONS"
      :key="opt.code"
      type="button"
      class="lt-opt"
      :class="{ 'lt-opt--on': locale === opt.code }"
      :aria-pressed="locale === opt.code"
      :disabled="isSwitchingLocale"
      @click="select(opt.code)"
    >
      {{ opt.label }}
    </button>
  </div>
  <button
    v-else
    type="button"
    class="lt-icon"
    :disabled="isSwitchingLocale"
    :aria-label="t('navigation.switchLang')"
    :title="t('navigation.switchLang')"
    @click="toggleLocale()"
  >
    <Globe :size="16" aria-hidden="true" />
    <span>{{ locale === "th" ? "EN" : "TH" }}</span>
  </button>
</template>

<script setup lang="ts">
import { Globe } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import { useLocale } from "@/i18n"

withDefaults(defineProps<{ variant?: "icon" | "segmented" }>(), { variant: "icon" })

const { t } = useI18n()
const { locale, isSwitchingLocale, toggleLocale } = useLocale()

const OPTIONS = [
  // hardcode-ok: a language is always named in its own language (endonym)
  { code: "th", label: "ไทย" },
  { code: "en", label: "English" },
] as const

function select(code: string) {
  if (code !== locale.value) toggleLocale()
}
</script>

<style scoped>
.lt-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  height: 2.5rem;
  min-width: 2.5rem;
  padding: 0 0.6rem;
  border: none;
  border-radius: 9999px;
  background: none;
  color: var(--color-text-secondary);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  cursor: pointer;
  transition: color var(--motion-fast), background var(--motion-fast);
}
.lt-icon:disabled {
  opacity: 0.6;
  cursor: progress;
}
@media (hover: hover) {
  .lt-icon:hover:not(:disabled) {
    color: var(--color-text-primary);
    background: color-mix(in srgb, var(--color-text-primary) 8%, transparent);
  }
}
.lt-seg {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.25rem;
  padding: 0.25rem;
  border-radius: 0.75rem;
  background: var(--color-surface-3);
}
.lt-opt {
  min-height: 2.75rem;
  border: none;
  border-radius: 0.5rem;
  background: none;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: background var(--motion-fast), color var(--motion-fast);
}
.lt-opt--on {
  background: var(--color-surface);
  color: var(--color-text-primary);
  box-shadow: 0 0 0 1px var(--color-border-strong);
}
</style>
