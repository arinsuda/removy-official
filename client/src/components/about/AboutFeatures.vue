<template>
  <section class="features" aria-labelledby="about-features">
    <p class="features__label">{{ t("about.features.label") }}</p>
    <h2 id="about-features" class="features__title">{{ t("about.features.title") }}</h2>
    <ul class="features__grid">
      <li v-for="f in features" :key="f.key">
        <RouterLink :to="f.to" class="feature">
          <component :is="f.icon" :size="22" :stroke-width="1.8" class="feature__icon" aria-hidden="true" />
          <h3 class="feature__title">{{ t(`about.features.items.${f.key}.title`) }}</h3>
          <p class="feature__desc">{{ t(`about.features.items.${f.key}.desc`) }}</p>
        </RouterLink>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, type Component } from "vue"
import type { RouteLocationRaw } from "vue-router"
import { useI18n } from "vue-i18n"
import { Bookmark, Star, Trophy, Users } from "lucide-vue-next"
import { useAuthStore } from "@/stores/auth"

const { t } = useI18n()
const auth = useAuthStore()

interface Feature {
  key: "reviews" | "watchlist" | "achievements" | "community"
  icon: Component
  to: RouteLocationRaw
}

// Personal pages need a user id; guests are sent to register instead.
const features = computed<Feature[]>(() => {
  const userId = auth.user?.id
  const register: RouteLocationRaw = { name: "register" }
  return [
    { key: "reviews", icon: Star, to: { name: "movies" } },
    { key: "watchlist", icon: Bookmark, to: userId ? { name: "user-library", params: { userId } } : register },
    { key: "achievements", icon: Trophy, to: userId ? { name: "user-achievements", params: { userId } } : register },
    { key: "community", icon: Users, to: { name: "feed" } },
  ]
})
</script>

<style scoped>
.features {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.features__label {
  margin: 0;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-muted);
}
.features__title {
  margin: 0 0 0.75rem;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.features__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.75rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.feature {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  height: 100%;
  padding: 1.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  color: inherit;
  text-decoration: none;
  transition: background var(--motion-fast) var(--ease-out-soft), border-color var(--motion-fast) var(--ease-out-soft);
}
.feature:hover {
  background: var(--color-surface-3);
  border-color: var(--color-border-strong);
}
.feature__icon {
  color: var(--color-brand);
}
.feature__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
}
.feature__desc {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.6;
  color: var(--color-text-secondary);
}
@media (min-width: 640px) {
  .features__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (min-width: 1024px) {
  .features__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
