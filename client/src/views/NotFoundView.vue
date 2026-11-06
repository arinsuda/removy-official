<template>
  <main class="not-found">
    <h1 class="not-found__title">{{ $t("common.notFound.heading") }}</h1>
    <p class="not-found__text">{{ $t("common.notFound.message") }}</p>
    <AppButton to="/" size="lg">{{ $t("common.notFound.home") }}</AppButton>
  </main>
</template>

<script setup lang="ts">
import AppButton from "@/components/ui/AppButton.vue"
</script>

<style scoped>
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  min-height: 100vh;
  padding: 24px 16px;
  text-align: center;
  color: var(--color-text-primary);
}

.not-found__title {
  font-size: 24px;
  font-weight: 800;
}

.not-found__text {
  max-width: 28rem;
  color: var(--color-text-secondary);
}

@media (min-width: 768px) {
  .not-found__title {
    font-size: 32px;
  }
}
</style>
