<template>
  <div class="upcoming-page" ref="pageRef">
    <div class="ambient-bg">
      <div
        class="ambient-blob"
        :style="
          featuredMovie
            ? `background-image: url(https://image.tmdb.org/t/p/w780${featuredMovie.backdrop_path})`
            : ''
        "
      />
      <div class="ambient-overlay" />
    </div>

    <main class="page-main">
      <header class="page-header" ref="introRef">
        <div class="page-header__eyebrow">
          <span class="eyebrow-dot" />
          <span class="eyebrow-text">POPCORN</span>
        </div>
        <div class="page-header__row">
          <div>
            <h1 class="page-header__title">อัปเดตภาพยนตร์ล่าสุด</h1>
            <p class="page-header__sub">
              ติดตามภาพยนตร์ที่กำลังฉายและกำลังจะเข้าฉายเร็วๆ นี้
            </p>
          </div>
        </div>
      </header>

      <section class="hero" ref="heroRef" v-if="featuredMovie">
        <div class="hero__backdrop">
          <img
            :src="`https://image.tmdb.org/t/p/w1280${featuredMovie.backdrop_path}`"
            :alt="featuredMovie.title"
            class="hero__backdrop-img"
          />
          <div class="hero__backdrop-grad" />
        </div>
        <div class="hero__inner">
          <div class="hero__poster-col">
            <div class="hero__poster-wrap">
              <img
                :src="`https://image.tmdb.org/t/p/w342${featuredMovie.poster_path}`"
                :alt="featuredMovie.title"
                class="hero__poster"
              />
              <div class="hero__poster-glow" />
            </div>
          </div>
          <div class="hero__info">
            <span class="hero__badge">
              <i
                class="pi pi-star-fill"
                style="font-size: 0.65rem; margin-right: 0.25rem"
              ></i>
              เรื่องที่เลือก
            </span>
            <h2 class="hero__title">{{ featuredMovie.title }}</h2>
            <p class="hero__overview">
              {{ featuredMovie.overview?.slice(0, 220)
              }}{{ (featuredMovie.overview?.length ?? 0) > 220 ? "…" : "" }}
            </p>
            <div class="hero__meta-row">
              <span class="meta-tag">
                <i class="pi pi-calendar" style="font-size: 0.75rem"></i>
                {{ formatDate(featuredMovie.release_date) }}
              </span>
            </div>

            <div
              class="countdown-wrapper"
              v-if="!isReleased(featuredMovie.release_date)"
            >
              <div class="countdown-title">
                <i
                  class="pi pi-clock"
                  style="font-size: 0.7rem; margin-right: 0.3rem"
                ></i>
                นับถอยหลังวันเข้าฉาย
              </div>
              <div class="countdown-clock">
                <div class="countdown-box">
                  <span class="countdown-num">{{ countdown.days }}</span>
                  <span class="countdown-label">วัน</span>
                </div>
                <div class="countdown-box">
                  <span class="countdown-num">{{
                    formatNumber(countdown.hours)
                  }}</span>
                  <span class="countdown-label">ชม.</span>
                </div>
                <div class="countdown-box">
                  <span class="countdown-num">{{
                    formatNumber(countdown.minutes)
                  }}</span>
                  <span class="countdown-label">นาที</span>
                </div>
                <div class="countdown-box">
                  <span class="countdown-num countdown-num--sec">{{
                    formatNumber(countdown.seconds)
                  }}</span>
                  <span class="countdown-label">วิ</span>
                </div>
              </div>
            </div>

            <div class="released-banner" v-else>
              <span class="meta-tag meta-tag--green">
                <i class="pi pi-ticket" style="font-size: 0.75rem"></i>
                เข้าฉายในโรงภาพยนตร์เรียบร้อยแล้ว
              </span>
            </div>

            <button class="btn-detail" @click="goToDetail(featuredMovie.id)">
              <span>ดูรายละเอียด</span>
              <i class="pi pi-arrow-right" style="font-size: 0.75rem"></i>
            </button>
          </div>
        </div>
      </section>

      <section class="movie-section" ref="nowPlayingRef">
        <div class="category-divider">
          <h2 class="category-title">
            <i
              class="pi pi-play-circle"
              style="
                color: var(--red);
                margin-right: 0.5rem;
                font-size: 1.1rem;
                vertical-align: middle;
              "
            ></i>
            กำลังฉายในโรงภาพยนตร์ (Latest)
          </h2>
          <span class="category-count"
            >{{ nowPlayingMovies.length }} เรื่อง</span
          >
        </div>

        <div class="movie-grid">
          <button
            v-for="movie in nowPlayingMovies"
            :key="`now-${movie.id}`"
            class="movie-card"
            :class="{ 'movie-card--active': movie.id === featuredMovie?.id }"
            @click="setFeatured(movie)"
          >
            <div class="movie-card__poster-wrap">
              <img
                v-if="movie.poster_path"
                :src="`https://image.tmdb.org/t/p/w185${movie.poster_path}`"
                :alt="movie.title"
                class="movie-card__poster"
                loading="lazy"
              />
              <div v-else class="movie-card__poster movie-card__poster--empty">
                <i class="pi pi-image" style="font-size: 1.2rem"></i>
              </div>
              <div class="movie-card__overlay">
                <i
                  class="pi pi-play"
                  style="font-size: 1.25rem; color: #fff"
                ></i>
              </div>
              <div class="movie-card__badge movie-card__badge--out">
                เข้าฉายแล้ว
              </div>
            </div>
            <div class="movie-card__info">
              <span class="movie-card__title">{{ movie.title }}</span>
              <span class="movie-card__date">{{
                formatDateShort(movie.release_date)
              }}</span>
            </div>
          </button>
        </div>

        <div class="load-more-section" v-if="nowPlayingHasMore">
          <button
            class="btn-load-more"
            @click="loadMoreNowPlaying"
            :disabled="isNowPlayingLoading"
          >
            <AppSpinner v-if="isNowPlayingLoading" />
            <span v-else>เพิ่มเติมสำหรับกําลังฉาย</span>
          </button>
        </div>
      </section>

      <section class="movie-section" ref="upcomingRef" style="margin-top: 4rem">
        <div class="category-divider">
          <h2 class="category-title">
            <i
              class="pi pi-hourglass"
              style="
                color: var(--gold);
                margin-right: 0.5rem;
                font-size: 1.1rem;
                vertical-align: middle;
              "
            ></i>
            ภาพยนตร์เร็วๆ นี้ (Coming Soon)
          </h2>
          <span class="category-count">{{ upcomingMovies.length }} เรื่อง</span>
        </div>

        <div class="movie-grid">
          <button
            v-for="movie in upcomingMovies"
            :key="`up-${movie.id}`"
            class="movie-card"
            :class="{ 'movie-card--active': movie.id === featuredMovie?.id }"
            @click="setFeatured(movie)"
          >
            <div class="movie-card__poster-wrap">
              <img
                v-if="movie.poster_path"
                :src="`https://image.tmdb.org/t/p/w185${movie.poster_path}`"
                :alt="movie.title"
                class="movie-card__poster"
                loading="lazy"
              />
              <div v-else class="movie-card__poster movie-card__poster--empty">
                <i class="pi pi-image" style="font-size: 1.2rem"></i>
              </div>
              <div class="movie-card__overlay">
                <i
                  class="pi pi-play"
                  style="font-size: 1.25rem; color: #fff"
                ></i>
              </div>
              <div class="movie-card__badge movie-card__badge--soon">
                อีก {{ getDaysLeft(movie.release_date) }} วัน
              </div>
            </div>
            <div class="movie-card__info">
              <span class="movie-card__title">{{ movie.title }}</span>
              <span class="movie-card__date">{{
                formatDateShort(movie.release_date)
              }}</span>
            </div>
          </button>
        </div>

        <div class="load-more-section" v-if="upcomingHasMore">
          <button
            class="btn-load-more"
            @click="loadMoreUpcoming"
            :disabled="isUpcomingLoading"
          >
            <AppSpinner v-if="isUpcomingLoading" />
            <span v-else>เพิ่มเติมสำหรับเร็วๆ นี้</span>
          </button>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import { ref, onMounted, onUnmounted, watch } from "vue"
  import { useRouter } from "vue-router"
  import { movieApi } from "@/api/api"
  import type { Movie } from "@/types"
  import { gsap } from "gsap"
  import { ScrollTrigger } from "gsap/ScrollTrigger"

  // สังเกตว่าเราไม่ต้อง import ไอคอนรายตัวแล้ว เพราะ PrimeIcons เรียกใช้ผ่าน class string ใน template ได้เลย!

  gsap.registerPlugin(ScrollTrigger)
  const router = useRouter()

  const pageRef = ref<HTMLElement | null>(null)
  const introRef = ref<HTMLElement | null>(null)
  const heroRef = ref<HTMLElement | null>(null)
  const nowPlayingRef = ref<HTMLElement | null>(null)
  const upcomingRef = ref<HTMLElement | null>(null)

  const featuredMovie = ref<Movie | null>(null)

  const countdown = ref({ days: 0, hours: 0, minutes: 0, seconds: 0 })
  let countdownTimer: number | null = null

  const nowPlayingMovies = ref<Movie[]>([])
  const nowPlayingPage = ref(1)
  const nowPlayingHasMore = ref(true)
  const isNowPlayingLoading = ref(false)

  const upcomingMovies = ref<Movie[]>([])
  const upcomingPage = ref(1)
  const upcomingHasMore = ref(true)
  const isUpcomingLoading = ref(false)

  function parseLocalDate(dateStr: string): Date {
    if (!dateStr) return new Date(NaN)
    const [year, month, day] = dateStr.split("-").map(Number)
    return new Date(year!, month! - 1, day!)
  }

  function isReleased(dateStr: string): boolean {
    if (!dateStr) return true
    const release = parseLocalDate(dateStr)
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    return release.getTime() <= today.getTime()
  }

  function getDaysLeft(dateStr: string): number {
    if (!dateStr) return 0
    const release = parseLocalDate(dateStr)
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const diff = release.getTime() - today.getTime()
    return Math.max(0, Math.ceil(diff / 86400000))
  }

  function formatNumber(num: number): string {
    return num < 10 ? `0${num}` : `${num}`
  }

  function startCountdown(dateStr: string) {
    if (countdownTimer) clearInterval(countdownTimer)

    const targetDate = parseLocalDate(dateStr)
    targetDate.setHours(0, 0, 0, 0)

    const updateClock = () => {
      const now = new Date().getTime()
      const diff = targetDate.getTime() - now

      if (diff <= 0) {
        countdown.value = { days: 0, hours: 0, minutes: 0, seconds: 0 }
        if (countdownTimer) clearInterval(countdownTimer)
        return
      }

      countdown.value = {
        days: Math.floor(diff / (1000 * 60 * 60 * 24)),
        hours: Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
        minutes: Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)),
        seconds: Math.floor((diff % (1000 * 60)) / 1000),
      }
    }

    updateClock()
    countdownTimer = window.setInterval(updateClock, 1000)
  }

  watch(
    () => featuredMovie.value?.release_date,
    newDate => {
      if (newDate && !isReleased(newDate)) {
        startCountdown(newDate)
      } else {
        if (countdownTimer) clearInterval(countdownTimer)
      }
    },
  )

  function formatDate(date: string): string {
    if (!date) return ""
    return parseLocalDate(date).toLocaleDateString("th-TH", {
      day: "numeric",
      month: "long",
      year: "numeric",
    })
  }

  function formatDateShort(date: string): string {
    if (!date) return ""
    return parseLocalDate(date).toLocaleDateString("th-TH", {
      day: "numeric",
      month: "short",
    })
  }

  function goToDetail(id: number) {
    router.push({ name: "movie-detail", params: { id } })
  }

  function setFeatured(movie: Movie) {
    if (movie.id === featuredMovie.value?.id) return
    const hero = heroRef.value
    gsap.to(hero, {
      opacity: 0,
      y: 10,
      duration: 0.2,
      ease: "power2.in",
      onComplete: () => {
        featuredMovie.value = movie
        gsap.fromTo(
          hero,
          { opacity: 0, y: -10 },
          { opacity: 1, y: 0, duration: 0.35, ease: "power3.out" },
        )
      },
    })
  }

  async function fetchNowPlaying(page = 1) {
    isNowPlayingLoading.value = true
    try {
      const res = await movieApi.getNowPlaying(page)
      const results = res.data.results as Movie[]

      if (page === 1) {
        nowPlayingMovies.value = results
        if (!featuredMovie.value && results.length)
          featuredMovie.value = results[0]!
      } else {
        nowPlayingMovies.value.push(...results)
      }
      nowPlayingHasMore.value = res.data.page < res.data.total_pages
      nowPlayingPage.value = page
    } catch (err) {
      console.error("Error fetching now playing:", err)
    } finally {
      isNowPlayingLoading.value = false
    }
  }

  async function fetchUpcoming(page = 1) {
    isUpcomingLoading.value = true
    try {
      const res = await movieApi.getUpcoming(page)
      const results = res.data.results as Movie[]

      if (page === 1) {
        upcomingMovies.value = results
      } else {
        upcomingMovies.value.push(...results)
      }
      upcomingHasMore.value = res.data.page < res.data.total_pages
      upcomingPage.value = page
    } catch (err) {
      console.error("Error fetching upcoming:", err)
    } finally {
      isUpcomingLoading.value = false
    }
  }

  async function loadMoreNowPlaying() {
    if (isNowPlayingLoading.value || !nowPlayingHasMore.value) return
    await fetchNowPlaying(nowPlayingPage.value + 1)
    triggerCardAnimation(nowPlayingRef.value)
  }

  async function loadMoreUpcoming() {
    if (isUpcomingLoading.value || !upcomingHasMore.value) return
    await fetchUpcoming(upcomingPage.value + 1)
    triggerCardAnimation(upcomingRef.value)
  }

  function triggerCardAnimation(container: HTMLElement | null) {
    if (!container) return
    setTimeout(() => {
      const newCards = container.querySelectorAll(
        ".movie-card:not(.gsap-revealed)",
      )
      if (newCards.length) {
        gsap.fromTo(
          newCards,
          { opacity: 0, y: 24 },
          {
            opacity: 1,
            y: 0,
            duration: 0.4,
            stagger: 0.04,
            ease: "power3.out",
            onComplete: () =>
              newCards.forEach(c => c.classList.add("gsap-revealed")),
          },
        )
      }
    }, 50)
  }

  function setupAnimations() {
    const tl = gsap.timeline({ defaults: { ease: "power3.out" } })
    tl.fromTo(
      introRef.value,
      { opacity: 0, y: 20 },
      { opacity: 1, y: 0, duration: 0.55 },
    )
    tl.fromTo(
      heroRef.value,
      { opacity: 0, y: 36, scale: 0.97 },
      { opacity: 1, y: 0, scale: 1, duration: 0.7 },
      "-=0.3",
    )

    ScrollTrigger.batch(".movie-card", {
      onEnter: batch =>
        gsap.fromTo(
          batch,
          { opacity: 0, scale: 0.94 },
          {
            opacity: 1,
            scale: 1,
            duration: 0.35,
            stagger: 0.04,
            ease: "power2.out",
            onComplete: () =>
              batch.forEach((el: Element) => el.classList.add("gsap-revealed")),
          },
        ),
      start: "top 94%",
      once: true,
    })
  }

  onMounted(async () => {
    gsap.set([introRef.value, heroRef.value], { opacity: 0 })
    await Promise.all([fetchNowPlaying(1), fetchUpcoming(1)])
    setupAnimations()
  })

  onUnmounted(() => {
    ScrollTrigger.getAll().forEach(t => t.kill())
    if (countdownTimer) clearInterval(countdownTimer)
  })
</script>

<style scoped>
  @import url("https://fonts.googleapis.com/css2?family=Noto+Sans+Thai:wght@300;400;500;600;700&display=swap");

  *,
  *::before,
  *::after {
    box-sizing: border-box;
  }

  .upcoming-page {
    --red: var(--color-brand);
    --red-dim: rgba(245, 67, 44, 0.15);
    --gold: #f5c518;
    --green: #22c55e;
    --bg: #080808;
    --surface: #101010;
    --surface2: #181818;
    --surface3: #202020;
    --border: rgba(255, 255, 255, 0.07);
    --border-md: rgba(255, 255, 255, 0.12);
    --text: #f0f0f0;
    --muted: #5a5a5a;
    --muted2: #8a8a8a;

    font-family: "Noto Sans Thai", sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    position: relative;
    overflow-x: hidden;
  }

  /* ── Ambient ── */
  .ambient-bg {
    position: fixed;
    inset: 0;
    pointer-events: none;
    z-index: 0;
  }
  .ambient-blob {
    position: absolute;
    inset: 0;
    background-size: cover;
    background-position: center top;
    filter: blur(100px) brightness(0.1) saturate(1.8);
    transform: scale(1.15);
    transition: background-image 1s ease;
  }
  .ambient-overlay {
    position: absolute;
    inset: 0;
    background: radial-gradient(
      ellipse at 50% 0%,
      rgba(8, 8, 8, 0.3) 0%,
      #080808 65%
    );
  }

  .page-main {
    position: relative;
    z-index: 1;
    max-width: 1120px;
    margin: 0 auto;
    padding: 3rem 1.75rem 6rem;
    display: flex;
    flex-direction: column;
  }

  /* ── Header ── */
  .page-header {
    padding-bottom: 2rem;
    border-bottom: 1px solid var(--border);
    margin-bottom: 2.5rem;
  }
  .page-header__eyebrow {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin-bottom: 0.9rem;
  }
  .eyebrow-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--red);
    box-shadow: 0 0 8px var(--red);
  }
  .eyebrow-text {
    font-size: 0.68rem;
    letter-spacing: 0.45em;
    color: var(--red);
    font-weight: 600;
  }
  .page-header__title {
    font-size: clamp(2rem, 4.5vw, 3.2rem);
    font-weight: 700;
    color: #fff;
    margin: 0;
  }
  .page-header__sub {
    color: var(--muted2);
    font-size: 0.875rem;
    margin: 0.3rem 0 0;
    font-weight: 300;
  }

  /* ── Category Dividers ── */
  .category-divider {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.5rem;
    border-left: 4px solid var(--red);
    padding-left: 10px;
  }
  .category-title {
    font-size: 1.25rem;
    font-weight: 600;
    color: #fff;
    margin: 0;
  }
  .category-count {
    font-size: 0.75rem;
    color: var(--muted2);
    background: var(--surface2);
    padding: 0.2rem 0.6rem;
    border-radius: 9999px;
    border: 1px solid var(--border);
  }

  /* ── Hero ── */
  .hero {
    position: relative;
    border-radius: 16px;
    overflow: hidden;
    border: 1px solid var(--border);
    background: var(--surface);
    min-height: 340px;
    margin-bottom: 3.5rem;
  }
  .hero__backdrop {
    position: absolute;
    inset: 0;
  }
  .hero__backdrop-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center 25%;
  }
  .hero__backdrop-grad {
    position: absolute;
    inset: 0;
    background:
      linear-gradient(
        to right,
        rgba(8, 8, 8, 0.98) 35%,
        rgba(8, 8, 8, 0.6) 65%,
        rgba(8, 8, 8, 0.1) 100%
      ),
      linear-gradient(to top, rgba(8, 8, 8, 0.85) 0%, transparent 45%);
  }
  .hero__inner {
    position: relative;
    z-index: 1;
    display: flex;
    gap: 2.5rem;
    padding: 2.5rem;
    align-items: flex-start;
  }
  .hero__poster-wrap {
    position: relative;
    width: 155px;
  }
  .hero__poster {
    width: 155px;
    border-radius: 10px;
    display: block;
    box-shadow: 0 24px 64px rgba(0, 0, 0, 0.85);
  }
  .hero__poster-glow {
    position: absolute;
    inset: -6px;
    border-radius: 14px;
    background: linear-gradient(135deg, var(--red), transparent 55%);
    opacity: 0.35;
    z-index: -1;
    filter: blur(10px);
  }
  .hero__info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
  }
  .hero__badge {
    display: inline-flex;
    align-items: center;
    background: var(--red-dim);
    border: 1px solid rgba(245, 67, 44, 0.28);
    color: #ff7070;
    font-size: 0.68rem;
    font-weight: 700;
    padding: 0.2rem 0.7rem;
    border-radius: 9999px;
    width: fit-content;
  }
  .hero__title {
    font-size: clamp(1.8rem, 3.5vw, 2.8rem);
    font-weight: 700;
    color: #fff;
    margin: 0;
  }
  .hero__overview {
    color: var(--muted2);
    font-size: 0.85rem;
    line-height: 1.75;
    margin: 0;
    max-width: 500px;
  }
  .hero__meta-row {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  .meta-tag {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border);
    color: var(--text);
    font-size: 0.75rem;
    padding: 0.25rem 0.75rem;
    border-radius: 6px;
  }
  .meta-tag--green {
    background: rgba(34, 197, 94, 0.12);
    border-color: rgba(34, 197, 94, 0.3);
    color: var(--green);
    font-weight: 500;
  }

  /* ── Countdown Blocks ── */
  .countdown-wrapper {
    background: rgba(0, 0, 0, 0.4);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 0.85rem 1.2rem;
    max-width: 320px;
    backdrop-filter: blur(8px);
    margin: 0.2rem 0;
  }
  .countdown-title {
    display: flex;
    align-items: center;
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--gold);
    font-weight: 600;
    margin-bottom: 0.5rem;
  }
  .countdown-clock {
    display: flex;
    gap: 0.6rem;
  }
  .countdown-box {
    background: var(--surface2);
    border: 1px solid rgba(255, 255, 255, 0.05);
    min-width: 58px;
    padding: 0.4rem 0.2rem;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .countdown-num {
    font-size: 1.35rem;
    font-weight: 700;
    color: #fff;
    font-family: monospace;
    line-height: 1.1;
  }
  .countdown-num--sec {
    color: var(--gold);
  }
  .countdown-label {
    font-size: 0.6rem;
    color: var(--muted2);
    margin-top: 0.15rem;
    font-weight: 500;
  }
  .released-banner {
    margin: 0.2rem 0;
  }

  .btn-detail {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    background: var(--red);
    color: #fff;
    border: none;
    border-radius: 9999px;
    padding: 0.6rem 1.35rem;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    width: fit-content;
    transition:
      background 0.2s,
      transform 0.15s;
    margin-top: 0.25rem;
  }
  .btn-detail:hover {
    background: var(--color-brand-hover);
    transform: translateY(-1px);
  }

  /* ── Movie Grid ── */
  .movie-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
    gap: 14px;
  }
  .movie-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 10px;
    overflow: hidden;
    cursor: pointer;
    padding: 0;
    text-align: left;
    position: relative;
    transition:
      border-color 0.2s,
      box-shadow 0.2s,
      transform 0.2s;
  }
  .movie-card:hover {
    border-color: var(--border-md);
    transform: translateY(-4px);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
  }
  .movie-card--active {
    border-color: var(--red) !important;
    box-shadow:
      0 0 0 1px var(--red),
      0 10px 28px rgba(245, 67, 44, 0.18) !important;
  }
  .movie-card__poster-wrap {
    position: relative;
    width: 100%;
    aspect-ratio: 2/3;
    overflow: hidden;
  }
  .movie-card__poster {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.3s;
  }
  .movie-card:hover .movie-card__poster {
    transform: scale(1.06);
  }
  .movie-card__poster--empty {
    background: var(--surface2);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--muted);
    height: 100%;
  }
  .movie-card__overlay {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.48);
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.2s;
  }
  .movie-card:hover .movie-card__overlay {
    opacity: 1;
  }

  .movie-card__badge {
    position: absolute;
    bottom: 6px;
    left: 6px;
    font-size: 0.58rem;
    font-weight: 700;
    padding: 0.18rem 0.45rem;
    border-radius: 4px;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(4px);
  }
  .movie-card__badge--soon {
    color: var(--gold);
  }
  .movie-card__badge--out {
    color: var(--green);
  }

  .movie-card__info {
    padding: 0.55rem 0.6rem 0.65rem;
    display: flex;
    flex-direction: column;
    gap: 0.18rem;
  }
  .movie-card__title {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }
  .movie-card__date {
    font-size: 0.65rem;
    color: var(--muted2);
  }

  /* ── Load more ── */
  .load-more-section {
    display: flex;
    justify-content: center;
    padding-top: 1.5rem;
  }
  .btn-load-more {
    background: transparent;
    border: 1px solid var(--border-md);
    color: var(--muted2);
    font-size: 0.83rem;
    padding: 0.65rem 2.2rem;
    border-radius: 9999px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 160px;
    justify-content: center;
    transition:
      border-color 0.2s,
      color 0.2s,
      background 0.2s;
  }
  .btn-load-more:hover:not(:disabled) {
    border-color: rgba(255, 255, 255, 0.2);
    color: #fff;
    background: rgba(255, 255, 255, 0.04);
  }
  .btn-load-more:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  @media (max-width: 720px) {
    .hero__inner {
      flex-direction: column;
      gap: 1.4rem;
      padding: 1.5rem;
    }
    .hero__poster-wrap,
    .hero__poster {
      width: 120px;
    }
    .movie-grid {
      grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
      gap: 10px;
    }
    .countdown-wrapper {
      max-width: 100%;
    }
  }
</style>
