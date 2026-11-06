<template>
  <div class="list-page">
    <div class="page-header">
      <h1 class="page-title">TV SERIES</h1>
      <div class="tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-btn"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          {{ tab.label }}
        </button>
      </div>
      <div class="filter-bar">
        <div class="filter-select" ref="genreRef">
          <button class="filter-trigger" @click="genreOpen = !genreOpen">
            <Tag :size="14" />
            {{
              selectedGenre
                ? genres.find(g => g.id === selectedGenre)?.name
                : "All Genres"
            }}
            <ChevronDown
              :size="13"
              :class="{ 'rotate-180': genreOpen }"
              class="chevron"
            />
          </button>
          <div class="filter-dropdown" v-if="genreOpen">
            <button
              class="filter-option"
              :class="{ active: !selectedGenre }"
              @click="selectGenre(null)"
            >
              All Genres
            </button>
            <button
              v-for="g in genres"
              :key="g.id"
              class="filter-option"
              :class="{ active: selectedGenre === g.id }"
              @click="selectGenre(g.id)"
            >
              {{ g.name }}
            </button>
          </div>
        </div>
        <div class="filter-select" ref="sortRef">
          <button class="filter-trigger" @click="sortOpen = !sortOpen">
            <ArrowUpDown :size="14" />
            {{ sortOptions.find(s => s.key === sortBy)?.label }}
            <ChevronDown
              :size="13"
              :class="{ 'rotate-180': sortOpen }"
              class="chevron"
            />
          </button>
          <div class="filter-dropdown" v-if="sortOpen">
            <button
              v-for="s in sortOptions"
              :key="s.key"
              class="filter-option"
              :class="{ active: sortBy === s.key }"
              @click="selectSort(s.key)"
            >
              {{ s.label }}
            </button>
          </div>
        </div>
        <div class="search-inline">
          <Search :size="14" class="search-ic" />
          <input
            v-model="searchQuery"
            class="search-input"
            placeholder="Search TV series..."
            @keydown.enter="doSearch"
          />
          <button v-if="searchQuery" class="search-clear" @click="clearSearch">
            <X :size="13" />
          </button>
        </div>
      </div>
    </div>

    <div class="content-area">
      <div class="movie-grid" v-if="series.length">
        <div
          v-for="(show, i) in series"
          :key="show.id"
          class="poster-wrap"
          @mouseenter="onCardEnter(show.id, show)"
          @mouseleave="onCardLeave(show.id)"
        >
          <RouterLink
            :to="{ name: 'tv-detail', params: { id: show.id } }"
            class="poster-card"
          >
            <img
              :src="
                show.poster_path
                  ? `https://image.tmdb.org/t/p/w342${show.poster_path}`
                  : '/placeholder.jpg'
              "
              :alt="show.name"
              loading="lazy"
            />
          </RouterLink>
          <Transition name="popup">
            <div
              v-if="hoveredId === show.id"
              class="hover-popup"
              :class="getPopupPos(i)"
              @mouseenter="onPopupEnter"
              @mouseleave="onPopupLeave(show.id)"
            >
              <PopupCard
                :movie="show"
                :media-type="'tv'"
                :current-trailer="getState(show.id).currentTrailer.value"
                :trailer-unavailable="
                  getState(show.id).trailerUnavailable.value
                "
                :is-iframe-mounted="getState(show.id).isIframeMounted.value"
                :is-iframe-loaded="getState(show.id).isIframeLoaded.value"
                :show-skeleton="getState(show.id).showSkeleton.value"
                :show-fallback="getState(show.id).showFallback.value"
                :attach-player="getState(show.id).attachPlayer"
              />
            </div>
          </Transition>
        </div>
      </div>
      <div class="empty-state" v-else-if="!isLoading">
        <Tv :size="48" class="empty-icon" />
        <p>No TV series found</p>
      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button
          class="page-btn page-nav"
          :disabled="currentPage === 1"
          @click="goToPage(currentPage - 1)"
        >
          <ChevronLeft :size="16" />
        </button>
        <button
          v-for="p in paginationPages"
          :key="p"
          class="page-btn"
          :class="{
            'page-btn--active': p === currentPage,
            'page-btn--ellipsis': p === '...',
          }"
          :disabled="p === '...'"
          @click="typeof p === 'number' && goToPage(p)"
        >
          {{ p }}
        </button>
        <button
          class="page-btn page-nav"
          :disabled="currentPage === Math.min(totalPages, 500)"
          @click="goToPage(currentPage + 1)"
        >
          <ChevronRight :size="16" />
        </button>
      </div>
    </div>
    <div class="loading-overlay" v-if="isLoading"><div class="spinner" /></div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, onUnmounted, watch } from "vue"
  import { useRoute, useRouter } from "vue-router"
  import { useQuery } from "@tanstack/vue-query"
  import { movieApi } from "@/api/api"
  import type { TVSeries, Genre, Movie } from "@/types"
  import {
    Search,
    X,
    Tag,
    ArrowUpDown,
    ChevronDown,
    Tv,
    ChevronLeft,
    ChevronRight,
  } from "lucide-vue-next"
  import {
    resolveTrailer,
    useTrailerPreview,
    type ResolvedTrailer,
  } from "@/composables/useTrailerPreview"
  import PopupCard from "@/components/movie/PopupCard.vue"

  const route = useRoute()
  const router = useRouter()

  type TabKey = "popular" | "now_airing" | "top_rated"
  const tabs = [
    { key: "popular" as TabKey, label: "Popular" },
    { key: "now_airing" as TabKey, label: "Now Airing" },
    { key: "top_rated" as TabKey, label: "Top Rated" },
  ]
  const sortOptions = [
    { key: "default", label: "Default" },
    { key: "rating", label: "Rating" },
    { key: "date", label: "First Air Date" },
    { key: "popular", label: "Popularity" },
  ]

  const activeTab = ref<TabKey>("popular")
  const sortBy = ref("default")
  const selectedGenre = ref<number | null>(
    route.query.genre ? Number(route.query.genre) : null
  )
  const searchQuery = ref((route.query.q as string) ?? "")
  const currentPage = ref(1)
  const hoveredId = ref<number | null>(null)
  const cardStates = new Map<number, ReturnType<typeof useTrailerPreview>>()
  const cardTrailers = new Map<number, ResolvedTrailer | null>()
  let insidePopup = false
  let showTimer: ReturnType<typeof setTimeout> | null = null
  const SHOW_DELAY = 200
  const genres = ref<Genre[]>([])
  const genreOpen = ref(false)
  const sortOpen = ref(false)
  const genreRef = ref<HTMLElement | null>(null)
  const sortRef = ref<HTMLElement | null>(null)

  const isSearchMode = computed(() => !!searchQuery.value.trim())
  const queryKey = computed(() => [
    "tv",
    activeTab.value,
    currentPage.value,
    searchQuery.value,
  ])

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: async () => {
      if (isSearchMode.value)
        return movieApi
          .searchSeries(searchQuery.value.trim(), currentPage.value)
          .then(r => r.data)
      const fn = {
        popular: () => movieApi.getPopularSeries(currentPage.value),
        now_airing: () => movieApi.getNowAiringSeries(currentPage.value),
        top_rated: () => movieApi.getTopRatedSeries(currentPage.value),
      }[activeTab.value]
      return fn().then(r => r.data)
    },
  })

  const rawSeries = computed<TVSeries[]>(() => data.value?.results ?? [])
  const totalPages = computed(() => data.value?.total_pages ?? 1)

  const series = computed(() => {
    let list = [...rawSeries.value]
    if (selectedGenre.value)
      list = list.filter(s => s.genre_ids?.includes(selectedGenre.value!))
    if (sortBy.value === "rating")
      list.sort((a, b) => b.vote_average - a.vote_average)
    if (sortBy.value === "date")
      list.sort((a, b) =>
        (b.first_air_date ?? "").localeCompare(a.first_air_date ?? ""),
      )
    if (sortBy.value === "popular")
      list.sort((a, b) => b.popularity - a.popularity)
    return list
  })

  const paginationPages = computed(() => {
    const total = Math.min(totalPages.value, 500)
    const cur = currentPage.value
    const p: (number | string)[] = []
    if (cur > 2) p.push(1)
    if (cur > 3) p.push("...")
    for (let i = Math.max(1, cur - 1); i <= Math.min(total, cur + 1); i++)
      p.push(i)
    if (cur < total - 2) p.push("...")
    if (cur < total - 1) p.push(total)
    return [...new Set(p)]
  })

  /**
   * Map TVSeries fields ให้ตรงกับ Movie shape ที่ PopupCard คาดหวัง
   * - title / original_title  ← name
   * - release_date            ← first_air_date
   * - media_type              = "tv"  (ใช้ใน API calls ของ PopupCard)
   */
  function toMovieShape(show: TVSeries): Movie {
    return {
      ...show,
      title: show.name,
      original_title: show.name,
      release_date: show.first_air_date ?? "",
      media_type: "tv",
    } as unknown as Movie
  }

  function getPopupPos(i: number) {
    const col = i % 5
    if (col === 0) return "popup--right"
    if (col === 4) return "popup--left"
    return "popup--center"
  }
  function switchTab(key: TabKey) {
    activeTab.value = key
    currentPage.value = 1
    searchQuery.value = ""
  }
  function doSearch() {
    currentPage.value = 1
  }
  function clearSearch() {
    searchQuery.value = ""
    currentPage.value = 1
  }
  function selectGenre(id: number | null) {
    selectedGenre.value = id
    genreOpen.value = false
    const query = { ...route.query }
    if (id) {
      query.genre = String(id)
    } else {
      delete query.genre
    }
    router.replace({ query })
  }
  function selectSort(key: string) {
    sortBy.value = key
    sortOpen.value = false
  }
  function goToPage(p: number) {
    currentPage.value = p
    window.scrollTo({ top: 0, behavior: "smooth" })
  }

  function onClickOutside(e: MouseEvent) {
    const t = e.target as Node
    if (genreRef.value && !genreRef.value.contains(t)) genreOpen.value = false
    if (sortRef.value && !sortRef.value.contains(t)) sortOpen.value = false
  }

  function getState(seriesId: number) {
    if (!cardStates.has(seriesId)) {
      cardStates.set(seriesId, useTrailerPreview({ mountDelay: 500 }))
    }
    return cardStates.get(seriesId)!
  }

  function getTrailer(seriesId: number): ResolvedTrailer | null {
    return cardTrailers.get(seriesId) ?? null
  }

  async function fetchAndCacheTrailer(show: TVSeries) {
    if (cardTrailers.has(show.id)) return
    cardTrailers.set(show.id, null)
    try {
      const res = await movieApi.getTVVideos(show.id)
      const videos = res.data?.results ?? []
      const trailer = resolveTrailer(videos)
      cardTrailers.set(show.id, trailer)
      if (hoveredId.value === show.id && trailer) {
        getState(show.id).scheduleMount()
      }
    } catch {
      cardTrailers.set(show.id, null)
    }
  }

  function onCardEnter(seriesId: number, show: TVSeries) {
    clearTimeout(showTimer ?? undefined)
    insidePopup = false
    fetchAndCacheTrailer(show)
    showTimer = setTimeout(() => {
      hoveredId.value = seriesId
      const trailer = getTrailer(seriesId)
      if (trailer) getState(seriesId).scheduleMount()
    }, SHOW_DELAY)
  }

  function onCardLeave(seriesId: number) {
    clearTimeout(showTimer ?? undefined)
    setTimeout(() => {
      if (!insidePopup) closeCard(seriesId)
    }, 80)
  }

  function onPopupEnter() {
    insidePopup = true
  }

  function onPopupLeave(seriesId: number) {
    insidePopup = false
    closeCard(seriesId)
  }

  function closeCard(seriesId: number) {
    if (hoveredId.value !== seriesId) return
    hoveredId.value = null
    cardStates.get(seriesId)?.unmount()
  }

  watch(
    () => route.query.q,
    q => {
      searchQuery.value = (q as string) ?? ""
      currentPage.value = 1
    },
  )
  watch(
    () => route.query.genre,
    g => {
      selectedGenre.value = g ? Number(g) : null
    },
    { immediate: true },
  )
  onMounted(async () => {
    document.addEventListener("click", onClickOutside)
    const res = await movieApi.getSeriesGenres()
    genres.value = res.data.genres
  })
  onUnmounted(() => {
    document.removeEventListener("click", onClickOutside)
    clearTimeout(showTimer ?? undefined)
    cardStates.forEach(s => s.unmount())
    cardStates.clear()
  })
</script>

<style scoped>
  .list-page {
    background: #141414;
    min-height: 100vh;
  }
  .page-header {
    padding: 5.5rem 1.5rem 0;
    max-width: 1100px;
    margin: 0 auto;
  }
  .page-title {
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 1.75rem;
    font-weight: 900;
    letter-spacing: 2px;
    color: #fff;
    margin: 0 0 1.25rem;
  }
  .tabs {
    display: flex;
    gap: 4px;
    margin-bottom: 1rem;
  }
  .tab-btn {
    padding: 0.4rem 1rem;
    background: #1f1f1f;
    border: 1px solid #2a2a2a;
    border-radius: 9999px;
    font-size: 0.82rem;
    font-weight: 600;
    color: #a3a3a3;
    cursor: pointer;
    transition: all 0.2s;
  }
  .tab-btn:hover {
    color: #fff;
    border-color: #444;
  }
  .tab-btn.active {
    background: var(--color-brand);
    border-color: var(--color-brand);
    color: #fff;
  }
  .filter-bar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding-bottom: 1.25rem;
    flex-wrap: wrap;
  }
  .filter-select {
    position: relative;
  }
  .filter-trigger {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.45rem 0.875rem;
    background: #1f1f1f;
    border: 1px solid #2a2a2a;
    border-radius: 8px;
    color: #ccc;
    font-size: 0.82rem;
    cursor: pointer;
    transition: border-color 0.2s;
    white-space: nowrap;
  }
  .filter-trigger:hover {
    border-color: #555;
    color: #fff;
  }
  .chevron {
    transition: transform 0.2s;
  }
  .rotate-180 {
    transform: rotate(180deg);
  }
  .filter-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    min-width: 160px;
    max-height: 280px;
    overflow-y: auto;
    background: #1f1f1f;
    border: 1px solid #2a2a2a;
    border-radius: 10px;
    padding: 0.35rem;
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6);
    z-index: 30;
  }
  .filter-option {
    display: block;
    width: 100%;
    padding: 0.45rem 0.75rem;
    background: none;
    border: none;
    color: #ccc;
    font-size: 0.82rem;
    cursor: pointer;
    border-radius: 6px;
    text-align: left;
    transition:
      background 0.15s,
      color 0.15s;
  }
  .filter-option:hover {
    background: rgba(255, 255, 255, 0.07);
    color: #fff;
  }
  .filter-option.active {
    color: var(--color-brand);
    font-weight: 600;
  }
  .search-inline {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: #1f1f1f;
    border: 1px solid #2a2a2a;
    border-radius: 8px;
    padding: 0.4rem 0.75rem;
    flex: 1;
    max-width: 320px;
  }
  .search-inline:focus-within {
    border-color: #555;
  }
  .search-ic {
    color: #666;
    flex-shrink: 0;
  }
  .search-input {
    background: none;
    border: none;
    outline: none;
    color: #fff;
    font-size: 0.875rem;
    width: 100%;
  }
  .search-input::placeholder {
    color: #555;
  }
  .search-clear {
    background: none;
    border: none;
    cursor: pointer;
    color: #666;
    display: flex;
    padding: 0;
  }
  .search-clear:hover {
    color: #ccc;
  }
  .content-area {
    padding: 0 1.5rem 3rem;
    max-width: 1100px;
    margin: 0 auto;
  }
  .movie-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 12px;
    margin-bottom: 1.5rem;
  }
  .poster-wrap {
    position: relative;
  }
  .poster-card {
    display: block;
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition:
      transform 0.2s,
      box-shadow 0.2s;
  }
  .poster-card:hover {
    transform: scale(1.03);
    box-shadow: 0 6px 24px rgba(0, 0, 0, 0.5);
  }
  .poster-card img {
    width: 100%;
    aspect-ratio: 2/3;
    object-fit: cover;
    display: block;
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
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 5rem 0;
    color: #555;
  }
  .empty-icon {
    opacity: 0.4;
  }
  .pagination {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    flex-wrap: wrap;
  }
  .page-btn {
    min-width: 36px;
    height: 36px;
    padding: 0 0.5rem;
    background: #1f1f1f;
    border: 1px solid #2a2a2a;
    color: #a3a3a3;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
    transition:
      background 0.15s,
      color 0.15s;
  }
  .page-btn:hover:not(:disabled):not(.page-btn--ellipsis) {
    background: #2a2a2a;
    color: #fff;
  }
  .page-btn--active {
    background: var(--color-brand);
    border-color: var(--color-brand);
    color: #fff;
  }
  .page-btn--ellipsis {
    cursor: default;
  }
  .page-nav {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .page-nav:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }
  .loading-overlay {
    position: fixed;
    inset: 0;
    background: rgba(20, 20, 20, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
  }
  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.15);
    border-top-color: var(--color-brand);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
