<template>
  <header class="manga-hero">
    <img v-if="manga.banner_url" class="manga-hero__banner" :src="manga.banner_url" alt="" aria-hidden="true" loading="eager" decoding="async" />
    <div class="manga-hero__body">
      <div class="manga-hero__cover">
        <img v-if="manga.cover_url" :src="manga.cover_url" :alt="manga.display_title" decoding="async" />
        <div v-else class="manga-hero__fallback" aria-hidden="true">{{ manga.display_title }}</div>
      </div>
      <div class="manga-hero__text">
        <h1 class="manga-hero__title">{{ manga.display_title }}</h1>
        <p v-if="romaji" class="manga-hero__sub">{{ romaji }}</p>
        <p v-if="native" class="manga-hero__sub" lang="ja">{{ native }}</p>
        <p v-if="score !== null" class="manga-hero__score">{{ t("manga.scoreShort", { score }) }}</p>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { anilistScore, type MangaDetail } from "@/types/manga"

const props = defineProps<{ manga: MangaDetail }>()

const { t } = useI18n()
/** Romaji and native names are shown only when they differ from the display title. */
const romaji = computed(() => (props.manga.titles.romaji && props.manga.titles.romaji !== props.manga.display_title ? props.manga.titles.romaji : ""))
const native = computed(() => (props.manga.titles.native && props.manga.titles.native !== props.manga.display_title ? props.manga.titles.native : ""))
const score = computed(() => anilistScore(props.manga.ratings))
</script>

<style scoped>
.manga-hero {
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.manga-hero__banner {
  display: block;
  width: 100%;
  height: 9rem;
  object-fit: cover;
  opacity: 0.45;
}
.manga-hero__body {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
  padding: 1rem;
}
.manga-hero__banner + .manga-hero__body {
  margin-top: -5rem;
}
.manga-hero__cover {
  flex-shrink: 0;
  width: 7.5rem;
  aspect-ratio: 2 / 3;
  overflow: hidden;
  border-radius: var(--radius-card);
  background: var(--color-surface-3);
}
.manga-hero__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.manga-hero__fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 0.5rem;
  text-align: center;
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
}
.manga-hero__text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}
.manga-hero__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  line-height: 1.25;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}
.manga-hero__sub {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}
.manga-hero__score {
  margin: 0.25rem 0 0;
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-rating);
}
@media (min-width: 768px) {
  .manga-hero__banner {
    height: 14rem;
  }
  .manga-hero__body {
    padding: 1.5rem;
  }
  .manga-hero__banner + .manga-hero__body {
    margin-top: -7rem;
  }
  .manga-hero__cover {
    width: 10rem;
  }
  .manga-hero__title {
    font-size: var(--text-page);
  }
}
</style>
