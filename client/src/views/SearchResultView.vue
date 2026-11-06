<template>
  <div class="search-page">
    <div class="search-page__inner">
      <!-- ── Header ── -->
      <div class="search-header">
        <h1 class="search-title">
          <i class="pi pi-search"></i>
          <span v-if="queryText">ผลการค้นหา: "{{ queryText }}"</span>
          <span v-else>ค้นหาภาพยนตร์และซีรีส์</span>
        </h1>

        <p class="search-subtitle" v-if="!isLoading">
          <template v-if="hasResults">
            พบ {{ filteredResults.length.toLocaleString() }} รายการ
          </template>
          <template v-else-if="queryText">
            ไม่พบผลลัพธ์ที่ตรงกับคำค้นหาของคุณ
          </template>
        </p>
      </div>

      <!-- ── Filter Tabs ── -->
      <div class="filter-tabs" v-if="!isLoading && rawResults.length">
        <button
          class="filter-tab"
          :class="{ active: activeFilter === 'all' }"
          @click="activeFilter = 'all'"
        >
          ทั้งหมด
          <span class="tab-count">{{ rawResults.length }}</span>
        </button>
        <button
          class="filter-tab"
          :class="{ active: activeFilter === 'movie' }"
          @click="activeFilter = 'movie'"
        >
          <i class="pi pi-video"></i>
          ภาพยนตร์
          <span class="tab-count">{{ movieCount }}</span>
        </button>
        <button
          class="filter-tab"
          :class="{ active: activeFilter === 'tv' }"
          @click="activeFilter = 'tv'"
        >
          <i class="pi pi-desktop"></i>
          ซีรีส์
          <span class="tab-count">{{ tvCount }}</span>
        </button>
      </div>

      <!-- ── Loading ── -->
      <div class="loading-state" v-if="isLoading">
        <i class="pi pi-spin pi-spinner spinner-icon"></i>
        <p>กำลังค้นหา...</p>
      </div>

      <!-- ── Results Grid ── -->
      <template v-else>
        <div class="media-grid" v-if="filteredResults.length">
          <div
            v-for="(item, i) in filteredResults"
            :key="`${item.mediaType}-${item.id}`"
            class="media-card-wrap"
            @mouseenter="onCardEnter(item)"
            @mouseleave="onCardLeave(item.id)"
          >
            <RouterLink
              :to="
                item.mediaType === 'movie'
                  ? `/movies/${item.id}`
                  : `/tv/${item.id}`
              "
              class="media-card"
            >
              <div class="media-poster">
                <img
                  v-if="item.posterPath"
                  :src="getTmdbImageUrlOrPlaceholder(item.posterPath)"
                  :alt="item.title"
                  loading="lazy"
                />
                <div v-else class="media-poster--empty">
                  <i class="pi pi-image"></i>
                </div>

                <span class="media-type-badge" :class="item.mediaType">
                  <i
                    :class="
                      item.mediaType === 'movie'
                        ? 'pi pi-video'
                        : 'pi pi-desktop'
                    "
                  ></i>
                  {{ item.mediaType === "movie" ? "หนัง" : "ซีรีส์" }}
                </span>

                <span class="media-rating" v-if="item.voteAverage > 0">
                  <i class="pi pi-star-fill"></i>
                  {{ item.voteAverage.toFixed(1) }}
                </span>

                <div class="media-overlay">
                  <span class="overlay-play"
                    ><i class="pi pi-play-circle"></i
                  ></span>
                </div>
              </div>

              <div class="media-info">
                <span class="media-title">{{ item.title }}</span>
                <span class="media-year" v-if="item.year">{{ item.year }}</span>
              </div>
            </RouterLink>

            <Transition name="popup">
              <div
                v-if="hoveredId === item.id"
                class="hover-popup"
                :class="getPopupPos(i)"
                @mouseenter="onPopupEnter"
                @mouseleave="onPopupLeave(item.id)"
              >
                <PopupCard
                  :movie="item.raw"
                  :media-type="'movie'"
                  :current-trailer="getState(item.id).currentTrailer.value"
                  :trailer-unavailable="
                    getState(item.id).trailerUnavailable.value
                  "
                  :is-iframe-mounted="getState(item.id).isIframeMounted.value"
                  :is-iframe-loaded="getState(item.id).isIframeLoaded.value"
                  :show-skeleton="getState(item.id).showSkeleton.value"
                  :show-fallback="getState(item.id).showFallback.value"
                  :attach-player="getState(item.id).attachPlayer"
                />
              </div>
            </Transition>
          </div>
        </div>

        <!-- ── Empty State + Suggestions ── -->
        <div class="empty-block" v-else>
          <div class="empty-state">
            <i class="pi pi-inbox empty-icon"></i>
            <h2>ไม่พบผลลัพธ์{{ queryText ? ` สำหรับ "${queryText}"` : "" }}</h2>
            <p>ลองค้นหาด้วยคำอื่น หรือดูรายการยอดนิยมด้านล่างแทน</p>
          </div>

          <div class="suggestion-section" v-if="suggestions.length">
            <h3 class="suggestion-title">
              <i class="pi pi-sparkles"></i>
              อาจถูกใจคุณ
            </h3>
            <div class="media-grid">
              <div
                v-for="(item, i) in suggestions"
                :key="`sugg-${item.mediaType}-${item.id}`"
                class="media-card-wrap"
                @mouseenter="onCardEnter(item)"
                @mouseleave="onCardLeave(item.id)"
              >
                <RouterLink
                  :to="
                    item.mediaType === 'movie'
                      ? `/movies/${item.id}`
                      : `/tv/${item.id}`
                  "
                  class="media-card"
                >
                  <div class="media-poster">
                    <img
                      v-if="item.posterPath"
                      :src="getTmdbImageUrlOrPlaceholder(item.posterPath)"
                      :alt="item.title"
                      loading="lazy"
                    />
                    <div v-else class="media-poster--empty">
                      <i class="pi pi-image"></i>
                    </div>

                    <span class="media-type-badge" :class="item.mediaType">
                      <i
                        :class="
                          item.mediaType === 'movie'
                            ? 'pi pi-video'
                            : 'pi pi-desktop'
                        "
                      ></i>
                      {{ item.mediaType === "movie" ? "หนัง" : "ซีรีส์" }}
                    </span>

                    <span class="media-rating" v-if="item.voteAverage > 0">
                      <i class="pi pi-star-fill"></i>
                      {{ item.voteAverage.toFixed(1) }}
                    </span>

                    <div class="media-overlay">
                      <span class="overlay-play"
                        ><i class="pi pi-play-circle"></i
                      ></span>
                    </div>
                  </div>

                  <div class="media-info">
                    <span class="media-title">{{ item.title }}</span>
                    <span class="media-year" v-if="item.year">{{
                      item.year
                    }}</span>
                  </div>
                </RouterLink>

                <Transition name="popup">
                  <div
                    v-if="hoveredId === item.id"
                    class="hover-popup"
                    :class="getPopupPos(i)"
                    @mouseenter="onPopupEnter"
                    @mouseleave="onPopupLeave(item.id)"
                  >
                    <PopupCard
                      :movie="item.raw"
                      :media-type="item.mediaType"
                      :current-trailer="getState(item.id).currentTrailer.value"
                      :trailer-unavailable="
                        getState(item.id).trailerUnavailable.value
                      "
                      :is-iframe-mounted="
                        getState(item.id).isIframeMounted.value
                      "
                      :is-iframe-loaded="getState(item.id).isIframeLoaded.value"
                      :show-skeleton="getState(item.id).showSkeleton.value"
                      :show-fallback="getState(item.id).showFallback.value"
                      :attach-player="getState(item.id).attachPlayer"
                    />
                  </div>
                </Transition>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useRoute } from "vue-router";
import { movieApi } from "@/api/api";
import { getTmdbImageUrlOrPlaceholder } from "@/utils/image";
import PopupCard from "@/components/movie/PopupCard.vue";
import { useHoverPreviewGrid } from "@/composables/useHoverPreviewGrid";
import type { Movie, TVSeries } from "@/types";

type MediaType = "movie" | "tv";

interface SearchItem {
  id: number;
  mediaType: MediaType;
  title: string;
  posterPath: string | null;
  voteAverage: number;
  year: string;
  raw: Movie | TVSeries;
}

const route = useRoute();
const queryText = computed(() => (route.query.q as string) || "");
const isLoading = ref(false);
const rawResults = ref<SearchItem[]>([]);
const suggestions = ref<SearchItem[]>([]);
const activeFilter = ref<"all" | "movie" | "tv">("all");
const hasResults = computed(() => rawResults.value.length > 0);

const movieCount = computed(
  () => rawResults.value.filter((i) => i.mediaType === "movie").length,
);
const tvCount = computed(
  () => rawResults.value.filter((i) => i.mediaType === "tv").length,
);

const filteredResults = computed(() => {
  if (activeFilter.value === "all") return rawResults.value;
  return rawResults.value.filter((i) => i.mediaType === activeFilter.value);
});

const {
  hoveredId,
  getState,
  getTrailer,
  onCardEnter,
  onCardLeave,
  onPopupEnter,
  onPopupLeave,
  getPopupPos,
} = useHoverPreviewGrid({ columns: 5 });

function normalizeMovie(m: any): SearchItem {
  return {
    id: m.id,
    mediaType: "movie",
    title: m.title || m.original_title || "ไม่ทราบชื่อ",
    posterPath: m.poster_path || null,
    voteAverage: m.vote_average || 0,
    year: m.release_date ? m.release_date.split("-")[0] : "",
    raw: m,
  };
}

function normalizeSeries(s: any): SearchItem {
  return {
    id: s.id,
    mediaType: "tv",
    title: s.name || s.original_name || "ไม่ทราบชื่อ",
    posterPath: s.poster_path || null,
    voteAverage: s.vote_average || 0,
    year: s.first_air_date ? s.first_air_date.split("-")[0] : "",
    raw: s,
  };
}

function sortByRelevance(items: SearchItem[]): SearchItem[] {
  return [...items].sort((a, b) => b.voteAverage - a.voteAverage);
}

async function performSearch(query: string) {
  isLoading.value = true;
  rawResults.value = [];
  suggestions.value = [];
  activeFilter.value = "all";

  if (!query.trim()) {
    isLoading.value = false;
    return;
  }

  try {
    const [movieRes, seriesRes] = await Promise.allSettled([
      movieApi.search(query),
      movieApi.searchSeries(query),
    ]);

    const movies =
      movieRes.status === "fulfilled"
        ? movieRes.value.data.results.map(normalizeMovie)
        : [];
    const series =
      seriesRes.status === "fulfilled"
        ? seriesRes.value.data.results.map(normalizeSeries)
        : [];

    rawResults.value = sortByRelevance([...movies, ...series]);

    if (rawResults.value.length === 0) {
      await loadSuggestions();
    }
  } catch (err) {
    console.error("ค้นหาล้มเหลว:", err);
    await loadSuggestions();
  } finally {
    isLoading.value = false;
  }
}

async function loadSuggestions() {
  try {
    const [popMovies, popSeries] = await Promise.allSettled([
      movieApi.getPopular(1),
      movieApi.getPopularSeries(1),
    ]);

    const movies =
      popMovies.status === "fulfilled"
        ? popMovies.value.data.results.slice(0, 10).map(normalizeMovie)
        : [];
    const series =
      popSeries.status === "fulfilled"
        ? popSeries.value.data.results.slice(0, 10).map(normalizeSeries)
        : [];

    const merged: SearchItem[] = [];
    const maxLen = Math.max(movies.length, series.length);

    for (let i = 0; i < maxLen; i++) {
      const movie = movies[i];
      const show = series[i];
      if (movie) merged.push(movie);
      if (show) merged.push(show);
    }
    suggestions.value = merged;
  } catch (err) {
    console.error("โหลดรายการแนะนำล้มเหลว:", err);
  }
}

watch(
  () => route.query.q,
  (newQ) => {
    performSearch((newQ as string) || "");
  },
);

onMounted(() => {
  performSearch(queryText.value);
});
</script>

<style scoped>
.search-page {
  min-height: calc(100vh - 64px);
  background: #141414;
  color: #f0f0f0;
  font-family: "Noto Sans Thai", sans-serif;
}

.search-page__inner {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem 2rem 4rem;
  box-sizing: border-box;
}

/* ── Header ── */
.search-header {
  margin-bottom: 1.5rem;
}

.search-title {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: clamp(1.3rem, 2.5vw, 1.9rem);
  font-weight: 700;
  color: #fff;
  margin: 0 0 0.4rem;
  overflow-wrap: break-word;
}
.search-title i {
  color: var(--color-brand);
  font-size: 0.9em;
  flex-shrink: 0;
}

.search-subtitle {
  font-size: 0.9rem;
  color: #a3a3a3;
  margin: 0;
}

/* ── Filter Tabs ── */
.filter-tabs {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
  margin-bottom: 1.75rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 1rem;
}

.filter-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #a3a3a3;
  font-size: 0.85rem;
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 9999px;
  cursor: pointer;
  transition:
    background 0.2s,
    border-color 0.2s,
    color 0.2s;
  white-space: nowrap;
}
.filter-tab:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}
.filter-tab.active {
  background: rgba(245, 67, 44, 0.15);
  border-color: rgba(245, 67, 44, 0.4);
  color: #ff3b47;
}
.tab-count {
  background: rgba(255, 255, 255, 0.1);
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.1rem 0.5rem;
  border-radius: 9999px;
}
.filter-tab.active .tab-count {
  background: rgba(245, 67, 44, 0.3);
}

/* ── Loading ── */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 5rem 1rem;
  color: #a3a3a3;
}
.spinner-icon {
  font-size: 2rem;
  color: var(--color-brand);
}

/* ── Media Grid (responsive, auto-fill) ── */
.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 1.5rem;
}

.media-card {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  text-decoration: none;
  color: inherit;
  cursor: pointer;
}

.media-poster {
  position: relative;
  aspect-ratio: 2/3;
  border-radius: 10px;
  overflow: hidden;
  background: #1c1c1c;
  border: 1px solid rgba(255, 255, 255, 0.06);
  transition:
    transform 0.25s ease,
    box-shadow 0.25s ease;
}
.media-card:hover .media-poster {
  transform: translateY(-4px) scale(1.02);
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.55);
}
.media-poster img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.media-poster--empty {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #555;
  font-size: 2rem;
}

.media-type-badge {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0.22rem 0.5rem;
  border-radius: 6px;
  backdrop-filter: blur(6px);
  color: #fff;
}
.media-type-badge.movie {
  background: rgba(245, 67, 44, 0.75);
}
.media-type-badge.tv {
  background: rgba(46, 108, 255, 0.75);
}

.media-rating {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.7rem;
  font-weight: 700;
  background: rgba(0, 0, 0, 0.6);
  color: #f5c518;
  padding: 0.22rem 0.5rem;
  border-radius: 6px;
  backdrop-filter: blur(6px);
}

.media-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.35);
  opacity: 0;
  transition: opacity 0.2s;
}
.media-card:hover .media-overlay {
  opacity: 1;
}
.overlay-play {
  font-size: 2.2rem;
  color: #fff;
}

.media-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  padding: 0 0.1rem;
}
.media-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.35;
  min-height: 2.4em;
}
.media-year {
  font-size: 0.75rem;
  color: #8a8a8a;
}

/* ── Empty State ── */
.empty-block {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.5rem;
  padding: 3rem 1rem 1rem;
}
.empty-icon {
  font-size: 3rem;
  color: #444;
  margin-bottom: 0.5rem;
}
.empty-state h2 {
  font-size: 1.15rem;
  font-weight: 600;
  color: #fff;
  margin: 0;
}
.empty-state p {
  font-size: 0.88rem;
  color: #8a8a8a;
  margin: 0;
}

.suggestion-section {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}
.suggestion-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: #fff;
  margin: 0;
}
.suggestion-title i {
  color: #f5c518;
}

/* ============================================================
   RESPONSIVE BREAKPOINTS
   ============================================================ */

/* ── Large desktop: ปรับขนาดขั้นต่ำของการ์ดให้ใหญ่ขึ้นเพื่อไม่ให้เยอะเกิน ── */
@media (min-width: 1600px) {
  .media-grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 1.75rem;
  }
}

/* ── Tablet landscape: 1024px ── */
@media (max-width: 1024px) {
  .search-page__inner {
    padding: 1.75rem 1.5rem 3.5rem;
  }
  .media-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 1.25rem;
  }
}

/* ── Tablet portrait / large phone landscape: 768px ── */
@media (max-width: 768px) {
  .search-page__inner {
    padding: 1.5rem 1.25rem 3rem;
  }
  .filter-tabs {
    gap: 0.5rem;
    margin-bottom: 1.4rem;
    padding-bottom: 0.85rem;
  }
  .filter-tab {
    font-size: 0.8rem;
    padding: 0.45rem 0.85rem;
  }
  .media-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 1rem;
  }
  .media-title {
    font-size: 0.82rem;
  }
}

/* ── Phone (portrait): 640px ── */
@media (max-width: 640px) {
  .search-page__inner {
    padding: 1.25rem 1rem 2.5rem;
  }
  .search-title {
    font-size: 1.2rem;
  }
  .search-subtitle {
    font-size: 0.82rem;
  }
  .media-grid {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 0.85rem;
  }
  .media-title {
    font-size: 0.78rem;
    -webkit-line-clamp: 2;
  }
  .media-year {
    font-size: 0.7rem;
  }
  .empty-state {
    padding: 2rem 0.5rem 0.5rem;
  }
  .empty-icon {
    font-size: 2.4rem;
  }
  .empty-state h2 {
    font-size: 1rem;
  }
}

/* ── Small phone: 480px ── */
@media (max-width: 480px) {
  .search-page__inner {
    padding: 1rem 0.75rem 2rem;
  }
  .filter-tabs {
    gap: 0.4rem;
  }
  .filter-tab {
    font-size: 0.75rem;
    padding: 0.4rem 0.7rem;
  }
  .filter-tab i {
    display: none; /* ประหยัดพื้นที่บนจอแคบ */
  }
  .media-grid {
    /* 2 คอลัมน์คงที่บนจอมือถือแคบสุด ให้การ์ดไม่เล็กจนอ่านไม่ออก */
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }
  .media-type-badge {
    font-size: 0.6rem;
    padding: 0.18rem 0.4rem;
  }
  .media-rating {
    font-size: 0.65rem;
    padding: 0.18rem 0.4rem;
  }
  .overlay-play {
    font-size: 1.8rem;
  }
  .suggestion-title {
    font-size: 0.95rem;
  }
}

/* ── Extra small phone: ≤360px ── */
@media (max-width: 360px) {
  .search-page__inner {
    padding: 0.85rem 0.6rem 1.75rem;
  }
  .search-title {
    font-size: 1.05rem;
    gap: 0.45rem;
  }
  .media-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.6rem;
  }
  .media-title {
    font-size: 0.72rem;
    min-height: 2.1em;
  }
  .media-year {
    font-size: 0.65rem;
  }
}

.media-card-wrap {
  position: relative;
}

.hover-popup {
  position: absolute;
  top: 0;
  z-index: 50;
  width: 280px;
  background: #1c1c1c;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.8);
}
.popup--right {
  left: calc(100% + 8px);
}
.popup--left {
  right: calc(100% + 8px);
}
.popup--center {
  left: 50%;
  transform: translateX(-50%);
}
.popup-enter-active {
  transition:
    opacity 0.15s,
    transform 0.15s;
}
.popup-leave-active {
  transition: opacity 0.1s;
}
.popup-enter-from {
  opacity: 0;
  transform: translateY(6px) scale(0.97);
}
.popup-leave-to {
  opacity: 0;
}
</style>
