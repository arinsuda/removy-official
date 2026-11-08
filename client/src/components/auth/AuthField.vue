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
        :aria-describedby="error ? errorId : undefined"
      />
      <button
        v-if="isPassword"
        type="button"
        class="field__toggle"
        :aria-label="revealed ? 'Hide password' : 'Show password'"
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
    <Transition name="field-error">
      <p v-if="error" :id="errorId" class="field__error">
        <IconAlertCircle :size="12" />
        {{ error }}
      </p>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useId, type Component } from "vue"
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
    valid?: boolean
  }>(),
  { type: "text", error: "", valid: false },
)

const model = defineModel<string>({ required: true })

const inputId = useId()
const errorId = `${inputId}-error`
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
  color: #d4d4d4;
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
  color: #737373;
  pointer-events: none;
  transition: color var(--motion-base) var(--ease-out-soft);
}

.field__check {
  right: 14px;
  color: #22c55e;
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
  color: #8a8a8a;
  cursor: pointer;
  transition: color var(--motion-fast) var(--ease-out-soft);
}

.field__input {
  width: 100%;
  min-height: 46px;
  padding: 0 2.75rem;
  box-sizing: border-box;
  background: rgb(255 255 255 / 5%);
  border: 1px solid rgb(255 255 255 / 12%);
  border-radius: 12px;
  outline: none;
  font-size: 0.9375rem;
  color: #fff;
  transition:
    background var(--motion-base) var(--ease-out-soft),
    border-color var(--motion-base) var(--ease-out-soft),
    box-shadow var(--motion-base) var(--ease-out-soft);
}

/* Keep browser autofill on the dark surface. */
.field__input:-webkit-autofill {
  -webkit-text-fill-color: #fff;
  caret-color: #fff;
  transition: background-color 9999s;
}

.field__input:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.field__input:focus {
  background: rgb(255 255 255 / 8%);
  border-color: transparent;
  box-shadow: var(--focus-input-shadow);
}

.field__control:focus-within .field__icon {
  color: var(--color-brand);
}

.is-valid .field__input:not(:focus) {
  border-color: rgb(34 197 94 / 55%);
}

.is-error .field__input:not(:focus) {
  background: rgb(245 67 44 / 7%);
  border-color: rgb(248 113 113 / 75%);
}

.field__error {
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 0;
  font-size: 0.75rem;
  color: #f87171;
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
    color: #fff;
  }

  .field__input:hover:not(:focus, :disabled) {
    border-color: rgb(255 255 255 / 24%);
  }
}
</style>
