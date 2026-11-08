<template>
  <section class="app-card">
    <header v-if="$slots.header" class="app-card__header"><slot name="header" /></header>
    <div class="app-card__body"><slot /></div>
    <footer v-if="$slots.footer" class="app-card__footer"><slot name="footer" /></footer>
  </section>
</template>

<style scoped>
.app-card {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow: hidden;
}
.app-card__header,
.app-card__footer {
  padding: 0.875rem 1.25rem;
}
.app-card__header {
  border-bottom: 1px solid var(--color-border);
}
.app-card__footer {
  border-top: 1px solid var(--color-border);
}
.app-card__body {
  padding: 1.25rem;
}
</style>
