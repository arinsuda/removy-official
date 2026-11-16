<template>
  <RouterLink :to="{ name: 'manga-detail', params: { id: item.id } }" class="manga-card">
    <div class="manga-card__cover">
      <img
        v-if="item.cover_url && !failed"
        :class="{ 'is-blurred': blurred }"
        :src="item.cover_url"
        :alt="item.display_title"
        loading="lazy"
        decoding="async"
        @error="failed = true"
      />
      <div v-else class="manga-card__fallback">
        <span>{{ item.display_title }}</span>
      </div>
      <span class="manga-card__kind">{{ t(`manga.kind.${item.kind}`) }}</span>
      <span v-if="item.is_adult" class="manga-card__adult">{{ t("manga.adult.badge") }}</span>
      <span v-if="score !== null" class="manga-card__score" :aria-label="t('manga.score', { score })">
        {{ t("manga.scoreShort", { score }) }}
      </span>
    </div>
    <div class="manga-card__info">
      <h3 class="manga-card__title">{{ item.display_title }}</h3>
      <span v-if="year" class="manga-card__meta">{{ year }}</span>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { useMangaAdult } from "@/composables/useMangaAdult"
import { knownNumber } from "./mangaFormat"
import { anilistScore, type MangaItem } from "@/types/manga"

const props = defineProps<{ item: MangaItem }>()

const { t } = useI18n()
const failed = ref(false)
const score = computed(() => anilistScore(props.item.ratings))
const year = computed(() => knownNumber(props.item.start_year))
/** Adult covers stay blurred until the viewer is eligible and has confirmed the 18+ modal. */
const { canShowAdult } = useMangaAdult()
const blurred = computed(() => props.item.is_adult && !canShowAdult.value)
</script>

<style scoped>
.manga-card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
  color: inherit;
  text-decoration: none;
}
.manga-card__cover {
  position: relative;
  aspect-ratio: 2 / 3;
  overflow: hidden;
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.manga-card__cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.manga-card__fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 0.75rem;
  text-align: center;
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
}
.manga-card__cover img.is-blurred {
  /* hardcode-ok: blur radius is a visual effect, not a design token */
  filter: blur(1rem);
  transform: scale(1.1);
}
.manga-card__kind,
.manga-card__adult,
.manga-card__score {
  position: absolute;
  top: 0.5rem;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-btn);
  background: var(--color-surface);
  font-size: var(--text-caption);
  color: var(--color-text-primary);
}
.manga-card__kind {
  left: 0.5rem;
}
.manga-card__adult {
  top: 2.25rem;
  left: 0.5rem;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-btn);
  background: var(--color-danger);
  font-size: var(--text-caption);
  font-weight: 700;
  color: var(--color-on-brand);
}
.manga-card__score {
  right: 0.5rem;
  color: var(--color-rating);
}
.manga-card__info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}
.manga-card__title {
  margin: 0;
  font-size: var(--text-small);
  font-weight: 600;
  line-height: 1.35;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}
.manga-card__meta {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
</style>
