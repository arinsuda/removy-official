<template>
  <div class="popup__media">
    <RouterLink
      :to="{ name: detailRouteName, params: { id: movie.id } }"
      class="popup__media-link"
    >
      <div v-if="showSkeleton" class="popup__skeleton" aria-hidden="true">
        <SkeletonBlock width="55%" height="7px" radius="4px" />
        <SkeletonBlock width="38%" height="5px" radius="4px" />
      </div>

      <template v-if="showFallback">
        <img
          v-if="backdropUrl"
          :src="backdropUrl"
          :alt="displayTitle"
          class="popup__backdrop"
        />
        <div v-else class="popup__no-media">
          <Film :size="28" class="popup__no-media-icon" />
        </div>
      </template>

      <!-- ── YouTube Player mount target ── -->
      <!-- ใช้ div ธรรมดาแทน <iframe :src="..."> เพราะต้องให้ YT.Player
           คุม element นี้เองถึงจะดัก onError ได้จริง -->
      <div
        v-if="isIframeMounted"
        :id="playerElId"
        class="popup__iframe"
        :class="{ 'popup__iframe--visible': isIframeLoaded }"
      />

      <div v-if="trailerUnavailable" class="popup__no-trailer">
        <VolumeX :size="12" />
        <span>{{ t('media.popup.noTrailer') }}</span>
      </div>

      <div v-if="isIframeMounted && currentTrailer" class="popup__badge">
        <span class="popup__badge-dot" />
        {{ currentTrailer.type }}
      </div>

      <div class="popup__media-fade" />
    </RouterLink>
  </div>

  <div class="popup__info">
    <div class="popup__title-row">
      <RouterLink
        :to="{ name: detailRouteName, params: { id: movie.id } }"
        class="popup__title-link"
      >
        <h3 class="popup__title">{{ displayTitle }}</h3>
      </RouterLink>
    </div>

    <MediaRating
      class="popup__rating"
      variant="detail"
      :ratings="ratings ?? movie.ratings"
      :external-loading="ratingsLoading"
      :vote-average="movie.vote_average"
    />

    <div class="popup__meta" v-if="releaseYear">
      <span class="popup__chip">{{ releaseYear }}</span>
      <span v-if="displayRuntime" class="popup__chip">
        {{ fmtRuntime(displayRuntime) }}
      </span>
    </div>

    <p class="popup__overview">{{ truncate(movie.overview, 130) }}</p>

    <div class="popup__actions">
      <!-- 1. Watched Toggle Button -->
      <button
        class="action-btn action-btn--watched"
        :class="{ 'action-btn--active': isWatched }"
        :title="isWatched ? t('media.popup.watchedOnTitle') : t('media.popup.watchedOffTitle')"
        @click.stop="handleWatchedToggle"
      >
        <CheckCircle2 :size="15" :fill="isWatched ? 'currentColor' : 'none'" />
        <span>{{ isWatched ? t('media.popup.watched') : t('media.popup.notWatched') }}</span>
      </button>

      <!-- 2. Views Count Indicator -->
      <div class="action-btn action-btn--views" :title="t('media.popup.views')">
        <Eye :size="15" />
        <span>{{ fmtCount(stats.view_count) }}</span>
      </div>

      <!-- 3. Review Button -->
      <RouterLink
        :to="{
          name: detailRouteName,
          params: { id: movie.id },
          query: { action: 'review' },
        }"
        class="action-btn action-btn--review"
        :title="t('media.popup.reviews')"
      >
        <PenLine :size="15" />
        <span>{{ fmtCount(stats.review_count) }}</span>
      </RouterLink>

      <!-- 4. Favorite/Like Button -->
      <button
        class="action-btn action-btn--favorite"
        :class="{ 'action-btn--active': isLiked }"
        :title="isLiked ? t('media.popup.likedOnTitle') : t('media.popup.likedOffTitle')"
        @click.stop="handleLikeToggle"
      >
        <Heart :size="15" :fill="isLiked ? 'currentColor' : 'none'" />
        <span>{{ isLiked ? t('media.popup.liked') : t('media.popup.like') }}</span>
      </button>

      <!-- 5. Watchlist Button -->
      <button
        class="action-btn action-btn--watchlist"
        :class="{ 'action-btn--active': isInWatchlist }"
        :title="isInWatchlist ? t('media.popup.watchlistOnTitle') : t('media.popup.watchlistOffTitle')"
        @click.stop="handleWatchlistToggle"
      >
        <BookmarkCheck v-if="isInWatchlist" :size="15" fill="currentColor" />
        <BookmarkPlus v-else :size="15" fill="none" />
        <span>{{ isInWatchlist ? t('media.popup.inWatchlist') : t('media.popup.watchlist') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import { useI18n } from "vue-i18n"
import { logApiError } from "@/utils/logApiError"
import {
  computed,
  ref,
  onMounted,
  onBeforeUnmount,
  watch,
  nextTick,
} from "vue";
import {
  Film,
  Eye,
  CheckCircle2,
  PenLine,
  Heart,
  BookmarkPlus,
  BookmarkCheck,
  VolumeX,
} from "lucide-vue-next";
import { useImageUrl } from "@/composables/useImageUrl";
import type { Movie, TVSeries } from "@/types";
import type { ResolvedTrailer } from "@/composables/useTrailerPreview";
import { libraryApi } from "@/api/endpoints/library";
import { useAuthStore } from "@/stores/auth";
import { mediaApi } from "@/api/endpoints/media";
import MediaRating from "./MediaRating.vue";
import type { MediaRatings } from "@/types/movie";

const props = defineProps<{
  movie: Movie | TVSeries;
  /** Fuller ratings (with IMDb / RT) fetched from the detail endpoint; overrides `movie.ratings`. */
  ratings?: MediaRatings;
  /** Detail ratings are still being fetched; IMDb / RT show a skeleton instead of an em dash. */
  ratingsLoading?: boolean;
  // ── Trailer/player state: มาจาก useTrailerPreview instance ของการ์ดนี้ ──
  currentTrailer: ResolvedTrailer | null;
  trailerUnavailable: boolean;
  isIframeMounted: boolean;
  isIframeLoaded: boolean;
  showSkeleton: boolean;
  showFallback: boolean;
  attachPlayer: (elementId: string) => void;
  mediaType?: "movie" | "tv";
}>();

const { t, n } = useI18n();
const authStore = useAuthStore();
const currentUserId = computed(() => authStore.user?.id ?? null);

// ─── Derived media type (default: "movie") ────────────────────────────────
const mt = computed(() => props.mediaType ?? "movie");

// ─── Route name ───────────────────────────────────────────────────────────
const detailRouteName = computed(() =>
  mt.value === "tv" ? "tv-detail" : "movie-detail",
);

// ─── Derived display fields (รองรับทั้ง Movie และ TVSeries) ───────────────
// Movie and TVSeries name the same fields differently; read both shapes without `any`.
type MediaShape = Partial<{ title: string; name: string; release_date: string; first_air_date: string; runtime: number; episode_run_time: number[] }>;
const media = computed(() => props.movie as MediaShape);
type Counts = Partial<{ view_count: number; like_count: number; review_count: number; watchlist_count: number; liked_at: string | null; watched_at: string | null }>;
type StatsShape = Counts & { stats?: Counts | null };

const displayTitle = computed(() => media.value.title ?? media.value.name ?? t("media.popup.unknownTitle"));

const displayReleaseDate = computed(() => {
  return media.value.release_date ?? media.value.first_air_date ?? "";
});

const releaseYear = computed(() =>
  displayReleaseDate.value ? displayReleaseDate.value.slice(0, 4) : null,
);

const displayRuntime = computed(() => {
  const m = media.value;
  if (typeof m.runtime === "number") return m.runtime;
  if (Array.isArray(m.episode_run_time) && m.episode_run_time.length) {
    return m.episode_run_time[0];
  }
  return null;
});

// ─── YouTube Player mount ───────────────────────────────────────────────────
// unique element id ต่อการ์ด กันชนกันเวลามีหลายป็อปอัพ instance พร้อมกันใน DOM
const playerElId = `yt-player-${props.movie.id}-${Math.random()
  .toString(36)
  .slice(2)}`;

// เมื่อ isIframeMounted กลายเป็น true (ไม่ว่าจะครั้งแรก หรือ retry candidate ถัดไป)
// ให้เรียก attachPlayer เพื่อสร้าง YT.Player ผูกกับ div ของการ์ดนี้
watch(
  () => props.isIframeMounted,
  async (mounted) => {
    if (!mounted) return;
    await nextTick(); // รอให้ div :id="playerElId" render เสร็จก่อน
    props.attachPlayer(playerElId);
  },
);

onMounted(() => {
  if (props.isIframeMounted) {
    nextTick(() => props.attachPlayer(playerElId));
  }
});


// ─── Reactive Stats State ─────────────────────────────────────────────────
const stats = ref({
  view_count: 0,
  like_count: 0,
  review_count: 0,
  watchlist_count: 0,
});
const isLiked = ref(false);
const isInWatchlist = ref(false);
const watchlistItemId = ref<number | null>(null);
const isWatched = ref(false);
const watchedItemId = ref<number | null>(null);

// Populate stats and like status from props reactively
watch(
  () => props.movie as (Movie | TVSeries) & StatsShape,
  (newMovie) => {
    if (newMovie) {
      const s = newMovie.stats;
      stats.value = {
        view_count: s?.view_count ?? newMovie.view_count ?? 0,
        like_count: s?.like_count ?? newMovie.like_count ?? 0,
        review_count: s?.review_count ?? newMovie.review_count ?? 0,
        watchlist_count: s?.watchlist_count ?? newMovie.watchlist_count ?? 0,
      };
      isLiked.value = !!(s?.liked_at ?? newMovie.liked_at);
      isWatched.value = !!(s?.watched_at ?? newMovie.watched_at);
    }
  },
  { immediate: true }
);

onMounted(async () => {
  try {
    if (!currentUserId.value) return;

    const resLibrary = await libraryApi.getOwnMediaStatus(
      props.movie.id,
      mt.value,
    );
    const watchlistInfo = resLibrary.data?.in_lists?.find(
      (item) => item.list_type === "watchlist",
    );

    if (watchlistInfo) {
      isInWatchlist.value = true;
      watchlistItemId.value = watchlistInfo.item_id;
    }

    const watchedInfo = resLibrary.data?.in_lists?.find(
      (item) => item.list_type === "watched",
    );

    if (watchedInfo) {
      isWatched.value = true;
      watchedItemId.value = watchedInfo.item_id;
    }
  } catch (err) {
    logApiError("Failed to load library status", err);
  }
});

// ─── Watched Toggle ───────────────────────────────────────────────────────
async function handleWatchedToggle() {
  if (!currentUserId.value) {
    window.$toast?.warning(t("media.popup.toast.loginRequired"), t("media.popup.toast.loginRequiredTitle"));
    return;
  }

  try {
    if (isWatched.value && watchedItemId.value) {
      await libraryApi.removeItem(watchedItemId.value);
      stats.value.view_count = Math.max(0, stats.value.view_count - 1);
      isWatched.value = false;
      watchedItemId.value = null;
      window.$toast?.info(t("media.popup.toast.removedWatched"), displayTitle.value);
    } else {
      const res = await libraryApi.addItem({
        media_id: props.movie.id,
        media_type: mt.value,
        list_type: "watched",
        watched_at: new Date().toISOString().split("T")[0],
      });

      if (res.data?.item) {
        watchedItemId.value = res.data.item.id;
      }

      stats.value.view_count++;
      isWatched.value = true;
      window.$toast?.success(
        t("media.popup.toast.markedWatched"),
        displayTitle.value,
      );
    }
  } catch (err) {
    logApiError("Watched toggle failed", err);
    window.$toast?.error(t("media.popup.toast.saveFailed"), t("common.brand"));
  }
}

// ─── Like Toggle ──────────────────────────────────────────────────────────
async function handleLikeToggle() {
  try {
    if (isLiked.value) {
      await mediaApi.unlikeMedia(mt.value, props.movie.id);
      stats.value.like_count = Math.max(0, stats.value.like_count - 1);
      isLiked.value = false;
      window.$toast?.info(t("media.popup.toast.removedLike"), displayTitle.value);
    } else {
      await mediaApi.likeMedia(mt.value, props.movie.id);
      stats.value.like_count++;
      isLiked.value = true;
      window.$toast?.success(
        t("media.popup.toast.addedLike"),
        displayTitle.value,
      );
    }
  } catch (err) {
    logApiError("Like toggle failed", err);
    window.$toast?.error(t("media.popup.toast.genericError"), t("common.brand"));
  }
}

// ─── Watchlist Toggle ─────────────────────────────────────────────────────
async function handleWatchlistToggle() {
  if (!currentUserId.value) {
    window.$toast?.warning(t("media.popup.toast.loginRequired"), t("media.popup.toast.loginRequiredTitle"));
    return;
  }

  try {
    if (isInWatchlist.value && watchlistItemId.value) {
      await libraryApi.removeItem(watchlistItemId.value);
      stats.value.watchlist_count = Math.max(
        0,
        stats.value.watchlist_count - 1,
      );
      isInWatchlist.value = false;
      watchlistItemId.value = null;
      window.$toast?.info(t("media.popup.toast.removedWatchlist"), displayTitle.value);
    } else {
      const res = await libraryApi.addItem({
        media_id: props.movie.id,
        media_type: mt.value,
        list_type: "watchlist",
      });

      if (res.data?.item) {
        watchlistItemId.value = res.data.item.id;
      }

      stats.value.watchlist_count++;
      isInWatchlist.value = true;
      window.$toast?.success(
        t("media.popup.toast.addedWatchlist"),
        displayTitle.value,
      );
    }
  } catch (err) {
    logApiError("Watchlist toggle failed", err);
    window.$toast?.error(t("media.popup.toast.saveFailed"), t("common.brand"));
  }
}

// ─── Composables & Computed ───────────────────────────────────────────────
const { backdropImage } = useImageUrl();

const backdropUrl = computed(() => {
  const path = props.movie.backdrop_path;
  return path ? backdropImage(path) : null;
});

function truncate(str: string, len: number) {
  if (!str) return "";
  return str.length > len ? str.slice(0, len) + "…" : str;
}

function fmtCount(count: number) {
  return n(count, "compact");
}

function fmtRuntime(mins: number) {
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return h > 0
    ? t("media.detail.meta.runtimeHours", { h, m })
    : t("media.detail.meta.runtimeMinutes", { minutes: m });
}
</script>

<style scoped>
.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-decoration: none;
  gap: 3px;
  padding: 0.45rem 0.15rem;
  border: none;
  border-radius: 7px;
  cursor: pointer;
  font-size: 0.59rem;
  font-weight: 700;
  color: var(--color-text-primary);
  transition:
    filter 0.15s,
    transform 0.12s,
    background-color 0.2s;
}
.action-btn:hover {
  filter: brightness(1.2);
  transform: translateY(-1px);
}
.action-btn--active {
  box-shadow: inset 0 0 0 2px var(--color-border-strong);
  filter: saturate(1.4) brightness(1.1);
}

.popup__media {
  position: relative;
  aspect-ratio: 16 / 9;
  background: var(--color-surface);
  overflow: hidden;
}
.popup__backdrop {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center top;
}
.popup__no-media {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-surface) 0%, var(--color-surface-2) 100%);
}
.popup__no-media-icon {
  color: var(--color-surface-3);
}
.popup__iframe {
  position: absolute;
  inset: -1px;
  width: calc(100% + 2px);
  height: calc(100% + 2px);
  border: none;
  opacity: 0;
  transition: opacity 0.45s ease;
  pointer-events: none;
}
.popup__iframe--visible {
  opacity: 1;
}
.popup__media-fade {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 55%;
  background: linear-gradient(to bottom, transparent, var(--color-surface-2));
  pointer-events: none;
}
.popup__no-trailer {
  position: absolute;
  bottom: 8px;
  left: 10px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.62rem;
  color: var(--color-text-muted);
  z-index: 2;
}
.popup__badge {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 0.58rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  background: var(--color-scrim);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 3px 6px;
  backdrop-filter: blur(6px);
  z-index: 3;
}
.popup__badge-dot {
  display: inline-block;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--color-brand);
  animation: pulse-dot 1.8s ease-in-out infinite;
}
.popup__media-link {
  display: block;
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 4;
  cursor: pointer;
}
@keyframes pulse-dot {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.35;
    transform: scale(0.65);
  }
}
.popup__skeleton {
  position: absolute;
  inset: 0;
  background: var(--color-surface);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  z-index: 1;
}
.popup__info {
  padding: 0.85rem;
  background: color-mix(in srgb, var(--color-surface-2) 95%, transparent);
  backdrop-filter: blur(12px);
}
.popup__title-row {
  display: flex;
  align-items: flex-start;
  gap: 0.4rem;
  margin-bottom: 0.4rem;
}
.popup__title-link {
  cursor: pointer;
  color: inherit;
  text-decoration: none;
}
.popup__title-link:hover .popup__title {
  color: var(--color-text-primary);
}
.popup__title {
  flex: 1;
  font-size: 0.88rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
  line-height: 1.3;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  transition: color 0.15s;
}
.popup__rating {
  margin-bottom: 0.5rem;
}
.popup__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 0.5rem;
}
.popup__chip {
  font-size: 0.6rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  background: var(--color-border);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 2px 6px;
}
.popup__overview {
  font-size: 0.7rem;
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin: 0 0 0.75rem;
}
.popup__actions {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
}

.action-btn {
  /* --tone drives the hover/active tint of each action */
  --tone: var(--color-text-secondary);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-decoration: none;
  gap: 3px;
  padding: 0.45rem 0.15rem;
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.58rem;
  font-weight: 700;
  backdrop-filter: blur(8px);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
}
.action-btn:hover {
  transform: translateY(-2px);
}

.action-btn--watched,
.action-btn--favorite,
.action-btn--watchlist {
  background: var(--color-border);
  border-color: var(--color-border-strong);
  color: var(--color-text-secondary);
}
.action-btn--watched:hover,
.action-btn--favorite:hover,
.action-btn--watchlist:hover {
  background: color-mix(in srgb, var(--tone) 18%, transparent);
  border-color: color-mix(in srgb, var(--tone) 45%, transparent);
  color: var(--tone);
  box-shadow: 0 4px 12px color-mix(in srgb, var(--tone) 20%, transparent);
}
.action-btn--watched.action-btn--active,
.action-btn--favorite.action-btn--active,
.action-btn--watchlist.action-btn--active {
  background: color-mix(in srgb, var(--tone) 28%, transparent);
  border-color: color-mix(in srgb, var(--tone) 85%, transparent);
  color: var(--tone);
  box-shadow: 0 0 12px color-mix(in srgb, var(--tone) 45%, transparent);
}
.action-btn--watched {
  --tone: var(--color-info);
}
.action-btn--favorite {
  --tone: var(--color-danger);
}
.action-btn--watchlist {
  --tone: var(--color-warning);
}

.action-btn--views {
  --tone: var(--color-info);
  background: color-mix(in srgb, var(--tone) 12%, transparent);
  border-color: color-mix(in srgb, var(--tone) 22%, transparent);
  color: var(--tone);
  cursor: default;
}
.action-btn--views:hover {
  transform: none;
}

.action-btn--review {
  --tone: var(--color-success);
  background: color-mix(in srgb, var(--tone) 12%, transparent);
  border-color: color-mix(in srgb, var(--tone) 25%, transparent);
  color: var(--tone);
}
.action-btn--review:hover {
  background: color-mix(in srgb, var(--tone) 22%, transparent);
  border-color: color-mix(in srgb, var(--tone) 45%, transparent);
  box-shadow: 0 4px 12px color-mix(in srgb, var(--tone) 25%, transparent);
}
</style>
