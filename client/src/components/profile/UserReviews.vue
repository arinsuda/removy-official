<template>
  <div class="reviews-root">
    <div class="section-head">
      <h2 class="eyebrow">{{ isOwner ? t("profile.reviews.titleOwn") : t("profile.reviews.titleOther") }}</h2>
      <span class="count-chip">{{ reviews.length }}</span>
      <div class="rule" />
    </div>

    <!-- Unified filter toolbar -->
    <div class="filter-toolbar">
      <AppTabs v-model="activeFilter" :tabs="filters" :controls="false" />

      <div class="toolbar-right">
        <!-- Visibility: compact icon-only segmented control -->
        <div class="seg-control" role="tablist" :aria-label="t('profile.reviews.filterVisibility')">
          <button
            v-for="v in visibilityOptions"
            :key="v.key"
            class="seg-btn"
            :class="{ 'seg-btn--active': visibilityFilter === v.key }"
            :aria-selected="visibilityFilter === v.key"
            :aria-label="v.label"
            :title="v.label"
            role="tab"
            @click="onVisibilityChange(v.key)"
          >
            <component :is="v.icon" :size="13" :stroke-width="2" />
          </button>
        </div>

        <!-- Date: icon trigger + popover -->
        <div class="date-popover-wrap" @click.stop>
          <button
            class="date-trigger"
            :class="{ 'date-trigger--active': hasDateFilter }"
            @click="isDatePopoverOpen = !isDatePopoverOpen"
            :aria-label="t('profile.reviews.filterDate')"
          >
            <Calendar :size="13" :stroke-width="2" />
            <span v-if="hasDateFilter" class="date-trigger-label">{{
              dateSummaryLabel
            }}</span>
            <button
              v-if="hasDateFilter"
              class="date-trigger-clear"
              type="button"
              :aria-label="t('profile.reviews.clearDate')"
              @click.stop="clearDateFilters"
            >
              <X :size="10" />
            </button>
          </button>

          <Transition name="dropdown">
            <div v-if="isDatePopoverOpen" class="date-popover">
              <span class="date-popover-label">{{ t("profile.reviews.dateRange") }}</span>
              <div class="date-filters">
                <input
                  v-model="dateFrom"
                  type="date"
                  class="date-input"
                  :max="dateTo || undefined"
                  :aria-label="t('profile.reviews.from')"
                  @change="onDateRangeChange"
                />
                <span class="date-sep">–</span>
                <input
                  v-model="dateTo"
                  type="date"
                  class="date-input"
                  :min="dateFrom || undefined"
                  :aria-label="t('profile.reviews.to')"
                  @change="onDateRangeChange"
                />
              </div>
              <span class="date-popover-label">{{ t("profile.reviews.orMonth") }}</span>
              <input
                v-model="monthFilter"
                type="month"
                class="date-input date-input--full"
                :aria-label="t('profile.reviews.month')"
                @change="onMonthFilterChange"
              />
            </div>
          </Transition>
        </div>
      </div>
    </div>

    <!-- States -->
    <SkeletonList v-if="loading" :count="4" />

    <AppErrorState v-else-if="loadError" :title="$t('profile.states.sectionError')" @retry="fetchReviews" />

    <div v-else-if="filteredReviews.length === 0" class="state-empty">
      <Star :size="28" :stroke-width="1.2" />
      <p>{{ t("profile.reviews.empty") }}</p>
    </div>

    <!-- List -->
    <div v-else class="reviews-list">
      <article
        v-for="(review, i) in displayReviews"
        :key="review.id"
        class="review-card"
        :style="{ '--i': i }"
      >
        <!-- Poster -->
        <div
          class="poster"
          @click="goToDetail(review.mediaId, review.mediaType)"
        >
          <img
            v-if="review.posterUrl"
            :src="review.posterUrl"
            :alt="review.targetName"
            class="poster-img"
          />
          <div v-else class="poster-fallback">
            <Film :size="16" :stroke-width="1.4" />
          </div>
        </div>

        <!-- Body -->
        <div class="review-body">
          <div class="title-rating-row">
            <h4
              class="review-title"
              @click="goToDetail(review.mediaId, review.mediaType)"
            >
              {{ review.targetName }}
            </h4>
            <div class="inline-rating">
              <span class="inline-rating-num">{{
                review.rating.toFixed(1)
              }}</span>
              <StarRatingInput
                :model-value="review.rating"
                readonly
                :show-label="false"
                :size="11"
              />
            </div>
          </div>

          <div class="meta-row">
            <time class="review-date">{{ review.createdAt }}</time>
            <component
              :is="review.visibility === 'private' ? Lock : review.visibility === 'followers' ? Users : Globe"
              class="private-icon"
              :size="12"
              :stroke-width="2.2"
              :title="t(`visibility.${review.visibility}`)"
            />
          </div>

          <p class="review-content">{{ review.content }}</p>

          <div v-if="review.tags?.length" class="tag-row">
            <span v-for="tag in review.tags" :key="tag" class="tag">{{
              tag
            }}</span>
          </div>

          <div class="stats-row">
            <span
              class="stat-item"
              :class="{ 'stat-item--active': review.isLiked }"
            >
              <Heart
                :size="11"
                :fill="review.isLiked ? 'currentColor' : 'none'"
              />
              {{ review.likeCount }}
            </span>
            <span
              class="stat-item"
              :class="{ 'stat-item--active': review.isHelpfulVoted }"
            >
              <ThumbsUp
                :size="11"
                :fill="review.isHelpfulVoted ? 'currentColor' : 'none'"
              />
              {{ t("profile.reviews.helpful", { count: review.helpfulCount }) }}
            </span>
          </div>
        </div>

        <!-- Actions Menu -->
        <div v-if="isOwner" class="menu-wrap" @click.stop>
          <button
            class="menu-trigger"
            :class="{ 'menu-trigger--open': openMenuId === review.id }"
            @click="toggleMenu(review.id)"
            :aria-label="t('profile.reviews.options', { title: review.targetName })"
          >
            <MoreHorizontal :size="14" />
          </button>
          <Transition name="dropdown">
            <div v-if="openMenuId === review.id" class="dropdown">
              <div v-if="review.isEdited" class="dropdown-info">
                <Pencil :size="10" :stroke-width="2" />
                <span>{{ t("profile.reviews.editedAt", { date: review.updatedLabel }) }}</span>
              </div>
              <button class="dropdown-item" @click="openEdit(review)">
                <Pencil :size="12" /> {{ t("profile.reviews.edit") }}
              </button>
              <button
                class="dropdown-item dropdown-item--del"
                @click="openDelete(review)"
              >
                <Trash2 :size="12" /> {{ t("profile.reviews.delete") }}
              </button>
            </div>
          </Transition>
        </div>
      </article>
    </div>

    <!-- Modals -->
    <EditReview
      v-if="editTarget"
      :userId="userId"
      :review="editTarget"
      @close="editTarget = null"
      @saved="onReviewSaved"
    />
    <DeleteReview
      v-if="deleteTarget"
      :userId="userId"
      :review="deleteTarget"
      @close="deleteTarget = null"
      @deleted="onReviewDeleted"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useAuthStore } from "@/stores/auth";
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import StarRatingInput from "@/components/movie/Starratinginput.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { computed, onMounted, onUnmounted, ref } from "vue";
import {
  Star,
  Film,
  Pencil,
  Trash2,
  MoreHorizontal,
  Heart,
  ThumbsUp,
  Calendar,
  Lock,
  Globe2,
  Globe,
  Users,
  Layers,
  X,
} from "lucide-vue-next";
import { reviewApi } from "@/api/api";
import type { GetUserReviewsParams } from "@/api/endpoints/social";
import { useRouter } from "vue-router";
import type { ReviewResponse } from "@/types";
import EditReview from "./components/EditReview.vue";
import DeleteReview from "./components/DeleteReview.vue";

const router = useRouter();

const props = defineProps<{ userId: number }>();

const { t, locale } = useI18n();
const auth = useAuthStore();
const isOwner = computed(() => auth.user?.id === props.userId);

const loading = ref(false);
const loadError = ref(false);

// ── Rating filter (client-side) ──────────────────────────────────
type FilterKey = "all" | "high" | "low";
const activeFilter = ref<FilterKey>("all");
const filters = computed<{ key: FilterKey; label: string }[]>(() => [
  { key: "all", label: t("profile.reviews.filters.all") },
  { key: "high", label: t("profile.reviews.filters.high") },
  { key: "low", label: t("profile.reviews.filters.low") },
]);

// ── Visibility filter (server-side, icon-only segmented control) ──
type VisibilityKey = "all" | "public" | "private";
const visibilityOptions = computed<{ key: VisibilityKey; label: string; icon: typeof Layers }[]>(() => [
  { key: "all", label: t("profile.reviews.visibility.all"), icon: Layers },
  { key: "public", label: t("profile.reviews.visibility.public"), icon: Globe2 },
  { key: "private", label: t("profile.reviews.visibility.private"), icon: Lock },
]);
const visibilityFilter = ref<VisibilityKey>("all");

function onVisibilityChange(key: VisibilityKey) {
  visibilityFilter.value = key;
  fetchReviews();
}

// ── Date filter (server-side, tucked into a popover) ──────────────
const dateFrom = ref<string>(""); // "YYYY-MM-DD"
const dateTo = ref<string>(""); // "YYYY-MM-DD"
const monthFilter = ref<string>(""); // "YYYY-MM" จาก <input type="month">
const isDatePopoverOpen = ref(false);

const hasDateFilter = computed(
  () => !!(dateFrom.value || dateTo.value || monthFilter.value),
);

const dateSummaryLabel = computed(() => {
  if (monthFilter.value) {
    const [year, month] = monthFilter.value.split("-").map(Number) as [
      number,
      number,
    ];
    return new Date(year, month - 1).toLocaleDateString(locale.value, {
      month: "short",
      year: "numeric",
    });
  }
  if (dateFrom.value && dateTo.value) return t("profile.reviews.dateRange");
  if (dateFrom.value) return t("profile.reviews.since");
  if (dateTo.value) return t("profile.reviews.until");
  return "";
});

function onMonthFilterChange() {
  // เลือกเดือน/ปีแล้ว ล้าง date range แบบ manual ทิ้ง กันตีกัน
  if (monthFilter.value) {
    dateFrom.value = "";
    dateTo.value = "";
  }
  fetchReviews();
}

function onDateRangeChange() {
  // แก้ date range เอง ล้าง month filter ทิ้ง
  if (dateFrom.value || dateTo.value) {
    monthFilter.value = "";
  }
  fetchReviews();
}

function clearDateFilters() {
  dateFrom.value = "";
  dateTo.value = "";
  monthFilter.value = "";
  isDatePopoverOpen.value = false;
  fetchReviews();
}

const reviews = ref<ReviewResponse[]>([]);
const openMenuId = ref<number | null>(null);
const editTarget = ref<ReviewResponse | null>(null);
const deleteTarget = ref<ReviewResponse | null>(null);

function goToDetail(mediaId: number, mediaType: string) {
  router.push({
    name: mediaType === "tv" ? "tv-detail" : "movie-detail",
    params: { id: mediaId },
  });
}

function toggleMenu(id: number) {
  openMenuId.value = openMenuId.value === id ? null : id;
}
function closeMenu() {
  openMenuId.value = null;
}
function closeDatePopover() {
  isDatePopoverOpen.value = false;
}
function openEdit(review: ReviewResponse) {
  closeMenu();
  editTarget.value = review;
}
function openDelete(review: ReviewResponse) {
  closeMenu();
  deleteTarget.value = review;
}
function onReviewSaved(updated: ReviewResponse) {
  const idx = reviews.value.findIndex((r) => r.id === updated.id);
  if (idx !== -1) reviews.value[idx] = updated;
  editTarget.value = null;
}
function onReviewDeleted(id: number) {
  reviews.value = reviews.value.filter((r) => r.id !== id);
  deleteTarget.value = null;
}

function handleOutsideClick() {
  closeMenu();
  closeDatePopover();
}

onMounted(() => document.addEventListener("click", handleOutsideClick));
onUnmounted(() => document.removeEventListener("click", handleOutsideClick));

const filteredReviews = computed(() => {
  if (activeFilter.value === "high")
    return reviews.value.filter((r) => r.rating >= 4);
  if (activeFilter.value === "low")
    return reviews.value.filter((r) => r.rating < 4);
  return reviews.value;
});

/**
 * ฟอร์แมตวันที่ดู (watched_at)
 * ถ้าเวลาไม่ใช่ 00:00 จะโชว์เวลาด้วย ถ้าเป็น 00:00 (ไม่ได้ระบุเวลา) จะโชว์แค่วัน
 * หมายเหตุ: ถ้าฝั่ง backend อยากแม่นยำ 100% ควรมี flag แยก เช่น
 * `watched_at_has_time: boolean` เพราะ user ที่ดูตอนเที่ยงคืนตรงเป๊ะจะถูกตัดเวลาออกไปด้วย
 * แต่สำหรับตอนนี้ heuristic นี้เพียงพอ
 */
function formatWatched(dateStr: string | null): string | null {
  if (!dateStr) return null;
  const d = new Date(dateStr);
  const hasTime = d.getHours() !== 0 || d.getMinutes() !== 0;
  return d.toLocaleString(locale.value, {
    year: "numeric",
    month: "short",
    day: "numeric",
    ...(hasTime ? { hour: "2-digit", minute: "2-digit" } : {}),
  });
}

function formatUpdatedAt(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleString(locale.value, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const displayReviews = computed(() =>
  filteredReviews.value.map((r) => ({
    ...r,
    targetName: r.media.title || `#${r.media.id}`,
    posterUrl: r.media.poster_url || null,
    mediaId: r.media.id,
    mediaType: r.media.media_type,
    createdAt: new Date(r.created_at).toLocaleDateString(locale.value, {
      year: "numeric",
      month: "short",
      day: "numeric",
    }),
    content: r.body,
    tags: [] as string[],
    likeCount: r.like_count ?? 0,
    helpfulCount: r.helpful_count ?? 0,
    isLiked: r.is_liked ?? false,
    isHelpfulVoted: r.is_helpful_voted ?? false,
    isEdited: r.updated_at !== r.created_at,
    updatedLabel: formatUpdatedAt(r.updated_at),
  })),
);

// ดึงรีวิวจาก backend ตาม filter ปัจจุบัน (visibility + date range/month)
async function fetchReviews() {
  try {
    loading.value = true;
    loadError.value = false;

    const params: GetUserReviewsParams = {};

    if (visibilityFilter.value !== "all") {
      params.visibility = visibilityFilter.value;
    }

    if (monthFilter.value) {
      const [year, month] = monthFilter.value.split("-").map(Number) as [
        number,
        number,
      ];
      params.year = year;
      params.month = month;
    } else {
      if (dateFrom.value) params.date_from = dateFrom.value;
      if (dateTo.value) params.date_to = dateTo.value;
    }

    const res = await reviewApi.getUserReviews(props.userId, params);
    reviews.value = res.data.reviews ?? [];
  } catch (err) {
    console.error("Fetch reviews failed:", err);
    loadError.value = true;
  } finally {
    loading.value = false;
  }
}

onMounted(fetchReviews);
</script>

<style scoped>
.reviews-root {
  --c-card: var(--color-surface-2);
  --c-border: var(--color-border);
  --c-border-h: var(--color-border-strong);
  --c-red: var(--color-brand);
  --c-text: var(--color-text-primary);
  --c-sub: var(--color-text-secondary);
  --c-muted: var(--color-text-muted);
  --c-star: var(--color-rating);
  --c-green: var(--color-success);
  --font: var(--font-sans);
  --ease: cubic-bezier(0.16, 1, 0.3, 1);
  font-family: var(--font-sans);
  color: var(--c-text);
}

.section-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
}
.eyebrow {
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
  white-space: nowrap;
}
.count-chip {
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--color-surface-3);
  color: var(--color-text-secondary);
  padding: 2px 10px;
  border-radius: 999px;
}
.rule {
  flex: 1;
  height: 1px;
  background: var(--c-border);
}

/* ── Unified filter toolbar ───────────────────────────────────── */
.filter-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 22px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--c-border);
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

/* Visibility: icon-only segmented control */
.seg-control {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--c-border);
  border-radius: 8px;
}
.seg-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: transparent;
  border: none;
  color: var(--c-sub);
  cursor: pointer;
  transition: all 0.15s var(--ease);
}
.seg-btn:hover {
  color: var(--c-text);
}
.seg-btn--active {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

/* Date filter trigger + popover */
.date-popover-wrap {
  position: relative;
}
.date-trigger {
  display: flex;
  align-items: center;
  gap: 5px;
  height: 30px;
  padding: 0 10px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--c-border);
  color: var(--c-sub);
  font-family: var(--font-sans);
  font-size: 0.7rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s var(--ease);
}
.date-trigger:hover {
  color: var(--c-text);
  border-color: var(--c-border-h);
}
.date-trigger--active {
  color: var(--c-text);
  background: rgba(225, 37, 27, 0.1);
  border-color: rgba(225, 37, 27, 0.3);
}
.date-trigger-label {
  white-space: nowrap;
}
.date-trigger-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
  border: none;
  color: inherit;
  cursor: pointer;
  margin-left: 2px;
}
.date-trigger-clear:hover {
  background: rgba(255, 69, 58, 0.25);
  color: #ff453a;
}

.date-popover {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 240px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: var(--color-surface-3);
  border: 1px solid var(--c-border-h);
  border-radius: 10px;
  padding: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  z-index: 50;
  transform-origin: top right;
}
.date-popover-label {
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--c-muted);
}
.date-filters {
  display: flex;
  align-items: center;
  gap: 6px;
}
.date-input {
  flex: 1;
  min-width: 0;
  font-family: var(--font-sans);
  font-size: 0.7rem;
  color: var(--c-text);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--c-border);
  border-radius: 6px;
  padding: 5px 6px;
  cursor: pointer;
  color-scheme: dark;
}
.date-input:hover,
.date-input:focus {
  border-color: var(--c-border-h);
  outline: none;
}
.date-input--full {
  width: 100%;
}
.date-sep {
  color: var(--c-muted);
  font-size: 0.7rem;
}

.state-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 56px 0;
  color: var(--c-muted);
}
.state-empty p {
  font-size: 0.82rem;
  margin: 0;
  color: var(--c-sub);
}

/* List */
.reviews-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.review-card {
  display: grid;
  grid-template-columns: 56px 1fr auto;
  gap: 0 16px;
  padding: 1.25rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  align-items: start;
  position: relative;
  animation: cardIn 0.4s var(--ease) calc(var(--i) * 50ms) both;
}

@keyframes cardIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.poster {
  width: 56px;
  height: 80px;
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
}
.poster-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.poster-fallback {
  color: var(--c-muted);
}

.review-body {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.title-rating-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.review-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: #fff;
  margin: 0;
  line-height: 1.3;
  cursor: pointer;
  transition: color 0.15s;
}
.review-title:hover {
  color: var(--c-red);
}

.inline-rating {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
}
.inline-rating-num {
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--c-star);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.02em;
}

/* Meta row: created date + private icon + watched date */
.meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.review-date {
  font-size: 0.68rem;
  color: var(--c-sub);
}
.private-icon {
  color: var(--c-sub);
  flex-shrink: 0;
}
.watched-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.65rem;
  font-weight: 500;
  color: var(--c-sub);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--c-border);
  padding: 1px 7px;
  border-radius: 20px;
}

.review-content {
  font-size: 0.8rem;
  color: rgba(240, 240, 240, 0.72);
  line-height: 1.58;
  margin: 2px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.tag-row {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
}
.tag {
  font-size: 0.6rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--c-sub);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--c-border);
  padding: 2px 7px;
  border-radius: 4px;
}

/* Stats row: likes + helpful */
.stats-row {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: 2px;
}
.stat-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.68rem;
  font-weight: 500;
  color: var(--c-sub);
}
.stat-item--active {
  color: var(--c-red);
}

/* Menu */
.menu-wrap {
  position: relative;
  align-self: start;
}
.menu-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--c-sub);
  cursor: pointer;
  transition: all 0.15s var(--ease);
}
.menu-trigger:hover,
.menu-trigger--open {
  color: var(--c-text);
  background: rgba(255, 255, 255, 0.06);
  border-color: var(--c-border-h);
}

.dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  min-width: 160px;
  background: var(--color-surface-3);
  border: 1px solid var(--c-border-h);
  border-radius: 8px;
  padding: 4px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  z-index: 50;
  transform-origin: top right;
}
.dropdown-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  margin-bottom: 4px;
  border-bottom: 1px solid var(--c-border);
  font-size: 0.65rem;
  font-weight: 500;
  color: var(--c-sub);
  white-space: nowrap;
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 7px 10px;
  border-radius: 5px;
  font-family: var(--font-sans);
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--c-sub);
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  transition: all 0.12s;
}
.dropdown-item:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.06);
}
.dropdown-item--del:hover {
  color: #ff453a;
  background: rgba(255, 69, 58, 0.08);
}

.dropdown-enter-active {
  transition:
    opacity 0.15s var(--ease),
    transform 0.15s var(--ease);
}
.dropdown-leave-active {
  transition:
    opacity 0.1s ease,
    transform 0.1s ease;
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.92);
}

@media (max-width: 520px) {
  .review-card {
    grid-template-columns: 56px 1fr auto;
  }
  .title-rating-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  .filter-toolbar {
    flex-wrap: wrap;
  }
  .date-popover {
    right: auto;
    left: 0;
  }
}
</style>
