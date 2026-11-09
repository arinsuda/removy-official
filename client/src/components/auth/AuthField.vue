<template>
  <div class="field form-item">
    <label :for="inputId" class="field__label">{{ label }}</label>
    <div class="field__control" :class="{ 'is-error': !!error, 'is-valid': valid }">
      <span class="field__icon" aria-hidden="true"><component :is="icon" :size="16" /></span>
      <input
        v-bind="$attrs"
        :id="inputId"
        v-model="model"
        :type="inputType"
        class="field__input"
        data-focus-custom
        :aria-invalid="!!error || undefined"
        :aria-describedby="describedBy"
      />
      <button
        v-if="isPassword"
        type="button"
        class="field__toggle"
        :aria-label="revealed ? t('auth.common.hidePassword') : t('auth.common.showPassword')"
        :aria-pressed="revealed"
        @click="revealed = !revealed"
      >
        <IconEyeOff v-if="revealed" :size="16" />
        <IconEye v-else :size="16" />
      </button>
      <span v-else-if="valid" class="field__check" aria-hidden="true">
        <IconCheck :size="15" />
      </span>
    </div>
    <slot />
    <p v-if="hint && !error" :id="hintId" class="field__hint">{{ hint }}</p>
    <Transition name="field-error">
      <p v-if="error" :id="errorId" class="field__error" role="alert">
        <IconAlertCircle :size="12" />
        {{ error }}
      </p>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useId, type Component } from "vue"
import { useI18n } from "vue-i18n"
import {
  AlertCircle as IconAlertCircle,
  Check as IconCheck,
  Eye as IconEye,
  EyeOff as IconEyeOff,
} from "lucide-vue-next"

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    label: string
    icon: Component
    type?: string
    error?: string
    hint?: string
    valid?: boolean
  }>(),
  { type: "text", error: "", hint: "", valid: false },
)

const model = defineModel<string>({ required: true })

const { t } = useI18n()
const inputId = useId()
const errorId = `${inputId}-error`
const hintId = `${inputId}-hint`
const describedBy = computed(() => (props.error ? errorId : props.hint ? hintId : undefined))
const revealed = ref(false)
const isPassword = computed(() => props.type === "password")
const inputType = computed(() => (isPassword.value && revealed.value ? "text" : props.type))
</script>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.field__label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.field__control {
  position: relative;
  display: flex;
  align-items: center;
}

.field__icon,
.field__check,
.field__toggle {
  position: absolute;
  display: flex;
  align-items: center;
  justify-content: center;
}

.field__icon {
  left: 14px;
  color: var(--color-text-muted);
  pointer-events: none;
  transition: color var(--motion-base) var(--ease-out-soft);
}

.field__check {
  right: 14px;
  color: var(--color-success);
  pointer-events: none;
}

.field__toggle {
  right: 6px;
  width: 34px;
  height: 34px;
  padding: 0;
  background: none;
  border: 0;
  border-radius: 50%;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color var(--motion-fast) var(--ease-out-soft);
}

.field__input {
  width: 100%;
  min-height: 48px;
  padding: 0 2.75rem;
  box-sizing: border-box;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  outline: none;
  font-size: 0.9375rem;
  color: var(--color-text-primary);
  transition:
    background var(--motion-base) var(--ease-out-soft),
    border-color var(--motion-base) var(--ease-out-soft),
    box-shadow var(--motion-base) var(--ease-out-soft);
}

/* Keep browser autofill on the dark surface. */
.field__input:-webkit-autofill {
  -webkit-text-fill-color: var(--color-text-primary);
  caret-color: var(--color-text-primary);
  transition: background-color 9999s;
}

.field__input:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.field__input:focus {
  background: var(--color-surface-3);
  border-color: transparent;
  box-shadow: var(--focus-input-shadow);
}

.field__control:focus-within .field__icon {
  color: var(--color-brand);
}

.is-valid .field__input:not(:focus) {
  border-color: color-mix(in srgb, var(--color-success) 60%, transparent);
}

.is-error .field__input:not(:focus) {
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  border-color: color-mix(in srgb, var(--color-danger) 75%, transparent);
}

.field__hint {
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.field__error {
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 0;
  font-size: 0.75rem;
  color: var(--color-danger);
}

.field-error-enter-active,
.field-error-leave-active {
  transition:
    opacity var(--motion-fast) var(--ease-out-soft),
    transform var(--motion-fast) var(--ease-out-soft);
}

.field-error-enter-from,
.field-error-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

@media (hover: hover) {
  .field__toggle:hover {
    color: var(--color-text-primary);
  }

  .field__input:hover:not(:focus, :disabled) {
    border-color: color-mix(in srgb, var(--color-text-primary) 28%, transparent);
  }
}
</style>
