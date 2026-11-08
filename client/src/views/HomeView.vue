<template>
  <div class="home">
    <!-- ── Hero Banner Section ───────────────────────────────────── -->
    <section class="hero" v-if="heroMovies.length">
      <div class="hero__slides">
        <TransitionGroup name="hero-fade">
          <div
            v-for="(movie, i) in heroMovies"
            :key="movie.id"
            v-show="heroIndex === i"
            class="hero__slide"
          >
            <img
              v-if="movie.backdrop_path && !failedBackdrops.has(movie.id)"
              :src="`https://image.tmdb.org/t/p/original${movie.backdrop_path}`"
              :alt="movie.title"
              class="hero__backdrop"
              @error="failedBackdrops.add(movie.id)"
            />
            <div v-else class="hero__backdrop hero__backdrop--fallback">
              <span class="hero__fallback-title">{{ movie.title }}</span>
            </div>

            <!-- Gradient Overlays for Cinematic Depth -->
            <div class="hero__overlay-bottom" />
            <div class="hero__overlay-side" />
          </div>
        </TransitionGroup>
      </div>

      <!-- Hero Content -->
      <div class="hero__content-wrapper" v-if="currentHero">
        <div class="hero__badge">
          <Sparkles :size="14" class="text-amber-400" />
          <span>แนะนำพิเศษ</span>
        </div>

        <h1 class="hero__title">{{ currentHero.title }}</h1>

        <div class="hero__meta">
          <span class="meta-item year" v-if="currentHero.release_date">
            {{ currentHero.release_date.slice(0, 4) }}
          </span>
          <span class="meta-dot" v-if="currentHero.release_date">·</span>
          <MediaRating
            class="meta-item rating"
            :ratings="currentHero.ratings"
            :vote-average="currentHero.vote_average"
          />
        </div>

        <p class="hero__overview" v-if="currentHero.overview">
          {{ currentHero.overview }}
        </p>

        <div class="hero__actions">
          <RouterLink
            :to="{ name: 'movie-detail', params: { id: currentHero.id } }"
            class="btn btn--primary"
          >
            <Play :size="16" class="fill-current" />
            <span>ดูข้อมูลหนัง</span>
          </RouterLink>

          <RouterLink
            :to="{ name: 'movies' }"
            class="btn btn--secondary"
          >
            <Compass :size="16" />
            <span>สำรวจทั้งหมด</span>
          </RouterLink>
        </div>
      </div>

      <!-- Hero Slide Indicators / Dots -->
      <div class="hero__dots">
        <button
          v-for="(_, i) in heroMovies"
          :key="i"
          class="hero__dot"
          :class="{ 'hero__dot--active': heroIndex === i }"
          @click="handleDotClick(i)"
          :aria-label="`Slide ${i + 1}`"
        >
          <span class="hero__dot-progress" />
        </button>
      </div>
    </section>

    <!-- ── Quick Category Chips Navigation ─────────────────────── -->
    <div class="category-bar-wrapper">
      <div class="category-bar">
        <RouterLink
          v-for="genre in quickGenres"
          :key="genre.id"
          :to="{ name: 'movies', query: { genre: genre.id } }"
          class="category-chip"
        >
          <component :is="genre.icon" :size="15" class="chip-icon" />
          <span>{{ genre.name }}</span>
        </RouterLink>
      </div>
    </div>

    <!-- ── Main Content Area ───────────────────────────────────── -->
    <div class="content-area">
      <!-- Section 1: Recommended Movies (User Personalized) -->
      <section class="movie-section" v-if="hasGenres">
        <div class="section-header">
          <div class="header-title-group">
            <div class="icon-badge icon-badge--amber">
              <Sparkles :size="18" />
            </div>
            <div>
              <h2 class="section-title">แนะนำสำหรับคุณ</h2>
              <p class="section-subtitle">คัดสรรจากแนวหนังที่คุณชื่นชอบ</p>
            </div>
          </div>
        </div>

        <div class="movie-grid" v-if="recommendedMovies.length">
          <div
            v-for="(movie, i) in recommendedMovies"
            :key="`rec-${movie.id}`"
            class="poster-wrap"
            @mouseenter="onCardEnter(movie.id, i, 'rec')"
            @mouseleave="onCardLeave(movie.id, 'rec')"
          >
            <RouterLink
              :to="{ name: 'movie-detail', params: { id: movie.id } }"
              class="poster-card"
            >
              <div class="poster-image-container">
                <img
                  v-if="movie.poster_path && !failedPosters.has(movie.id)"
                  :src="`https://image.tmdb.org/t/p/w342${movie.poster_path}`"
                  :alt="movie.title"
                  loading="lazy"
                  @error="failedPosters.add(movie.id)"
                />
                <div v-else class="poster-fallback">
                  <span class="poster-fallback__title">{{ movie.title }}</span>
                </div>

                <!-- Rating Badge Overlay -->
                <div class="rating-badge">
                  <MediaRating :ratings="movie.ratings" :vote-average="movie.vote_average" />
                </div>
              </div>

              <!-- Movie Metadata Under Poster (Visible on Mobile & Desktop) -->
              <div class="poster-info">
                <h3 class="poster-title">{{ movie.title }}</h3>
                <span class="poster-year" v-if="movie.release_date">
                  {{ movie.release_date.slice(0, 4) }}
                </span>
              </div>
            </RouterLink>

            <!-- Hover Preview Popup (Desktop Only) -->
            <Transition name="popup">
              <div
                v-if="hoveredMovieKey === `rec-${movie.id}` && isDesktop"
                class="hover-popup"
                :class="getPopupPosition(i)"
                @mouseenter="onPopupEnter"
                @mouseleave="onPopupLeave(movie.id, 'rec')"
              >
                <PopupCard
                  :movie="movie"
                  :current-trailer="getState(movie.id).currentTrailer.value"
                  :trailer-unavailable="getState(movie.id).trailerUnavailable.value"
                  :is-iframe-mounted="getState(movie.id).isIframeMounted.value"
                  :is-iframe-loaded="getState(movie.id).isIframeLoaded.value"
                  :show-skeleton="getState(movie.id).showSkeleton.value"
                  :show-fallback="getState(movie.id).showFallback.value"
                  :attach-player="getState(movie.id).attachPlayer"
                />
              </div>
            </Transition>
          </div>
        </div>

        <div class="movie-grid" v-else-if="isLoadingRecommended">
          <SkeletonBlock v-for="i in 10" :key="i" aspect="2 / 3" radius="12px" />
        </div>
      </section>

      <!-- Section 2: Popular Movies (🔥 ยอดนิยม) -->
      <section class="movie-section">
        <div class="section-header">
          <div class="header-title-group">
            <div class="icon-badge icon-badge--red">
              <Flame :size="18" />
            </div>
            <div>
              <h2 class="section-title">หนังฮิตยอดนิยม</h2>
              <p class="section-subtitle">ภาพยนตร์ที่มีผู้เข้าชมมากที่สุดขณะนี้</p>
            </div>
          </div>
          <RouterLink :to="{ name: 'movies' }" class="see-all-btn link-nudge">
            <span>ดูทั้งหมด</span>
            <ChevronRight :size="16" />
          </RouterLink>
        </div>

        <div class="movie-grid" v-if="popularMovies.length">
          <div
            v-for="(movie, i) in popularMovies"
            :key="`pop-${movie.id}`"
            class="poster-wrap"
            @mouseenter="onCardEnter(movie.id, i, 'pop')"
            @mouseleave="onCardLeave(movie.id, 'pop')"
          >
            <RouterLink
              :to="{ name: 'movie-detail', params: { id: movie.id } }"
              class="poster-card"
            >
              <div class="poster-image-container">
                <img
                  v-if="movie.poster_path && !failedPosters.has(movie.id)"
                  :src="`https://image.tmdb.org/t/p/w342${movie.poster_path}`"
                  :alt="movie.title"
                  loading="lazy"
                  @error="failedPosters.add(movie.id)"
                />
                <div v-else class="poster-fallback">
                  <span class="poster-fallback__title">{{ movie.title }}</span>
                </div>

                <div class="rating-badge">
                  <MediaRating :ratings="movie.ratings" :vote-average="movie.vote_average" />
                </div>
              </div>

              <div class="poster-info">
                <h3 class="poster-title">{{ movie.title }}</h3>
                <span class="poster-year" v-if="movie.release_date">
                  {{ movie.release_date.slice(0, 4) }}
                </span>
              </div>
            </RouterLink>

            <Transition name="popup">
              <div
                v-if="hoveredMovieKey === `pop-${movie.id}` && isDesktop"
                class="hover-popup"
                :class="getPopupPosition(i)"
                @mouseenter="onPopupEnter"
                @mouseleave="onPopupLeave(movie.id, 'pop')"
              >
                <PopupCard
                  :movie="movie"
                  :current-trailer="getState(movie.id).currentTrailer.value"
                  :trailer-unavailable="getState(movie.id).trailerUnavailable.value"
                  :is-iframe-mounted="getState(movie.id).isIframeMounted.value"
                  :is-iframe-loaded="getState(movie.id).isIframeLoaded.value"
                  :show-skeleton="getState(movie.id).showSkeleton.value"
                  :show-fallback="getState(movie.id).showFallback.value"
                  :attach-player="getState(movie.id).attachPlayer"
                />
              </div>
            </Transition>
          </div>
        </div>

        <div class="movie-grid" v-else-if="isLoadingPopular">
          <SkeletonBlock v-for="i in 10" :key="i" aspect="2 / 3" radius="12px" />
        </div>
      </section>

      <!-- Section 3: Now Playing / Fresh Releases (🎬 หนังมาใหม่) -->
      <section class="movie-section">
        <div class="section-header">
          <div class="header-title-group">
            <div class="icon-badge icon-badge--blue">
              <Clapperboard :size="18" />
            </div>
            <div>
              <h2 class="section-title">หนังมาใหม่เข้าโรง</h2>
              <p class="section-subtitle">อัปเดตภาพยนตร์ฉายล่าสุดประจำสัปดาห์</p>
            </div>
          </div>
          <RouterLink :to="{ name: 'movies' }" class="see-all-btn link-nudge">
            <span>ดูทั้งหมด</span>
            <ChevronRight :size="16" />
          </RouterLink>
        </div>

        <div class="movie-grid" v-if="nowPlayingMovies.length">
          <div
            v-for="(movie, i) in nowPlayingMovies"
            :key="`now-${movie.id}`"
            class="poster-wrap"
            @mouseenter="onCardEnter(movie.id, i, 'now')"
            @mouseleave="onCardLeave(movie.id, 'now')"
          >
            <RouterLink
              :to="{ name: 'movie-detail', params: { id: movie.id } }"
              class="poster-card"
            >
              <div class="poster-image-container">
                <img
                  v-if="movie.poster_path && !failedPosters.has(movie.id)"
                  :src="`https://image.tmdb.org/t/p/w342${movie.poster_path}`"
                  :alt="movie.title"
                  loading="lazy"
                  @error="failedPosters.add(movie.id)"
                />
                <div v-else class="poster-fallback">
                  <span class="poster-fallback__title">{{ movie.title }}</span>
                </div>

                <div class="rating-badge">
                  <MediaRating :ratings="movie.ratings" :vote-average="movie.vote_average" />
                </div>
              </div>

              <div class="poster-info">
                <h3 class="poster-title">{{ movie.title }}</h3>
                <span class="poster-year" v-if="movie.release_date">
                  {{ movie.release_date.slice(0, 4) }}
                </span>
              </div>
            </RouterLink>

            <Transition name="popup">
              <div
                v-if="hoveredMovieKey === `now-${movie.id}` && isDesktop"
                class="hover-popup"
                :class="getPopupPosition(i)"
                @mouseenter="onPopupEnter"
                @mouseleave="onPopupLeave(movie.id, 'now')"
              >
                <PopupCard
                  :movie="movie"
                  :current-trailer="getState(movie.id).currentTrailer.value"
                  :trailer-unavailable="getState(movie.id).trailerUnavailable.value"
                  :is-iframe-mounted="getState(movie.id).isIframeMounted.value"
                  :is-iframe-loaded="getState(movie.id).isIframeLoaded.value"
                  :show-skeleton="getState(movie.id).showSkeleton.value"
                  :show-fallback="getState(movie.id).showFallback.value"
                  :attach-player="getState(movie.id).attachPlayer"
                />
              </div>
            </Transition>
          </div>
        </div>

        <div class="movie-grid" v-else-if="isLoadingNowPlaying">
          <SkeletonBlock v-for="i in 10" :key="i" aspect="2 / 3" radius="12px" />
        </div>
      </section>

      <!-- Section 4: Top Rated Classics (⭐ คะแนนสูงสุด) -->
      <section class="movie-section">
        <div class="section-header">
          <div class="header-title-group">
            <div class="icon-badge icon-badge--emerald">
              <TrendingUp :size="18" />
            </div>
            <div>
              <h2 class="section-title">หนังการันตีคุณภาพ</h2>
              <p class="section-subtitle">ได้รับคะแนนรีวิวสูงสุดจากผู้ชมทั่วโลก</p>
            </div>
          </div>
          <RouterLink :to="{ name: 'movies' }" class="see-all-btn link-nudge">
            <span>ดูทั้งหมด</span>
            <ChevronRight :size="16" />
          </RouterLink>
        </div>

        <div class="movie-grid" v-if="topRatedMovies.length">
          <div
            v-for="(movie, i) in topRatedMovies"
            :key="`top-${movie.id}`"
            class="poster-wrap"
            @mouseenter="onCardEnter(movie.id, i, 'top')"
            @mouseleave="onCardLeave(movie.id, 'top')"
          >
            <RouterLink
              :to="{ name: 'movie-detail', params: { id: movie.id } }"
              class="poster-card"
            >
              <div class="poster-image-container">
                <img
                  v-if="movie.poster_path && !failedPosters.has(movie.id)"
                  :src="`https://image.tmdb.org/t/p/w342${movie.poster_path}`"
                  :alt="movie.title"
                  loading="lazy"
                  @error="failedPosters.add(movie.id)"
                />
                <div v-else class="poster-fallback">
                  <span class="poster-fallback__title">{{ movie.title }}</span>
                </div>

                <div class="rating-badge">
                  <MediaRating :ratings="movie.ratings" :vote-average="movie.vote_average" />
                </div>
              </div>

              <div class="poster-info">
                <h3 class="poster-title">{{ movie.title }}</h3>
                <span class="poster-year" v-if="movie.release_date">
                  {{ movie.release_date.slice(0, 4) }}
                </span>
              </div>
            </RouterLink>

            <Transition name="popup">
              <div
                v-if="hoveredMovieKey === `top-${movie.id}` && isDesktop"
                class="hover-popup"
                :class="getPopupPosition(i)"
                @mouseenter="onPopupEnter"
                @mouseleave="onPopupLeave(movie.id, 'top')"
              >
                <PopupCard
                  :movie="movie"
                  :current-trailer="getState(movie.id).currentTrailer.value"
                  :trailer-unavailable="getState(movie.id).trailerUnavailable.value"
                  :is-iframe-mounted="getState(movie.id).isIframeMounted.value"
                  :is-iframe-loaded="getState(movie.id).isIframeLoaded.value"
                  :show-skeleton="getState(movie.id).showSkeleton.value"
                  :show-fallback="getState(movie.id).showFallback.value"
                  :attach-player="getState(movie.id).attachPlayer"
                />
              </div>
            </Transition>
          </div>
        </div>

        <div class="movie-grid" v-else-if="isLoadingTopRated">
          <SkeletonBlock v-for="i in 10" :key="i" aspect="2 / 3" radius="12px" />
        </div>
      </section>
    </div>

    <!-- Global Loading Overlay -->
    <Transition name="fade">
      <div
        class="loading-overlay"
        v-if="isLoadingNowPlaying && !nowPlayingMovies.length"
      >
        <AppSpinner />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
import MediaRating from "@/components/movie/MediaRating.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
  import { ref, computed, onMounted, onUnmounted } from "vue"
  import { useQuery } from "@tanstack/vue-query"
  import { movieApi } from "@/api/api"
  import { useAuthStore } from "@/stores/auth"
  import {
    Sparkles,
    Clapperboard,
    ChevronRight,
    Flame,
    TrendingUp,
    Play,
    Compass,
    Film,
    Tv,
    Ghost,
    Smile,
    Rocket,
    ShieldAlert
  } from "lucide-vue-next"
  import type { Movie } from "@/types"
  import PopupCard from "@/components/movie/PopupCard.vue"
  import {
    resolveTrailerCandidates,
    useTrailerPreview,
  } from "@/composables/useTrailerPreview"

  // ── Auth & Personalized Preferences ────────────────────────
  const authStore = useAuthStore()

  const hasGenres = computed(() => {
    const g = authStore.user?.favorite_genres
    if (!g || g === "null" || g === "skip") return false
    try {
      const parsed = JSON.parse(g)
      return Array.isArray(parsed) && parsed.length > 0
    } catch {
      return g.length > 0
    }
  })

  const genreQuery = computed(() => {
    if (!hasGenres.value) return ""
    const raw = authStore.user?.favorite_genres
    if (!raw) return ""
    try {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) return parsed.join(",")
    } catch {}
    return raw
  })

  // ── Quick Category / Genre Chips Configuration ─────────────
  const quickGenres = [
    { id: 28, name: "แอคชัน", icon: Flame },
    { id: 878, name: "ไซไฟ", icon: Rocket },
    { id: 35, name: "ตลก", icon: Smile },
    { id: 27, name: "สยองขวัญ", icon: Ghost },
    { id: 18, name: "ดราม่า", icon: Film },
    { id: 16, name: "แอนิเมชัน", icon: Sparkles },
    { id: 53, name: "ระทึกขวัญ", icon: ShieldAlert },
  ]

  // ── Responsive Check ───────────────────────────────────────
  const isDesktop = ref(true)

  function updateViewport() {
    isDesktop.value = window.innerWidth >= 768
  }

  // ── Hero & Popup State ─────────────────────────────────────
  const heroIndex = ref(0)
  const currentHero = computed(() => heroMovies.value[heroIndex.value] ?? null)
  const hoveredMovieKey = ref<string | null>(null)
  const hoveredMovieId = ref<number | null>(null)
  const failedPosters = ref(new Set<number>())
  const failedBackdrops = ref(new Set<number>())
  let heroTimer: ReturnType<typeof setInterval> | null = null

  // ── Per-card Trailer State Cache ───────────────────────────
  const cardStates = new Map<number, ReturnType<typeof useTrailerPreview>>()
  const videoCache = new Map<number, any[]>()

  let insidePopup = false
  let showTimer: ReturnType<typeof setTimeout> | null = null
  const SHOW_DELAY = 220

  // ── Queries ────────────────────────────────────────────────
  const { data: recommendedData, isLoading: isLoadingRecommended } = useQuery({
    queryKey: computed(() => ["movies-recommended", genreQuery.value]),
    queryFn: () => movieApi.getRecommended(genreQuery.value).then(r => r.data),
    enabled: hasGenres,
  })

  const { data: nowPlayingData, isLoading: isLoadingNowPlaying } = useQuery({
    queryKey: ["movies-now-playing-home"],
    queryFn: () => movieApi.getNowPlaying(1).then(r => r.data),
  })

  const { data: popularData, isLoading: isLoadingPopular } = useQuery({
    queryKey: ["movies-popular-home"],
    queryFn: () => movieApi.getPopular(1).then(r => r.data),
  })

  const { data: topRatedData, isLoading: isLoadingTopRated } = useQuery({
    queryKey: ["movies-top-rated-home"],
    queryFn: () => movieApi.getTopRated(1).then(r => r.data),
  })

  // ── Computed Lists ─────────────────────────────────────────
  const recommendedMovies = computed<Movie[]>(
    () => recommendedData.value?.results?.slice(0, 10) ?? [],
  )

  const nowPlayingMovies = computed<Movie[]>(
    () => nowPlayingData.value?.results?.slice(0, 10) ?? [],
  )

  const popularMovies = computed<Movie[]>(
    () => popularData.value?.results?.slice(0, 10) ?? [],
  )

  const topRatedMovies = computed<Movie[]>(
    () => topRatedData.value?.results?.slice(0, 10) ?? [],
  )

  const heroMovies = computed<Movie[]>(() => {
    const source = recommendedMovies.value.length
      ? recommendedMovies.value
      : popularMovies.value.length
      ? popularMovies.value
      : nowPlayingMovies.value
    return source.slice(0, 5)
  })

  // ── Trailer Helpers ────────────────────────────────────────
  function getState(movieId: number): ReturnType<typeof useTrailerPreview> {
    if (!cardStates.has(movieId)) {
      cardStates.set(movieId, useTrailerPreview({ mountDelay: 500 }))
    }
    return cardStates.get(movieId)!
  }

  async function fetchAndCacheTrailer(movie: Movie) {
    if (videoCache.has(movie.id)) return
    videoCache.set(movie.id, [])

    try {
      const res = await movieApi.getVideos(movie.id)
      const videos = res.data?.results ?? []
      videoCache.set(movie.id, videos)
      const candidates = resolveTrailerCandidates(videos)
      getState(movie.id).setCandidates(candidates)

      if (hoveredMovieId.value === movie.id && candidates.length > 0) {
        getState(movie.id).scheduleMount()
      }
    } catch {
      getState(movie.id).setCandidates([])
    }
  }

  // ── Hover Handlers ─────────────────────────────────────────
  function onCardEnter(
    movieId: number,
    index: number,
    section: "rec" | "now" | "pop" | "top",
  ) {
    if (!isDesktop.value) return

    clearTimeout(showTimer ?? undefined)
    insidePopup = false

    hoveredMovieId.value = movieId

    const allMovies = [
      ...recommendedMovies.value,
      ...popularMovies.value,
      ...nowPlayingMovies.value,
      ...topRatedMovies.value,
    ]
    const movie = allMovies.find(m => m.id === movieId)
    if (movie) fetchAndCacheTrailer(movie)

    showTimer = setTimeout(() => {
      hoveredMovieKey.value = `${section}-${movieId}`

      if (getState(movieId).currentTrailer.value) {
        getState(movieId).scheduleMount()
      }
    }, SHOW_DELAY)
  }

  function onCardLeave(movieId: number, section: "rec" | "now" | "pop" | "top") {
    if (!isDesktop.value) return
    clearTimeout(showTimer ?? undefined)
    setTimeout(() => {
      if (!insidePopup) closeCard(movieId, section)
    }, 80)
  }

  function onPopupEnter() {
    insidePopup = true
  }

  function onPopupLeave(movieId: number, section: "rec" | "now" | "pop" | "top") {
    insidePopup = false
    closeCard(movieId, section)
  }

  function closeCard(movieId: number, section: "rec" | "now" | "pop" | "top") {
    const targetKey = `${section}-${movieId}`
    if (hoveredMovieKey.value !== targetKey) return

    hoveredMovieKey.value = null
    hoveredMovieId.value = null

    const state = cardStates.get(movieId)
    if (state) state.unmount()
  }

  // ── Popup Alignment Calculator ────────────────────────────
  function getPopupPosition(index: number): string {
    const col = index % 5
    if (col === 0) return "popup--right"
    if (col === 4) return "popup--left"
    return "popup--center"
  }

  // ── Hero Navigation & Timer ────────────────────────────────
  function handleDotClick(index: number) {
    heroIndex.value = index
    resetTimer()
  }

  function resetTimer() {
    if (heroTimer) clearInterval(heroTimer)
    heroTimer = setInterval(() => {
      heroIndex.value =
        (heroIndex.value + 1) % Math.max(heroMovies.value.length, 1)
    }, 5000)
  }

  onMounted(() => {
    updateViewport()
    window.addEventListener("resize", updateViewport)
    resetTimer()
  })

  onUnmounted(() => {
    window.removeEventListener("resize", updateViewport)
    if (heroTimer) clearInterval(heroTimer)
    clearTimeout(showTimer ?? undefined)
    cardStates.forEach(s => s.unmount())
    cardStates.clear()
  })
</script>

<style scoped>
  .home {
    background: var(--color-surface);
    min-height: 100vh;
    color: #e2e8f0;
  }

  /* ── Hero Banner Section ───────────────────────────────────── */
  .hero {
    position: relative;
    width: 100%;
    height: 480px;
    overflow: hidden;
    background: #050507;
  }

  @media (max-width: 768px) {
    .hero {
      height: 380px;
    }
  }

  @media (max-width: 480px) {
    .hero {
      height: 320px;
    }
  }

  .hero__slides {
    position: absolute;
    inset: 0;
  }

  .hero__slide {
    position: absolute;
    inset: 0;
  }

  .hero__backdrop {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center 25%;
    filter: brightness(0.9);
  }

  .hero__backdrop--fallback {
    background: linear-gradient(135deg, #18181f 0%, #0d0d0f 100%);
    display: flex;
    align-items: flex-end;
    padding: 2.5rem;
  }

  .hero__fallback-title {
    font-size: 2rem;
    font-weight: 800;
    color: #334155;
  }

  /* Multi-layered Gradients for Deep Contrast */
  .hero__overlay-bottom {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to top,
      #0d0d0f 0%,
      rgba(13, 13, 15, 0.8) 25%,
      rgba(13, 13, 15, 0.2) 60%,
      transparent 100%
    );
  }

  .hero__overlay-side {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to right,
      rgba(13, 13, 15, 0.95) 0%,
      rgba(13, 13, 15, 0.6) 45%,
      transparent 80%
    );
  }

  .hero__content-wrapper {
    position: absolute;
    bottom: 50px;
    left: 50%;
    transform: translateX(-50%);
    width: 100%;
    max-width: 1240px;
    padding: 0 1.5rem;
    z-index: 10;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  @media (max-width: 768px) {
    .hero__content-wrapper {
      bottom: 36px;
      padding: 0 1rem;
      gap: 0.5rem;
    }
  }

  .hero__badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: 20px;
    background: rgba(245, 158, 11, 0.15);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: #f59e0b;
    font-size: 0.75rem;
    font-weight: 600;
    width: fit-content;
    backdrop-filter: blur(8px);
  }

  .hero__title {
    font-size: 2.25rem;
    font-weight: 800;
    color: #ffffff;
    margin: 0;
    line-height: 1.15;
    text-shadow: 0 2px 10px rgba(0, 0, 0, 0.7);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
    max-width: 680px;
  }

  @media (max-width: 768px) {
    .hero__title {
      font-size: 1.5rem;
    }
  }

  .hero__meta {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.875rem;
    color: #cbd5e1;
    font-weight: 500;
  }

  .meta-item.rating {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #fbbf24;
    font-weight: 700;
  }

  .meta-dot {
    color: #64748b;
  }

  .hero__overview {
    font-size: 0.925rem;
    color: #94a3b8;
    line-height: 1.5;
    max-width: 600px;
    margin: 0;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  @media (max-width: 768px) {
    .hero__overview {
      display: none;
    }
  }

  .hero__actions {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 0.25rem;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
    border-radius: 10px;
    font-size: 0.875rem;
    font-weight: 600;
    text-decoration: none;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    cursor: pointer;
  }

  .btn--primary {
    background: var(--color-brand);
    color: #ffffff;
    box-shadow: 0 4px 14px rgba(245, 67, 44, 0.4);
  }

  .btn--primary:hover {
    background: var(--color-brand-hover);
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(245, 67, 44, 0.6);
  }

  .btn--secondary {
    background: rgba(255, 255, 255, 0.1);
    color: #ffffff;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.15);
  }

  .btn--secondary:hover {
    background: rgba(255, 255, 255, 0.2);
    transform: translateY(-2px);
  }

  @media (max-width: 480px) {
    .btn {
      padding: 8px 14px;
      font-size: 0.8rem;
    }
  }

  /* Slide Indicators */
  .hero__dots {
    position: absolute;
    bottom: 16px;
    right: 50%;
    transform: translateX(50%);
    display: flex;
    align-items: center;
    gap: 8px;
    z-index: 10;
  }

  .hero__dot {
    position: relative;
    width: 24px;
    height: 4px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.25);
    border: none;
    cursor: pointer;
    overflow: hidden;
    padding: 0;
    transition: width 0.3s;
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

  /* ── Quick Category Chips Navigation (Fully Responsive) ─────── */
  .category-bar-wrapper {
    max-width: 1240px;
    margin: 1.75rem auto 0;
    padding: 0 1.5rem;
    width: 100%;
  }

  .category-bar {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    gap: 10px;
    width: 100%;
  }

  .category-chip {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 18px;
    border-radius: 9999px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #cbd5e1;
    font-size: 0.85rem;
    font-weight: 500;
    text-decoration: none;
    white-space: nowrap;
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    flex-shrink: 0;
  }

  .category-chip:hover {
    background: rgba(245, 67, 44, 0.18);
    border-color: rgba(245, 67, 44, 0.45);
    color: #ffffff;
    transform: translateY(-2px);
    box-shadow: 0 4px 14px rgba(245, 67, 44, 0.25);
  }

  .category-chip:active {
    transform: translateY(0);
  }

  .chip-icon {
    color: var(--color-brand);
    transition: transform 0.2s ease;
  }

  .category-chip:hover .chip-icon {
    transform: scale(1.15);
  }

  /* Responsive Adjustments for Mobile & Tablet */
  @media (max-width: 768px) {
    .category-bar-wrapper {
      margin-top: 1.25rem;
      padding: 0;
    }

    .category-bar {
      justify-content: flex-start;
      flex-wrap: nowrap;
      overflow-x: auto;
      scroll-behavior: smooth;
      -webkit-overflow-scrolling: touch;
      padding: 4px 1rem 12px;
      gap: 8px;
      scrollbar-width: none;
    }

    .category-bar::-webkit-scrollbar {
      display: none;
    }

    .category-chip {
      padding: 7px 14px;
      font-size: 0.8rem;
      gap: 6px;
    }
  }

  @media (max-width: 480px) {
    .category-chip {
      padding: 6px 12px;
      font-size: 0.775rem;
      gap: 5px;
    }
  }

  /* ── Main Content Area ───────────────────────────────────── */
  .content-area {
    padding: 1.5rem 1.5rem 4rem;
    max-width: 1240px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 3rem;
  }

  @media (max-width: 768px) {
    .content-area {
      padding: 1rem 1rem 3rem;
      gap: 2rem;
    }
  }

  /* ── Section Header ───────────────────────────────────────── */
  .section-header {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    margin-bottom: 1.25rem;
  }

  .header-title-group {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .icon-badge {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    border-radius: 10px;
    flex-shrink: 0;
  }

  .icon-badge--amber {
    background: rgba(245, 158, 11, 0.15);
    color: #f59e0b;
    border: 1px solid rgba(245, 158, 11, 0.25);
  }

  .icon-badge--red {
    background: rgba(245, 67, 44, 0.15);
    color: var(--color-brand);
    border: 1px solid rgba(245, 67, 44, 0.25);
  }

  .icon-badge--blue {
    background: rgba(59, 130, 246, 0.15);
    color: #3b82f6;
    border: 1px solid rgba(59, 130, 246, 0.25);
  }

  .icon-badge--emerald {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
    border: 1px solid rgba(16, 185, 129, 0.25);
  }

  .section-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: #ffffff;
    margin: 0;
    line-height: 1.2;
  }

  .section-subtitle {
    font-size: 0.8rem;
    color: #64748b;
    margin: 2px 0 0;
  }

  @media (max-width: 640px) {
    .section-title {
      font-size: 1.05rem;
    }
    .section-subtitle {
      display: none;
    }
    .icon-badge {
      width: 32px;
      height: 32px;
    }
  }

  .see-all-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.85rem;
    font-weight: 600;
    color: #94a3b8;
    text-decoration: none;
    transition: color 0.2s;
  }

  .see-all-btn:hover {
    color: var(--color-brand);
  }

  /* ── Responsive Movie Grid (Critical Mobile Fix!) ─────────── */
  /* ── Responsive Movie Grid (100% Equal Cards Fix!) ─────────── */
  .movie-grid {
    display: grid;
    /* Desktop default: 5 equal columns with strict minmax(0, 1fr) */
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 16px;
    align-items: start;
    width: 100%;
  }

  /* Laptop & Small Screens: 4 columns */
  @media (max-width: 1024px) {
    .movie-grid {
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 14px;
    }
  }

  /* Tablet: 3 columns */
  @media (max-width: 768px) {
    .movie-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 12px;
    }
  }

  /* Mobile: 2 columns */
  @media (max-width: 520px) {
    .movie-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 12px;
    }
  }

  .poster-wrap {
    position: relative;
    width: 100%;
    min-width: 0;
  }

  .poster-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    text-decoration: none;
    color: inherit;
    border-radius: 12px;
    transition: transform 0.25s ease, box-shadow 0.25s ease;
    width: 100%;
    min-width: 0;
  }

  .poster-card:hover {
    transform: translateY(-4px);
  }

  .poster-image-container {
    position: relative;
    width: 100%;
    aspect-ratio: 2 / 3;
    flex-shrink: 0;
    border-radius: 12px;
    overflow: hidden;
    background: #18181f;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .poster-image-container img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transition: transform 0.35s ease;
  }

  .poster-card:hover .poster-image-container img {
    transform: scale(1.05);
  }

  /* Rating Badge on Image */
  .rating-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    border-radius: 6px;
    background: rgba(13, 13, 15, 0.75);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #f1f5f9;
    font-size: 0.725rem;
    font-weight: 700;
  }

  /* Movie Title & Info Under Poster */
  .poster-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 0 2px;
    min-height: 40px;
    width: 100%;
    min-width: 0;
  }

  .poster-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: #f1f5f9;
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: 1.3;
    width: 100%;
    max-width: 100%;
  }

  .poster-year {
    font-size: 0.75rem;
    color: #64748b;
  }

  @media (max-width: 520px) {
    .poster-title {
      font-size: 0.825rem;
    }
    .poster-year {
      font-size: 0.7rem;
    }
  }

  /* ── Fallback Card ────────────────────────────────────────── */
  .poster-fallback {
    width: 100%;
    height: 100%;
    background: #18181f;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }

  .poster-fallback__title {
    font-size: 0.8rem;
    color: #64748b;
    text-align: center;
    line-height: 1.4;
  }

  /* ── Hover Popup (Desktop) ────────────────────────────────── */
  .hover-popup {
    position: absolute;
    top: 0;
    z-index: 60;
    width: 320px;
    background: #18181f;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 14px;
    overflow: hidden;
    box-shadow:
      0 24px 70px rgba(0, 0, 0, 0.85),
      0 0 0 1px rgba(255, 255, 255, 0.05);
  }

  .popup--right {
    left: calc(100% + 10px);
  }

  .popup--left {
    right: calc(100% + 10px);
  }

  .popup--center {
    left: 50%;
    transform: translateX(-50%);
  }

  .popup-enter-active {
    transition:
      opacity 0.2s,
      transform 0.2s cubic-bezier(0.2, 0, 0.13, 1.35);
  }

  .popup-leave-active {
    transition: opacity 0.15s;
  }

  .popup-enter-from {
    opacity: 0;
    transform: scale(0.92) translateY(6px);
  }

  .popup--center.popup-enter-from {
    transform: translateX(-50%) scale(0.92) translateY(6px);
  }

  .popup-leave-to {
    opacity: 0;
  }

  /* ── Loading Overlay ──────────────────────────────────────── */
  .loading-overlay {
    position: fixed;
    inset: 0;
    background: rgba(13, 13, 15, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    backdrop-filter: blur(4px);
  }

  .fade-enter-active,
  .fade-leave-active {
    transition: opacity 0.25s;
  }
  .fade-enter-from,
  .fade-leave-to {
    opacity: 0;
  }
</style>
