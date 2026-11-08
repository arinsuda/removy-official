<template>
  <div v-if="!items.length && loading" class="hero hero--loading" aria-hidden="true" />
  <section
    v-else-if="items.length"
    class="hero"
    :aria-label="t('home.hero.label')"
    aria-roledescription="carousel"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false"
    @focusin="focused = true"
    @focusout="focused = false"
  >
    <div class="hero__slide">
      <Transition name="hero-fade" mode="out-in">
        <div :key="current.key" class="hero__media">
          <img
            v-if="current.backdrop_path && !failed.has(current.key)"
            :src="backdropSrc(current.backdrop_path)"
            alt=""
            class="hero__backdrop"
            :fetchpriority="index === 0 ? 'high' : 'auto'"
            decoding="async"
            @error="failed.add(current.key)"
          />
          <div v-else class="hero__backdrop hero__backdrop--fallback" />
        </div>
      </Transition>
      <div class="hero__overlay" />
    </div>

    <div class="hero__content" :aria-live="playing ? 'off' : 'polite'">
      <span class="hero__eyebrow" :data-source="source">
        <Flame v-if="source === 'hot'" :size="14" aria-hidden="true" />
        <Sparkles v-else :size="14" aria-hidden="true" />
        {{ t(`home.hero.eyebrow.${source}`) }}
      </span>
      <h1 class="hero__title">{{ current.title }}</h1>
      <div class="hero__meta">
        <MediaRating variant="detail" :ratings="current.ratings" :vote-average="current.vote_average" />
        <span v-if="current.date" class="hero__year">{{ current.date.slice(0, 4) }}</span>
      </div>
      <p v-if="current.overview" class="hero__overview">{{ current.overview }}</p>
      <div class="hero__actions">
        <AppButton :to="detailRoute(current)">
          <Info :size="16" aria-hidden="true" />
          {{ t("home.hero.details") }}
        </AppButton>
        <AppButton :to="{ name: current.type === 'tv' ? 'tv' : 'movies' }" variant="outline">
          <Compass :size="16" aria-hidden="true" />
          {{ t("home.hero.explore") }}
        </AppButton>
      </div>
    </div>

    <div v-if="items.length > 1" class="hero__controls">
      <AppButton
        variant="ghost"
        size="sm"
        icon-only
        :aria-label="t(userPaused ? 'home.hero.play' : 'home.hero.pause')"
        @click="userPaused = !userPaused"
      >
        <Play v-if="userPaused" :size="14" aria-hidden="true" />
        <Pause v-else :size="14" aria-hidden="true" />
      </AppButton>
      <div class="hero__dots" role="group" :aria-label="t('home.hero.slides')">
        <button
          v-for="(item, i) in items"
          :key="item.key"
          type="button"
          class="hero__dot"
          :class="{ 'hero__dot--active': index === i }"
          :aria-label="t('home.hero.slide', { n: i + 1, title: item.title })"
          :aria-current="index === i"
          @click="index = i"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Compass, Flame, Info, Pause, Play, Sparkles } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import MediaRating from "@/components/movie/MediaRating.vue"
import { useCarouselAutoplay } from "@/composables/useCarouselAutoplay"
import { backdropSrc, detailRoute, type HomeMedia } from "./homeMedia"

const props = defineProps<{
  items: HomeMedia[]
  /** "hot" = ป๊อปร้อน titles, "popular" = TMDB fallback. */
  source: "hot" | "popular"
  loading?: boolean
}>()
const { t } = useI18n()

const failed = ref(new Set<string>())
const { index, userPaused, hovering, focused, playing } = useCarouselAutoplay(() => props.items.length)

const current = computed(() => props.items[index.value] ?? props.items[0]!)
</script>

<style scoped>
.hero {
  position: relative;
  width: 100%;
  height: 380px;
  overflow: hidden;
  background: var(--color-surface);
}
.hero__slide,
.hero__media {
  position: absolute;
  inset: 0;
}
.hero__backdrop {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 25%;
}
.hero__backdrop--fallback,
.hero__overlay {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(to top, var(--color-surface) 0%, color-mix(in srgb, var(--color-surface) 75%, transparent) 35%, transparent 100%),
    linear-gradient(to right, color-mix(in srgb, var(--color-surface) 90%, transparent) 0%, transparent 75%);
}
.hero__content {
  position: absolute;
  bottom: 52px;
  left: 50%;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
  max-width: 1240px;
  padding: 0 1rem;
  transform: translateX(-50%);
}
.hero__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  width: fit-content;
  padding: 4px 12px;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  background: color-mix(in srgb, var(--color-surface) 60%, transparent);
  color: var(--color-text-primary);
  font-size: var(--text-caption);
  font-weight: 600;
}
.hero__eyebrow[data-source="hot"] {
  border-color: var(--color-brand);
  color: var(--color-accent);
}
.hero__title {
  display: -webkit-box;
  max-width: 680px;
  margin: 0;
  overflow: hidden;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.15;
  color: var(--color-text-primary);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.hero__meta {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--color-text-secondary);
}
.hero__overview {
  display: none;
  max-width: 600px;
  margin: 0;
  overflow: hidden;
  font-size: 0.925rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
.hero__actions {
  display: flex;
  gap: 12px;
  margin-top: 0.25rem;
}
.hero__controls {
  position: absolute;
  bottom: 12px;
  left: 50%;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 8px;
  transform: translateX(-50%);
}
.hero__dots {
  display: flex;
  gap: 8px;
}
.hero__dot {
  width: 24px;
  height: 4px;
  padding: 0;
  border: none;
  border-radius: 2px;
  background: var(--color-border-strong);
  cursor: pointer;
  transition: width 0.3s, background-color 0.3s;
}
.hero__dot--active {
  width: 36px;
  background: var(--color-brand);
}
.hero-fade-enter-active,
.hero-fade-leave-active {
  transition: opacity 0.5s ease-in-out;
}
.hero-fade-enter-from,
.hero-fade-leave-to {
  opacity: 0;
}
@media (prefers-reduced-motion: reduce) {
  .hero-fade-enter-active,
  .hero-fade-leave-active {
    transition: none;
  }
}
@media (min-width: 768px) {
  .hero {
    height: 500px;
  }
  .hero__content {
    bottom: 56px;
    gap: 0.75rem;
    padding: 0 1.5rem;
  }
  .hero__title {
    font-size: var(--text-page);
  }
  .hero__overview {
    display: -webkit-box;
  }
}
</style>
