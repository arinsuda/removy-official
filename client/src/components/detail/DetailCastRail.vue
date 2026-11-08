<template>
  <section class="cast" :aria-labelledby="titleId">
    <h2 :id="titleId" class="detail-heading">{{ t("media.detail.cast.title") }}</h2>
    <ul class="cast__track">
      <li v-for="actor in cast" :key="actor.id" class="cast__cell">
        <RouterLink :to="`/actors/${actor.id}`" class="cast__card">
          <span class="cast__photo">
            <img v-if="actor.profile_path" :src="tmdbImage(actor.profile_path, 'w185')" :alt="actor.name" loading="lazy" decoding="async" />
            <User v-else :size="28" aria-hidden="true" />
          </span>
          <span class="cast__name">{{ actor.name }}</span>
          <span v-if="actor.character" class="cast__role">{{ actor.character }}</span>
        </RouterLink>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink } from "vue-router"
import { User } from "lucide-vue-next"
import { useImageUrl } from "@/composables/useImageUrl"
import type { Cast } from "@/types/movie"

defineProps<{ cast: Cast[] }>()
const { t } = useI18n()
const { tmdbImage } = useImageUrl()
const titleId = useId()
</script>

<style scoped>
.cast__track { display: flex; gap: 0.75rem; margin: 0.75rem -1rem 0; padding: 0 1rem 0.5rem; overflow-x: auto; list-style: none; scroll-snap-type: x proximity; scrollbar-width: thin; }
.cast__cell { flex: none; width: 96px; scroll-snap-align: start; }
.cast__card { display: flex; flex-direction: column; gap: 0.2rem; color: inherit; text-decoration: none; }
.cast__photo { display: grid; aspect-ratio: 2 / 3; place-items: center; overflow: hidden; color: var(--color-text-muted); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); transition: border-color var(--motion-fast) var(--ease-water); }
.cast__photo img { width: 100%; height: 100%; object-fit: cover; }
.cast__name { margin-top: 0.2rem; font-size: var(--text-caption); font-weight: 700; color: var(--color-text-primary); overflow-wrap: anywhere; }
.cast__role { font-size: var(--text-caption); color: var(--color-text-muted); overflow-wrap: anywhere; }
@media (hover: hover) {
  .cast__card:hover .cast__photo { border-color: var(--color-brand); }
}
@media (min-width: 640px) {
  .cast__cell { width: 120px; }
  .cast__track { margin-inline: 0; padding-inline: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .cast__photo { transition: none; }
}
</style>
