<template>
  <section class="genre-cta" :aria-labelledby="titleId">
    <h2 :id="titleId" class="genre-cta__title">{{ t("home.recommended.ctaTitle") }}</h2>
    <p class="genre-cta__desc">{{ t("home.recommended.ctaDescription") }}</p>
    <AppButton :to="{ name: 'onboarding' }" size="sm">{{ t("home.recommended.ctaAction") }}</AppButton>
  </section>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"

const { t } = useI18n()
const titleId = useId()
</script>

<style scoped>
.genre-cta {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  padding: 1.25rem;
  border: 1px dashed var(--color-border-strong);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.genre-cta__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.genre-cta__desc {
  margin: 0 0 4px;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
</style>
