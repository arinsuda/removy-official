<template>
  <section v-if="hero" class="up-hero" :aria-labelledby="titleId">
    <img v-if="hero.backdrop_path && !failed" :src="backdropSrc(hero.backdrop_path)" alt="" class="up-hero__backdrop" decoding="async" @error="failed = true" />
    <div class="up-hero__scrim" />
    <div class="up-hero__body">
      <img v-if="hero.poster_path" :src="posterSrc(hero.poster_path)" alt="" class="up-hero__poster" decoding="async" />
      <div class="up-hero__info">
        <span class="up-hero__eyebrow">
          <CalendarClock :size="14" aria-hidden="true" />
          {{ t("upcoming.hero.eyebrow") }} · {{ whenLabel }}
        </span>
        <h2 :id="titleId" class="up-hero__title">{{ hero.title }}</h2>
        <p v-if="hero.overview" class="up-hero__overview">{{ hero.overview }}</p>
        <AppButton :to="detailRoute(hero)" class="up-hero__cta">
          <Info :size="16" aria-hidden="true" />
          {{ t("upcoming.hero.details") }}
        </AppButton>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { CalendarClock, Info } from "lucide-vue-next"
import { backdropSrc, detailRoute, posterSrc, type HomeMedia } from "@/components/home/homeMedia"
import AppButton from "@/components/ui/AppButton.vue"
import { daysUntil, parseLocalDate, pickHero } from "./upcomingGroups"

const props = defineProps<{ items: HomeMedia[]; today: Date }>()
const { t } = useI18n()
const titleId = useId()
const failed = ref(false)

const hero = computed(() => pickHero(props.items, props.today))
// Day-level only: no per-second timer, so nothing re-renders or announces itself.
const whenLabel = computed(() => {
  const date = hero.value && parseLocalDate(hero.value.date)
  if (!date) return ""
  const n = daysUntil(date, props.today)
  return t("upcoming.hero.when", { n }, n)
})
</script>

<style scoped>
.up-hero {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.up-hero__backdrop,
.up-hero__scrim {
  display: none;
}
.up-hero__body {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
}
.up-hero__poster {
  width: 9rem;
  aspect-ratio: 2 / 3;
  object-fit: cover;
  border-radius: var(--radius-card);
  border: 1px solid var(--color-border);
}
.up-hero__info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
  min-width: 0;
}
.up-hero__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border: 1px solid var(--color-brand);
  border-radius: 9999px;
  color: var(--color-accent);
  font-size: var(--text-caption);
  font-weight: 600;
}
.up-hero__title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.2;
  color: var(--color-text-primary);
}
.up-hero__overview {
  display: -webkit-box;
  max-width: 40rem;
  margin: 0;
  overflow: hidden;
  font-size: var(--text-small);
  line-height: 1.5;
  color: var(--color-text-secondary);
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}
.up-hero__cta {
  margin-top: 0.25rem;
}
@media (min-width: 768px) {
  .up-hero__body {
    flex-direction: row;
    align-items: center;
    gap: 1.5rem;
    padding: 1.5rem;
  }
  .up-hero__poster {
    flex: none;
    width: 11rem;
  }
}
@media (min-width: 1024px) {
  .up-hero {
    min-height: 22rem;
    display: flex;
    align-items: center;
  }
  .up-hero__backdrop,
  .up-hero__scrim {
    display: block;
    position: absolute;
    inset: 0;
  }
  .up-hero__backdrop {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center 25%;
  }
  .up-hero__scrim {
    background:
      linear-gradient(to top, var(--color-surface) 0%, transparent 60%),
      linear-gradient(to right, color-mix(in srgb, var(--color-surface) 92%, transparent) 0%, color-mix(in srgb, var(--color-surface) 55%, transparent) 55%, transparent 100%);
  }
  .up-hero__body {
    width: 100%;
    padding: 2rem 2.5rem;
  }
  .up-hero__poster {
    width: 12rem;
  }
  .up-hero__title {
    font-size: var(--text-display);
  }
}
</style>
