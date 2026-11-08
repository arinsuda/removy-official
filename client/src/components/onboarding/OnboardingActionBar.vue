<template>
  <div class="bar" :class="{ 'bar--ready': ready }">
    <div class="bar__inner">
      <p v-if="failed" class="bar__error" role="alert">{{ t("onboarding.saveError") }}</p>
      <div class="bar__row">
        <AppButton variant="ghost" :disabled="busy" class="bar__skip" @click="emit('skip')">
          {{ t("onboarding.skip") }}
        </AppButton>
        <span class="bar__go">
          <AppButton block size="lg" :disabled="!ready" :loading="busy" @click="emit('continue')">
          <Transition name="morph" mode="out-in">
            <span v-if="ready" key="go" class="bar__label">
              {{ t("onboarding.continue") }}
              <ArrowRight :size="18" aria-hidden="true" />
            </span>
            <span v-else key="wait" class="bar__label">{{ t("onboarding.pickMore", { n: remaining }) }}</span>
          </Transition>
          </AppButton>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ArrowRight } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"

defineProps<{ ready: boolean; busy: boolean; remaining: number; failed: boolean }>()
const emit = defineEmits<{ continue: []; skip: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.bar {
  position: sticky;
  bottom: 0;
  z-index: 5;
  margin: 0 -1rem;
  padding: 0.75rem 1rem calc(0.75rem + env(safe-area-inset-bottom));
  border-top: 1px solid var(--color-border);
  background: color-mix(in srgb, var(--color-surface) 82%, transparent);
  backdrop-filter: blur(16px);
}

.bar__inner {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 36rem;
  margin: 0 auto;
}

.bar__row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.bar__skip {
  min-height: 44px;
  flex-shrink: 0;
}

.bar__go {
  position: relative;
  flex: 1;
  border-radius: var(--radius-btn);
}

.bar__label {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.bar__error {
  margin: 0;
  font-size: 0.8125rem;
  text-align: center;
  color: var(--color-danger);
}

/* The label swaps with a short vertical morph; the button itself keeps its size. */
.morph-enter-active,
.morph-leave-active {
  transition:
    opacity var(--motion-fast) var(--ease-out-soft),
    transform var(--motion-fast) var(--ease-out-soft);
}
.morph-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.morph-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* One pulse when the minimum is reached. */
.bar--ready .bar__go::after {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-brand) 60%, transparent);
  animation: bar-ping 900ms var(--ease-out-soft) 1;
  pointer-events: none;
}

@keyframes bar-ping {
  to {
    box-shadow: 0 0 0 14px transparent;
  }
}

@media (prefers-reduced-motion: reduce) {
  .morph-enter-active,
  .morph-leave-active {
    transition-duration: 1ms;
  }
  .bar--ready .bar__go::after {
    animation: none;
  }
}
</style>
