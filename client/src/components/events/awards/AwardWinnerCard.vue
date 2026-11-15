<template>
  <RouterLink :to="detailRoute" class="award-card">
    <div class="award-card__poster">
      <img
        v-if="posterUrl"
        :src="posterUrl"
        alt=""
        loading="lazy"
        class="award-card__img"
      />
      <div v-else class="award-card__empty">
        <ImageOff :size="28" aria-hidden="true" />
      </div>
    </div>

    <div class="award-card__body">
      <h3 class="award-card__title">{{ award.media_title }}</h3>
      <p class="award-card__meta">
        <span v-if="coWinner" class="award-card__co-winner">{{ t("awards.card.coWinner") }}</span>
        <span>{{ t("awards.card.votes", { count: award.votes }, award.votes) }}</span>
      </p>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { ImageOff } from "lucide-vue-next"
import type { EventAward } from "@/types/event"
import { AWARD_POSTER_SIZE } from "@/utils/events"
import { getTmdbImageUrl } from "@/utils/image"

const props = defineProps<{ award: EventAward; coWinner: boolean }>()
const { t } = useI18n()

const posterUrl = computed(() => getTmdbImageUrl(props.award.poster_path, AWARD_POSTER_SIZE))
const detailRoute = computed(() => ({
  name: props.award.media_type === "movie" ? "movie-detail" : "tv-detail",
  params: { id: props.award.media_id },
}))
</script>

<style scoped>
.award-card {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  color: var(--color-text-primary);
  text-decoration: none;
  background-color: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
}

.award-card:hover .award-card__title {
  color: var(--color-brand);
}

.award-card__poster {
  aspect-ratio: 2 / 3;
  background-color: var(--color-surface-3);
}

.award-card__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.award-card__empty {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  color: var(--color-text-muted);
}

.award-card__body {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.375rem;
  padding: 0.75rem;
}

.award-card__title {
  font-size: var(--text-small);
  font-weight: 600;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.award-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem 0.5rem;
  margin-top: auto;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.award-card__co-winner {
  color: var(--color-rating);
  font-weight: 600;
}
</style>
