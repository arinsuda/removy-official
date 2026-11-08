<template>
  <section class="mission" aria-labelledby="about-mission">
    <Heart :size="24" class="mission__icon" aria-hidden="true" />
    <p class="mission__label">{{ t("about.mission.label") }}</p>
    <h2 id="about-mission" class="mission__title">{{ t("about.mission.title") }}</h2>
    <p class="mission__text">{{ t("about.mission.paragraph1") }}</p>
    <p class="mission__text">{{ t("about.mission.paragraph2") }}</p>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Heart } from "lucide-vue-next"

const { t } = useI18n()
</script>

<style scoped>
.mission {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.mission__icon {
  color: var(--color-brand);
}
.mission__label {
  margin: 0;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-muted);
}
.mission__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.mission__text {
  max-width: 46rem;
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.7;
  color: var(--color-text-secondary);
}
@media (min-width: 768px) {
  .mission {
    padding: 2rem;
  }
}
</style>
