<template>
  <Transition name="auth-alert">
    <div v-if="message" class="auth-alert" role="alert">
      <IconAlertCircle :size="16" class="auth-alert__icon" aria-hidden="true" />
      <span class="auth-alert__text">{{ message }}</span>
      <button class="auth-alert__close touch-target" type="button" :aria-label="t('auth.common.dismiss')" @click="emit('dismiss')">
        <IconX :size="14" />
      </button>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { AlertCircle as IconAlertCircle, X as IconX } from "lucide-vue-next"

defineProps<{ message: string }>()
const { t } = useI18n()
const emit = defineEmits<{ dismiss: [] }>()
</script>

<style scoped>
.auth-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0.625rem 0.75rem;
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-danger) 40%, transparent);
  border-radius: var(--radius-card);
  font-size: 0.8125rem;
  line-height: 1.4;
  color: var(--color-text-primary);
}

.auth-alert__icon {
  flex-shrink: 0;
  color: var(--color-danger);
}

.auth-alert__text {
  flex: 1;
}

.auth-alert__close {
  display: flex;
  flex-shrink: 0;
  padding: 2px;
  background: none;
  border: 0;
  border-radius: 4px;
  color: inherit;
  cursor: pointer;
}

.auth-alert-enter-active,
.auth-alert-leave-active {
  transition:
    opacity var(--motion-base) var(--ease-out-soft),
    transform var(--motion-base) var(--ease-out-soft);
}

.auth-alert-enter-from,
.auth-alert-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
