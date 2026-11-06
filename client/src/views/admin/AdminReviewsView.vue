<template>
  <div class="admin-reviews-view">
    <div class="view-header">
      <h1 class="page-title">{{ $t("admin.reviews.title") }}</h1>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <div class="search-input-wrapper">
        <Search :size="16" class="search-icon" />
        <input
          v-model="search"
          type="text"
          class="search-input"
          :placeholder="$t('admin.reviews.searchPlaceholder')"
          @input="onSearchInput"
        />
      </div>

      <div class="filter-controls">
        <select v-model="mediaTypeFilter" class="filter-select" @change="fetchData(1)">
          <option value="all">{{ $t("admin.reviews.typeAll") }}</option>
          <option value="movie">{{ $t("admin.reviews.typeMovie") }}</option>
          <option value="tv">{{ $t("admin.reviews.typeTv") }}</option>
        </select>

        <select v-model="visibilityFilter" class="filter-select" @change="fetchData(1)">
          <option value="all">{{ $t("admin.reviews.visibilityAll") }}</option>
          <option value="public">{{ $t("admin.reviews.visibilityPublic") }}</option>
          <option value="private">{{ $t("admin.reviews.visibilityPrivate") }}</option>
        </select>

        <select v-model="sortBy" class="filter-select" @change="fetchData(1)">
          <option value="created_at">Sort by Date</option>
          <option value="rating">Sort by Rating</option>
          <option value="like_count">Sort by Likes</option>
        </select>

        <button class="sort-order-btn" @click="toggleSortOrder">
          <ArrowUpDown :size="16" />
          <span>{{ sortOrder.toUpperCase() }}</span>
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="table-container">
      <table class="admin-table">
        <thead>
          <tr>
            <th>{{ $t("admin.reviews.colId") }}</th>
            <th>{{ $t("admin.reviews.colAuthor") }}</th>
            <th>{{ $t("admin.reviews.colType") }}</th>
            <th>{{ $t("admin.reviews.colRating") }}</th>
            <th>{{ $t("admin.reviews.colBody") }}</th>
            <th>{{ $t("admin.reviews.colLikes") }}</th>
            <th>{{ $t("admin.reviews.colCreated") }}</th>
            <th class="text-right">{{ $t("admin.reviews.colActions") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="adminStore.isLoadingReviews">
            <td colspan="8" class="text-center py-8">
              <div class="spinner inline" />
            </td>
          </tr>
          <tr v-else-if="reviews.length === 0">
            <td colspan="8" class="text-center py-8 text-muted">
              No reviews found.
            </td>
          </tr>
          <tr v-for="r in reviews" :key="r.id" class="review-row">
            <td>#{{ r.id }}</td>
            <td>
              <div class="author-cell">
                <span class="font-semibold">@{{ r.username }}</span>
                <span v-if="!r.is_public" class="visibility-badge private" title="Private Review">
                  <Lock :size="12" /> Private
                </span>
                <span v-else class="visibility-badge public" title="Public Review">
                  <Globe :size="12" /> Public
                </span>
              </div>
            </td>
            <td>
              <RouterLink
                :to="r.media_type === 'movie' ? '/movies/' + r.media_id : '/tv/' + r.media_id"
                class="media-info-link"
                target="_blank"
                :title="'View ' + (r.media_title || r.media_type) + ' detail page'"
              >
                <img
                  v-if="r.poster_url"
                  :src="r.poster_url"
                  class="media-poster-thumb"
                  :alt="r.media_title || 'poster'"
                  loading="lazy"
                  decoding="async"
                />
                <div v-else class="media-poster-thumb placeholder">
                  <i class="pi pi-film" />
                </div>
                <div class="media-details">
                  <span class="media-title-text">{{ r.media_title || (r.media_type === 'movie' ? 'Movie' : 'TV') + ' #' + r.media_id }}</span>
                  <span class="media-type-tag" :class="r.media_type">
                    {{ r.media_type === "movie" ? "Movie" : "TV Series" }}
                  </span>
                </div>
                <ExternalLink :size="12" class="ext-icon" />
              </RouterLink>
            </td>
            <td>
              <div class="rating-cell">
                <Star :size="14" fill="#f5c518" color="#f5c518" />
                <span>{{ r.rating.toFixed(1) }}</span>
              </div>
            </td>
            <td class="review-body-cell" :title="r.body" @click="openDetail(r)">
              <span class="body-preview">{{ r.body }}</span>
            </td>
            <td>{{ r.like_count }}</td>
            <td class="text-secondary">{{ formatDate(r.created_at) }}</td>
            <td class="text-right">
              <div class="action-buttons-group">
                <button
                  class="action-btn info"
                  title="View Full Details"
                  @click="openDetail(r)"
                >
                  <Eye :size="14" />
                </button>
                <button
                  class="action-btn danger"
                  title="Delete Review"
                  @click="handleDelete(r)"
                >
                  <Trash2 :size="14" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Bar -->
    <div v-if="adminStore.reviewsTotalPages > 1" class="pagination-bar">
      <span class="pagination-info">
        {{ $t("admin.pagination.showing", {
          from: (adminStore.reviewsPage - 1) * adminStore.reviewsLimit + 1,
          to: Math.min(adminStore.reviewsPage * adminStore.reviewsLimit, adminStore.reviewsTotal),
          total: adminStore.reviewsTotal
        }) }}
      </span>

      <div class="pagination-buttons">
        <button
          class="page-btn"
          :disabled="adminStore.reviewsPage <= 1"
          @click="fetchData(adminStore.reviewsPage - 1)"
        >
          {{ $t("admin.pagination.previous") }}
        </button>
        <span class="page-num">{{ adminStore.reviewsPage }} / {{ adminStore.reviewsTotalPages }}</span>
        <button
          class="page-btn"
          :disabled="adminStore.reviewsPage >= adminStore.reviewsTotalPages"
          @click="fetchData(adminStore.reviewsPage + 1)"
        >
          {{ $t("admin.pagination.next") }}
        </button>
      </div>
    </div>

    <!-- Review Detail Modal -->
    <Teleport to="body">
      <div v-if="selectedReview" class="modal-overlay" @click.self="selectedReview = null">
        <div class="review-detail-modal" role="dialog" aria-modal="true">
          <div class="detail-modal-header">
            <h3><i class="pi pi-comments" /> รายละเอียดรีวิว #{{ selectedReview.id }}</h3>
            <button class="btn-close-modal" aria-label="Close detail modal" @click="selectedReview = null">
              <X :size="18" />
            </button>
          </div>

          <div class="detail-modal-content">
            <!-- Media Banner / Info -->
            <div class="media-banner-card">
              <img
                v-if="selectedReview.poster_url"
                :src="selectedReview.poster_url"
                class="modal-poster"
                :alt="selectedReview.media_title || 'Poster'"
                loading="lazy"
                decoding="async"
              />
              <div v-else class="modal-poster placeholder">
                <i class="pi pi-film" style="font-size: 2.5rem" />
              </div>
              <div class="modal-media-info">
                <span class="modal-media-type" :class="selectedReview.media_type">
                  {{ selectedReview.media_type === 'movie' ? 'ภาพยนตร์ (Movie)' : 'ซีรีส์ (TV Series)' }}
                </span>
                <h2 class="modal-media-title">
                  {{ selectedReview.media_title || ('ID #' + selectedReview.media_id) }}
                </h2>
                <span class="modal-media-id">TMDB ID: #{{ selectedReview.media_id }}</span>
                <RouterLink
                  :to="selectedReview.media_type === 'movie' ? '/movies/' + selectedReview.media_id : '/tv/' + selectedReview.media_id"
                  class="btn-view-media"
                  target="_blank"
                >
                  <span>ไปที่หน้ารายละเอียด</span>
                  <ExternalLink :size="14" />
                </RouterLink>
              </div>
            </div>

            <!-- Meta Grid -->
            <div class="meta-grid">
              <div class="meta-item">
                <span class="meta-label">ผู้เขียน (Author)</span>
                <span class="meta-value author">@{{ selectedReview.username }}</span>
              </div>
              <div class="meta-item">
                <span class="meta-label">คะแนนรีวิว (Rating)</span>
                <div class="meta-value rating">
                  <Star :size="16" fill="#f5c518" color="#f5c518" />
                  <span>{{ selectedReview.rating.toFixed(1) }} / 5.0</span>
                </div>
              </div>
              <div class="meta-item">
                <span class="meta-label">สิทธิ์การมองเห็น (Visibility)</span>
                <span class="meta-value">
                  <span v-if="!selectedReview.is_public" class="visibility-badge private">
                    <Lock :size="12" /> รีวิวส่วนตัว (Private)
                  </span>
                  <span v-else class="visibility-badge public">
                    <Globe :size="12" /> รีวิวสาธารณะ (Public)
                  </span>
                </span>
              </div>
              <div class="meta-item">
                <span class="meta-label">จำนวนถูกใจ (Likes)</span>
                <span class="meta-value">{{ selectedReview.like_count }} คน</span>
              </div>
              <div class="meta-item">
                <span class="meta-label">วันที่เขียน (Date)</span>
                <span class="meta-value">{{ formatDate(selectedReview.created_at) }}</span>
              </div>
            </div>

            <!-- Full Review Body -->
            <div class="full-review-body">
              <span class="body-label">ข้อความรีวิวทั้งหมด:</span>
              <div class="review-text-box">
                {{ selectedReview.body }}
              </div>
            </div>
          </div>

          <div class="detail-modal-footer">
            <button class="btn-modal-delete" @click="handleDelete(selectedReview)">
              <Trash2 :size="16" />
              <span>ลบรีวิวนี้</span>
            </button>
            <button class="btn-modal-close" @click="selectedReview = null">
              ปิดหน้าต่าง
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Official Admin Confirm Modal for Delete -->
    <AdminConfirmModal
      v-model="deleteModal.show"
      title="ยืนยันการลบรีวิว"
      description="คุณแน่ใจหรือไม่ที่จะลบรีวิวนี้? ระบบจะทำการลบแบบ Soft Delete"
      :target-name="deleteModal.review ? `รีวิว #${deleteModal.review.id} (@${deleteModal.review.username})` : ''"
      variant="danger"
      confirm-text="ลบรีวิว"
      :loading="deleteModal.loading"
      @confirm="onConfirmDelete"
      @cancel="deleteModal.show = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue"
import { RouterLink } from "vue-router"
import { useAdminStore } from "@/stores/admin"
import { useToast } from "@/composables/useToast"
import type { AdminReviewRow } from "@/types"
import { Search, ArrowUpDown, Star, Trash2, Eye, ExternalLink, Lock, Globe, X } from "lucide-vue-next"
import AdminConfirmModal from "@/components/admin/AdminConfirmModal.vue"

const adminStore = useAdminStore()
const toast = useToast()
const reviews = computed(() => adminStore.reviews)

const selectedReview = ref<AdminReviewRow | null>(null)

interface DeleteModalState {
  show: boolean
  review: AdminReviewRow | null
  loading: boolean
}

const deleteModal = ref<DeleteModalState>({
  show: false,
  review: null,
  loading: false,
})

function openDetail(review: AdminReviewRow) {
  selectedReview.value = review
}

const search = ref("")
const mediaTypeFilter = ref("all")
const visibilityFilter = ref("all")
const sortBy = ref("created_at")
const sortOrder = ref<"asc" | "desc">("desc")

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => fetchData(1), 300)
}

function toggleSortOrder() {
  sortOrder.value = sortOrder.value === "asc" ? "desc" : "asc"
  fetchData(1)
}

function fetchData(page = 1) {
  adminStore.fetchReviews({
    page,
    limit: 20,
    search: search.value.trim(),
    media_type: mediaTypeFilter.value,
    visibility: visibilityFilter.value,
    sort_by: sortBy.value,
    sort_order: sortOrder.value,
  })
}

function handleDelete(review: AdminReviewRow) {
  deleteModal.value = {
    show: true,
    review,
    loading: false,
  }
}

async function onConfirmDelete(payload: { reason: string }) {
  if (!deleteModal.value.review) return
  deleteModal.value.loading = true
  const review = deleteModal.value.review
  const reason = payload.reason || undefined

  try {
    await adminStore.deleteReview(review.id, reason)
    toast.success(`ลบรีวิว #${review.id} เรียบร้อยแล้ว`)
    deleteModal.value.show = false
    if (selectedReview.value?.id === review.id) {
      selectedReview.value = null
    }
    fetchData(adminStore.reviewsPage)
  } catch (err: unknown) {
    const e = err as { response?: { data?: { error?: string } } }
    toast.error(e.response?.data?.error || "เกิดข้อผิดพลาดในการลบรีวิว")
  } finally {
    deleteModal.value.loading = false
  }
}

function formatDate(isoStr: string) {
  if (!isoStr) return "-"
  return new Date(isoStr).toLocaleDateString()
}

onMounted(() => {
  fetchData(1)
})
</script>

<style scoped>
.admin-reviews-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-title {
  font-size: 1.6rem;
  font-weight: 800;
  margin: 0;
  color: var(--color-text-primary);
}

.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  justify-content: space-between;
  align-items: center;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  left: 0.85rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-secondary);
}

.search-input {
  width: 100%;
  padding: 0.6rem 0.85rem 0.6rem 2.4rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  outline: none;
  font-size: 0.875rem;
  box-sizing: border-box;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.filter-select {
  padding: 0.6rem 0.85rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-size: 0.85rem;
  outline: none;
  cursor: pointer;
}

.sort-order-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.6rem 0.85rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-size: 0.85rem;
  cursor: pointer;
}

.table-container {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  overflow-x: auto;
  contain: layout style;
}

.admin-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.875rem;
}

.admin-table th {
  padding: 0.85rem 1rem;
  background: rgba(255, 255, 255, 0.02);
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  font-weight: 600;
}

.admin-table td {
  padding: 0.85rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  color: var(--color-text-primary);
}

.font-semibold { font-weight: 600; }

.review-row {
  transition: background-color 0.15s ease;
  will-change: background-color;
}

.review-row:hover {
  background-color: rgba(255, 255, 255, 0.02);
}

.author-cell {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.visibility-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
  width: fit-content;
}

.visibility-badge.public {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.visibility-badge.private {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.media-info-link {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  text-decoration: none;
  color: var(--color-text-primary);
  padding: 0.25rem 0.4rem;
  border-radius: 0.35rem;
  transition: background-color 0.15s ease;
}

.media-info-link:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.media-info-link:hover .media-title-text {
  color: var(--color-brand);
  text-decoration: underline;
}

.media-poster-thumb {
  width: 32px;
  height: 48px;
  object-fit: cover;
  border-radius: 0.25rem;
  flex-shrink: 0;
}

.media-poster-thumb.placeholder {
  background: var(--color-surface-3);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
}

.media-details {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.media-title-text {
  font-weight: 600;
  font-size: 0.875rem;
  max-width: 180px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.15s ease;
}

.media-type-tag {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 0.25rem;
  background: var(--color-surface-3);
  color: var(--color-text-secondary);
  width: fit-content;
}

.media-type-tag.movie {
  background: rgba(245, 67, 44, 0.15);
  color: #ff4d4d;
}

.media-type-tag.tv {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.ext-icon {
  color: var(--color-text-secondary);
  opacity: 0;
  transition: opacity 0.15s ease;
}

.media-info-link:hover .ext-icon {
  opacity: 1;
}

.rating-cell {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-weight: 700;
  color: #f5c518;
}

.review-body-cell {
  max-width: 300px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
}

.review-body-cell:hover {
  color: #ffffff;
}

.body-preview {
  opacity: 0.9;
}

.action-buttons-group {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.4rem;
}

.action-btn {
  padding: 0.4rem;
  background: var(--color-surface-3);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  border-radius: 0.35rem;
  cursor: pointer;
  transition: color 0.15s ease, border-color 0.15s ease, background-color 0.15s ease, transform 0.15s ease;
  transform: translateZ(0);
}

.action-btn.info:hover {
  color: #3b82f6;
  border-color: #3b82f6;
  background-color: rgba(59, 130, 246, 0.1);
  transform: translateY(-1px) translateZ(0);
}

.action-btn.danger:hover {
  color: #ef4444;
  border-color: #ef4444;
  background-color: rgba(239, 68, 68, 0.1);
  transform: translateY(-1px) translateZ(0);
}

.text-right { text-align: right; }
.text-center { text-align: center; }
.text-secondary { color: var(--color-text-secondary); }
.text-muted { color: var(--color-text-muted); }
.py-8 { padding-top: 2rem; padding-bottom: 2rem; }

.pagination-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.pagination-buttons {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.page-btn {
  padding: 0.4rem 0.85rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  border-radius: 0.35rem;
  cursor: pointer;
  transition: background-color 0.15s ease;
}
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* ── Review Detail Modal ── */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  will-change: opacity;
  transform: translateZ(0);
}

.review-detail-modal {
  background: #141414;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.7);
  overflow: hidden;
  color: #ffffff;
  will-change: transform, opacity;
  transform: translateZ(0);
  contain: layout style;
}

.detail-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.2rem 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.detail-modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 0.5rem;

  i {
    color: var(--color-brand);
  }
}

.btn-close-modal {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  padding: 0.2rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s ease;
}

.btn-close-modal:hover {
  color: #ffffff;
}

.detail-modal-content {
  padding: 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.media-banner-card {
  display: flex;
  gap: 1rem;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0.85rem;
  align-items: center;
}

.modal-poster {
  width: 72px;
  height: 108px;
  object-fit: cover;
  border-radius: 6px;
  flex-shrink: 0;
}

.modal-poster.placeholder {
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.3);
}

.modal-media-info {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  flex: 1;
}

.modal-media-type {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  width: fit-content;
}

.modal-media-type.movie {
  background: rgba(245, 67, 44, 0.2);
  color: #ff4d4d;
}

.modal-media-type.tv {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}

.modal-media-title {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 800;
  color: #ffffff;
}

.modal-media-id {
  font-size: 0.8rem;
  color: rgba(255, 255, 255, 0.5);
}

.btn-view-media {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  margin-top: 0.2rem;
  font-size: 0.825rem;
  font-weight: 600;
  color: #3b82f6;
  text-decoration: none;
  transition: color 0.15s ease;
}

.btn-view-media:hover {
  color: #60a5fa;
  text-decoration: underline;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.85rem;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  padding: 1rem;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.meta-label {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.5);
}

.meta-value {
  font-size: 0.9rem;
  font-weight: 600;
}

.meta-value.author {
  color: #ffffff;
}

.meta-value.rating {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  color: #f5c518;
}

.full-review-body {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.body-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.8);
}

.review-text-box {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 1rem;
  font-size: 0.9375rem;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.95);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 240px;
  overflow-y: auto;
}

.detail-modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.2);
}

.btn-modal-delete {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1rem;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #ef4444;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s ease, border-color 0.15s ease, transform 0.15s ease;
  transform: translateZ(0);
}

.btn-modal-delete:hover {
  background-color: rgba(239, 68, 68, 0.25);
  border-color: #ef4444;
  transform: translateY(-1px) translateZ(0);
}

.btn-modal-close {
  padding: 0.55rem 1.2rem;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #ffffff;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.btn-modal-close:hover {
  background-color: rgba(255, 255, 255, 0.15);
}
</style>
