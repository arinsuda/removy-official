<template>
  <div class="app-input">
    <label :for="inputId" class="app-input__label">{{ label }}</label>
    <input
      v-bind="$attrs"
      :id="inputId"
      v-model="model"
      :type="type"
      class="app-input__field"
      :class="{ 'is-error': !!error }"
      :disabled="disabled"
      :aria-invalid="!!error || undefined"
      :aria-describedby="describedBy"
    />
    <p v-if="error" :id="errorId" class="app-input__error" role="alert">{{ error }}</p>
    <p v-else-if="hint" :id="hintId" class="app-input__hint">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    label: string
    type?: string
    hint?: string
    error?: string
    disabled?: boolean
  }>(),
  { type: "text", hint: "", error: "", disabled: false },
)

const model = defineModel<string>({ required: true })

// The global focus ring (motion.css) applies: no data-focus-custom here.
const inputId = useId()
const errorId = `${inputId}-error`
const hintId = `${inputId}-hint`
const describedBy = computed(() => (props.error ? errorId : props.hint ? hintId : undefined))
</script>

<style scoped>
.app-input {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.app-input__label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.app-input__field {
  width: 100%;
  min-height: 46px;
  padding: 0 14px;
  box-sizing: border-box;
  background: var(--color-surface-3);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-btn);
  font-size: 0.9375rem;
  color: var(--color-text-primary);
  transition: border-color var(--motion-fast) var(--ease-out-soft);
}

.app-input__field::placeholder {
  color: var(--color-text-muted);
}

.app-input__field:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.app-input__field.is-error {
  border-color: var(--color-danger);
}

.app-input__error,
.app-input__hint {
  margin: 0;
  font-size: 0.75rem;
}

.app-input__error {
  color: var(--color-danger);
}

.app-input__hint {
  color: var(--color-text-muted);
}

@media (hover: hover) {
  .app-input__field:hover:not(:focus, :disabled) {
    border-color: var(--color-text-muted);
  }
}
</style>
