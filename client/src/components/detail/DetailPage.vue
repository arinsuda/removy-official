<template>
  <div class="detail" :aria-busy="status === 'loading'">
    <LoadingBar v-if="isSwitching" class="detail__progress" />

    <SkeletonMediaDetail v-if="status === 'loading' && !data" />

    <div v-else-if="status === 'notFound'" class="detail__state" data-test="detail-not-found">
      <h1 class="detail__state-title">{{ t(`media.detail.notFound.${config.type}`) }}</h1>
      <p class="detail__state-text">{{ t("media.detail.notFound.hint") }}</p>
      <AppButton :to="config.fallbackRoute">{{ t(`media.detail.notFound.browse.${config.type}`) }}</AppButton>
    </div>

    <AppErrorState
      v-else-if="status === 'error' && !data"
      :title="t('media.detail.loadFailed')"
      @retry="reload"
    />

    <template v-else-if="data">
      <DetailHero :data="data" :has-trailer="Boolean(trailerUrl)" @back="goBack" @trailer="openTrailer(trailerUrl)">
        <template #actions>
          <DetailActionBar
            :media-id="data.id"
            :media-type="config.type"
            :watchlisted="activity.isWatchlisted.value"
            :watchlist-pending="activity.pending.value.has('watchlist')"
            :has-own-review="Boolean(myReview)"
            :require-login="activity.requireLogin"
            @rate="onRate"
            @watchlist="activity.toggleWatchlist"
            @share="showShare = true"
            @logged="activity.markLogged"
          />
        </template>
      </DetailHero>

      <div class="detail__body">
        <aside class="detail__aside">
          <PopcornMeter
            class="detail__meter"
            :average="meter.average"
            :count="meter.count"
            :ratings="data.ratings"
            :vote-average="data.voteAverage"
            :liked="activity.isLiked.value"
            :like-count="activity.stats.value?.like_count ?? 0"
            :like-pending="activity.pending.value.has('like')"
            @toggle-like="activity.toggleLike"
          />
          <YourActivityPanel
            class="detail__activity"
            :logged-in="auth.isLoggedIn"
            :my-rating="myReview?.rating ?? null"
            :watch-summary="activity.watchSummary.value"
            :is-watchlisted="activity.isWatchlisted.value"
            :is-watched="activity.isWatched.value"
            :is-liked="activity.isLiked.value"
          />
          <DetailWatchProviders class="detail__providers" :type="config.type" :id="data.id" />
          <RatingDistribution class="detail__distribution" :distribution="activity.ratingDistribution.value" />
          <DetailFacts class="detail__facts" :facts="data.facts" :homepage="data.homepage" />
        </aside>

        <div class="detail__main">
          <DetailOverview class="detail__overview" :text="data.overview" />
          <TvSeasonsRail v-if="data.tvSeasons.length" class="detail__seasons" :seasons="data.tvSeasons" />
          <DetailCastRail v-if="data.cast.length" class="detail__cast" :cast="data.cast" />
          <DetailVideos v-if="videos.length" class="detail__videos" :videos="videos" @play="v => openTrailer(buildEmbedUrl(v.key))" />
          <section id="reviews" class="detail__reviews" aria-labelledby="detail-reviews-title">
            <h2 id="detail-reviews-title" class="detail-heading">{{ t("media.detail.reviews") }}</h2>
            <MovieReviews
              ref="reviews"
              :key="`${config.type}:${data.id}`"
              :movie-id="data.id"
              :media-type="config.type"
              :media-poster-path="data.posterPath"
              :media-title="data.englishTitle || data.title"
              @loaded="list => (reviewList = list)"
              @review-submitted="activity.refreshStats"
            />
          </section>
          <HomeRail
            class="detail__similar"
            :title="t('media.detail.similar')"
            :loading="similarLoading"
            :error="similarError"
            :empty="!similarLoading && !similarError && similar.length === 0"
            :skeleton-count="6"
            @retry="loadSimilar"
          >
            <HomePosterCard v-for="item in similar" :key="item.key" :item="item" rail="similar" />
          </HomeRail>
        </div>
      </div>

      <Teleport to="body">
        <PopupTrailer v-if="trailerOpen" :trailer-url="trailerOpen" @close="trailerOpen = null" />
      </Teleport>
      <ShareImageModal v-if="shareContext" v-model="showShare" share-type="poster" :media="shareContext" :stats="shareStats" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useRouter } from "vue-router"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import LoadingBar from "@/components/ui/LoadingBar.vue"
import SkeletonMediaDetail from "@/components/ui/skeleton/SkeletonMediaDetail.vue"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import HomeRail from "@/components/home/HomeRail.vue"
import type { HomeMedia } from "@/components/home/homeMedia"
import MovieReviews from "@/components/movie/MovieReviews.vue"
import PopupTrailer from "@/components/movie/PopupTrailer.vue"
import ShareImageModal from "@/components/share/ShareImageModal.vue"
import { buildEmbedUrl, resolveTrailer } from "@/composables/useTrailerPreview"
import { provideHomePopup } from "@/composables/useHomePopup"
import { useMediaActivity } from "@/composables/useMediaActivity"
import { useMediaDetail } from "@/composables/useMediaDetail"
import { useAuthStore } from "@/stores/auth"
import type { ReviewResponse } from "@/types/review"
import type { ShareMediaContext, ShareStatsContext } from "@/types/share"
import DetailActionBar from "./DetailActionBar.vue"
import DetailCastRail from "./DetailCastRail.vue"
import DetailFacts from "./DetailFacts.vue"
import DetailHero from "./DetailHero.vue"
import DetailOverview from "./DetailOverview.vue"
import DetailVideos from "./DetailVideos.vue"
import DetailWatchProviders from "./DetailWatchProviders.vue"
import PopcornMeter from "./PopcornMeter.vue"
import RatingDistribution from "./RatingDistribution.vue"
import TvSeasonsRail from "./TvSeasonsRail.vue"
import YourActivityPanel from "./YourActivityPanel.vue"
import type { DetailConfig } from "./detailConfig"

const props = defineProps<{ config: DetailConfig; id: number }>()

// HomePosterCard (similar titles) reads its hover popup from an ancestor.
provideHomePopup()

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()

const idRef = computed(() => props.id)
const { status, data, reload } = useMediaDetail(props.config, idRef)
// Activity follows the title on screen, so actions never hit a half-switched id.
const displayId = computed(() => data.value?.id ?? 0)
const activity = useMediaActivity(props.config.type, displayId)

const isSwitching = computed(() => status.value === "loading" && data.value !== null)

// Reviews
const reviewList = ref<ReviewResponse[]>([])
const reviewsRef = useTemplateRef<InstanceType<typeof MovieReviews>>("reviews")
const myReview = computed(() => reviewList.value.find(r => r.user.id === auth.user?.id) ?? null)

watch(displayId, () => {
  reviewList.value = []
})

function onRate() {
  document.getElementById("reviews")?.scrollIntoView({
    behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth",
    block: "start",
  })
  reviewsRef.value?.focusComposer()
}

// POPCORN numbers: live stats once loaded, the title payload before that.
const meter = computed(() => {
  const s = activity.stats.value
  if (s) return { average: s.has_rating ? s.average_rating : 0, count: s.has_rating ? s.review_count : 0 }
  const p = data.value?.ratings?.popcorn
  return { average: p?.available ? (p.average ?? 0) : 0, count: p?.available ? p.count : 0 }
})

// Trailer and videos
const trailerOpen = ref<string | null>(null)
const trailerUrl = computed(() => resolveTrailer(data.value?.videos)?.embedUrl ?? null)
const videos = computed(() => (data.value?.videos ?? []).filter(v => v.site === "YouTube").slice(0, 12))

function openTrailer(url: string | null) {
  if (url) trailerOpen.value = url
}

// Share
const showShare = ref(false)
const shareContext = computed<ShareMediaContext | null>(() => {
  const d = data.value
  if (!d) return null
  const s = activity.stats.value
  return {
    id: d.id,
    mediaType: props.config.type,
    title: d.englishTitle || d.title,
    posterPath: d.posterPath,
    posterUrl: null,
    releaseYear: d.year,
    genres: d.genres,
    voteAverage: d.voteAverage,
    popcornRating: s?.has_rating ? s.average_rating : null,
    popcornCount: s?.has_rating ? s.review_count : null,
  }
})
// The signed-in user's own activity only; guests and other people's data never reach the card.
const shareStats = computed<ShareStatsContext | null>(() => {
  if (!auth.isLoggedIn) return null
  const mine = myReview.value
  const watch = activity.watchSummary.value
  const stats: ShareStatsContext = {
    rating: mine?.rating ?? null,
    watchCount: watch?.watch_count ?? 0,
    lastWatchedOn: watch?.last_watched_on ?? null,
    reviewBody: mine?.body ?? null,
    reviewVisibility: mine?.visibility ?? null,
  }
  return stats.rating || stats.watchCount > 0 ? stats : null
})

// Similar titles
const similar = ref<HomeMedia[]>([])
const similarLoading = ref(false)
const similarError = ref(false)
let similarSeq = 0

async function loadSimilar() {
  const seq = ++similarSeq
  const id = displayId.value
  similar.value = []
  similarError.value = false
  if (!id) return
  similarLoading.value = true
  try {
    const items = await props.config.fetchSimilar(id)
    if (seq === similarSeq) similar.value = items.filter(i => i.poster_path).slice(0, 18)
  } catch (err) {
    if (seq !== similarSeq) return
    console.error("load similar failed:", err)
    similarError.value = true
  } finally {
    if (seq === similarSeq) similarLoading.value = false
  }
}
watch(displayId, loadSimilar, { immediate: true })

// Navigation
function goBack() {
  if (window.history.state?.back) router.back()
  else router.push(props.config.fallbackRoute)
}

// A new title starts at the top and closes anything tied to the old one.
watch(displayId, (_id, prev) => {
  trailerOpen.value = null
  showShare.value = false
  if (prev) window.scrollTo(0, 0)
})
</script>

<style>
/* Shared by the detail section components. */
.detail-heading {
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
</style>

<style scoped>
.detail {
  position: relative;
  min-height: 100vh; min-height: 100dvh;
  padding-bottom: calc(5.5rem + env(safe-area-inset-bottom));
  color: var(--color-text-primary);
  background: var(--color-surface);
  overflow-x: clip;
}
.detail__progress { position: fixed; top: 0; left: 0; z-index: var(--z-above-header); width: 100%; border-radius: 0; }
.detail__state { display: flex; flex-direction: column; align-items: center; gap: 0.75rem; padding: 8rem 1.5rem; text-align: center; }
.detail__state-title { font-size: var(--text-heading); font-weight: 800; }
.detail__state-text { max-width: 28rem; color: var(--color-text-secondary); }

/* Below 1024px the aside and main wrappers dissolve so children can be interleaved with `order`. */
.detail__body { display: flex; flex-direction: column; gap: 1.5rem; max-width: 1200px; margin: 0 auto; padding: 1.5rem 1rem 0; }
.detail__aside, .detail__main { display: contents; }
.detail__meter { order: 1; }
.detail__activity { order: 2; }
.detail__overview { order: 3; }
.detail__providers { order: 4; }
.detail__seasons { order: 5; }
.detail__cast { order: 6; }
.detail__videos { order: 7; }
.detail__reviews { order: 8; scroll-margin-top: 5rem; }
.detail__distribution { order: 9; }
.detail__facts { order: 10; }
.detail__similar { order: 11; margin-inline: -1rem; }

@media (min-width: 640px) {
  .detail__body { padding-inline: 1.5rem; }
  .detail__similar { margin-inline: -1.5rem; }
}
@media (min-width: 768px) {
  .detail { padding-bottom: 3rem; }
  /* Tablet: two-column grid; paired cards sit side by side, everything else spans the row. */
  .detail__body { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1.5rem; }
  :where(.detail__aside, .detail__main) > * { grid-column: 1 / -1; }
  .detail__meter, .detail__activity, .detail__distribution, .detail__facts { grid-column: auto; }
  .detail__meter:not(:has(~ .detail__activity)) { grid-column: 1 / -1; }
}
@media (min-width: 1024px) {
  .detail__body { display: grid; grid-template-columns: minmax(0, 1fr) 320px; align-items: start; gap: 2rem 2.5rem; }
  .detail__main { display: flex; flex-direction: column; grid-column: 1; grid-row: 1; gap: 2rem; min-width: 0; }
  .detail__aside { display: flex; flex-direction: column; grid-column: 2; grid-row: 1; gap: 1rem; position: sticky; top: 5rem; max-height: calc(100vh - 6rem); max-height: calc(100dvh - 6rem); overflow-y: auto; scrollbar-width: thin; }
  .detail__similar { margin-inline: 0; }
}
</style>
