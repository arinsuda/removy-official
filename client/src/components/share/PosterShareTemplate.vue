<template>
  <div class="poster-share-card" ref="cardRef">
    <!-- Background poster (blurred, dimmed) -->
    <div class="bg-layer">
      <img
        v-if="displayPosterUrl"
        :src="displayPosterUrl"
        :alt="media.title"
        class="bg-poster"
        crossorigin="anonymous"
      />
      <div class="bg-gradient" />
    </div>

    <!-- Main content -->
    <div class="card-content">
      <!-- Poster image -->
      <div class="poster-frame">
        <img
          v-if="displayPosterUrl"
          :src="displayPosterUrl"
          :alt="media.title"
          class="poster-img"
          crossorigin="anonymous"
        />
        <div v-else class="poster-fallback">
          <i class="pi pi-image" />
        </div>
      </div>

      <!-- Media type badge -->
      <div class="type-badge">
        {{ media.mediaType === 'movie' ? 'MOVIE' : 'TV SERIES' }}
      </div>

      <!-- Title -->
      <h1
        class="media-title"
        :class="{
          'title-md': media.title.length > 30 && media.title.length <= 50,
          'title-sm': media.title.length > 50,
        }"
      >
        {{ media.title }}
      </h1>

      <!-- Year -->
      <p v-if="media.releaseYear" class="media-year">
        {{ media.releaseYear }}
      </p>

      <!-- Genres -->
      <div v-if="displayGenres.length" class="genre-row">
        <span v-for="genre in displayGenres" :key="genre.id" class="genre-chip">
          {{ genre.name }}
        </span>
      </div>

      <!-- Rating -->
      <div v-if="ratingDisplay" class="rating-section">
        <div class="rating-badge" :class="ratingDisplay.badgeClass">
          <i class="pi pi-star-fill rating-star" />
          <span class="rating-value">{{ ratingDisplay.value }}</span>
          <span class="rating-scale">{{ ratingDisplay.scale }}</span>
          <span class="rating-source">{{ ratingDisplay.source }}</span>
        </div>
      </div>
    </div>

    <!-- Branding footer -->
    <div class="brand-footer">
      <div class="brand-divider" />
      <div class="brand-logo">
        <span class="logo-re">POP</span><span class="logo-mov">CORN</span>
      </div>
      <span class="brand-handle">POPCORN</span>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from "vue"
  import type { ShareMediaContext } from "@/types/share"
  import { getTmdbImageUrl } from "@/utils/image"

  const props = defineProps<{
    media: ShareMediaContext
    posterBlobUrl: string | null
  }>()

  const displayPosterUrl = computed(() => {
    if (props.posterBlobUrl) return props.posterBlobUrl
    if (props.media.posterUrl) return props.media.posterUrl
    if (props.media.posterPath) return getTmdbImageUrl(props.media.posterPath, "w500")
    return null
  })

  const ENGLISH_GENRES: Record<number, string> = {
    28: "Action",
    12: "Adventure",
    16: "Animation",
    35: "Comedy",
    80: "Crime",
    99: "Documentary",
    18: "Drama",
    10751: "Family",
    14: "Fantasy",
    36: "History",
    27: "Horror",
    10402: "Music",
    9648: "Mystery",
    10749: "Romance",
    878: "Sci-Fi",
    10770: "TV Movie",
    53: "Thriller",
    10752: "War",
    37: "Western",
    10759: "Action & Adventure",
    10762: "Kids",
    10763: "News",
    10764: "Reality",
    10765: "Sci-Fi & Fantasy",
    10766: "Soap",
    10767: "Talk",
    10768: "War & Politics",
  }

  const displayGenres = computed(() =>
    props.media.genres.slice(0, 3).map((g) => ({
      id: g.id,
      name: ENGLISH_GENRES[g.id] || g.name,
    })),
  )

  const ratingDisplay = computed(() => {
    if (props.media.popcornRating !== null && props.media.popcornRating > 0) {
      return {
        value: props.media.popcornRating.toFixed(1),
        scale: "/5",
        source: "POPCORN",
        badgeClass: "badge-popcorn",
      }
    }
    if (props.media.voteAverage > 0) {
      return {
        value: props.media.voteAverage.toFixed(1),
        scale: "/10",
        source: "TMDB",
        badgeClass: "badge-tmdb",
      }
    }
    return null
  })
</script>

<style scoped>
  .poster-share-card {
    width: 1080px;
    height: 1920px;
    position: relative;
    overflow: hidden;
    background: #0a0a0a;
    font-family: "Noto Sans Thai", "Inter", system-ui, sans-serif;
    -webkit-font-smoothing: antialiased;
    text-rendering: optimizeLegibility;
    display: flex;
    flex-direction: column;
  }

  /* ── Background ── */
  .bg-layer {
    position: absolute;
    inset: 0;
    z-index: 0;
  }

  .bg-poster {
    width: 100%;
    height: 100%;
    object-fit: cover;
    opacity: 0.25;
    filter: blur(20px) saturate(1.2);
  }

  .bg-gradient {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      180deg,
      rgba(10, 10, 10, 0.3) 0%,
      rgba(10, 10, 10, 0.7) 40%,
      rgba(10, 10, 10, 0.95) 65%,
      #0a0a0a 100%
    );
  }

  /* ── Content ── */
  .card-content {
    position: relative;
    z-index: 1;
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 60px 40px;
    gap: 28px;
  }

  /* ── Poster ── */
  .poster-frame {
    width: 420px;
    height: 630px;
    border-radius: 16px;
    overflow: hidden;
    box-shadow:
      0 20px 60px rgba(0, 0, 0, 0.6),
      0 0 0 1px rgba(255, 255, 255, 0.08);
    flex-shrink: 0;
  }

  .poster-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .poster-fallback {
    width: 100%;
    height: 100%;
    background: #1a1a2e;
    display: flex;
    align-items: center;
    justify-content: center;
    color: rgba(255, 255, 255, 0.2);
    font-size: 80px;
  }

  /* ── Type Badge ── */
  .type-badge {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.8);
    font-size: 20px;
    font-weight: 600;
    padding: 6px 20px;
    border-radius: 20px;
    text-transform: uppercase;
    letter-spacing: 2px;
  }

  /* ── Title ── */
  .media-title {
    color: #ffffff;
    font-size: 48px;
    font-weight: 800;
    text-align: center;
    line-height: 1.2;
    margin: 0;
    max-width: 900px;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
  }

  .media-title.title-md {
    font-size: 36px;
  }

  .media-title.title-sm {
    font-size: 28px;
  }

  /* ── Year ── */
  .media-year {
    color: rgba(255, 255, 255, 0.6);
    font-size: 24px;
    font-weight: 500;
    margin: 0;
  }

  /* ── Genres ── */
  .genre-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    justify-content: center;
  }

  .genre-chip {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: #ffffff;
    font-size: 20px;
    font-weight: 500;
    padding: 6px 18px;
    border-radius: 20px;
  }

  /* ── Rating ── */
  .rating-section {
    margin-top: 8px;
  }

  .rating-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 24px;
    border-radius: 30px;
    font-weight: 700;
  }

  .badge-popcorn {
    background: linear-gradient(135deg, #ff512f, #dd2476);
    color: #ffffff;
  }

  .badge-tmdb {
    background: #f5c518;
    color: #000000;
  }

  .rating-star {
    font-size: 22px;
  }

  .rating-value {
    font-size: 28px;
  }

  .rating-scale {
    font-size: 20px;
    opacity: 0.8;
  }

  .rating-source {
    font-size: 18px;
    opacity: 0.7;
    margin-left: 4px;
  }

  /* ── Branding ── */
  .brand-footer {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding-bottom: 60px;
  }

  .brand-divider {
    width: 60px;
    height: 2px;
    background: rgba(255, 255, 255, 0.2);
  }

  .brand-logo {
    font-size: 36px;
    font-weight: 800;
    letter-spacing: 4px;
  }

  .logo-re {
    color: #ffffff;
  }

  .logo-mov {
    color: #f5432c;
  }

  .brand-handle {
    color: rgba(255, 255, 255, 0.4);
    font-size: 18px;
    font-weight: 500;
  }
</style>
