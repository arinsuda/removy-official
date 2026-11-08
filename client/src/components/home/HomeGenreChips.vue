<template>
  <nav class="genre-chips" :aria-label="t('home.genres.label')">
    <RouterLink
      v-for="genre in GENRES"
      :key="genre.id"
      :to="{ name: 'movies', query: { genre: genre.id } }"
      class="genre-chip"
    >
      <component :is="genre.icon" :size="15" class="genre-chip__icon" aria-hidden="true" />
      <span>{{ t(`home.genres.${genre.key}`) }}</span>
    </RouterLink>
  </nav>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Film, Flame, Ghost, Rocket, ShieldAlert, Smile, Sparkles } from "lucide-vue-next"

const { t } = useI18n()

const GENRES = [
  { id: 28, key: "action", icon: Flame },
  { id: 878, key: "scifi", icon: Rocket },
  { id: 35, key: "comedy", icon: Smile },
  { id: 27, key: "horror", icon: Ghost },
  { id: 18, key: "drama", icon: Film },
  { id: 16, key: "animation", icon: Sparkles },
  { id: 53, key: "thriller", icon: ShieldAlert },
] as const
</script>

<style scoped>
.genre-chips {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  scrollbar-width: none;
  padding: 2px 0 4px;
}
.genre-chips::-webkit-scrollbar {
  display: none;
}
.genre-chip {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--color-border);
  border-radius: 9999px;
  background: var(--color-surface-2);
  color: var(--color-text-secondary);
  font-size: 0.775rem;
  font-weight: 500;
  text-decoration: none;
  white-space: nowrap;
  transition: color 0.2s, border-color 0.2s, transform var(--motion-fast, 0.2s) var(--ease-out-soft);
}
.genre-chip__icon {
  color: var(--color-brand);
}
.genre-chip:hover {
  color: var(--color-text-primary);
  border-color: var(--color-border-strong);
  transform: translateY(-2px);
}
@media (min-width: 768px) {
  .genre-chips {
    flex-wrap: wrap;
    justify-content: center;
    overflow: visible;
  }
  .genre-chip {
    padding: 8px 18px;
    font-size: 0.85rem;
  }
}
@media (prefers-reduced-motion: reduce) {
  .genre-chip:hover {
    transform: none;
  }
}
</style>
