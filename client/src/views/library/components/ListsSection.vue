<template>
  <section class="list-section fade-in-up">
    <!-- Advanced Filter Bar -->
    <div class="filter-bar">
      <!-- Title Search -->
      <div class="filter-input-wrapper search-wrapper">
        <Search :size="16" class="filter-icon" />
        <input
          v-model="filters.search"
          type="text"
          :placeholder="$t('library.filters.search')"
          class="filter-input"
        />
      </div>

      <!-- Media Type Select -->
      <div class="filter-select-wrapper">
        <Film :size="14" class="select-icon" />
        <select v-model="filters.mediaType" class="filter-select">
          <option value="all">{{ $t("library.filters.allMedia") }}</option>
          <option value="movie">{{ $t("library.filters.movies") }}</option>
          <option value="tv">{{ $t("library.filters.tvSeries") }}</option>
        </select>
      </div>

      <!-- Genre Filter -->
      <div class="filter-select-wrapper">
        <ArrowUpDown :size="14" class="select-icon" />
        <select v-model="filters.genre" class="filter-select">
          <option value="all">{{ $t("library.filters.allGenres") }}</option>
          <option
            v-for="genre in availableGenres"
            :key="genre"
            :value="genre"
          >{{ genre }}</option>
        </select>
      </div>

      <!-- Sort Order -->
      <div class="filter-select-wrapper">
        <ArrowUpDown :size="14" class="select-icon" />
        <select v-model="filters.sortBy" class="filter-select">
          <option value="newest">{{ $t("library.filters.sort.newest") }}</option>
          <option value="oldest">{{ $t("library.filters.sort.oldest") }}</option>
          <option v-if="activeTab === 'reviews'" value="rating_high">{{ $t("library.filters.sort.ratingHigh") }}</option>
          <option v-if="activeTab === 'reviews'" value="rating_low">{{ $t("library.filters.sort.ratingLow") }}</option>
          <option value="title_az">{{ $t("library.filters.sort.titleAZ") }}</option>
          <option value="title_za">{{ $t("library.filters.sort.titleZA") }}</option>
        </select>
      </div>
    </div>

    <!-- Render List Items -->
    <!-- Tab: Reviews -->
    <div v-if="activeTab === 'reviews'" class="reviews-list">
      <div
        v-for="review in filteredReviews"
        :key="review.id"
        class="review-library-card"
      >
        <router-link
          :to="review.media.media_type === 'tv' ? `/tv/${review.media.id}` : `/movies/${review.media.id}`"
          class="review-media-poster"
        >
          <img
            v-if="review.media.poster_url"
            :src="`${TMDB_IMG}${review.media.poster_url}`"
            :alt="review.media.title"
            loading="lazy"
          />
          <div v-else class="poster-fallback">
            <FilmIcon :size="18" />
          </div>
        </router-link>
        <div class="review-card-content">
          <div class="review-card-header">
            <router-link
              :to="review.media.media_type === 'tv' ? `/tv/${review.media.id}` : `/movies/${review.media.id}`"
              class="review-media-title"
            >
              {{ review.media.title }}
            </router-link>
            <div class="review-card-stars">
              <span class="rating-num">{{ review.rating.toFixed(1) }}</span>
              <Star :size="12" class="stars-icon" />
            </div>
          </div>
          <p class="review-card-body">{{ review.body }}</p>
          <div class="review-card-footer">
            <span class="review-date">{{ formatDate(review.created_at) }}</span>
            <div class="review-meta-actions">
              <span class="action-count"><Heart :size="12" class="liked-heart" /> {{ review.like_count }}</span>
            </div>
          </div>
        </div>
      </div>
      <div v-if="filteredReviews.length === 0" class="empty-state-card fade-in-up">
        <div class="empty-state-visual">
          <div class="empty-glow-backdrop" />
          <div class="empty-icon-badge">
            <Star :size="36" class="empty-trophy-icon" />
          </div>
        </div>
        <h3 class="empty-state-title">{{ $t("library.noReviewsTitle") || 'ยังไม่มีการเขียนรีวิว' }}</h3>
        <p class="empty-state-description">{{ $t("library.empty_reviews") }}</p>
      </div>
    </div>

    <!-- Tab: Library Items (Watchlist, Watched, Likes) -->
    <div v-else class="media-grid">
      <div
        v-for="item in filteredLibraryItems"
        :key="item.id"
        class="media-library-card"
      >
        <router-link
          :to="item.media.media_type === 'tv' ? `/tv/${item.media.id}` : `/movies/${item.media.id}`"
          class="media-poster-frame"
        >
          <img
            v-if="item.media.poster_url"
            :src="`${TMDB_IMG}${item.media.poster_url}`"
            :alt="item.media.title"
            loading="lazy"
          />
          <div v-else class="poster-fallback">
            <FilmIcon :size="24" />
          </div>
          <span class="media-type-chip">{{ item.media.media_type === 'tv' ? 'TV' : 'Movie' }}</span>
        </router-link>
        <div class="media-card-meta">
          <router-link
            :to="item.media.media_type === 'tv' ? `/tv/${item.media.id}` : `/movies/${item.media.id}`"
            class="media-title"
          >
            {{ item.media.title }}
          </router-link>
          <div class="media-details-row">
            <span class="media-rating-badge"><MediaRating :ratings="item.media.ratings" :vote-average="item.media.vote_average" /></span>
            <span class="save-date">{{ formatDate(item.created_at) }}</span>
          </div>
          <!-- User tags if present -->
          <div v-if="item.tags && item.tags.length > 0" class="tags-wrapper">
            <span v-for="t in item.tags" :key="t" class="tag-badge">{{ t }}</span>
          </div>

          <!-- Watch Log Bar for Watched Tab -->
          <div v-if="activeTab === 'watched'" class="watch-log-bar">
            <button
              class="watch-count-chip"
              :title="$t('watchLog.watchHistory')"
              @click.prevent="openWatchHistory(item)"
            >
              <Eye :size="12" />
              <span>{{ getWatchCountText(item) }}</span>
            </button>
            <button
              v-if="isOwner"
              class="btn-log-watch-mini"
              :title="$t('watchLog.recordAnother')"
              @click.prevent="openWatchLogModal(item)"
            >
              <Plus :size="12" />
              <span>{{ $t('watchLog.recordWatch') }}</span>
            </button>
          </div>
        </div>
      </div>
      <div v-if="filteredLibraryItems.length === 0" class="empty-state-card fade-in-up">
        <div class="empty-state-visual">
          <div class="empty-glow-backdrop" />
          <div class="empty-icon-badge">
            <component
              :is="activeTab === 'watchlist' ? Bookmark : activeTab === 'likes' ? Heart : Film"
              :size="36"
              class="empty-trophy-icon"
            />
          </div>
        </div>
        <h3 class="empty-state-title">
          {{ activeTab === 'watchlist' ? 'ยังไม่มีรายการที่ต้องการดู' : activeTab === 'likes' ? 'ยังไม่มีรายการที่ถูกใจ' : 'ยังไม่มีรายการที่รับชมแล้ว' }}
        </h3>
        <p class="empty-state-description">{{ $t("library.empty") }}</p>
      </div>
    </div>

    <!-- Modals for Watch Log & History -->
    <WatchLogModal
      v-if="showLogModal && activeLogMedia"
      :media-id="activeLogMedia.id"
      :media-type="activeLogMedia.type"
      @close="showLogModal = false"
      @logged="onWatchLogged"
    />

    <WatchHistoryPanel
      v-if="showHistoryPanel && activeHistoryMedia"
      :media-id="activeHistoryMedia.id"
      :media-type="activeHistoryMedia.type"
      @close="showHistoryPanel = false"
      @changed="onWatchHistoryChanged"
    />
  </section>
</template>

<script setup lang="ts">
import MediaRating from "@/components/movie/MediaRating.vue"
import { ref, computed, watch, onMounted } from "vue"
import { useI18n } from "vue-i18n"
import { Search, Film, ArrowUpDown, Heart, Star, Bookmark, Film as FilmIcon, Eye, Plus, History } from "lucide-vue-next"
import type { LibraryItemResponse, ReviewResponse, LibraryListTab } from "@/types"
import { watchLogApi } from "@/api/endpoints/watchLog"
import WatchLogModal from "@/components/movie/WatchLogModal.vue"
import WatchHistoryPanel from "@/components/movie/WatchHistoryPanel.vue"

const { t, locale } = useI18n()

function formatDate(dateStr: string): string {
  if (!dateStr) return ""
  const d = new Date(dateStr)
  return d.toLocaleDateString(locale.value === "th" ? "th-TH" : "en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

const props = defineProps<{
  activeTab: LibraryListTab
  userId: number
  isOwner: boolean
  watchlistItems: LibraryItemResponse[]
  watchedItems: LibraryItemResponse[]
  likedItems: LibraryItemResponse[]
  userReviews: ReviewResponse[]
}>()

const TMDB_IMG = "https://image.tmdb.org/t/p/w342"

// Filter States
const filters = ref({
  search: "",
  mediaType: "all",
  genre: "all",
  sortBy: "newest"
})

function resetFilters() {
  filters.value.search = ""
  filters.value.mediaType = "all"
  filters.value.genre = "all"
  filters.value.sortBy = "newest"
}

// Reset filters on activeTab or userId changes
watch([() => props.activeTab, () => props.userId], resetFilters)

// Dynamic Genres for Dropdown Filter
const availableGenres = computed(() => {
  const genresMap = new Set<string>()
  const items =
    props.activeTab === "reviews"
      ? props.userReviews.map(r => r.media)
      : props.activeTab === "watchlist"
      ? props.watchlistItems.map(i => i.media)
      : props.activeTab === "watched"
      ? props.watchedItems.map(i => i.media)
      : props.likedItems.map(i => i.media)

  items.forEach(media => {
    if (media && media.genres) {
      media.genres.forEach(g => genresMap.add(g.name))
    }
  })

  return Array.from(genresMap).sort()
})

// Filter & Sort Logic for Library Items
const filteredLibraryItems = computed(() => {
  let list =
    props.activeTab === "watchlist"
      ? props.watchlistItems
      : props.activeTab === "watched"
      ? props.watchedItems
      : props.likedItems

  // Deduplicate items by media id and media type
  const seen = new Set<string>()
  const uniqueList: typeof list = []
  for (const item of list) {
    if (!item.media) continue
    const key = `${item.media.media_type}_${item.media.id}`
    if (!seen.has(key)) {
      seen.add(key)
      uniqueList.push(item)
    }
  }
  list = uniqueList

  // 1. Search Query
  if (filters.value.search.trim()) {
    const query = filters.value.search.toLowerCase()
    list = list.filter(item => item.media.title.toLowerCase().includes(query))
  }

  // 2. Media Type
  if (filters.value.mediaType !== "all") {
    list = list.filter(item => item.media.media_type === filters.value.mediaType)
  }

  // 3. Genre
  if (filters.value.genre !== "all") {
    list = list.filter(
      item => item.media.genres && item.media.genres.some(g => g.name === filters.value.genre)
    )
  }

  // 4. Sort
  const sorted = [...list]
  if (filters.value.sortBy === "newest") {
    sorted.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  } else if (filters.value.sortBy === "oldest") {
    sorted.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
  } else if (filters.value.sortBy === "title_az") {
    sorted.sort((a, b) => a.media.title.localeCompare(b.media.title))
  } else if (filters.value.sortBy === "title_za") {
    sorted.sort((a, b) => b.media.title.localeCompare(a.media.title))
  }

  return sorted
})

// Filter & Sort Logic for Reviews
const filteredReviews = computed(() => {
  let list = props.userReviews

  // 1. Search Query
  if (filters.value.search.trim()) {
    const query = filters.value.search.toLowerCase()
    list = list.filter(r => r.media.title.toLowerCase().includes(query))
  }

  // 2. Media Type
  if (filters.value.mediaType !== "all") {
    list = list.filter(r => r.media.media_type === filters.value.mediaType)
  }

  // 3. Genre
  if (filters.value.genre !== "all") {
    list = list.filter(
      r => r.media.genres && r.media.genres.some(g => g.name === filters.value.genre)
    )
  }

  // 4. Sort
  const sorted = [...list]
  if (filters.value.sortBy === "newest") {
    sorted.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  } else if (filters.value.sortBy === "oldest") {
    sorted.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
  } else if (filters.value.sortBy === "rating_high") {
    sorted.sort((a, b) => b.rating - a.rating)
  } else if (filters.value.sortBy === "rating_low") {
    sorted.sort((a, b) => a.rating - b.rating)
  } else if (filters.value.sortBy === "title_az") {
    sorted.sort((a, b) => a.media.title.localeCompare(b.media.title))
  } else if (filters.value.sortBy === "title_za") {
    sorted.sort((a, b) => b.media.title.localeCompare(a.media.title))
  }

  return sorted
})

// Watch Log & History Management for Watched Tab
const watchSummaries = ref<Record<string, number>>({})
const showLogModal = ref(false)
const activeLogMedia = ref<{ id: number; type: "movie" | "tv" } | null>(null)
const showHistoryPanel = ref(false)
const activeHistoryMedia = ref<{ id: number; type: "movie" | "tv" } | null>(null)

function getWatchMediaKey(mediaId: number, mediaType: "movie" | "tv"): string {
  return `${mediaType}_${mediaId}`
}

function getWatchCount(item: LibraryItemResponse): number {
  const key = getWatchMediaKey(item.media.id, item.media.media_type)
  return watchSummaries.value[key] ?? 1
}

function getWatchCountText(item: LibraryItemResponse): string {
  const count = getWatchCount(item)
  if (count === 1) {
    return t("watchLog.watchedOnce")
  }
  return t("watchLog.watchedTimes", { count })
}

async function fetchWatchSummaryForItem(mediaId: number, mediaType: "movie" | "tv") {
  try {
    const res = await watchLogApi.getMyWatchLogs(mediaType, mediaId)
    const key = getWatchMediaKey(mediaId, mediaType)
    if (res.data) {
      const count = res.data.summary?.watch_count ?? res.data.logs?.length
      if (count !== undefined) {
        watchSummaries.value[key] = count
      }
    }
  } catch (err) {
    // Fallback gracefully
  }
}

async function loadAllWatchSummaries() {
  if (props.activeTab !== "watched") return
  const items = props.watchedItems
  for (const item of items) {
    fetchWatchSummaryForItem(item.media.id, item.media.media_type)
  }
}

function openWatchLogModal(item: LibraryItemResponse) {
  activeLogMedia.value = { id: item.media.id, type: item.media.media_type }
  showLogModal.value = true
}

function onWatchLogged() {
  showLogModal.value = false
  if (activeLogMedia.value) {
    const key = getWatchMediaKey(activeLogMedia.value.id, activeLogMedia.value.type)
    watchSummaries.value[key] = (watchSummaries.value[key] || 1) + 1
    fetchWatchSummaryForItem(activeLogMedia.value.id, activeLogMedia.value.type)
  }
}

function openWatchHistory(item: LibraryItemResponse) {
  activeHistoryMedia.value = { id: item.media.id, type: item.media.media_type }
  showHistoryPanel.value = true
}

function onWatchHistoryChanged() {
  if (activeHistoryMedia.value) {
    fetchWatchSummaryForItem(activeHistoryMedia.value.id, activeHistoryMedia.value.type)
  }
}

watch(
  () => props.activeTab,
  (newTab) => {
    if (newTab === "watched") {
      loadAllWatchSummaries()
    }
  },
  { immediate: true }
)

watch(
  () => props.watchedItems,
  () => {
    if (props.activeTab === "watched") {
      loadAllWatchSummaries()
    }
  },
  { deep: true }
)
</script>

<style scoped>
/* Filters styling */
.filter-bar {
  position: relative;
  z-index: 10;
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  background: rgba(20, 20, 22, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.03);
  border-radius: 16px;
  padding: 1rem;
  margin-bottom: 2rem;
  align-items: center;
}

.filter-input-wrapper {
  display: flex;
  align-items: center;
  background: #161618;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 0 1rem;
  height: 42px;
}

.search-wrapper {
  flex: 1;
  min-width: 200px;
}

.filter-icon, .select-icon {
  color: #9ca3af;
}

.filter-input {
  background: transparent;
  border: none;
  color: #ffffff;
  padding-left: 0.75rem;
  font-size: 0.875rem;
  width: 100%;
  outline: none;
}

.filter-select-wrapper {
  display: flex;
  align-items: center;
  background: #161618;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 0 1rem;
  height: 42px;
  min-width: 150px;
  position: relative;
}

.filter-select {
  background: transparent;
  border: none;
  color: #ffffff;
  padding-left: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  width: 100%;
  outline: none;
  cursor: pointer;
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
}

.filter-select option {
  background: #161618;
  color: #ffffff;
}

/* Grid lists and cards */
.media-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
}

@media (min-width: 640px) {
  .media-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 768px) {
  .media-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (min-width: 1024px) {
  .media-grid {
    grid-template-columns: repeat(5, 1fr);
  }
}

.media-library-card {
  background: rgba(20, 20, 22, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 16px;
  overflow: hidden;
  transition: transform 0.25s, border-color 0.25s;
  display: flex;
  flex-direction: column;
}

.media-library-card:hover {
  transform: translateY(-4px);
  border-color: rgba(255, 255, 255, 0.12);
}

.media-poster-frame {
  position: relative;
  aspect-ratio: 2/3;
  background: #1c1c1e;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.media-poster-frame img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s;
}

.media-library-card:hover .media-poster-frame img {
  transform: scale(1.05);
}

.media-type-chip {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.75);
  color: #ffffff;
  font-size: 0.625rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.media-card-meta {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
}

.media-title {
  font-size: 0.875rem;
  font-weight: 700;
  color: #ffffff;
  text-decoration: none;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  transition: color 0.2s;
}

.media-title:hover {
  color: #e1251b;
}

.media-details-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
}

.media-rating-badge {
  color: #ffb800;
  font-weight: 600;
}

.save-date {
  color: #71717a;
}

.tags-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.tag-badge {
  font-size: 0.675rem;
  background: rgba(255, 255, 255, 0.05);
  color: #a1a1aa;
  padding: 2px 6px;
  border-radius: 4px;
}

/* Reviews List card styling */
.reviews-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.review-library-card {
  background: rgba(20, 20, 22, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 20px;
  padding: 1.5rem;
  display: flex;
  gap: 1.5rem;
}

.review-media-poster {
  width: 80px;
  height: 120px;
  border-radius: 12px;
  overflow: hidden;
  flex-shrink: 0;
  background: #1c1c1e;
  display: flex;
  align-items: center;
  justify-content: center;
}

.review-media-poster img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.review-card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.review-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.review-media-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #ffffff;
  text-decoration: none;
  transition: color 0.2s;
}

.review-media-title:hover {
  color: #e1251b;
}

.review-card-stars {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: rgba(255, 184, 0, 0.08);
  border: 1px solid rgba(255, 184, 0, 0.15);
  padding: 4px 10px;
  border-radius: 8px;
}

.rating-num {
  font-size: 0.8125rem;
  font-weight: 700;
  color: #ffb800;
}

.stars-icon {
  color: #ffb800;
  font-size: 0.8125rem;
}

.review-card-body {
  font-size: 0.875rem;
  color: #d1d5db;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
}

.review-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
  padding-top: 0.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
}

.review-date {
  font-size: 0.75rem;
  color: #71717a;
}

.review-meta-actions {
  display: flex;
  gap: 1rem;
}

.action-count {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  color: #a1a1aa;
}

.liked-heart {
  color: #ef4444;
}

/* Rich Empty state styling */
.empty-state-card {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background: rgba(20, 20, 24, 0.65);
  border: 1px dashed rgba(225, 37, 27, 0.25);
  border-radius: 24px;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.35);
  text-align: center;
  margin: 1rem 0;
  position: relative;
  z-index: 1;
  overflow: hidden;
}

.empty-state-visual {
  position: relative;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-glow-backdrop {
  position: absolute;
  width: 100px;
  height: 100px;
  background: radial-gradient(circle, rgba(225, 37, 27, 0.35) 0%, rgba(225, 37, 27, 0) 70%);
  border-radius: 50%;
  filter: blur(12px);
  animation: pulseGlow 3.5s ease-in-out infinite alternate;
}

@keyframes pulseGlow {
  0% {
    transform: scale(0.85);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.15);
    opacity: 1;
  }
}

.empty-icon-badge {
  width: 76px;
  height: 76px;
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(35, 35, 42, 0.9), rgba(20, 20, 24, 0.95));
  border: 1px solid rgba(225, 37, 27, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 2;
  box-shadow: 0 8px 24px rgba(225, 37, 27, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.empty-trophy-icon {
  color: #e1251b;
  filter: drop-shadow(0 2px 8px rgba(225, 37, 27, 0.5));
}

.empty-state-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 0.5rem 0;
  letter-spacing: -0.01em;
}

.empty-state-description {
  font-size: 0.925rem;
  color: #9ca3af;
  max-width: 440px;
  margin: 0 0 0.5rem 0;
  line-height: 1.6;
}

/* Poster fallback */
.poster-fallback {
  color: #3f3f46;
}

/* entrance animation */
.fade-in-up {
  opacity: 0;
  transform: translateY(15px);
  animation: fadeInUp 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes fadeInUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Watch Log Bar inside media-card-meta */
.watch-log-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  flex-wrap: wrap;
}

.watch-count-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: rgba(79, 70, 229, 0.12);
  border: 1px solid rgba(99, 102, 241, 0.25);
  color: #a5b4fc;
  padding: 0.3rem 0.65rem;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.watch-count-chip:hover {
  background: rgba(79, 70, 229, 0.22);
  border-color: rgba(99, 102, 241, 0.45);
  color: #c7d2fe;
}

.btn-log-watch-mini {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #e4e4e7;
  padding: 0.3rem 0.65rem;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.btn-log-watch-mini:hover {
  background: rgba(225, 37, 27, 0.15);
  border-color: rgba(225, 37, 27, 0.4);
  color: #ffffff;
}
</style>
