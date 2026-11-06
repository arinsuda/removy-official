<template>
  <div class="onboarding-page">
    <div class="onboarding-bg"><div class="onboarding-bg__overlay" /></div>

    <div class="onboarding-container" ref="containerRef">
      <!-- Header -->
      <div class="onboarding-header" ref="headerRef">
        <div class="brand-logo">
          <span class="brand-movie">MOVIE</span
          ><span class="brand-hub">HUB</span>
        </div>
        <h1 class="onboarding-title">เลือกหนังที่คุณชอบ</h1>
        <p class="onboarding-sub">
          เลือกอย่างน้อย {{ MIN_PICKS }} เรื่องที่คุณชอบหรือเคยดูแล้ว
          เพื่อให้เราแนะนำหนังที่ใช่สำหรับคุณ
        </p>

        <!-- Progress bar -->
        <div class="progress-bar-wrap">
          <div
            class="progress-bar-fill"
            ref="progressBarRef"
            :style="{
              width: `${Math.min((selectedIds.length / MIN_PICKS) * 100, 100)}%`,
            }"
            :class="{
              'progress-bar-fill--done': selectedIds.length >= MIN_PICKS,
            }"
          />
        </div>
        <p class="progress-label">
          <span :class="{ 'count--done': selectedIds.length >= MIN_PICKS }">
            {{ selectedIds.length }}
          </span>
          / {{ MIN_PICKS }} เรื่องขึ้นไป
        </p>

        <!-- Genre tags display -->
        <div class="genre-tags" v-if="genreLabels.length">
          <TransitionGroup name="tag-pop">
            <span v-for="label in genreLabels" :key="label" class="genre-tag">{{
              label
            }}</span>
          </TransitionGroup>
        </div>
      </div>

      <!-- Movie Grid -->
      <div class="movie-grid" ref="gridRef" v-if="movies.length">
        <button
          v-for="(movie, index) in movies"
          :key="movie.id"
          class="movie-card"
          :class="{ 'movie-card--selected': selectedIds.includes(movie.id) }"
          :data-index="index"
          @click="toggleMovie(movie)"
          @mouseenter="onCardHover($event, true)"
          @mouseleave="onCardHover($event, false)"
        >
          <img
            v-if="movie.poster_path && !failedIds.has(movie.id)"
            :src="`https://image.tmdb.org/t/p/w185${movie.poster_path}`"
            :alt="movie.title"
            loading="lazy"
            class="movie-poster"
            @error="failedIds.add(movie.id)"
          />
          <div v-else class="movie-poster movie-poster--placeholder">
            <Film :size="28" />
            <span class="placeholder-title">{{ movie.title }}</span>
          </div>

          <!-- Genre badge -->
          <div
            class="movie-genre-badge"
            v-if="getGenreName(movie.genre_ids?.[0])"
          >
            {{ getGenreName(movie.genre_ids?.[0]) }}
          </div>

          <!-- Overlay on selected -->
          <Transition name="check-fade">
            <div v-if="selectedIds.includes(movie.id)" class="movie-check">
              <Check :size="22" />
            </div>
          </Transition>

          <!-- Title tooltip on hover -->
          <div class="movie-title-bar">
            <span class="movie-title">{{ movie.title }}</span>
          </div>
        </button>
      </div>

      <!-- Skeleton -->
      <div class="movie-grid" v-else>
        <div
          v-for="i in TOTAL_MOVIES"
          :key="i"
          class="movie-card movie-card--skeleton"
        />
      </div>

      <!-- Footer -->
      <div class="onboarding-footer" ref="footerRef">
        <button
          class="btn-confirm"
          :disabled="selectedIds.length < MIN_PICKS || isSaving"
          @click="handleConfirm"
        >
          <span v-if="isSaving" class="spinner" />
          <span v-else>เริ่มต้นใช้งาน</span>
        </button>
        <button class="btn-skip" @click="handleSkip" :disabled="isSaving">
          ข้ามไปก่อน
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, nextTick } from "vue"
  import { useRouter } from "vue-router"
  import { movieApi, userApi } from "@/api/api"
  import { useAuthStore } from "@/stores/auth"
  import { Film, Check } from "lucide-vue-next"
  import type { Movie } from "@/types"
  import { gsap } from "gsap"

  const router = useRouter()
  const authStore = useAuthStore()

  const MIN_PICKS = 5
  const TOTAL_MOVIES = 20

  // TMDB Genre map (most common ones)
  const GENRE_MAP: Record<number, string> = {
    28: "แอ็กชัน",
    12: "ผจญภัย",
    16: "แอนิเมชัน",
    35: "คอมเมดี้",
    80: "อาชญากรรม",
    99: "สารคดี",
    18: "ดราม่า",
    10751: "ครอบครัว",
    14: "แฟนตาซี",
    36: "ประวัติศาสตร์",
    27: "สยองขวัญ",
    10402: "ดนตรี",
    9648: "ลึกลับ",
    10749: "โรแมนติก",
    878: "ไซไฟ",
    10770: "ทีวีมูฟวี่",
    53: "ระทึกขวัญ",
    10752: "สงคราม",
    37: "คาวบอย",
  }

  // Target genre IDs to ensure variety — pick 1 representative movie per genre bucket
  const TARGET_GENRES = [28, 35, 18, 878, 27, 14, 12, 80, 10749, 53]

  const movies = ref<Movie[]>([])
  const selectedIds = ref<number[]>([])
  const selectedMovies = ref<Movie[]>([])
  const isSaving = ref(false)
  const failedIds = ref(new Set<number>())

  const containerRef = ref<HTMLElement | null>(null)
  const headerRef = ref<HTMLElement | null>(null)
  const gridRef = ref<HTMLElement | null>(null)
  const footerRef = ref<HTMLElement | null>(null)
  const progressBarRef = ref<HTMLElement | null>(null)

  // Computed genre labels from selected movies
  const genreLabels = computed(() => {
    const ids = new Set<number>()
    for (const movie of selectedMovies.value) {
      for (const gid of movie.genre_ids ?? []) {
        ids.add(gid)
      }
    }
    return [...ids]
      .filter(id => GENRE_MAP[id])
      .slice(0, 6)
      .map(id => GENRE_MAP[id]!)
  })

  function getGenreName(id?: number): string {
    if (!id) return ""
    return GENRE_MAP[id] ?? ""
  }

  function toggleMovie(movie: Movie) {
    const idx = selectedIds.value.indexOf(movie.id)
    if (idx !== -1) {
      selectedIds.value.splice(idx, 1)
      selectedMovies.value = selectedMovies.value.filter(m => m.id !== movie.id)
      // Animate deselect
      const card = gridRef.value?.querySelector(
        `[data-index="${movies.value.findIndex(m => m.id === movie.id)}"]`,
      )
      if (card) {
        gsap.to(card, {
          scale: 0.95,
          duration: 0.1,
          yoyo: true,
          repeat: 1,
          ease: "power2.inOut",
        })
      }
    } else {
      selectedIds.value.push(movie.id)
      selectedMovies.value.push(movie)
      // Animate select — punch effect
      const card = gridRef.value?.querySelector(
        `[data-index="${movies.value.findIndex(m => m.id === movie.id)}"]`,
      )
      if (card) {
        gsap.fromTo(
          card,
          { scale: 0.92 },
          { scale: 1, duration: 0.35, ease: "back.out(2)" },
        )
      }
    }
  }

  function extractGenres(): number[] {
    const freq: Record<number, number> = {}
    for (const movie of selectedMovies.value) {
      for (const gid of movie.genre_ids ?? []) {
        freq[gid] = (freq[gid] ?? 0) + 1
      }
    }
    return Object.entries(freq)
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5)
      .map(([id]) => Number(id))
  }

  async function handleConfirm() {
    if (selectedIds.value.length < MIN_PICKS || isSaving.value) return
    isSaving.value = true

    // Animate confirm button
    const btn = document.querySelector(".btn-confirm")
    if (btn) gsap.to(btn, { scale: 0.96, duration: 0.1, yoyo: true, repeat: 1 })

    try {
      const genres = extractGenres()
      await userApi.updateFavoriteGenres(authStore.user!.id, genres)
      authStore.setUser({
        ...authStore.user!,
        favorite_genres: JSON.stringify(genres),
      })
      router.push({ name: "home" })
    } catch {
      // handle error
    } finally {
      isSaving.value = false
    }
  }

  async function handleSkip() {
    try {
      await userApi.updateFavoriteGenres(authStore.user!.id, [])

      authStore.setUser({
        ...authStore.user!,
        favorite_genres: "skip",
      })
    } catch {
      /* ไม่ block user */
    }
    authStore.setUser({ ...authStore.user!, favorite_genres: "skip" })
    router.push({ name: "home" })
  }

  function onCardHover(event: MouseEvent, entering: boolean) {
    const card = event.currentTarget as HTMLElement
    if (entering) {
      gsap.to(card, { y: -6, scale: 1.05, duration: 0.25, ease: "power2.out" })
    } else {
      gsap.to(card, { y: 0, scale: 1, duration: 0.2, ease: "power2.out" })
    }
  }

  /** Build a genre-diverse pool from multiple API pages */
  async function buildDiversePool(): Promise<Movie[]> {
    // Fetch from 3 sources: popular (2 pages) + top-rated (1 page)
    const [pop1, pop2, top1] = await Promise.all([
      movieApi.getPopular(1).then(r => r.data.results as Movie[]),
      movieApi.getPopular(2).then(r => r.data.results as Movie[]),
      movieApi.getTopRated(1).then(r => r.data.results as Movie[]),
    ])

    // Deduplicate by id
    const seen = new Set<number>()
    const allMovies: Movie[] = []
    for (const m of [...pop1, ...pop2, ...top1]) {
      if (!seen.has(m.id) && m.poster_path) {
        seen.add(m.id)
        allMovies.push(m as Movie)
      }
    }

    // Pick best representative per target genre (highest vote_average + vote_count combo)
    const pickedIds = new Set<number>()
    const result: Movie[] = []

    // Score = vote_average * log(vote_count)
    const score = (m: Movie) => m.vote_average * Math.log(m.vote_count + 1)

    for (const genreId of TARGET_GENRES) {
      const candidates = allMovies
        .filter(m => !pickedIds.has(m.id) && m.genre_ids?.includes(genreId))
        .sort((a, b) => score(b) - score(a))

      // Pick top 2 per genre for variety
      const picks = candidates.slice(0, 2)
      for (const p of picks) {
        pickedIds.add(p.id)
        result.push(p)
      }
    }

    // Fill remaining slots with best-scored movies not yet picked
    const remaining = allMovies
      .filter(m => !pickedIds.has(m.id))
      .sort((a, b) => score(b) - score(a))

    for (const m of remaining) {
      if (result.length >= TOTAL_MOVIES) break
      result.push(m)
      pickedIds.add(m.id)
    }

    // Shuffle the final result so it doesn't feel ordered by genre
    for (let i = result.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1))
      ;[result[i], result[j]] = [result[j]!, result[i]!]
    }

    return result.slice(0, TOTAL_MOVIES)
  }

  // ── GSAP entrance animation ──────────────────────────────────────
  function runEntranceAnimation() {
    const tl = gsap.timeline({ defaults: { ease: "power3.out" } })

    // Header elements
    tl.fromTo(
      headerRef.value,
      { opacity: 0, y: -30 },
      { opacity: 1, y: 0, duration: 0.7 },
    )

    // Grid cards stagger
    const cards = gridRef.value?.querySelectorAll(".movie-card")
    if (cards?.length) {
      tl.fromTo(
        cards,
        { opacity: 0, y: 40, scale: 0.88 },
        {
          opacity: 1,
          y: 0,
          scale: 1,
          duration: 0.5,
          stagger: {
            amount: 0.8,
            from: "random",
            ease: "power1.inOut",
          },
        },
        "-=0.3",
      )
    }

    // Footer
    tl.fromTo(
      footerRef.value,
      { opacity: 0, y: 20 },
      { opacity: 1, y: 0, duration: 0.5 },
      "-=0.3",
    )
  }

  onMounted(async () => {
    // Set initial states
    gsap.set(headerRef.value, { opacity: 0 })
    gsap.set(footerRef.value, { opacity: 0 })

    try {
      movies.value = await buildDiversePool()
      await nextTick()
      runEntranceAnimation()
    } catch {
      // fallback: stay empty
    }
  })
</script>

<style scoped>
  /* ── Page shell ───────────────────────────────────────── */
  .onboarding-page {
    position: relative;
    min-height: 100vh;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    background: #0a0a0a;
    padding: 2rem 1rem 3rem;
  }

  .onboarding-bg {
    position: fixed;
    inset: 0;
    background:
      url("https://image.tmdb.org/t/p/w300/qJ2tW6WMUDux911r6m7haRef0WH.jpg") 0
        0 / 20% auto repeat-y,
      url("https://image.tmdb.org/t/p/w300/udDclJoHjfjb8Ekgsd4FDteOkCU.jpg") 20%
        0 / 20% auto repeat-y,
      url("https://image.tmdb.org/t/p/w300/rAiYTfKGqDCRIIqo664sY9XZIvQ.jpg") 40%
        0 / 20% auto repeat-y,
      url("https://image.tmdb.org/t/p/w300/b0PlSFdDwbyK0cf5RxwDpaOJQvQ.jpg") 60%
        0 / 20% auto repeat-y,
      url("https://image.tmdb.org/t/p/w300/74xTEgt7R36Fpooo50r9T25onhq.jpg") 80%
        0 / 20% auto repeat-y;
    filter: brightness(0.18) saturate(0.4);
    pointer-events: none;
  }

  .onboarding-bg__overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to bottom,
      rgba(10, 10, 10, 0.5) 0%,
      rgba(10, 10, 10, 0.8) 100%
    );
  }

  /* ── Container ────────────────────────────────────────── */
  .onboarding-container {
    position: relative;
    z-index: 1;
    width: 100%;
    max-width: 900px;
    display: flex;
    flex-direction: column;
    gap: 1.75rem;
  }

  /* ── Header ───────────────────────────────────────────── */
  .onboarding-header {
    text-align: center;
  }

  .brand-logo {
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 2rem;
    font-weight: 900;
    margin-bottom: 1rem;
  }
  .brand-movie {
    color: #fff;
  }
  .brand-hub {
    color: var(--color-brand);
  }

  .onboarding-title {
    font-size: 1.6rem;
    font-weight: 700;
    color: #fff;
    margin: 0 0 0.5rem;
  }

  .onboarding-sub {
    font-size: 0.875rem;
    color: #a3a3a3;
    margin: 0 0 1.25rem;
    line-height: 1.6;
  }

  /* Progress */
  .progress-bar-wrap {
    width: 100%;
    max-width: 320px;
    margin: 0 auto 0.4rem;
    height: 4px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 9999px;
    overflow: hidden;
  }
  .progress-bar-fill {
    height: 100%;
    background: var(--color-brand);
    border-radius: 9999px;
    transition:
      width 0.4s cubic-bezier(0.34, 1.56, 0.64, 1),
      background 0.3s;
  }
  .progress-bar-fill--done {
    background: #22c55e;
  }

  .progress-label {
    font-size: 0.8rem;
    color: #666;
    margin: 0 0 0.75rem;
  }
  .count--done {
    color: #22c55e;
    font-weight: 700;
  }

  /* ── Genre tags ───────────────────────────────────────── */
  .genre-tags {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.4rem;
    min-height: 28px;
  }

  .genre-tag {
    background: rgba(245, 67, 44, 0.15);
    border: 1px solid rgba(245, 67, 44, 0.4);
    color: #ff6b6b;
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.2rem 0.6rem;
    border-radius: 9999px;
    letter-spacing: 0.03em;
  }

  .tag-pop-enter-active {
    transition: all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
  }
  .tag-pop-leave-active {
    transition: all 0.15s ease;
  }
  .tag-pop-enter-from {
    opacity: 0;
    transform: scale(0.6) translateY(4px);
  }
  .tag-pop-leave-to {
    opacity: 0;
    transform: scale(0.8);
  }

  /* ── Movie Grid ───────────────────────────────────────── */
  .movie-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 10px;
  }

  @media (max-width: 680px) {
    .movie-grid {
      grid-template-columns: repeat(4, 1fr);
    }
  }
  @media (max-width: 480px) {
    .movie-grid {
      grid-template-columns: repeat(3, 1fr);
    }
  }

  /* ── Movie Card ───────────────────────────────────────── */
  .movie-card {
    position: relative;
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    border: 2px solid transparent;
    background: #1a1a1a;
    padding: 0;
    /* GSAP handles hover transform — CSS transition only for border/shadow */
    transition:
      border-color 0.2s,
      box-shadow 0.2s;
    aspect-ratio: 2/3;
    display: flex;
    flex-direction: column;
    will-change: transform;
  }

  .movie-card--selected {
    border-color: var(--color-brand) !important;
    box-shadow:
      0 0 0 1px var(--color-brand),
      0 8px 28px rgba(245, 67, 44, 0.3) !important;
  }

  .movie-card--skeleton {
    background: linear-gradient(90deg, #1f1f1f 25%, #2a2a2a 50%, #1f1f1f 75%);
    background-size: 200% 100%;
    animation: shimmer 1.2s infinite;
    cursor: default;
    border-color: transparent !important;
    aspect-ratio: 2/3;
  }

  @keyframes shimmer {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }

  /* Poster image */
  .movie-poster {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    flex: 1;
  }

  .movie-poster--placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    background: #1a1a1a;
    color: #333;
    flex: 1;
  }

  /* Genre badge */
  .movie-genre-badge {
    position: absolute;
    top: 6px;
    left: 6px;
    background: rgba(0, 0, 0, 0.75);
    backdrop-filter: blur(4px);
    color: #e0e0e0;
    font-size: 0.55rem;
    font-weight: 700;
    padding: 0.15rem 0.45rem;
    border-radius: 4px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    pointer-events: none;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  /* Checkmark overlay */
  .movie-check {
    position: absolute;
    inset: 0;
    background: rgba(245, 67, 44, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    backdrop-filter: blur(1px);
  }

  .placeholder-title {
    font-size: 0.6rem;
    color: #555;
    text-align: center;
    padding: 0 0.4rem;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .check-fade-enter-active {
    transition: opacity 0.15s;
  }
  .check-fade-leave-active {
    transition: opacity 0.1s;
  }
  .check-fade-enter-from,
  .check-fade-leave-to {
    opacity: 0;
  }

  /* Title bar (hover) */
  .movie-title-bar {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    padding: 1.5rem 0.5rem 0.4rem;
    background: linear-gradient(
      to top,
      rgba(0, 0, 0, 0.85) 0%,
      transparent 100%
    );
    opacity: 0;
    transition: opacity 0.2s;
    pointer-events: none;
  }

  .movie-card:hover .movie-title-bar {
    opacity: 1;
  }

  .movie-title {
    display: block;
    font-size: 0.65rem;
    font-weight: 600;
    color: #fff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    line-height: 1.3;
  }

  /* ── Footer ───────────────────────────────────────────── */
  .onboarding-footer {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
    padding-top: 0.25rem;
  }

  .btn-confirm {
    width: 100%;
    max-width: 320px;
    padding: 0.85rem;
    background: var(--color-brand);
    color: #fff;
    border: none;
    border-radius: 9999px;
    font-family: "Noto Sans Thai", "Arial Black", sans-serif;
    font-size: 1.1rem;
    font-weight: 900;
    letter-spacing: 2px;
    cursor: pointer;
    transition:
      background 0.2s,
      opacity 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 48px;
    will-change: transform;
  }
  .btn-confirm:hover:not(:disabled) {
    background: var(--color-brand-hover);
  }
  .btn-confirm:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .btn-skip {
    background: none;
    border: none;
    color: #555;
    font-size: 0.82rem;
    cursor: pointer;
    text-decoration: underline;
    text-underline-offset: 3px;
    transition: color 0.2s;
  }
  .btn-skip:hover {
    color: #a3a3a3;
  }

  .spinner {
    width: 20px;
    height: 20px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: #fff;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
