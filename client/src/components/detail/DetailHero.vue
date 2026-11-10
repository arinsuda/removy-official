<template>
  <header class="hero">
    <div class="hero__backdrop" aria-hidden="true">
      <img v-if="data.backdropPath" :src="tmdbImage(data.backdropPath, 'w780')" alt="" class="hero__img" fetchpriority="high" decoding="async" />
      <div class="hero__scrim" />
    </div>

    <div class="hero__inner">
      <AppButton variant="ghost" size="sm" class="hero__back" @click="emit('back')">
        <ArrowLeft :size="16" aria-hidden="true" />
        {{ t("media.detail.back") }}
      </AppButton>

      <div class="hero__row">
        <div class="hero__poster">
          <img v-if="data.posterPath" :src="tmdbImage(data.posterPath, 'w500')" :alt="data.title" decoding="async" />
          <div v-else class="hero__poster-empty"><ImageOff :size="32" aria-hidden="true" /></div>
        </div>

        <div class="hero__info">
          <h1 class="hero__title">{{ data.title }}</h1>
          <p v-if="data.englishTitle && data.englishTitle !== data.title" class="hero__sub">{{ data.englishTitle }}</p>
          <p v-if="meta.length" class="hero__meta" data-test="hero-meta">
            <template v-for="(m, i) in meta" :key="i">
              <span v-if="i > 0" aria-hidden="true">·</span>
              <span>{{ m }}</span>
            </template>
          </p>
          <ul v-if="data.genres.length" class="hero__genres">
            <li v-for="g in data.genres" :key="g.id">{{ g.name }}</li>
          </ul>
          <p v-if="data.tagline" class="hero__tagline">“{{ data.tagline }}”</p>
          <p v-if="data.crewHighlights.length" class="hero__crew">
            <span class="hero__crew-role">{{ t(`media.detail.crew.${data.crewHighlights[0]!.role}`) }}</span>
            {{ data.crewHighlights.map(c => c.name).join(", ") }}
          </p>
          <AppButton v-if="hasTrailer" variant="outline" size="sm" class="hero__trailer" data-test="hero-trailer" @click="emit('trailer')">
            <Play :size="16" aria-hidden="true" />
            {{ t("media.detail.actions.playTrailer") }}
          </AppButton>
          <div class="hero__actions"><slot name="actions" /></div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { ArrowLeft, ImageOff, Play } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import { useImageUrl } from "@/composables/useImageUrl"
import { statusKey, type DetailData } from "./detailConfig"

const props = defineProps<{ data: DetailData; hasTrailer: boolean }>()
const emit = defineEmits<{ back: []; trailer: [] }>()

const { t, te } = useI18n()
const { tmdbImage } = useImageUrl()

function formatRuntime(min: number): string {
  const h = Math.floor(min / 60)
  const m = min % 60
  return h > 0 ? t("media.detail.meta.runtimeHours", { h, m }) : t("media.detail.meta.runtimeMinutes", { minutes: m })
}

const meta = computed(() => {
  const d = props.data
  const out: string[] = []
  if (d.year) out.push(d.year)
  if (d.runtimeMin) out.push(formatRuntime(d.runtimeMin))
  if (d.seasonCount) out.push(t("media.detail.meta.seasons", { count: d.seasonCount }, d.seasonCount))
  if (d.episodeCount) out.push(t("media.detail.meta.episodes", { count: d.episodeCount }, d.episodeCount))
  if (d.status) out.push(te(`media.detail.status.${statusKey(d.status)}`) ? t(`media.detail.status.${statusKey(d.status)}`) : d.status)
  return out
})
</script>

<style scoped>
.hero { position: relative; }
.hero__backdrop { position: absolute; inset: 0 0 auto; height: 56svh; overflow: hidden; }
.hero__img { width: 100%; height: 100%; object-fit: cover; object-position: center 20%; animation: hero-in var(--motion-wave) var(--ease-out-soft) both; }
.hero__scrim {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(to bottom, color-mix(in srgb, var(--color-surface) 55%, transparent), transparent 40%),
    linear-gradient(to top, var(--color-surface), color-mix(in srgb, var(--color-surface) 35%, transparent) 70%, transparent);
}
.hero__inner { position: relative; max-width: 1200px; margin: 0 auto; padding: 4.5rem 1rem 0; }
.hero__back { position: absolute; top: 4rem; left: 0.5rem; }
.hero__row { display: flex; align-items: flex-end; gap: 1rem; margin-top: calc(56svh - 7rem - 4.5rem - 1rem); }
.hero__poster { flex: none; width: 112px; aspect-ratio: 2 / 3; overflow: hidden; background: var(--color-surface-2); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card); box-shadow: 0 12px 32px color-mix(in srgb, black 55%, transparent); animation: hero-rise var(--motion-wave) var(--ease-out-soft) both; }
.hero__poster img { width: 100%; height: 100%; object-fit: cover; }
.hero__poster-empty { display: grid; width: 100%; height: 100%; place-items: center; color: var(--color-text-muted); }
.hero__info { display: flex; flex: 1; flex-direction: column; gap: 0.5rem; min-width: 0; }
.hero__title { font-size: var(--text-heading); font-weight: 800; line-height: 1.2; text-wrap: balance; overflow-wrap: anywhere; }
.hero__sub { font-size: var(--text-small); color: var(--color-text-secondary); }
.hero__meta { display: flex; flex-wrap: wrap; gap: 0.1rem 0.5rem; font-size: var(--text-small); color: var(--color-text-secondary); }
.hero__genres { display: flex; flex-wrap: wrap; gap: 0.4rem; margin: 0; padding: 0; list-style: none; }
.hero__genres li { padding: 0.15rem 0.65rem; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-primary); background: color-mix(in srgb, var(--color-surface-3) 80%, transparent); border: 1px solid var(--color-border-strong); border-radius: var(--radius-btn); }
.hero__tagline { display: none; font-size: var(--text-small); font-style: italic; color: var(--color-text-secondary); }
.hero__crew { font-size: var(--text-small); color: var(--color-text-primary); }
.hero__crew-role { margin-right: 0.35rem; color: var(--color-text-muted); }
.hero__trailer { align-self: flex-start; }
.hero__actions { margin-top: 0.25rem; }

@keyframes hero-in { from { opacity: 0; transform: scale(1.04); } to { opacity: 1; transform: none; } }
@keyframes hero-rise { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: none; } }

@media (min-width: 640px) {
  .hero__inner { padding-inline: 1.5rem; }
  .hero__poster { width: 140px; }
  .hero__row { gap: 1.5rem; margin-top: calc(56svh - 10rem - 4.5rem - 1rem); }
  .hero__title { font-size: var(--text-page); }
  .hero__tagline { display: block; }
}
@media (min-width: 768px) {
  .hero__actions { margin-top: 0.75rem; }
}
@media (min-width: 1024px) {
  .hero__poster { width: 200px; }
  .hero__row { align-items: center; margin-top: 1.5rem; }
  .hero__info { padding-bottom: 0; }
}
@media (min-width: 1280px) {
  .hero__poster { width: 240px; }
  .hero__title { font-size: var(--text-display); }
}
@media (prefers-reduced-motion: reduce) {
  .hero__img, .hero__poster { animation: none; }
}
</style>
