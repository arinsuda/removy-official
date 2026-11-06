<template>
  <Transition name="auth-alert">
    <div v-if="message" class="auth-alert" role="alert">
      <IconAlertCircle :size="16" class="auth-alert__icon" />
      <span class="auth-alert__text">{{ message }}</span>
      <button class="auth-alert__close" type="button" aria-label="Dismiss" @click="emit('dismiss')">
        <IconX :size="14" />
      </button>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { AlertCircle as IconAlertCircle, X as IconX } from "lucide-vue-next"

defineProps<{ message: string }>()
const emit = defineEmits<{ dismiss: [] }>()
</script>

<style scoped>
.auth-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0.625rem 0.75rem;
  background: rgb(245 67 44 / 12%);
  border: 1px solid rgb(245 67 44 / 35%);
  border-radius: 12px;
  font-size: 0.8125rem;
  line-height: 1.4;
  color: #fca5a5;
}

.auth-alert__icon {
  flex-shrink: 0;
  color: var(--color-brand);
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
