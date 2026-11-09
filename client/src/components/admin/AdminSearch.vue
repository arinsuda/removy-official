<template>
  <div class="admin-search">
    <label :for="id" class="sr-only">{{ label }}</label>
    <Search :size="16" class="admin-search__icon" aria-hidden="true" />
    <input :id="id" v-model="model" type="search" class="admin-search__input" :placeholder="placeholder" autocomplete="off" />
  </div>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { Search } from "lucide-vue-next"

defineProps<{ label: string; placeholder: string }>()
const model = defineModel<string>({ required: true })
const id = useId()
</script>

<style scoped>
.admin-search {
  position: relative;
  flex: 1 1 16rem;
  min-width: 0;
}

.admin-search__icon {
  position: absolute;
  top: 50%;
  left: 0.875rem;
  color: var(--color-text-muted);
  transform: translateY(-50%);
  pointer-events: none;
}

.admin-search__input {
  width: 100%;
  min-height: 2.5rem;
  padding: 0 0.875rem 0 2.5rem;
  box-sizing: border-box;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  font-size: var(--text-small);
  color: var(--color-text-primary);
}

.admin-search__input::placeholder {
  color: var(--color-text-muted);
}
</style>
