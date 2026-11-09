<template>
  <div class="admin-select">
    <label :for="id" class="sr-only">{{ label }}</label>
    <select :id="id" v-model="model" class="admin-select__input">
      <option v-for="o in options" :key="o.value" :value="o.value">{{ o.label }}</option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { useId } from "vue"

defineProps<{ label: string; options: { value: string; label: string }[] }>()
const model = defineModel<string>({ required: true })
const id = useId()
</script>

<style scoped>
.admin-select__input {
  min-height: 2.5rem;
  max-width: 100%;
  padding: 0 0.75rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  font-size: var(--text-small);
  color: var(--color-text-primary);
  cursor: pointer;
}
</style>
