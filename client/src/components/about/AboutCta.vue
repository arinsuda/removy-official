<template>
  <section class="cta" aria-labelledby="about-cta">
    <div class="cta__card">
      <div class="cta__text">
        <h2 id="about-cta" class="cta__title">{{ t("about.cta.title") }}</h2>
        <p class="cta__desc">{{ t("about.cta.desc") }}</p>
      </div>
      <AppButton :to="{ name: 'donate' }">
        <Heart :size="16" aria-hidden="true" />
        {{ t("about.cta.btn") }}
      </AppButton>
    </div>
    <p class="cta__built">{{ t("about.builtWith") }}</p>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Heart } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"

const { t } = useI18n()
</script>

<style scoped>
.cta {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.cta__card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.5rem;
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.cta__title {
  margin: 0 0 0.25rem;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.cta__desc {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
.cta__built {
  margin: 0;
  text-align: center;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
@media (min-width: 640px) {
  .cta__card {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}
</style>
