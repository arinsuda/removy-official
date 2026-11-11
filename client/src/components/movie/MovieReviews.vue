<template>
  <div class="reviews">
    <OwnReviewCard v-if="ownReview" ref="own" :review="ownReview" @updated="onOwnUpdated" />
    <ReviewComposer
      v-else-if="auth.isLoggedIn && ownLoaded"
      ref="composer"
      :media-id="movieId"
      :media-type="mediaType"
      @submitted="onSubmitted"
    />
    <p v-else-if="auth.isLoggedIn && ownFailed" class="reviews__guest" role="alert" data-test="own-failed">
      {{ t("reviews.ownLoadFailed") }}
      <AppButton variant="ghost" size="sm" data-test="own-retry" @click="loadOwn(loadSeq)">{{ t("common.retry") }}</AppButton>
    </p>
    <p v-else-if="!auth.isLoggedIn" class="reviews__guest">
      {{ t("reviews.guestPrompt") }}
      <RouterLink :to="{ name: 'login', query: buildAuthQuery(route.fullPath) }" class="reviews__link">
        {{ t("media.detail.activity.signIn") }}
      </RouterLink>
    </p>

    <SkeletonList v-if="isLoadingReviews" avatar :count="3" />
    <AppErrorState v-else-if="loadFailed" :title="t('reviews.loadFailed')" @retry="loadReviews" />
    <p v-else-if="others.length" class="reviews__total" data-test="reviews-total">{{ t("reviews.list.total", { count: totalCount }, totalCount) }}</p>
    <ul v-if="!isLoadingReviews && !loadFailed && others.length" class="reviews__list">
      <li v-for="item in others" :key="item.id">
        <ReviewCard
          :review="item"
          :follow-status="followStatusFor(item.user.id)"
          :like-pending="pendingLike.has(item.id)"
          :helpful-pending="pendingHelpful.has(item.id)"
          @like="toggleLike"
          @helpful="toggleHelpful"
          @share="openReviewShare"
          @profile="goToProfile"
        />
      </li>
    </ul>
    <div v-if="!isLoadingReviews && hasMore" class="reviews__more">
      <AppButton variant="outline" :loading="isLoadingMore" data-test="reviews-more" @click="loadMore">{{ t("reviews.list.loadMore") }}</AppButton>
    </div>
    <p v-else-if="!isLoadingReviews && !loadFailed && !others.length && !ownReview" class="reviews__empty">{{ t("reviews.list.empty") }}</p>

    <ShareImageModal
      v-if="selectedShareMediaContext && selectedShareReviewContext"
      v-model="showReviewShareModal"
      share-type="review"
      :media="selectedShareMediaContext"
      :review="selectedShareReviewContext"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink, useRoute } from "vue-router"
import { httpStatus } from "@/components/lists/listErrors"
import { followApi, reviewApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { useAuthGate } from "@/composables/useAuthGate"
import { useProfileNav } from "@/composables/useProfileNav"
import { buildAuthQuery } from "@/utils/safeRedirect"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import ShareImageModal from "@/components/share/ShareImageModal.vue"
import OwnReviewCard from "@/components/detail/reviews/OwnReviewCard.vue"
import ReviewCard from "@/components/detail/reviews/ReviewCard.vue"
import ReviewComposer from "@/components/detail/reviews/ReviewComposer.vue"
import type { ReviewResponse } from "@/types/review"
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
  /** The visible reviews changed (loaded, created, edited). */
  (e: "loaded", reviews: ReviewResponse[], total: number): void
}>()

const { t } = useI18n()
const auth = useAuthStore()
const route = useRoute()
const { requireAuth } = useAuthGate()
const { goToProfile } = useProfileNav()
const composer = useTemplateRef<InstanceType<typeof ReviewComposer>>("composer")
const own = useTemplateRef<InstanceType<typeof OwnReviewCard>>("own")

type FollowStatus = "none" | "pending" | "accepted"
const followStatusByUserId = ref<Record<number, FollowStatus>>({})
const isLoadingReviews = ref(false)
const loadFailed = ref(false)
let loadSeq = 0

// Pages arrive newest first and may or may not contain the signed-in user's review, so that
// review is looked up on its own (the user's reviews, filtered by title) and never depends on page 1.
const pageReviews = ref<ReviewResponse[]>([])
const ownReview = ref<ReviewResponse | null>(null)
const ownLoaded = ref(false)
const ownFailed = ref(false)
const hasMore = ref(false)
const page = ref(1)
const totalCount = ref(0)
const isLoadingMore = ref(false)
const others = computed(() => pageReviews.value.filter(r => r.user.id !== auth.user?.id))
// The own review first, then everyone else; this is what the detail page reads.
const reviews = computed(() => (ownReview.value ? [ownReview.value, ...others.value] : others.value))

const pendingLike = ref(new Set<number>())
const pendingHelpful = ref(new Set<number>())

// Review share state
const showReviewShareModal = ref(false)
const selectedShareReview = ref<ReviewResponse | null>(null)

function openReviewShare(item: ReviewResponse) {
  selectedShareReview.value = item
  showReviewShareModal.value = true
}

const selectedShareMediaContext = computed<ShareMediaContext | null>(() => {
  const r = selectedShareReview.value
  if (!r) return null
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
  const r = selectedShareReview.value
  if (!r) return null
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

function normalizeFollowStatus(data: unknown): FollowStatus {
  const v = data as { status?: string; is_following?: boolean; following?: boolean; is_pending?: boolean }
  if (v.status === "accepted" || v.status === "following" || v.is_following || v.following) return "accepted"
  if (v.status === "pending" || v.is_pending) return "pending"
  return "none"
}

function followStatusFor(userId: number): FollowStatus | null {
  if (!auth.user || auth.user.id === userId) return null
  return followStatusByUserId.value[userId] ?? "none"
}

/** Follow status of authors not seen yet; known ones are kept. Replies for a stale load are dropped. */
async function loadReviewAuthorFollowStatuses(list: ReviewResponse[], seq: number) {
  if (!auth.user) {
    followStatusByUserId.value = {}
    return
  }
  const ids = [...new Set(list.map(r => r.user.id).filter(id => id !== auth.user?.id && !(id in followStatusByUserId.value)))]
  const entries = await Promise.all(
    ids.map(async userId => {
      try {
        return [userId, normalizeFollowStatus((await followApi.getFollowStatus(userId)).data)] as const
      } catch (err) {
        console.error(`getFollowStatus failed for user ${userId}, status:`, httpStatus(err))
        return [userId, "none"] as const
      }
    }),
  )
  if (seq !== loadSeq) return
  followStatusByUserId.value = { ...followStatusByUserId.value, ...Object.fromEntries(entries) }
}

/** Own review via its own endpoint, independent of page 1. A failure never reads as "no review". */
async function loadOwn(seq: number) {
  if (!auth.user) {
    ownReview.value = null
    ownLoaded.value = true
    return
  }
  ownFailed.value = false
  try {
    const res = await reviewApi.getMyReview(props.mediaType, props.movieId)
    if (seq !== loadSeq) return
    ownReview.value = res.data.review
    ownLoaded.value = true
  } catch (err) {
    if (seq !== loadSeq) return
    console.error("load own review failed, status:", httpStatus(err))
    ownFailed.value = true
  }
}

async function loadReviews() {
  if (!props.movieId) return
  const seq = ++loadSeq
  isLoadingReviews.value = true
  loadFailed.value = false
  ownLoaded.value = false
  ownFailed.value = false
  ownReview.value = null
  followStatusByUserId.value = {}
  pageReviews.value = []
  hasMore.value = false
  page.value = 1
  totalCount.value = 0
  try {
    const [res] = await Promise.all([reviewApi.getMediaReviews(props.mediaType, props.movieId, { page: 1 }), loadOwn(seq)])
    if (seq !== loadSeq) return
    pageReviews.value = res.data.reviews
    hasMore.value = res.data.pagination?.has_more ?? false
    totalCount.value = res.data.pagination?.total ?? res.data.reviews.length
    await loadReviewAuthorFollowStatuses(res.data.reviews, seq)
  } catch (err) {
    if (seq !== loadSeq) return
    console.error("getMediaReviews failed, status:", httpStatus(err))
    loadFailed.value = true
  } finally {
    if (seq === loadSeq) isLoadingReviews.value = false
  }
}

async function loadMore() {
  if (!hasMore.value || isLoadingMore.value) return
  const seq = loadSeq
  isLoadingMore.value = true
  try {
    const res = await reviewApi.getMediaReviews(props.mediaType, props.movieId, { page: page.value + 1 })
    if (seq !== loadSeq) return
    const known = new Set(pageReviews.value.map(r => r.id))
    pageReviews.value = [...pageReviews.value, ...res.data.reviews.filter(r => !known.has(r.id))]
    page.value = res.data.pagination.page
    hasMore.value = res.data.pagination.has_more
    totalCount.value = res.data.pagination.total
    await loadReviewAuthorFollowStatuses(res.data.reviews, seq)
  } catch (err) {
    console.error("load more reviews failed, status:", httpStatus(err))
  } finally {
    isLoadingMore.value = false
  }
}

function onSubmitted(created: ReviewResponse) {
  ownReview.value = created
  totalCount.value += 1
  emit("review-submitted")
}

function onOwnUpdated(updated: ReviewResponse) {
  ownReview.value = updated
  emit("review-submitted")
}

// Optimistic like with rollback on error.
async function toggleLike(item: ReviewResponse) {
  if (!requireAuth("like")) return
  if (pendingLike.value.has(item.id)) return
  pendingLike.value.add(item.id)
  const wasLiked = item.is_liked
  item.is_liked = !wasLiked
  item.like_count += wasLiked ? -1 : 1
  try {
    await (wasLiked ? reviewApi.unlikeReview(item.id) : reviewApi.likeReview(item.id))
  } catch (err) {
    item.is_liked = wasLiked
    item.like_count += wasLiked ? 1 : -1
    console.error("toggleLike failed, status:", httpStatus(err))
  } finally {
    pendingLike.value.delete(item.id)
  }
}

async function toggleHelpful(item: ReviewResponse) {
  if (!requireAuth("helpful")) return
  if (pendingHelpful.value.has(item.id)) return
  pendingHelpful.value.add(item.id)
  const wasVoted = item.is_helpful_voted
  item.is_helpful_voted = !wasVoted
  item.helpful_count += wasVoted ? -1 : 1
  try {
    await (wasVoted ? reviewApi.unmarkHelpful(item.id) : reviewApi.markHelpful(item.id))
  } catch (err) {
    item.is_helpful_voted = wasVoted
    item.helpful_count += wasVoted ? 1 : -1
    console.error("toggleHelpful failed, status:", httpStatus(err))
  } finally {
    pendingHelpful.value.delete(item.id)
  }
}

/** Rate button: edit your review if you have one, otherwise focus the composer. */
function focusComposer() {
  if (ownReview.value) own.value?.startEdit()
  else composer.value?.focus()
}

watch([reviews, totalCount], () => emit("loaded", [...reviews.value], totalCount.value), { deep: true })
watch(() => [props.movieId, props.mediaType], loadReviews)
onMounted(loadReviews)

defineExpose({ focusComposer })
</script>

<style scoped>
.reviews { display: flex; flex-direction: column; gap: 1rem; }
.reviews__list { display: flex; flex-direction: column; gap: 0.75rem; margin: 0; padding: 0; list-style: none; }
.reviews__total { margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.reviews__more { display: flex; justify-content: center; }
.reviews__empty, .reviews__guest { font-size: var(--text-small); color: var(--color-text-muted); }
.reviews__link { color: var(--color-brand); font-weight: 600; text-decoration: underline; text-underline-offset: 3px; }
</style>
