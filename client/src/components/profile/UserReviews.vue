<template>
  <div class="ur">
    <SkeletonList v-if="loading" :count="4" />

    <AppErrorState v-else-if="loadError" :title="t('profile.states.sectionError')" @retry="fetchReviews" />

    <TabEmptyState
      v-else-if="reviews.length === 0 && !serverFiltered"
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
        v-model:query="list.query.value"
        v-model:sort="list.sort.value"
        v-model:filter="list.filter.value"
        :count-label="t('profile.tabContent.reviews.count', { n: list.filtered.value.length })"
        :filters="filters"
        :sorts="sorts"
      >
        <template #extra>
          <label v-if="isOwner" class="ur__field">
            <span class="ur__sr">{{ t("profile.reviews.filterVisibility") }}</span>
            <select v-model="visibilityFilter" class="ur__select" :aria-label="t('profile.reviews.filterVisibility')" @change="fetchReviews">
              <option value="all">{{ t("profile.reviews.visibility.all") }}</option>
              <option value="public">{{ t("profile.reviews.visibility.public") }}</option>
              <option value="private">{{ t("profile.reviews.visibility.private") }}</option>
            </select>
          </label>
          <MonthPicker v-model="monthFilter" @change="fetchReviews" />
        </template>
      </TabToolbar>

      <TabEmptyState v-if="list.filtered.value.length === 0" filtered :icon="Star" :is-owner="isOwner" @clear="clearAll" />
      <template v-else>
        <ul class="ur__list" :aria-label="t('profile.tabs.reviews')">
          <li v-for="(review, i) in list.visible.value" :key="review.id" class="ur__item" :style="{ '--i': Math.min(i, 10) }">
            <ReviewTabCard :review="review" :is-owner="isOwner" @edit="editTarget = $event" @delete="deleteTarget = $event" />
          </li>
        </ul>
        <div v-if="list.hasMore.value" class="ur__more">
          <AppButton variant="outline" @click="list.loadMore">{{ t("profile.tabContent.common.loadMore") }}</AppButton>
        </div>
      </template>
    </template>

    <EditReview v-if="editTarget" :userId="userId" :review="editTarget" @close="editTarget = null" @saved="onReviewSaved" />
    <DeleteReview v-if="deleteTarget" :userId="userId" :review="deleteTarget" @close="deleteTarget = null" @deleted="onReviewDeleted" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Star } from "lucide-vue-next"
import { reviewApi } from "@/api/api"
import type { GetUserReviewsParams } from "@/api/endpoints/social"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useAuthStore } from "@/stores/auth"
import type { ReviewResponse } from "@/types"
import { logApiError } from "@/utils/logApiError"
import DeleteReview from "./components/DeleteReview.vue"
import EditReview from "./components/EditReview.vue"
import MonthPicker from "./tabs/MonthPicker.vue"
import ReviewTabCard from "./tabs/ReviewTabCard.vue"
import TabEmptyState from "./tabs/TabEmptyState.vue"
import TabToolbar from "./tabs/TabToolbar.vue"
import { REVIEW_FILTERS, REVIEW_SORTS, REVIEW_SORT_KEYS, summarizeReviews } from "./tabs/reviewList"
import { useTabList } from "./tabs/useTabList"

const props = defineProps<{ userId: number }>()

const { t } = useI18n()
const auth = useAuthStore()
const isOwner = computed(() => auth.user?.id === props.userId)

const loading = ref(false)
const loadError = ref(false)
const reviews = ref<ReviewResponse[]>([])
const editTarget = ref<ReviewResponse | null>(null)
const deleteTarget = ref<ReviewResponse | null>(null)

// Visibility and month are server-side filters; rating, text, search and sort run in the browser.
type VisibilityKey = "all" | "public" | "private"
const visibilityFilter = ref<VisibilityKey>("all")
const monthFilter = ref("") // "YYYY-MM"
const serverFiltered = computed(() => visibilityFilter.value !== "all" || !!monthFilter.value)

const list = useTabList<ReviewResponse>({
  items: reviews,
  searchText: (r) => `${r.media.title} ${r.body}`,
  filters: REVIEW_FILTERS,
  sorts: REVIEW_SORTS,
  defaultSort: "newest",
  pageSize: 20,
})
const summary = computed(() => summarizeReviews(reviews.value))

const filters = computed(() => [
  { key: "all", label: t("profile.reviews.filters.all"), count: list.filterCounts.value.all },
  { key: "high", label: t("profile.reviews.filters.high"), count: list.filterCounts.value.high },
  { key: "low", label: t("profile.reviews.filters.low"), count: list.filterCounts.value.low },
  { key: "withText", label: t("profile.reviews.filters.withText"), count: list.filterCounts.value.withText },
])
const sorts = computed(() => REVIEW_SORT_KEYS.map((key) => ({ key, label: t(`profile.tabContent.reviews.sorts.${key}`) })))

function clearAll() {
  list.clear()
  if (serverFiltered.value) {
    visibilityFilter.value = "all"
    monthFilter.value = ""
    void fetchReviews()
  }
}

function onReviewSaved(updated: ReviewResponse) {
  const idx = reviews.value.findIndex((r) => r.id === updated.id)
  if (idx !== -1) reviews.value[idx] = updated
  editTarget.value = null
}

function onReviewDeleted(id: number) {
  reviews.value = reviews.value.filter((r) => r.id !== id)
  deleteTarget.value = null
}

async function fetchReviews() {
  try {
    loading.value = true
    loadError.value = false
    const params: GetUserReviewsParams = {}
    if (visibilityFilter.value !== "all") params.visibility = visibilityFilter.value
    if (monthFilter.value) {
      const [year, month] = monthFilter.value.split("-").map(Number) as [number, number]
      params.year = year
      params.month = month
    }
    const res = await reviewApi.getUserReviews(props.userId, params)
    reviews.value = res.data.reviews ?? []
  } catch (err) {
    logApiError("fetchReviews", err)
    loadError.value = true
  } finally {
    loading.value = false
  }
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
