<template>
  <div class="ur">
    <SkeletonList v-if="!loaded" :count="4" />

    <AppErrorState v-else-if="loadError" :title="t('profile.states.sectionError')" @retry="fetchReviews" />

    <TabEmptyState
      v-else-if="reviews.length === 0 && !isFiltered"
      :icon="Star"
      :is-owner="isOwner"
      :owner-title="t('profile.tabContent.reviews.emptyOwnerTitle')"
      :owner-body="t('profile.tabContent.reviews.emptyOwnerBody')"
      :visitor-title="t('profile.tabContent.reviews.emptyVisitorTitle')"
      :visitor-body="t('profile.tabContent.reviews.emptyVisitorBody')"
      :cta-label="t('profile.tabContent.reviews.browse')"
    />

    <template v-else>
      <ul class="ur__summary" :aria-label="isOwner ? t('profile.reviews.titleOwn') : t('profile.reviews.titleOther')">
        <li class="ur__stat">
          <span class="ur__stat-num">{{ summary.count }}</span>
          <span class="ur__stat-lbl">{{ t("profile.tabs.reviews") }}</span>
        </li>
        <li class="ur__stat">
          <span class="ur__stat-num ur__stat-num--rating"><Star :size="16" fill="currentColor" aria-hidden="true" />{{ summary.average.toFixed(1) }}</span>
          <span class="ur__stat-lbl">{{ t("profile.tabContent.reviews.average") }}</span>
        </li>
        <li v-if="summary.topGenre" class="ur__stat">
          <span class="ur__stat-num ur__stat-num--text">{{ summary.topGenre }}</span>
          <span class="ur__stat-lbl">{{ t("profile.tabContent.reviews.topGenre") }}</span>
        </li>
      </ul>

      <TabToolbar
        v-model:query="query"
        v-model:sort="sort"
        v-model:filter="filter"
        :count-label="t('profile.tabContent.reviews.count', { n: shownCount })"
        :search-placeholder="t('profile.reviews.searchPlaceholder')"
        :filters="filters"
        :sorts="sorts"
      >
        <template #extra>
          <label v-if="isOwner" class="ur__field">
            <span class="ur__sr">{{ t("profile.reviews.filterVisibility") }}</span>
            <select v-model="visibilityFilter" class="ur__select" :aria-label="t('profile.reviews.filterVisibility')">
              <option value="all">{{ t("profile.reviews.visibility.all") }}</option>
              <option value="public">{{ t("profile.reviews.visibility.public") }}</option>
              <option value="private">{{ t("profile.reviews.visibility.private") }}</option>
            </select>
          </label>
          <MonthPicker v-model="monthFilter" :min-year="joinYear" />
        </template>
      </TabToolbar>

      <TabEmptyState v-if="reviews.length === 0 && !refreshing" filtered :icon="Star" :is-owner="isOwner" @clear="clearAll" />
      <template v-else>
        <ul class="ur__list" :aria-label="t('profile.tabs.reviews')" :aria-busy="refreshing">
          <li v-for="(review, i) in reviews" :key="review.id" class="ur__item" :style="{ '--i': Math.min(i, 10) }">
            <ReviewTabCard :review="review" :is-owner="isOwner" @edit="editTarget = $event" @delete="deleteTarget = $event" />
          </li>
        </ul>
        <div v-if="hasMore" class="ur__more">
          <AppButton variant="outline" :loading="loadingMore" @click="loadMoreReviews">{{ t("profile.tabContent.common.loadMore") }}</AppButton>
        </div>
      </template>
    </template>

    <EditReview v-if="editTarget" :userId="userId" :review="editTarget" @close="editTarget = null" @saved="onReviewSaved" />
    <DeleteReview v-if="deleteTarget" :userId="userId" :review="deleteTarget" @close="deleteTarget = null" @deleted="onReviewDeleted" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Star } from "lucide-vue-next"
import { reviewApi } from "@/api/api"
import type { GetUserReviewsParams, ReviewRatingFilterParam, ReviewSortParam, ReviewVisibilityFilter } from "@/api/endpoints/social"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useAuthStore } from "@/stores/auth"
import type { ReviewResponse, UserReviewsSummary } from "@/types"
import { logApiError } from "@/utils/logApiError"
import DeleteReview from "./components/DeleteReview.vue"
import EditReview from "./components/EditReview.vue"
import MonthPicker from "./tabs/MonthPicker.vue"
import ReviewTabCard from "./tabs/ReviewTabCard.vue"
import TabEmptyState from "./tabs/TabEmptyState.vue"
import TabToolbar from "./tabs/TabToolbar.vue"
import { REVIEW_SEARCH_DEBOUNCE_MS, REVIEW_SORT_KEYS, summarizeReviews } from "./tabs/reviewList"

const props = defineProps<{ userId: number; createdAt?: string | null }>()

const { t } = useI18n()
const auth = useAuthStore()
const isOwner = computed(() => auth.user?.id === props.userId)
const joinYear = computed(() => {
  const raw = props.createdAt ?? (isOwner.value ? auth.user?.created_at : undefined)
  if (!raw) return undefined
  const y = new Date(raw).getFullYear()
  return Number.isFinite(y) ? y : undefined
})

const DEFAULT_SORT: ReviewSortParam = "newest"
const DEFAULT_FILTER: ReviewRatingFilterParam = "all"
const DEFAULT_VISIBILITY: ReviewVisibilityFilter = "all"

// Every control below is a request param: a change resets to page 1 and refetches from the server.
const sort = ref<ReviewSortParam>(DEFAULT_SORT)
const filter = ref<ReviewRatingFilterParam>(DEFAULT_FILTER)
const visibilityFilter = ref<ReviewVisibilityFilter>(DEFAULT_VISIBILITY)
const monthFilter = ref("") // "YYYY-MM"
const query = ref("") // raw input; `search` is its debounced, trimmed value
const search = ref("")
const isFiltered = computed(
  () => filter.value !== DEFAULT_FILTER || visibilityFilter.value !== DEFAULT_VISIBILITY || !!monthFilter.value || !!search.value,
)

const loaded = ref(false) // first response (or failure) arrived; drives the skeleton
const refreshing = ref(false) // refetching page 1 while the previous list stays visible
const loadingMore = ref(false)
const loadError = ref(false)
const reviews = ref<ReviewResponse[]>([])
const page = ref(1)
const hasMore = ref(false)
const total = ref(0)
const serverSummary = ref<UserReviewsSummary | null>(null)
const editTarget = ref<ReviewResponse | null>(null)
const deleteTarget = ref<ReviewResponse | null>(null)

const shownCount = computed(() => total.value)
const summary = computed(() => ({
  // Without a server summary (older API) fall back to the loaded list so the cards never contradict it.
  count: serverSummary.value?.count ?? total.value,
  average: serverSummary.value?.average_rating ?? summarizeReviews(reviews.value).average,
  // Genres come from the loaded page(s), so they are only representative for the unfiltered list.
  topGenre: filter.value === DEFAULT_FILTER && !search.value ? summarizeReviews(reviews.value).topGenre : null,
}))

const filters = computed(() => {
  const counts = serverSummary.value?.filter_counts
  return [
    { key: "all", label: t("profile.reviews.filters.all"), count: counts?.all },
    { key: "high", label: t("profile.reviews.filters.high"), count: counts?.high },
    { key: "low", label: t("profile.reviews.filters.low"), count: counts?.low },
    { key: "withText", label: t("profile.reviews.filters.withText"), count: counts?.withText },
  ]
})
const sorts = computed(() => REVIEW_SORT_KEYS.map((key) => ({ key, label: t(`profile.tabContent.reviews.sorts.${key}`) })))

function buildParams(pageNo: number): GetUserReviewsParams {
  const params: GetUserReviewsParams = {}
  if (sort.value !== DEFAULT_SORT) params.sort = sort.value
  if (filter.value !== DEFAULT_FILTER) params.filter = filter.value
  if (search.value) params.search = search.value
  if (visibilityFilter.value !== DEFAULT_VISIBILITY) params.visibility = visibilityFilter.value
  if (monthFilter.value) {
    const [year, month] = monthFilter.value.split("-").map(Number) as [number, number]
    params.year = year
    params.month = month
  }
  if (pageNo > 1) params.page = pageNo
  return params
}

// Each request takes a sequence id; a response is applied only if no newer request started meanwhile.
let requestSeq = 0

async function fetchReviews() {
  const seq = ++requestSeq
  refreshing.value = true
  loadingMore.value = false
  loadError.value = false
  try {
    const res = await reviewApi.getUserReviews(props.userId, buildParams(1))
    if (seq !== requestSeq) return
    reviews.value = res.data.reviews ?? []
    page.value = 1
    hasMore.value = Boolean(res.data.pagination?.has_more)
    total.value = res.data.pagination?.total ?? reviews.value.length
    serverSummary.value = res.data.summary ?? null
  } catch (err) {
    if (seq !== requestSeq) return
    logApiError("fetchReviews", err)
    loadError.value = true
  } finally {
    if (seq === requestSeq) {
      refreshing.value = false
      loaded.value = true
    }
  }
}

async function loadMoreReviews() {
  if (!hasMore.value || loadingMore.value || refreshing.value) return
  const seq = ++requestSeq
  loadingMore.value = true
  try {
    const nextPage = page.value + 1
    const res = await reviewApi.getUserReviews(props.userId, buildParams(nextPage))
    if (seq !== requestSeq) return
    reviews.value = [...reviews.value, ...(res.data.reviews ?? [])]
    page.value = nextPage
    hasMore.value = Boolean(res.data.pagination?.has_more)
  } catch (err) {
    if (seq !== requestSeq) return
    logApiError("loadMoreReviews", err)
  } finally {
    if (seq === requestSeq) loadingMore.value = false
  }
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(query, (value) => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => (search.value = value.trim()), REVIEW_SEARCH_DEBOUNCE_MS)
})
watch([sort, filter, visibilityFilter, monthFilter, search], () => void fetchReviews())
onBeforeUnmount(() => clearTimeout(searchTimer))

function clearAll() {
  clearTimeout(searchTimer)
  query.value = ""
  search.value = ""
  filter.value = DEFAULT_FILTER
  visibilityFilter.value = DEFAULT_VISIBILITY
  monthFilter.value = ""
}

// Edits and deletes change counts, average and order, so reload the list and summary.
function onReviewSaved() {
  editTarget.value = null
  void fetchReviews()
}

function onReviewDeleted() {
  deleteTarget.value = null
  void fetchReviews()
}

onMounted(fetchReviews)
</script>

<style scoped>
.ur__summary { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.5rem; margin: 0 0 1.25rem; padding: 0; list-style: none; }
.ur__stat { display: flex; flex-direction: column; gap: 0.125rem; min-width: 0; padding: 0.75rem 0.875rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.ur__stat-num { display: inline-flex; align-items: center; gap: 0.3rem; font-size: 1.25rem; font-weight: 700; line-height: 1.2; color: var(--color-text-primary); }
.ur__stat-num--rating { color: var(--color-rating); }
.ur__stat-num--text { overflow: hidden; font-size: var(--text-small); text-overflow: ellipsis; white-space: nowrap; }
.ur__stat-lbl { font-size: var(--text-caption); color: var(--color-text-muted); }
.ur__list { display: flex; flex-direction: column; gap: 0.75rem; margin: 0; padding: 0; list-style: none; }
.ur__item { animation: ur-in var(--motion-base) var(--ease-out-soft) calc(var(--i) * 40ms) both; }
@keyframes ur-in { from { opacity: 0; transform: translateY(8px); } }
.ur__select { min-height: 2.25rem; padding: 0 0.75rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); cursor: pointer; color-scheme: dark; }
.ur__sr { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
.ur__more { display: flex; justify-content: center; padding: 1.5rem 0 0.5rem; }
@media (prefers-reduced-motion: reduce) { .ur__item { animation: none; } }
</style>
