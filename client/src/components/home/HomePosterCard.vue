<template>
  <div
    ref="wrap"
    class="poster-wrap"
    @mouseenter="popup.show(item, popupKey, wrap!)"
    @mouseleave="popup.leave(popupKey)"
    @focusin="popup.focus(item, popupKey, wrap!)"
    @focusout="popup.leave(popupKey)"
  >
    <RouterLink :to="detailRoute(item)" class="poster-card" :data-tier="tier">
      <div class="poster-image">
        <img
          v-if="item.poster_path && !failed"
          :src="posterSrc(item.poster_path)"
          :srcset="posterSrcset(item.poster_path)"
          :sizes="sizes"
          :alt="item.title"
          loading="lazy"
          decoding="async"
          @error="failed = true"
        />
        <div v-else class="poster-fallback">
          <span>{{ item.title }}</span>
        </div>
        <div class="rating-badge">
          <PopcornScoreBadge :average="popcorn.average" :count="popcorn.count" />
        </div>
        <span v-if="tmdb" class="tmdb-pill" :aria-label="t('media.rating.tmdbScore', { score: tmdb })">
          <span aria-hidden="true">TMDB {{ tmdb }}</span>
        </span>
      </div>
      <div class="poster-info">
        <h3 class="poster-title">{{ item.title }}</h3>
        <span v-if="caption" class="poster-year">{{ caption }}</span>
        <span v-else-if="item.date" class="poster-year">{{ item.date.slice(0, 4) }}</span>
      </div>
    </RouterLink>

    <Teleport to="body">
      <Transition name="popup">
        <div
          v-if="popup.openKey.value === popupKey && popup.isDesktop.value"
          class="hover-popup"
          :style="{ top: `${popup.anchor.value.top}px`, left: `${popup.anchor.value.left}px` }"
          @mouseenter="popup.popupEnter()"
          @mouseleave="popup.popupLeave(popupKey)"
        >
          <PopupCard
            :movie="item.source"
            :ratings="popup.detailRatings(item)"
            :media-type="item.type"
            :current-trailer="state.currentTrailer.value"
            :trailer-unavailable="state.trailerUnavailable.value"
            :is-iframe-mounted="state.isIframeMounted.value"
            :is-iframe-loaded="state.isIframeLoaded.value"
            :show-skeleton="state.showSkeleton.value"
            :show-fallback="state.showFallback.value"
            :attach-player="state.attachPlayer"
          />
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import PopcornScoreBadge from "@/components/movie/PopcornScoreBadge.vue"
import PopupCard from "@/components/movie/PopupCard.vue"
import { popcornTier } from "@/components/movie/popcornTier"
import { useHomePopup } from "@/composables/useHomePopup"
import { detailRoute, posterSrc, posterSrcset, type HomeMedia } from "./homeMedia"

const props = withDefaults(
  defineProps<{
    item: HomeMedia
    /** Rail id, keeps popup keys unique when a title appears in two rails. */
    rail: string
    /** img `sizes` hint; grids with other column counts pass their own. */
    sizes?: string
    /** Replaces the release year under the title (e.g. an upcoming release date). */
    caption?: string
  }>(),
  { sizes: "(min-width: 1024px) 180px, (min-width: 640px) 28vw, 42vw" },
)

const { t } = useI18n()
const popup = useHomePopup()
const wrap = useTemplateRef<HTMLElement>("wrap")
const failed = ref(false)

const popupKey = computed(() => `${props.rail}:${props.item.key}`)
const state = computed(() => popup.getState(props.item))
const popcorn = computed(() => props.item.ratings?.popcorn ?? { average: 0, count: 0 })
const tmdb = computed(() => {
  const src = props.item.ratings?.tmdb
  const v = src ? (src.available ? src.average : 0) : props.item.vote_average
  return v && v > 0 ? v.toFixed(1) : null
})
const tier = computed(() => {
  const p = props.item.ratings?.popcorn
  return popcornTier(p?.average, p?.count)
})
</script>

<style scoped>
.poster-wrap {
  position: relative;
  width: 100%;
  min-width: 0;
}
.poster-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
  color: inherit;
  text-decoration: none;
  border-radius: var(--radius-card);
  transition: transform var(--motion-fast, 0.25s) var(--ease-out-soft);
}
.poster-card:hover {
  transform: translateY(-4px);
}
.poster-image {
  position: relative;
  aspect-ratio: 2 / 3;
  overflow: hidden;
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
}
/* Heat rim: the poster edge and bottom scrim carry the POPCORN tier. */
.poster-image::after {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
  border-radius: inherit;
  transition: box-shadow 0.25s;
}
.poster-card[data-tier="hot"] .poster-image {
  border-color: var(--color-brand);
}
.poster-card[data-tier="hot"] .poster-image::after {
  background: linear-gradient(to top, color-mix(in srgb, var(--color-brand) 45%, transparent) 0%, transparent 38%);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--color-accent) 55%, transparent);
}
.poster-card[data-tier="hot"]:hover .poster-image,
.poster-card[data-tier="hot"]:focus-visible .poster-image {
  box-shadow: 0 0 22px color-mix(in srgb, var(--color-brand) 55%, transparent);
}
.poster-card[data-tier="fresh"] .poster-image {
  border-color: var(--color-border-strong);
}
.poster-card[data-tier="stale"] .poster-image img {
  filter: saturate(0.6);
}
.poster-card[data-tier="stale"] .poster-image::after {
  background: linear-gradient(to top, color-mix(in srgb, var(--color-surface) 70%, transparent) 0%, transparent 38%);
}
.poster-card[data-tier="none"] .poster-image {
  border-style: dashed;
}
.poster-image img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.35s ease;
}
.poster-card:hover img {
  transform: scale(1.05);
}
.rating-badge {
  position: absolute;
  top: 6px;
  left: 6px;
}
.tmdb-pill {
  position: absolute;
  top: 6px;
  right: 6px;
  padding: 2px 6px;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--color-surface) 80%, transparent);
  color: var(--color-text-secondary);
  font-size: 10px;
  font-weight: 600;
  line-height: 1.4;
  white-space: nowrap;
}
.poster-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 1rem;
  text-align: center;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.poster-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-height: 40px;
  padding: 0 2px;
}
.poster-title {
  margin: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: var(--text-small);
  font-weight: 600;
  line-height: 1.3;
  color: var(--color-text-primary);
}
.poster-year {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.hover-popup {
  position: fixed;
  z-index: var(--z-popover);
  width: 320px;
  overflow: hidden;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border-strong);
  border-radius: 14px;
  box-shadow: 0 24px 70px rgb(0 0 0 / 85%);
}
.popup-enter-active {
  transition: opacity 0.2s, transform 0.2s var(--ease-out-soft);
}
.popup-leave-active {
  transition: opacity 0.15s;
}
.popup-enter-from {
  opacity: 0;
  transform: scale(0.94) translateY(6px);
}
.popup-leave-to {
  opacity: 0;
}
@media (prefers-reduced-motion: reduce) {
  .poster-card:hover,
  .poster-card:hover img {
    transform: none;
  }
  .popup-enter-from {
    transform: none;
  }
}
</style>
