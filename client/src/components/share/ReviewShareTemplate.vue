<template>
  <div class="review-share-card">
    <!-- Background -->
    <div class="bg-layer">
      <img
        v-if="displayPosterUrl"
        :src="displayPosterUrl"
        :alt="media.title"
        class="bg-poster"
      />
      <div class="bg-gradient" />
    </div>

    <!-- Content -->
    <div class="card-content">
      <!-- Top section: poster + media info -->
      <div class="media-section">
        <div class="poster-frame">
          <img
            v-if="displayPosterUrl"
            :src="displayPosterUrl"
            :alt="media.title"
            class="poster-img"
          />
          <div v-else class="poster-fallback">
            <i class="pi pi-image" />
          </div>
        </div>
        <div class="media-info">
          <div class="type-badge">
            {{ media.mediaType === 'movie' ? 'MOVIE' : 'TV SERIES' }}
          </div>
          <h2
            class="media-title"
            :class="{ 'title-sm': media.title.length > 30 }"
          >
            {{ media.title }}
          </h2>
          <p v-if="media.releaseYear" class="media-year">
            {{ media.releaseYear }}
          </p>
        </div>
      </div>

      <!-- Divider -->
      <div class="section-divider" />

      <!-- Review section -->
      <div class="review-section">
        <!-- Author info -->
        <div class="author-row">
          <div class="author-avatar">
            <img
              v-if="displayAvatarUrl"
              :src="displayAvatarUrl"
              :alt="review.authorDisplayName"
              class="avatar-img"
            />
            <div v-else class="avatar-fallback">
              <i class="pi pi-user" />
            </div>
          </div>
          <div class="author-info">
            <span class="author-name">
              {{ review.authorDisplayName || review.authorUsername }}
            </span>
            <span class="author-username">@{{ review.authorUsername }}</span>
          </div>
          <!-- Rating -->
          <div class="review-rating">
            <div class="stars-row">
              <i
                v-for="star in 5"
                :key="star"
                class="pi"
                :class="star <= Math.round(review.rating) ? 'pi-star-fill star-filled' : 'pi-star star-empty'"
              />
            </div>
            <span class="rating-text">{{ review.rating.toFixed(1) }}/5</span>
          </div>
        </div>

        <!-- Review body -->
        <div class="review-body">
          <span class="quote-mark">&ldquo;</span>
          <p class="review-text">{{ truncatedBody }}</p>
          <span class="quote-mark quote-end">&rdquo;</span>
        </div>


      </div>
    </div>

    <!-- Branding footer -->
    <div class="brand-footer">
      <div class="brand-divider" />
      <span class="reviewed-label">REVIEWED ON</span>
      <div class="brand-logo">
        <span class="logo-re">POP</span><span class="logo-mov">CORN</span>
      </div>
      <span class="brand-handle">POPCORN</span>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from "vue"
  import type { ShareMediaContext, ShareReviewContext } from "@/types/share"
  import { useShareImage } from "@/composables/useShareImage"

  const props = defineProps<{
    media: ShareMediaContext
    review: ShareReviewContext
    posterBlobUrl: string | null
    avatarBlobUrl: string | null
  }>()

  const { truncateReview } = useShareImage()

  const truncatedBody = computed(() => truncateReview(props.review.body, 200))

  const displayPosterUrl = computed(() => {
    if (props.posterBlobUrl) return props.posterBlobUrl
    return null
  })

  const displayAvatarUrl = computed(() => {
    if (props.avatarBlobUrl) return props.avatarBlobUrl
    return null
  })
</script>

<style scoped>
  .review-share-card {
    width: 1080px;
    height: 1920px;
    position: relative;
    overflow: hidden;
    background: #0a0a0a;
    font-family: var(--font-sans);
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
    opacity: 0.15;
    filter: blur(30px) saturate(1.3);
  }

  .bg-gradient {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      180deg,
      rgba(10, 10, 10, 0.5) 0%,
      rgba(10, 10, 10, 0.85) 50%,
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
    padding: 80px 60px 40px;
    gap: 40px;
  }

  /* ── Media Section ── */
  .media-section {
    display: flex;
    gap: 36px;
    align-items: flex-start;
  }

  .poster-frame {
    width: 300px;
    height: 450px;
    border-radius: 14px;
    overflow: hidden;
    flex-shrink: 0;
    box-shadow:
      0 16px 48px rgba(0, 0, 0, 0.5),
      0 0 0 1px rgba(255, 255, 255, 0.08);
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
    font-size: 60px;
  }

  .media-info {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding-top: 20px;
    flex: 1;
    min-width: 0;
  }

  .type-badge {
    display: inline-block;
    align-self: flex-start;
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.8);
    font-size: 18px;
    font-weight: 600;
    padding: 4px 16px;
    border-radius: 16px;
    text-transform: uppercase;
    letter-spacing: 1.5px;
  }

  .media-title {
    color: #ffffff;
    font-size: 40px;
    font-weight: 800;
    line-height: 1.2;
    margin: 0;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
  }

  .media-title.title-sm {
    font-size: 32px;
  }

  .media-year {
    color: rgba(255, 255, 255, 0.5);
    font-size: 22px;
    margin: 0;
    font-weight: 500;
  }

  /* ── Divider ── */
  .section-divider {
    width: 100%;
    height: 1px;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.15),
      transparent
    );
  }

  /* ── Review Section ── */
  .review-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 32px;
  }

  .author-row {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .author-avatar {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    overflow: hidden;
    flex-shrink: 0;
    border: 2px solid rgba(255, 255, 255, 0.15);
  }

  .avatar-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .avatar-fallback {
    width: 100%;
    height: 100%;
    background: #1a1a2e;
    display: flex;
    align-items: center;
    justify-content: center;
    color: rgba(255, 255, 255, 0.3);
    font-size: 28px;
  }

  .author-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
    min-width: 0;
  }

  .author-name {
    color: #ffffff;
    font-size: 26px;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .author-username {
    color: rgba(255, 255, 255, 0.5);
    font-size: 20px;
    font-weight: 400;
  }

  .review-rating {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
  }

  .stars-row {
    display: flex;
    gap: 4px;
  }

  .star-filled {
    color: #f5c518;
    font-size: 24px;
  }

  .star-empty {
    color: rgba(255, 255, 255, 0.2);
    font-size: 24px;
  }

  .rating-text {
    color: rgba(255, 255, 255, 0.7);
    font-size: 20px;
    font-weight: 600;
  }

  /* ── Review Body ── */
  .review-body {
    position: relative;
    padding: 24px 32px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 16px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    overflow: hidden;
    max-height: 500px;
  }

  .quote-mark {
    color: rgba(245, 67, 44, 0.4);
    font-size: 64px;
    font-weight: 800;
    line-height: 1;
    position: absolute;
    top: 8px;
    left: 16px;
  }

  .quote-end {
    left: auto;
    right: 16px;
    top: auto;
    bottom: 0;
  }

  .review-text {
    color: rgba(255, 255, 255, 0.9);
    font-size: 26px;
    line-height: 1.6;
    margin: 24px 0 8px;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 8;
    -webkit-box-orient: vertical;
  }

  .watched-date {
    color: rgba(255, 255, 255, 0.4);
    font-size: 20px;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* ── Branding ── */
  .brand-footer {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding-bottom: 60px;
  }

  .brand-divider {
    width: 60px;
    height: 2px;
    background: rgba(255, 255, 255, 0.2);
  }

  .reviewed-label {
    color: rgba(255, 255, 255, 0.5);
    font-size: 20px;
    font-weight: 500;
    letter-spacing: 1px;
  }

  .brand-logo {
    font-size: 32px;
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
    color: rgba(255, 255, 255, 0.35);
    font-size: 16px;
    font-weight: 500;
  }
</style>
