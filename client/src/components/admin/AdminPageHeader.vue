<template>
  <header class="admin-page-header">
    <div>
      <h1 class="admin-page-header__title">{{ title }}</h1>
      <p v-if="subtitle" class="admin-page-header__subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.default" class="admin-page-header__actions"><slot /></div>
  </header>
</template>

<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>()
</script>

<style scoped>
.admin-page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-page-header__title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.3;
  color: var(--color-text-primary);
}

.admin-page-header__subtitle {
  margin: 0.125rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
</style>
