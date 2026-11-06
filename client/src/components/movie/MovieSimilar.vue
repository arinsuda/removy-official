<template>
  <div class="similar-block">
    <!-- Loading Skeleton -->
    <div v-if="isLoading" class="similar-track">
      <div
        v-for="n in 6"
        :key="`skeleton-${n}`"
        class="similar-card similar-card--skeleton"
      >
        <div class="skeleton-poster"></div>
        <div class="skeleton-line skeleton-line--title"></div>
        <div class="skeleton-line skeleton-line--year"></div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else-if="hasError" class="similar-empty">
      <i class="pi pi-exclamation-circle"></i>
      <p>ไม่สามารถโหลด{{ mediaLabel }}ที่คล้ายกันได้ในขณะนี้</p>
      <button type="button" class="similar-retry-btn" @click="fetchSimilar">
        <i class="pi pi-refresh"></i>
        <span>ลองอีกครั้ง</span>
      </button>
    </div>

    <!-- Empty State -->
    <div v-else-if="!similarItems.length" class="similar-empty">
      <i class="pi pi-inbox"></i>
      <p>ยังไม่พบ{{ mediaLabel }}ที่คล้ายกับเรื่องนี้</p>
    </div>

    <!-- Loaded -->
    <div v-else class="similar-carousel">
      <button
        v-show="canScrollLeft"
        type="button"
        class="similar-nav similar-nav--left"
        :aria-label="`เลื่อนดู${mediaLabel}ก่อนหน้า`"
        @click="scrollByCards(-1)"
      >
        <i class="pi pi-chevron-left"></i>
      </button>

      <div class="similar-track" ref="trackRef" @scroll="updateNavState">
        <button
          v-for="item in similarItems"
          :key="`${props.movieType}-${item.id}`"
          type="button"
          class="similar-card"
          @click="goToMedia(item.id)"
        >
          <div class="similar-poster-wrap">
            <img
              v-if="item.poster_path"
              :src="`https://image.tmdb.org/t/p/w342${item.poster_path}`"
              :alt="getMediaTitle(item)"
              class="similar-poster"
              loading="lazy"
            />
            <div v-else class="similar-poster similar-poster--empty">
              <i class="pi pi-image"></i>
            </div>

            <span class="similar-rating" v-if="hasRating(item)">
              <i class="pi pi-star-fill"></i>
              {{ formatRating(item.vote_average) }}
            </span>
          </div>

          <div class="similar-meta">
            <span class="similar-title">{{ getMediaTitle(item) }}</span>
            <span class="similar-year">
              {{ formatYear(getMediaDate(item)) }}
            </span>
          </div>
        </button>
      </div>

      <button
        v-show="canScrollRight"
        type="button"
        class="similar-nav similar-nav--right"
        :aria-label="`เลื่อนดู${mediaLabel}ถัดไป`"
        @click="scrollByCards(1)"
      >
        <i class="pi pi-chevron-right"></i>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, onMounted, onUnmounted, watch, nextTick } from "vue"
  import { useRouter } from "vue-router"
  import { gsap } from "gsap"
  import { movieApi } from "@/api/api"

  type MediaType = "movie" | "tv"

  type SimilarMediaItem = {
    id: number
    title?: string
    name?: string
    poster_path?: string | null
    release_date?: string | null
    first_air_date?: string | null
    vote_average?: number | null
  }

  const props = defineProps<{
    movieId: number
    movieType: MediaType
  }>()

  const router = useRouter()

  const similarItems = ref<SimilarMediaItem[]>([])
  const isLoading = ref(true)
  const hasError = ref(false)

  const trackRef = ref<HTMLElement | null>(null)
  const canScrollLeft = ref(false)
  const canScrollRight = ref(false)

  const mediaLabel = computed(() =>
    props.movieType === "tv" ? "ซีรีส์" : "ภาพยนตร์",
  )

  function getMediaTitle(item: SimilarMediaItem): string {
    return item.title || item.name || "ไม่มีชื่อ"
  }

  function getMediaDate(item: SimilarMediaItem): string | null | undefined {
    return item.release_date || item.first_air_date
  }

  function formatYear(dateStr?: string | null): string {
    if (!dateStr) return "-"
    return dateStr.split("-")[0] || "-"
  }

  function hasRating(item: SimilarMediaItem): boolean {
    return typeof item.vote_average === "number" && item.vote_average > 0
  }

  function formatRating(value?: number | null): string {
    if (typeof value !== "number") return "-"
    return value.toFixed(1)
  }

  function goToMedia(id: number) {
    if (props.movieType === "tv") {
      router.push({ name: "tv-detail", params: { id } })
      return
    }

    router.push({ name: "movie-detail", params: { id } })
  }

  function updateNavState() {
    const el = trackRef.value
    if (!el) return

    canScrollLeft.value = el.scrollLeft > 8
    canScrollRight.value = el.scrollLeft + el.clientWidth < el.scrollWidth - 8
  }

  function scrollByCards(direction: 1 | -1) {
    const el = trackRef.value
    if (!el) return

    const card = el.querySelector<HTMLElement>(".similar-card")
    const cardWidth = card?.offsetWidth ?? 160
    const gap = 16
    const distance = (cardWidth + gap) * 3 * direction

    el.scrollBy({ left: distance, behavior: "smooth" })
  }

  function animateCards() {
    const el = trackRef.value
    if (!el) return

    const cards = el.querySelectorAll(".similar-card")
    if (!cards.length) return

    gsap.fromTo(
      cards,
      { opacity: 0, y: 16 },
      { opacity: 1, y: 0, duration: 0.4, stagger: 0.05, ease: "power2.out" },
    )
  }

  async function fetchSimilar() {
    if (!props.movieId) {
      similarItems.value = []
      isLoading.value = false
      return
    }

    isLoading.value = true
    hasError.value = false

    try {
      const res =
        props.movieType === "tv"
          ? await movieApi.getSimilarSeries(props.movieId)
          : await movieApi.getSimilar(props.movieId)

      const results = Array.isArray(res.data?.results) ? res.data.results : []

      similarItems.value = results.filter(
        (item: SimilarMediaItem) => item.id && item.id !== props.movieId,
      )
    } catch (err) {
      console.error(`ไม่สามารถดึงข้อมูล${mediaLabel.value}ที่คล้ายกันได้:`, err)
      hasError.value = true
      similarItems.value = []
    } finally {
      isLoading.value = false

      await nextTick()

      if (trackRef.value) {
        trackRef.value.scrollLeft = 0
      }

      updateNavState()

      if (!hasError.value && similarItems.value.length) {
        animateCards()
      }
    }
  }

  // รองรับกรณี component ถูกใช้ซ้ำข้ามหน้า detail
  // เช่น เปลี่ยนจาก movie -> tv หรือเปลี่ยนแค่ id
  watch(
    () => [props.movieId, props.movieType] as const,
    () => {
      fetchSimilar()
    },
  )

  onMounted(() => {
    fetchSimilar()
    window.addEventListener("resize", updateNavState)
  })

  onUnmounted(() => {
    window.removeEventListener("resize", updateNavState)
  })
</script>

<style scoped>
  .similar-block {
    --sim-red: var(--red, var(--color-brand));
    --sim-gold: var(--gold, #f5c518);
    --sim-surface: var(--surface, #121212);
    --sim-surface2: var(--surface2, #1c1c1c);
    --sim-border: var(--border, rgba(255, 255, 255, 0.08));
    --sim-text: var(--text, #f0f0f0);
    --sim-muted: var(--muted, #8a8a8a);

    position: relative;
    width: 100%;
  }

  .similar-carousel {
    position: relative;
  }

  .similar-track {
    display: flex;
    gap: 1rem;
    overflow-x: auto;
    scroll-behavior: smooth;
    scroll-snap-type: x proximity;
    padding: 0.25rem 0.1rem 1rem;
    scrollbar-width: thin;
    scrollbar-color: var(--sim-surface2) transparent;
    -webkit-overflow-scrolling: touch;
  }
  .similar-track::-webkit-scrollbar {
    height: 6px;
  }
  .similar-track::-webkit-scrollbar-thumb {
    background: var(--sim-surface2);
    border-radius: 999px;
  }

  .similar-nav {
    position: absolute;
    top: 40%;
    transform: translateY(-50%);
    z-index: 3;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 1px solid var(--sim-border);
    background: rgba(8, 8, 8, 0.75);
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    backdrop-filter: blur(6px);
    transition:
      background 0.2s,
      border-color 0.2s,
      transform 0.2s;
  }
  .similar-nav:hover {
    background: var(--sim-red);
    border-color: var(--sim-red);
    transform: translateY(-50%) scale(1.05);
  }
  .similar-nav--left {
    left: -8px;
  }
  .similar-nav--right {
    right: -8px;
  }

  .similar-card {
    all: unset;
    flex-shrink: 0;
    width: 160px;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    cursor: pointer;
    scroll-snap-align: start;
    box-sizing: border-box;
  }

  .similar-poster-wrap {
    position: relative;
    width: 100%;
    aspect-ratio: 2/3;
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid var(--sim-border);
    background: var(--sim-surface);
    transition:
      transform 0.25s ease,
      border-color 0.25s ease,
      box-shadow 0.25s ease;
  }
  .similar-card:hover .similar-poster-wrap {
    transform: translateY(-4px);
    border-color: rgba(255, 255, 255, 0.2);
    box-shadow: 0 14px 28px rgba(0, 0, 0, 0.55);
  }
  .similar-card:focus-visible .similar-poster-wrap {
    outline: 2px solid var(--sim-red);
    outline-offset: 2px;
  }

  .similar-poster {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .similar-poster--empty {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--sim-muted);
    font-size: 1.75rem;
  }

  .similar-rating {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    background: rgba(0, 0, 0, 0.7);
    border: 1px solid rgba(245, 197, 24, 0.35);
    color: var(--sim-gold);
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.2rem 0.5rem;
    border-radius: 999px;
    backdrop-filter: blur(4px);
  }

  .similar-meta {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;
    text-align: left;
  }
  .similar-title {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--sim-text);
    line-height: 1.35;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  .similar-card:hover .similar-title {
    color: var(--sim-red);
  }
  .similar-year {
    font-size: 0.72rem;
    color: var(--sim-muted);
  }

  .similar-card--skeleton {
    cursor: default;
  }
  .skeleton-poster {
    width: 100%;
    aspect-ratio: 2/3;
    border-radius: 10px;
    background: linear-gradient(
      100deg,
      var(--sim-surface) 30%,
      var(--sim-surface2) 50%,
      var(--sim-surface) 70%
    );
    background-size: 200% 100%;
    animation: similar-shimmer 1.4s ease-in-out infinite;
  }
  .skeleton-line {
    height: 0.7rem;
    border-radius: 4px;
    background: var(--sim-surface);
    margin-top: 0.5rem;
  }
  .skeleton-line--title {
    width: 85%;
  }
  .skeleton-line--year {
    width: 40%;
  }
  @keyframes similar-shimmer {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }

  .similar-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.65rem;
    padding: 2.5rem 1rem;
    color: var(--sim-muted);
    text-align: center;
    background: rgba(255, 255, 255, 0.02);
    border: 1px dashed var(--sim-border);
    border-radius: 12px;
  }
  .similar-empty i {
    font-size: 1.5rem;
  }
  .similar-empty p {
    margin: 0;
    font-size: 0.85rem;
  }
  .similar-retry-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--sim-border);
    color: var(--sim-text);
    font-size: 0.8rem;
    padding: 0.45rem 0.9rem;
    border-radius: 999px;
    cursor: pointer;
    transition:
      background 0.2s,
      border-color 0.2s;
  }
  .similar-retry-btn:hover {
    background: var(--sim-surface2);
    border-color: rgba(255, 255, 255, 0.2);
  }

  @media (max-width: 1024px) {
    .similar-card {
      width: 145px;
    }
  }
  @media (max-width: 820px) {
    .similar-card {
      width: 130px;
    }
    .similar-nav {
      width: 34px;
      height: 34px;
    }
  }
  @media (max-width: 640px) {
    .similar-card {
      width: 115px;
    }
    .similar-track {
      gap: 0.75rem;
    }
    .similar-nav {
      display: none;
    }
    .similar-title {
      font-size: 0.78rem;
    }
  }
  @media (max-width: 480px) {
    .similar-card {
      width: 100px;
    }
  }
  @media (max-width: 360px) {
    .similar-card {
      width: 88px;
    }
    .similar-rating {
      font-size: 0.62rem;
      padding: 0.15rem 0.4rem;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .skeleton-poster {
      animation: none;
    }
    .similar-card,
    .similar-poster-wrap,
    .similar-nav {
      transition: none;
    }
  }
</style>
