<template>
  <section class="videos" :aria-labelledby="titleId">
    <h2 :id="titleId" class="detail-heading">{{ t("media.detail.videos.title") }}</h2>
    <ul class="videos__track">
      <li v-for="v in videos" :key="v.id" class="videos__cell">
        <button type="button" class="videos__card" @click="emit('play', v)">
          <span class="videos__thumb">
            <img :src="`https://i.ytimg.com/vi/${v.key}/mqdefault.jpg`" alt="" loading="lazy" decoding="async" />
            <span class="videos__play"><Play :size="20" aria-hidden="true" /></span>
          </span>
          <span class="videos__name">{{ v.name }}</span>
          <span class="videos__type">{{ v.type }}</span>
        </button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import { Play } from "lucide-vue-next"
import type { Video } from "@/types/movie"

defineProps<{ videos: Video[] }>()
const emit = defineEmits<{ play: [video: Video] }>()
const { t } = useI18n()
const titleId = useId()
</script>

<style scoped>
.videos__track { display: flex; gap: 0.75rem; margin: 0.75rem -1rem 0; padding: 0 1rem 0.5rem; overflow-x: auto; list-style: none; scroll-snap-type: x proximity; scrollbar-width: thin; }
.videos__cell { flex: none; width: 220px; scroll-snap-align: start; }
.videos__card { display: flex; flex-direction: column; gap: 0.2rem; width: 100%; padding: 0; text-align: left; color: inherit; background: none; border: 0; cursor: pointer; }
.videos__thumb { position: relative; display: block; aspect-ratio: 16 / 9; overflow: hidden; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.videos__thumb img { width: 100%; height: 100%; object-fit: cover; }
.videos__play { position: absolute; inset: 0; display: grid; place-items: center; color: var(--color-on-brand); background: color-mix(in srgb, black 35%, transparent); transition: background-color var(--motion-fast) var(--ease-water); }
.videos__name { margin-top: 0.2rem; font-size: var(--text-caption); font-weight: 700; color: var(--color-text-primary); display: -webkit-box; -webkit-box-orient: vertical; -webkit-line-clamp: 2; line-clamp: 2; overflow: hidden; }
.videos__type { font-size: var(--text-caption); color: var(--color-text-muted); }
@media (hover: hover) {
  .videos__card:hover .videos__play { background: color-mix(in srgb, var(--color-brand) 55%, transparent); }
}
@media (min-width: 640px) {
  .videos__cell { width: 260px; }
  .videos__track { margin-inline: 0; padding-inline: 0; }
}
@media (prefers-reduced-motion: reduce) {
  .videos__play { transition: none; }
}
</style>
