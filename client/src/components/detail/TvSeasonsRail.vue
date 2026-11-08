<template>
  <section class="seasons" :aria-labelledby="titleId">
    <h2 :id="titleId" class="detail-heading">{{ t("media.detail.seasons.title") }}</h2>
    <ul class="seasons__track">
      <li v-for="s in seasons" :key="s.id" class="seasons__cell">
        <div class="seasons__poster">
          <img v-if="s.poster_path" :src="tmdbImage(s.poster_path, 'w342')" :alt="s.name" loading="lazy" decoding="async" />
          <Tv v-else :size="28" aria-hidden="true" />
        </div>
        <p class="seasons__name">{{ s.name }}</p>
        <p class="seasons__meta">
          {{ t("media.detail.meta.episodes", { count: s.episode_count }, s.episode_count) }}
          <template v-if="s.air_date"> · {{ s.air_date.slice(0, 4) }}</template>
        </p>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import { Tv } from "lucide-vue-next"
import { useImageUrl } from "@/composables/useImageUrl"
import type { TVSeason } from "@/types/movie"

defineProps<{ seasons: TVSeason[] }>()
const { t } = useI18n()
const { tmdbImage } = useImageUrl()
const titleId = useId()
</script>

<style scoped>
.seasons__track { display: flex; gap: 0.75rem; margin: 0.75rem -1rem 0; padding: 0 1rem 0.5rem; overflow-x: auto; list-style: none; scroll-snap-type: x proximity; scrollbar-width: thin; }
.seasons__cell { flex: none; width: 112px; scroll-snap-align: start; }
.seasons__poster { display: grid; aspect-ratio: 2 / 3; place-items: center; overflow: hidden; color: var(--color-text-muted); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.seasons__poster img { width: 100%; height: 100%; object-fit: cover; }
.seasons__name { margin-top: 0.4rem; font-size: var(--text-caption); font-weight: 700; color: var(--color-text-primary); }
.seasons__meta { font-size: var(--text-caption); color: var(--color-text-muted); }
@media (min-width: 640px) {
  .seasons__cell { width: 140px; }
  .seasons__track { margin-inline: 0; padding-inline: 0; }
}
</style>
