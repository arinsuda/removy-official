<template>
  <section v-if="items.length" class="hero" :aria-label="t('home.hero.label')">
    <div class="hero__slides">
      <TransitionGroup name="hero-fade">
        <div v-for="(item, i) in items" v-show="index === i" :key="item.key" class="hero__slide">
          <img
            v-if="item.backdrop_path && !failed.has(item.key)"
            :src="backdropSrc(item.backdrop_path)"
            alt=""
            class="hero__backdrop"
            @error="failed.add(item.key)"
          />
          <div v-else class="hero__backdrop hero__backdrop--fallback" />
          <div class="hero__overlay-bottom" />
          <div class="hero__overlay-side" />
        </div>
      </TransitionGroup>
    </div>

    <div v-if="current" class="hero__content">
      <span class="hero__badge">
        <Sparkles :size="14" aria-hidden="true" />
        {{ t("home.hero.badge") }}
      </span>
      <h1 class="hero__title">{{ current.title }}</h1>
      <div class="hero__meta">
        <span v-if="current.date">{{ current.date.slice(0, 4) }}</span>
        <MediaRating :ratings="current.ratings" :vote-average="current.vote_average" />
      </div>
      <p v-if="current.overview" class="hero__overview">{{ current.overview }}</p>
      <div class="hero__actions">
        <AppButton :to="detailRoute(current)">
          <Info :size="16" aria-hidden="true" />
          {{ t("home.hero.details") }}
        </AppButton>
        <AppButton :to="{ name: 'movies' }" variant="outline">
          <Compass :size="16" aria-hidden="true" />
          {{ t("home.hero.explore") }}
        </AppButton>
      </div>
    </div>

    <div class="hero__dots">
      <button
        v-for="(item, i) in items"
        :key="item.key"
        type="button"
        class="hero__dot"
        :class="{ 'hero__dot--active': index === i }"
        :aria-label="t('home.hero.slide', { n: i + 1, title: item.title })"
        :aria-current="index === i"
        @click="go(i)"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Compass, Info, Sparkles } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import MediaRating from "@/components/movie/MediaRating.vue"
import { backdropSrc, detailRoute, type HomeMedia } from "./homeMedia"

const props = defineProps<{ items: HomeMedia[] }>()
const { t } = useI18n()

const AUTOPLAY_MS = 5000
const index = ref(0)
const failed = ref(new Set<string>())
const current = computed(() => props.items[index.value] ?? null)
let timer: ReturnType<typeof setInterval> | null = null

function restart() {
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    index.value = (index.value + 1) % Math.max(props.items.length, 1)
  }, AUTOPLAY_MS)
}

function go(i: number) {
  index.value = i
  restart()
}

onMounted(restart)
onBeforeUnmount(() => timer && clearInterval(timer))
</script>

<style scoped>
.hero {
  position: relative;
  width: 100%;
  height: 320px;
  overflow: hidden;
  background: var(--color-surface);
}
.hero__slides,
.hero__slide {
  position: absolute;
  inset: 0;
}
.hero__backdrop {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 25%;
}
.hero__backdrop--fallback {
  background: var(--color-surface-2);
}
.hero__overlay-bottom {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, var(--color-surface) 0%, color-mix(in srgb, var(--color-surface) 80%, transparent) 25%, transparent 100%);
}
.hero__overlay-side {
  position: absolute;
  inset: 0;
  background: linear-gradient(to right, color-mix(in srgb, var(--color-surface) 95%, transparent) 0%, transparent 80%);
}
.hero__content {
  position: absolute;
  bottom: 36px;
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
.hero__badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  width: fit-content;
  padding: 4px 12px;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  color: var(--color-accent);
  font-size: var(--text-caption);
  font-weight: 600;
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
  gap: 8px;
  font-size: var(--text-small);
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
.hero__dots {
  position: absolute;
  bottom: 16px;
  left: 50%;
  z-index: 10;
  display: flex;
  gap: 8px;
  transform: translateX(-50%);
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
  transition: opacity 0.7s ease-in-out;
}
.hero-fade-enter-from,
.hero-fade-leave-to {
  opacity: 0;
}
@media (min-width: 480px) {
  .hero {
    height: 380px;
  }
}
@media (min-width: 768px) {
  .hero {
    height: 480px;
  }
  .hero__content {
    bottom: 50px;
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
