<template>
  <section class="about-hero" aria-labelledby="about-title">
    <p class="about-hero__eyebrow">{{ t("about.eyebrow") }}</p>
    <h1 id="about-title" class="about-hero__title">
      {{ t("about.title.line1") }}
      <span class="about-hero__accent">{{ t("about.title.line2") }}</span>
    </h1>
    <p class="about-hero__sub">{{ t("about.sub") }}</p>
    <AppButton :to="auth.isLoggedIn ? { name: 'home' } : { name: 'register' }" size="lg">
      {{ auth.isLoggedIn ? t("about.hero.cta.home") : t("about.hero.cta.register") }}
      <ArrowRight :size="16" aria-hidden="true" />
    </AppButton>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ArrowRight } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import { useAuthStore } from "@/stores/auth"

const { t } = useI18n()
const auth = useAuthStore()
</script>

<style scoped>
.about-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem 0 1rem;
  text-align: center;
}
.about-hero__eyebrow {
  margin: 0;
  padding: 4px 12px;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-secondary);
}
.about-hero__title {
  display: flex;
  flex-direction: column;
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.2;
  color: var(--color-text-primary);
}
.about-hero__accent {
  color: var(--color-brand);
}
.about-hero__sub {
  max-width: 36rem;
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.6;
  color: var(--color-text-secondary);
}
@media (min-width: 768px) {
  .about-hero {
    padding: 4rem 0 2rem;
  }
  .about-hero__title {
    font-size: var(--text-display);
  }
  .about-hero__sub {
    font-size: 1rem;
  }
}
</style>
