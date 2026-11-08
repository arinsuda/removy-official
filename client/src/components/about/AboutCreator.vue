<template>
  <section class="creator" aria-labelledby="about-creator">
    <p class="creator__label">{{ t("about.creator.label") }}</p>
    <div class="creator__card">
      <AppAvatar src="/arin.png" :name="t('about.creator.avatarAlt')" size="xl" />
      <div class="creator__body">
        <p class="creator__role">{{ t("about.creator.role") }}</p>
        <h2 id="about-creator" class="creator__name">{{ t("about.creator.name") }}</h2>
        <p class="creator__bio">{{ t("about.creator.bio") }}</p>
        <ul class="creator__links">
          <li v-for="link in links" :key="link.key">
            <a :href="link.href" class="creator__link" v-bind="link.external ? { target: '_blank', rel: 'noopener noreferrer' } : {}">
              <component :is="link.icon" :size="16" aria-hidden="true" />
              {{ t(`about.creator.links.${link.key}`) }}
              <span v-if="link.external" class="sr-only">({{ t("about.creator.newTab") }})</span>
            </a>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Github, Linkedin, Mail } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"

const { t } = useI18n()

const links = [
  { key: "github", icon: Github, href: "https://github.com/arinsuda", external: true },
  { key: "linkedin", icon: Linkedin, href: "https://linkedin.com/in/arinsuda", external: true },
  { key: "email", icon: Mail, href: "mailto:sixarin.thorn@gmail.com", external: false },
] as const
</script>

<style scoped>
.creator {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.creator__label {
  margin: 0;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-muted);
}
.creator__card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.creator__body {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  min-width: 0;
}
.creator__role {
  margin: 0;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-brand);
}
.creator__name {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.creator__bio {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.6;
  color: var(--color-text-secondary);
}
.creator__links {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0.5rem 0 0;
  padding: 0;
  list-style: none;
}
.creator__link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  min-height: 2rem;
  padding: 0 0.875rem;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  color: var(--color-text-secondary);
  font-size: var(--text-caption);
  text-decoration: none;
}
.creator__link:hover {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
@media (min-width: 640px) {
  .creator__card {
    flex-direction: row;
    align-items: center;
    gap: 1.5rem;
  }
}
</style>
