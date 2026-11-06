<template>
  <div class="movie-reviews-component">
    <div class="write-review-box">
      <textarea
        v-model="form.body.value"
        placeholder="แบ่งปันความคิดเห็นของคุณต่อภาพยนตร์เรื่องนี้..."
        rows="3"
      ></textarea>

      <div class="extra-fields-row">
        <VisibilitySelector v-model="form.visibility.value" size="sm" />
      </div>

      <div class="form-footer">
        <div class="rating-input">
          <span>ให้คะแนนของคุณ: </span>
          <StarRatingInput v-model="form.rating.value" />
        </div>
        <button
          class="btn-submit-review"
          :disabled="!form.canSubmit.value"
          @click="handleSubmit"
        >
          {{ form.isSubmitting.value ? "กำลังส่ง..." : "ส่งรีวิว" }}
        </button>
      </div>

      <p v-if="!form.isAuthenticated.value" class="auth-hint">
        เข้าสู่ระบบเพื่อเขียนรีวิว
      </p>
    </div>

    <div v-if="isLoadingReviews" class="reviews-status">กำลังโหลดรีวิว...</div>

    <div v-else-if="reviews.length" class="reviews-list">
      <div v-for="item in reviews" :key="item.id" class="review-card">
        <div class="review-header">
          <div
            class="user-info"
            role="button"
            tabindex="0"
            @click="goToProfile(item.user.id)"
            @keydown.enter="goToProfile(item.user.id)"
          >
            <div class="user-avatar">
              <img
                v-if="item.user.avatar_url"
                :src="item.user.avatar_url"
                :alt="item.user.username"
                loading="lazy"
              />
              <i v-else class="pi pi-user"></i>
            </div>
            <div class="user-meta">
              <div class="username-row">
                <span class="username">
                  {{ item.user.display_name || item.user.username }}
                </span>

                <span
                  v-if="shouldShowFollowBadge(item.user.id)"
                  class="follow-status-badge"
                  :class="followBadgeClass(item.user.id)"
                >
                  {{ followBadgeLabel(item.user.id) }}
                </span>
              </div>
              <div class="meta-badges">
                <span class="review-date">{{
                  formatDate(item.created_at)
                }}</span>

                <span v-if="item.visibility === 'private'" class="private-badge">
                  <i class="pi pi-lock"></i>
                  ส่วนตัว
                </span>
                <span v-else-if="item.visibility === 'followers'" class="followers-badge">
                  <i class="pi pi-users"></i>
                  ผู้ติดตาม
                </span>
              </div>
            </div>
          </div>
          <div class="user-score">
            <i class="pi pi-star-fill"></i>
            <span>{{ item.rating }}/5</span>
          </div>
        </div>

        <p class="review-comment">{{ item.body }}</p>

        <div class="review-actions">
          <button
            class="action-btn like-btn"
            :class="{ active: item.is_liked }"
            :disabled="pendingLike.has(item.id)"
            @click="toggleLike(item)"
          >
            <i :class="item.is_liked ? 'pi pi-heart-fill' : 'pi pi-heart'"></i>
            <span v-if="item.like_count > 0">{{ item.like_count }}</span>
            <span v-else class="action-label">ถูกใจ</span>
          </button>

          <button
            class="action-btn helpful-btn"
            :class="{ active: item.is_helpful_voted }"
            :disabled="pendingHelpful.has(item.id)"
            @click="toggleHelpful(item)"
          >
            <i
              :class="
                item.is_helpful_voted
                  ? 'pi pi-thumbs-up-fill'
                  : 'pi pi-thumbs-up'
              "
            ></i>
            <span class="action-label-full">
              รีวิวนี้มีประโยชน์{{
                item.helpful_count > 0 ? ` (${item.helpful_count})` : ""
              }}
            </span>
            <span class="action-label-short">
              มีประโยชน์{{
                item.helpful_count > 0 ? ` (${item.helpful_count})` : ""
              }}
            </span>
          </button>

          <span v-if="item.comment_count > 0" class="comment-count">
            <i class="pi pi-comment"></i>
            {{ item.comment_count }} ความคิดเห็น
          </span>

          <button
            v-if="item.visibility === 'public'"
            class="action-btn share-btn"
            @click="openReviewShare(item)"
          >
            <i class="pi pi-share-alt"></i>
            <span class="action-label">{{ $t("share.share") }}</span>
          </button>
        </div>
      </div>
    </div>

    <ShareImageModal
      v-if="selectedShareMediaContext && selectedShareReviewContext"
      v-model="showReviewShareModal"
      share-type="review"
      :media="selectedShareMediaContext"
      :review="selectedShareReviewContext"
    />

    <div v-else class="empty-reviews">
      <i class="pi pi-inbox"></i>
      <p>ยังไม่มีรีวิวสำหรับเรื่องนี้ มารีวิวเป็นคนแรกกันเถอะ!</p>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { followApi, reviewApi } from "@/api/api"
  import { useAuthStore } from "@/stores/auth"
  import { computed, onMounted, ref, watch } from "vue"
  import type { ReviewResponse } from "@/types/review"
  import StarRatingInput from "@/components/movie/Starratinginput.vue"
  import VisibilitySelector from "@/components/common/VisibilitySelector.vue"
  import { useReviewForm } from "@/composables/useReviewForm"
  import { useProfileNav } from "@/composables/useProfileNav"
  import ShareImageModal from "@/components/share/ShareImageModal.vue"
  import type { ShareMediaContext, ShareReviewContext } from "@/types/share"

  const props = withDefaults(
    defineProps<{
      movieId: number
      mediaType?: "movie" | "tv"
      mediaPosterPath?: string | null
      mediaTitle?: string
    }>(),
    { mediaType: "movie" },
  )

  const emit = defineEmits<{
    (e: "review-submitted"): void
  }>()

  const auth = useAuthStore()

  type FollowStatus = "none" | "pending" | "accepted"

  const followStatusByUserId = ref<Record<number, FollowStatus>>({})

  const reviews = ref<ReviewResponse[]>([])
  const isLoadingReviews = ref(false)

  // Review Share state
  const showReviewShareModal = ref(false)
  const selectedShareReview = ref<ReviewResponse | null>(null)

  function openReviewShare(reviewItem: ReviewResponse) {
    selectedShareReview.value = reviewItem
    showReviewShareModal.value = true
  }

  const selectedShareMediaContext = computed<ShareMediaContext | null>(() => {
    if (!selectedShareReview.value) return null
    const r = selectedShareReview.value
    return {
      id: props.movieId,
      mediaType: props.mediaType,
      title: props.mediaTitle || r.media?.title || "",
      posterPath: props.mediaPosterPath || null,
      posterUrl: r.media?.poster_url || null,
      releaseYear: null,
      genres: r.media?.genres || [],
      voteAverage: r.media?.vote_average || 0,
      popcornRating: null,
    }
  })

  const selectedShareReviewContext = computed<ShareReviewContext | null>(() => {
    if (!selectedShareReview.value) return null
    const r = selectedShareReview.value
    return {
      id: r.id,
      authorDisplayName: r.user.display_name || r.user.username,
      authorUsername: r.user.username,
      authorAvatarUrl: r.user.avatar_url,
      rating: r.rating,
      body: r.body,
      visibility: r.visibility || "public",
      createdAt: r.created_at,
    }
  })

  // ป้องกันการกดซ้ำระหว่างรอ response กลับมา
  const pendingLike = ref<Set<number>>(new Set())
  const pendingHelpful = ref<Set<number>>(new Set())

  const { goToProfile } = useProfileNav()

  const reviewSegment = computed(() =>
    props.mediaType === "tv" ? "tv" : "movies",
  )
  const form = useReviewForm(props.movieId, props.mediaType)

  const dateFormatter = new Intl.DateTimeFormat("th-TH", {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
  function formatDate(iso: string) {
    return dateFormatter.format(new Date(iso))
  }

  function normalizeFollowStatus(data: unknown): FollowStatus {
    const value = data as {
      status?: string
      is_following?: boolean
      following?: boolean
      is_pending?: boolean
    }

    if (value.status === "accepted" || value.status === "following") {
      return "accepted"
    }

    if (value.status === "pending") {
      return "pending"
    }

    if (value.is_following || value.following) {
      return "accepted"
    }

    if (value.is_pending) {
      return "pending"
    }

    return "none"
  }

  function shouldShowFollowBadge(userId: number) {
    return Boolean(auth.user && auth.user.id !== userId)
  }

  function getFollowStatus(userId: number): FollowStatus {
    return followStatusByUserId.value[userId] ?? "none"
  }

  function followBadgeLabel(userId: number) {
    const status = getFollowStatus(userId)

    if (status === "accepted") return "Following"
    if (status === "pending") return "Requested"

    return "Follow"
  }

  function followBadgeClass(userId: number) {
    const status = getFollowStatus(userId)

    return {
      "follow-status-badge--following": status === "accepted",
      "follow-status-badge--pending": status === "pending",
      "follow-status-badge--none": status === "none",
    }
  }

  async function loadReviewAuthorFollowStatuses() {
    if (!auth.user) {
      followStatusByUserId.value = {}
      return
    }

    const uniqueUserIds = Array.from(
      new Set(
        reviews.value
          .map(review => review.user.id)
          .filter(userId => userId !== auth.user?.id),
      ),
    )

    if (!uniqueUserIds.length) {
      followStatusByUserId.value = {}
      return
    }

    const entries = await Promise.all(
      uniqueUserIds.map(async userId => {
        try {
          const res = await followApi.getFollowStatus(userId)
          return [userId, normalizeFollowStatus(res.data)] as const
        } catch (err) {
          console.error(`getFollowStatus failed for user ${userId}:`, err)
          return [userId, "none"] as const
        }
      }),
    )

    followStatusByUserId.value = Object.fromEntries(entries)
  }

  async function loadReviews() {
    if (!props.movieId) return
    isLoadingReviews.value = true

    try {
      const res = await reviewApi.getMediaReviews(
        reviewSegment.value,
        props.movieId,
      )

      reviews.value = res.data.reviews
      await loadReviewAuthorFollowStatuses()
    } catch (err) {
      console.error("getMediaReviews failed:", err)
    } finally {
      isLoadingReviews.value = false
    }
  }

  async function handleSubmit() {
    const created = await form.submit()
    if (!created) return
    reviews.value.unshift(created)
    emit("review-submitted")
  }

  // ── กดไลค์รีวิว (optimistic update + rollback ถ้า error) ──────────
  async function toggleLike(item: ReviewResponse) {
    if (pendingLike.value.has(item.id)) return
    pendingLike.value.add(item.id)

    const wasLiked = item.is_liked
    item.is_liked = !wasLiked
    item.like_count += wasLiked ? -1 : 1

    try {
      if (wasLiked) {
        await reviewApi.unlikeReview(item.id)
      } else {
        await reviewApi.likeReview(item.id)
      }
    } catch (err) {
      // rollback ถ้า request ล้มเหลว
      item.is_liked = wasLiked
      item.like_count += wasLiked ? 1 : -1
      console.error("toggleLike failed:", err)
    } finally {
      pendingLike.value.delete(item.id)
    }
  }

  // ── โหวต "มีประโยชน์" แบบ Stack Overflow ─────────────────────────
  async function toggleHelpful(item: ReviewResponse) {
    if (pendingHelpful.value.has(item.id)) return
    pendingHelpful.value.add(item.id)

    const wasVoted = item.is_helpful_voted
    item.is_helpful_voted = !wasVoted
    item.helpful_count += wasVoted ? -1 : 1

    try {
      if (wasVoted) {
        await reviewApi.unmarkHelpful(item.id)
      } else {
        await reviewApi.markHelpful(item.id)
      }
    } catch (err) {
      item.is_helpful_voted = wasVoted
      item.helpful_count += wasVoted ? 1 : -1
      console.error("toggleHelpful failed:", err)
    } finally {
      pendingHelpful.value.delete(item.id)
    }
  }

  // Re-fetch if this component instance gets reused for a different movie
  // (e.g. router reuses the view between two movie detail pages).
  watch(() => props.movieId, loadReviews)
  onMounted(loadReviews)
</script>

<style scoped>
  .movie-reviews-component {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    margin-top: 0.5rem;
    font-family: "Noto Sans Thai", sans-serif;
  }

  .write-review-box {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 1rem;
  }

  textarea {
    width: 100%;
    background: rgba(0, 0, 0, 0.3);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 0.75rem;
    color: #fff;
    font-family: inherit;
    font-size: 0.9rem;
    resize: none;
    outline: none;
    box-sizing: border-box;
  }
  textarea:focus {
    border-color: var(--color-brand);
  }

  .username-row {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    flex-wrap: wrap;
    min-width: 0;
  }

  .follow-status-badge {
    display: inline-flex;
    align-items: center;
    border-radius: 999px;
    padding: 0.08rem 0.5rem;
    font-size: 0.66rem;
    font-weight: 650;
    line-height: 1.4;
    white-space: nowrap;
    border: 1px solid transparent;
  }

  .follow-status-badge--following {
    background: rgba(245, 67, 44, 0.1);
    color: var(--color-brand);
    border-color: rgba(245, 67, 44, 0.3);
  }
  .follow-status-badge--pending {
    background: rgba(255, 255, 255, 0.04);
    color: #8a8a8a;
    border-color: rgba(255, 255, 255, 0.08);
  }
  .follow-status-badge--none {
    background: rgba(255, 255, 255, 0.04);
    color: #8a8a8a;
    border-color: rgba(255, 255, 255, 0.08);
  }

  .form-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 0.75rem;
    gap: 0.75rem;
  }

  .rating-input {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.85rem;
    color: #8a8a8a;
    flex-wrap: wrap;
  }

  .btn-submit-review {
    background: var(--red, var(--color-brand));
    color: #fff;
    border: none;
    padding: 0.5rem 1.4rem;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.85rem;
    cursor: pointer;
    transition:
      background 0.2s,
      opacity 0.2s;
    white-space: nowrap;
    flex-shrink: 0;
  }
  .btn-submit-review:hover:not(:disabled) {
    background: var(--color-brand-hover);
  }
  .btn-submit-review:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .auth-hint {
    margin: 0.5rem 0 0;
    font-size: 0.75rem;
    color: #8a8a8a;
  }

  .reviews-status {
    text-align: center;
    padding: 1.5rem;
    color: #8a8a8a;
    font-size: 0.875rem;
  }

  .reviews-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .review-card {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 10px;
    padding: 1rem;
    transition: border-color 0.2s;
    min-width: 0; /* กัน flex/grid parent บีบจนล้น */
  }
  .review-card:hover {
    border-color: rgba(255, 255, 255, 0.1);
  }

  .review-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
    flex-wrap: wrap;
  }

  .user-info {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    min-width: 0;
    flex: 1 1 auto;
    cursor: pointer;
    border-radius: 8px;
  }
  .user-info:focus-visible {
    outline: 2px solid var(--red, var(--color-brand));
    outline-offset: 3px;
  }

  /* ── รูป user แบบวงกลม ── */
  .user-avatar {
    width: 40px;
    height: 40px;
    min-width: 40px;
    border-radius: 50%;
    overflow: hidden;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.12);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #8a8a8a;
    flex-shrink: 0;
  }
  .user-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 50%;
  }
  .user-avatar i {
    font-size: 1.1rem;
  }

  .user-meta {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    min-width: 0;
  }

  .username {
    font-size: 0.875rem;
    font-weight: 600;
    color: #fff;
    overflow-wrap: anywhere;
    transition: color 0.15s;
  }
  .user-info:hover .username {
    color: var(--color-brand);
    text-decoration: underline;
  }

  .meta-badges {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem 0.5rem;
  }

  .review-date {
    font-size: 0.75rem;
    color: #8a8a8a;
    white-space: nowrap;
  }

  .watched-badge,
  .private-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.72rem;
    color: #8a8a8a;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 999px;
    padding: 0.1rem 0.55rem;
    white-space: nowrap;
  }
  .watched-badge i,
  .private-badge i {
    font-size: 0.68rem;
  }
  .private-badge {
    color: #f5c518;
    background: rgba(245, 197, 24, 0.08);
    border-color: rgba(245, 197, 24, 0.18);
  }

  .user-score {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    background: rgba(245, 197, 24, 0.1);
    border: 1px solid rgba(245, 197, 24, 0.25);
    color: var(--gold, #f5c518);
    padding: 0.3rem 0.7rem;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .review-comment {
    font-size: 0.9rem;
    line-height: 1.6;
    color: #ccc;
    margin: 0 0 0.9rem;
    overflow-wrap: break-word;
    word-break: break-word;
  }

  /* ── action row: like / helpful / comments ── */
  .review-actions {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex-wrap: wrap;
    padding-top: 0.75rem;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    color: #8a8a8a;
    font-size: 0.78rem;
    font-weight: 500;
    padding: 0.35rem 0.75rem;
    border-radius: 999px;
    cursor: pointer;
    transition:
      background 0.15s,
      border-color 0.15s,
      color 0.15s;
    white-space: nowrap;
  }
  .action-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
    color: #fff;
  }
  .action-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .action-label-short {
    display: none;
  }
  .action-label-full {
    display: inline;
  }

  .like-btn.active {
    color: var(--color-brand);
    background: rgba(245, 67, 44, 0.1);
    border-color: rgba(245, 67, 44, 0.3);
  }

  .helpful-btn.active {
    color: #4ade80;
    background: rgba(74, 222, 128, 0.1);
    border-color: rgba(74, 222, 128, 0.3);
  }

  .comment-count {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.78rem;
    color: #8a8a8a;
    margin-left: auto;
    white-space: nowrap;
  }

  .empty-reviews {
    text-align: center;
    padding: 2.5rem;
    color: #8a8a8a;
    background: rgba(255, 255, 255, 0.01);
    border-radius: 12px;
    border: 1px dashed rgba(255, 255, 255, 0.05);
  }
  .empty-reviews i {
    font-size: 2rem;
    margin-bottom: 0.5rem;
  }
  .empty-reviews p {
    margin: 0;
    font-size: 0.875rem;
  }

  .extra-fields-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    margin-top: 0.85rem;
    padding-top: 0.85rem;
    border-top: 1px solid var(--border, rgba(255, 255, 255, 0.08));
    flex-wrap: wrap;
  }

  .watched-at-field {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    font-size: 0.78rem;
    color: var(--muted, #8a8a8a);
    flex: 1 1 220px;
    min-width: 0;
  }

  .watched-at-field label {
    white-space: nowrap;
  }

  .watched-at-field :deep(.dp__main) {
    flex: 1 1 auto;
    min-width: 0;
  }

  .watched-at-field :deep(.dp__input) {
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid var(--border, rgba(255, 255, 255, 0.08));
    border-radius: 6px;
    padding: 0.4rem 0.7rem 0.4rem 2.2rem;
    color: #fff;
    font-family: inherit;
    font-size: 0.8rem;
    min-height: unset;
    width: 150px;
    max-width: 100%;
    box-sizing: border-box;
  }

  .watched-at-field :deep(.dp__input:hover) {
    border-color: rgba(255, 255, 255, 0.2);
  }

  .watched-at-field :deep(.dp__input_focus) {
    border-color: var(--red, var(--color-brand));
  }

  .watched-at-field :deep(.dp__icon) {
    color: var(--muted, #8a8a8a);
  }

  .visibility-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0.85rem;
    border-radius: 9999px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border, rgba(255, 255, 255, 0.08));
    font-size: 0.78rem;
    color: var(--muted, #8a8a8a);
    cursor: pointer;
    user-select: none;
    transition:
      background 0.2s,
      border-color 0.2s,
      color 0.2s;
    white-space: nowrap;
    flex-shrink: 0;
  }
  .visibility-toggle:hover {
    background: rgba(255, 255, 255, 0.09);
    border-color: rgba(255, 255, 255, 0.2);
  }
  .visibility-toggle input[type="checkbox"] {
    accent-color: var(--red, var(--color-brand));
    cursor: pointer;
  }
  .visibility-toggle i {
    font-size: 0.8rem;
  }
  .visibility-toggle:has(input:checked) {
    background: rgba(245, 197, 24, 0.12);
    border-color: rgba(245, 197, 24, 0.3);
    color: var(--gold, #f5c518);
  }
  .visibility-toggle:has(input:checked) i {
    color: var(--gold, #f5c518);
  }

  /* ============================================================
   RESPONSIVE BREAKPOINTS
   Desktop (>1024px) = ค่าเริ่มต้นด้านบน
   ============================================================ */

  @media (max-width: 1024px) {
    .movie-reviews-component {
      gap: 1.25rem;
    }

    .write-review-box {
      padding: 0.9rem;
    }

    .review-card {
      padding: 0.9rem;
    }
  }

  @media (max-width: 768px) {
    .write-review-box {
      padding: 0.85rem;
      border-radius: 10px;
    }

    textarea {
      font-size: 0.85rem;
      padding: 0.65rem;
    }

    .extra-fields-row {
      gap: 0.75rem;
    }

    .watched-at-field {
      flex: 1 1 100%;
    }

    .watched-at-field :deep(.dp__input) {
      width: 100%;
    }

    .visibility-toggle {
      flex: 1 1 100%;
      justify-content: center;
    }

    .review-card {
      padding: 0.85rem;
      border-radius: 8px;
    }

    .user-avatar {
      width: 36px;
      height: 36px;
      min-width: 36px;
    }

    .review-comment {
      font-size: 0.87rem;
    }
  }

  @media (max-width: 640px) {
    .movie-reviews-component {
      gap: 1rem;
    }

    .form-footer {
      flex-direction: column;
      align-items: stretch;
      gap: 0.6rem;
    }

    .rating-input {
      justify-content: space-between;
    }

    .btn-submit-review {
      width: 100%;
      padding: 0.6rem 1.4rem;
    }

    .review-header {
      align-items: center;
    }

    .user-score {
      font-size: 0.75rem;
      padding: 0.25rem 0.6rem;
    }

    .review-actions {
      gap: 0.5rem;
    }

    .action-btn {
      padding: 0.4rem 0.7rem;
      font-size: 0.75rem;
    }

    .comment-count {
      margin-left: 0;
      flex-basis: 100%;
      justify-content: flex-start;
      padding-top: 0.3rem;
    }
  }

  @media (max-width: 480px) {
    .movie-reviews-component {
      gap: 0.85rem;
    }

    .write-review-box {
      padding: 0.75rem;
      border-radius: 8px;
    }

    textarea {
      rows: 3;
      font-size: 0.83rem;
      padding: 0.6rem;
    }

    .extra-fields-row {
      margin-top: 0.7rem;
      padding-top: 0.7rem;
      gap: 0.6rem;
    }

    .watched-at-field {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.35rem;
    }

    .rating-input {
      font-size: 0.8rem;
    }

    .review-card {
      padding: 0.7rem;
    }

    .review-header {
      gap: 0.5rem;
    }

    .user-info {
      gap: 0.55rem;
    }

    .user-avatar {
      width: 32px;
      height: 32px;
      min-width: 32px;
    }
    .user-avatar i {
      font-size: 0.95rem;
    }

    .username {
      font-size: 0.82rem;
    }

    .review-date {
      font-size: 0.7rem;
    }

    .watched-badge,
    .private-badge {
      font-size: 0.68rem;
      padding: 0.08rem 0.45rem;
    }

    .user-score {
      font-size: 0.72rem;
      padding: 0.22rem 0.55rem;
    }

    .review-comment {
      font-size: 0.85rem;
      margin-bottom: 0.7rem;
    }

    .action-label-full {
      display: none;
    }
    .action-label-short {
      display: inline;
    }

    .action-btn {
      flex: 1 1 auto;
      justify-content: center;
    }

    .empty-reviews {
      padding: 1.75rem 1rem;
    }
    .empty-reviews i {
      font-size: 1.6rem;
    }
    .empty-reviews p {
      font-size: 0.82rem;
    }
  }

  @media (max-width: 360px) {
    .write-review-box {
      padding: 0.6rem;
    }

    textarea {
      font-size: 0.8rem;
      padding: 0.5rem;
    }

    .btn-submit-review {
      font-size: 0.8rem;
      padding: 0.55rem 1rem;
    }

    .review-card {
      padding: 0.6rem;
    }

    .user-avatar {
      width: 28px;
      height: 28px;
      min-width: 28px;
    }

    .username {
      font-size: 0.78rem;
    }

    .review-comment {
      font-size: 0.82rem;
    }

    .review-actions {
      flex-direction: column;
      align-items: stretch;
      gap: 0.4rem;
    }

    .action-btn {
      width: 100%;
    }

    .comment-count {
      justify-content: center;
    }
  }

  /* ============================================================
   Minimal calendar redesign for VueDatePicker
   ============================================================ */

  :deep(.movie-dp-menu.dp__menu) {
    background: #111113;
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 16px;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.45);
    padding: 0.5rem;
    font-family: inherit;
    --dp-primary-color: var(--color-brand);
    --dp-primary-text-color: #fff;
    --dp-border-color: transparent;
    --dp-border-color-hover: transparent;
    --dp-menu-border-color: transparent;
  }

  :deep(.movie-dp-menu .dp__arrow_top) {
    display: none;
  }

  :deep(.movie-dp-menu .dp__calendar_header) {
    color: #6b6b6b;
    font-size: 0.7rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  :deep(.movie-dp-menu .dp__calendar_header_separator) {
    display: none;
  }

  :deep(.movie-dp-menu .dp__month_year_row) {
    padding: 0.35rem 0.4rem 0.75rem;
  }

  :deep(.movie-dp-menu .dp__month_year_select) {
    color: #fff;
    font-size: 0.9rem;
    font-weight: 600;
    border-radius: 8px;
    transition: background 0.15s;
  }
  :deep(.movie-dp-menu .dp__month_year_select:hover) {
    background: rgba(255, 255, 255, 0.06);
  }

  :deep(.movie-dp-menu .dp__inner_nav) {
    color: #8a8a8a;
    border-radius: 999px;
    transition:
      background 0.15s,
      color 0.15s;
  }
  :deep(.movie-dp-menu .dp__inner_nav:hover) {
    background: rgba(255, 255, 255, 0.08);
    color: #fff;
  }
  :deep(.movie-dp-menu .dp__inner_nav svg) {
    width: 14px;
    height: 14px;
  }

  :deep(.movie-dp-menu .dp__calendar_item) {
    padding: 1px;
  }

  :deep(.movie-dp-menu .dp__cell_inner) {
    height: 34px;
    width: 34px;
    font-size: 0.82rem;
    font-weight: 400;
    color: #d4d4d4;
    border-radius: 999px;
    border: none;
    transition:
      background 0.15s,
      color 0.15s;
  }

  :deep(.movie-dp-menu .dp__cell_inner:hover) {
    background: rgba(255, 255, 255, 0.08);
  }

  :deep(.movie-dp-menu .dp__cell_offset) {
    color: #3d3d3d;
  }

  :deep(.movie-dp-menu .dp__today) {
    border: 1px solid rgba(245, 197, 24, 0.5);
    color: #f5c518;
    font-weight: 600;
  }

  :deep(.movie-dp-menu .dp__active_date) {
    background: var(--red, var(--color-brand));
    color: #fff;
    font-weight: 600;
  }
  :deep(.movie-dp-menu .dp__active_date:hover) {
    background: var(--color-brand-hover);
  }

  :deep(.movie-dp-menu .dp__cell_disabled) {
    color: #3d3d3d;
    opacity: 0.5;
  }

  :deep(.movie-dp-menu .dp__action_row) {
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    padding: 0.6rem 0.4rem 0.2rem;
    margin-top: 0.35rem;
  }

  :deep(.movie-dp-menu .dp__action_button) {
    font-size: 0.78rem;
    font-weight: 600;
    border-radius: 8px;
    padding: 0.4rem 0.9rem;
  }

  :deep(.movie-dp-menu .dp__select) {
    background: var(--red, var(--color-brand));
    color: #fff;
  }
  :deep(.movie-dp-menu .dp__select:hover) {
    background: var(--color-brand-hover);
  }

  @media (max-width: 420px) {
    :deep(.movie-dp-menu.dp__menu) {
      max-width: calc(100vw - 2rem);
    }
  }
</style>
