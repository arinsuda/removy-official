<template>
  <AdminPage>
    <AdminPageHeader :title="t('admin.reviews.title')" />

    <AdminFilters>
      <AdminSearch v-model="search" :label="t('admin.table.search')" :placeholder="t('admin.reviews.searchPlaceholder')" />
      <AdminSelect v-model="mediaTypeFilter" :label="t('admin.reviews.colType')" :options="typeOptions" />
      <AdminSelect v-model="visibilityFilter" :label="t('admin.reviews.visibilityAll')" :options="visibilityOptions" />
    </AdminFilters>

    <AdminTable
      :columns="columns"
      :label="t('admin.reviews.title')"
      :loading="adminStore.isLoadingReviews"
      :error="loadError"
      :empty="reviews.length === 0"
      :filtered="isFiltered"
      :empty-title="t('admin.reviews.emptyTitle')"
      :empty-desc="t('admin.reviews.emptyDesc')"
      :sort-by="sortBy"
      :sort-order="sortOrder"
      @sort="onSort"
      @retry="fetchData(adminStore.reviewsPage)"
      @clear="clearFilters"
    >
      <tr v-for="r in reviews" :key="r.id">
        <td class="cell-id">#{{ r.id }}</td>
        <td>
          <div class="author-cell">
            <span class="author-name">@{{ r.username }}</span>
            <AdminBadge :tone="r.is_public ? 'info' : 'neutral'">
              <Globe v-if="r.is_public" :size="12" aria-hidden="true" />
              <Lock v-else :size="12" aria-hidden="true" />
              {{ r.is_public ? t("admin.reviews.public") : t("admin.reviews.private") }}
            </AdminBadge>
          </div>
        </td>
        <td>
          <RouterLink :to="mediaPath(r)" class="media-link" target="_blank">
            <img v-if="r.poster_url" :src="r.poster_url" class="media-thumb" alt="" loading="lazy" decoding="async" />
            <span v-else class="media-thumb media-thumb--empty" aria-hidden="true"><Film :size="16" /></span>
            <span class="media-text">
              <span class="media-title">{{ mediaTitle(r) }}</span>
              <span class="media-type">{{ typeLabel(r.media_type) }}</span>
            </span>
            <ExternalLink :size="12" class="media-ext" aria-hidden="true" />
          </RouterLink>
        </td>
        <td>
          <span class="rating"><Star :size="14" aria-hidden="true" />{{ r.rating.toFixed(1) }}</span>
        </td>
        <td class="body-cell">
          <button type="button" class="body-button" @click="selectedReview = r">{{ r.body }}</button>
        </td>
        <td class="cell-num">{{ r.like_count }}</td>
        <td class="cell-muted">{{ formatDate(r.created_at) }}</td>
        <td class="is-end">
          <AdminRowMenu
            :label="t('admin.table.rowActions', { name: `#${r.id}` })"
            :items="menuItems"
            @select="(key) => onMenu(r, key)"
          />
        </td>
      </tr>
    </AdminTable>

    <AdminPagination
      :page="adminStore.reviewsPage"
      :total-pages="adminStore.reviewsTotalPages"
      :total="adminStore.reviewsTotal"
      :limit="adminStore.reviewsLimit"
      @change="fetchData"
    />

    <AppModal
      :model-value="!!selectedReview"
      :title="selectedReview ? t('admin.reviews.detailTitle', { id: selectedReview.id }) : ''"
      size="md"
      @update:model-value="selectedReview = null"
    >
      <dl v-if="selectedReview" class="detail">
        <div class="detail__media">
          <img v-if="selectedReview.poster_url" :src="selectedReview.poster_url" class="detail__poster" alt="" decoding="async" />
          <div>
            <p class="detail__title">{{ mediaTitle(selectedReview) }}</p>
            <p class="detail__sub">{{ typeLabel(selectedReview.media_type) }}</p>
            <AppButton :to="mediaPath(selectedReview)" variant="outline" size="sm" target="_blank">
              {{ t("admin.reviews.openMedia") }}
              <ExternalLink :size="14" aria-hidden="true" />
            </AppButton>
          </div>
        </div>
        <div class="detail__grid">
          <div>
            <dt>{{ t("admin.reviews.author") }}</dt>
            <dd>@{{ selectedReview.username }}</dd>
          </div>
          <div>
            <dt>{{ t("admin.reviews.colRating") }}</dt>
            <dd class="rating"><Star :size="14" aria-hidden="true" />{{ t("admin.reviews.ratingOutOf", { rating: selectedReview.rating.toFixed(1) }) }}</dd>
          </div>
          <div>
            <dt>{{ t("admin.reviews.visibility") }}</dt>
            <dd>{{ selectedReview.is_public ? t("admin.reviews.public") : t("admin.reviews.private") }}</dd>
          </div>
          <div>
            <dt>{{ t("admin.reviews.likes") }}</dt>
            <dd>{{ selectedReview.like_count }}</dd>
          </div>
          <div>
            <dt>{{ t("admin.reviews.date") }}</dt>
            <dd>{{ formatDate(selectedReview.created_at) }}</dd>
          </div>
        </div>
        <div>
          <dt>{{ t("admin.reviews.body") }}</dt>
          <dd class="detail__body">{{ selectedReview.body }}</dd>
        </div>
      </dl>
      <template #footer>
        <div class="detail__actions">
          <AppButton variant="outline" @click="selectedReview = null">{{ t("admin.reviews.close") }}</AppButton>
          <AppButton v-if="selectedReview" variant="danger" @click="askDelete(selectedReview)">
            <Trash2 :size="16" aria-hidden="true" />
            {{ t("admin.reviews.actionDelete") }}
          </AppButton>
        </div>
      </template>
    </AppModal>

    <AdminConfirmModal
      v-model="deleteModal.show"
      :title="t('admin.reviewModal.title')"
      :description="t('admin.reviewModal.desc')"
      :target-name="deleteModal.review ? t('admin.reviewModal.target', { id: deleteModal.review.id, username: deleteModal.review.username }) : ''"
      variant="danger"
      :confirm-text="t('admin.reviewModal.confirm')"
      :reason-placeholder="t('admin.userModal.reasonNote')"
      :loading="deleteModal.loading"
      @confirm="onConfirmDelete"
      @cancel="deleteModal.show = false"
    />
  </AdminPage>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink } from "vue-router"
import { ExternalLink, Eye, Film, Globe, Lock, Star, Trash2 } from "lucide-vue-next"
import AdminBadge from "@/components/admin/AdminBadge.vue"
import AdminConfirmModal from "@/components/admin/AdminConfirmModal.vue"
import AdminFilters from "@/components/admin/AdminFilters.vue"
import AdminPage from "@/components/admin/AdminPage.vue"
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue"
import AdminPagination from "@/components/admin/AdminPagination.vue"
import AdminRowMenu, { type RowMenuItem } from "@/components/admin/AdminRowMenu.vue"
import AdminSearch from "@/components/admin/AdminSearch.vue"
import AdminSelect from "@/components/admin/AdminSelect.vue"
import AdminTable, { type AdminColumn } from "@/components/admin/AdminTable.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useToast } from "@/composables/useToast"
import { useAdminStore } from "@/stores/admin"
import type { AdminReviewRow } from "@/types"
import { adminErrorKey } from "@/utils/adminError"

const { t, locale } = useI18n()
const adminStore = useAdminStore()
const toast = useToast()
const reviews = computed(() => adminStore.reviews)

const search = ref("")
const mediaTypeFilter = ref("all")
const visibilityFilter = ref("all")
const sortBy = ref("created_at")
const sortOrder = ref<"asc" | "desc">("desc")
const loadError = ref(false)
const selectedReview = ref<AdminReviewRow | null>(null)
const deleteModal = ref<{ show: boolean; review: AdminReviewRow | null; loading: boolean }>({
  show: false,
  review: null,
  loading: false,
})

const columns = computed<AdminColumn[]>(() => [
  { key: "id", label: t("admin.reviews.colId") },
  { key: "author", label: t("admin.reviews.colAuthor") },
  { key: "media", label: t("admin.reviews.colType") },
  { key: "rating", label: t("admin.reviews.colRating"), sortable: true },
  { key: "body", label: t("admin.reviews.colBody") },
  { key: "like_count", label: t("admin.reviews.colLikes"), sortable: true },
  { key: "created_at", label: t("admin.reviews.colCreated"), sortable: true },
  { key: "actions", label: t("admin.reviews.colActions"), align: "end" },
])
const typeOptions = computed(() => [
  { value: "all", label: t("admin.reviews.typeAll") },
  { value: "movie", label: t("admin.reviews.typeMovie") },
  { value: "tv", label: t("admin.reviews.typeTv") },
])
const visibilityOptions = computed(() => [
  { value: "all", label: t("admin.reviews.visibilityAll") },
  { value: "public", label: t("admin.reviews.visibilityPublic") },
  { value: "private", label: t("admin.reviews.visibilityPrivate") },
])
const menuItems = computed<RowMenuItem[]>(() => [
  { key: "view", label: t("admin.reviews.view"), icon: Eye },
  { key: "delete", label: t("admin.reviews.actionDelete"), icon: Trash2, tone: "danger" },
])
const isFiltered = computed(() => !!search.value.trim() || mediaTypeFilter.value !== "all" || visibilityFilter.value !== "all")

const typeLabel = (type: string) => t(type === "movie" ? "admin.reviews.mediaMovie" : "admin.reviews.mediaTv")
const mediaPath = (r: AdminReviewRow) => (r.media_type === "movie" ? `/movies/${r.media_id}` : `/tv/${r.media_id}`)
const mediaTitle = (r: AdminReviewRow) => r.media_title || `${typeLabel(r.media_type)} #${r.media_id}`

async function fetchData(page = 1) {
  loadError.value = false
  try {
    await adminStore.fetchReviews({
      page,
      limit: 20,
      search: search.value.trim(),
      media_type: mediaTypeFilter.value,
      visibility: visibilityFilter.value,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    })
  } catch {
    loadError.value = true
  }
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch(search, () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => fetchData(1), 300)
})
watch([mediaTypeFilter, visibilityFilter], () => fetchData(1))

function onSort(key: string) {
  if (sortBy.value === key) sortOrder.value = sortOrder.value === "asc" ? "desc" : "asc"
  else {
    sortBy.value = key
    sortOrder.value = "desc"
  }
  fetchData(1)
}

function clearFilters() {
  search.value = ""
  mediaTypeFilter.value = "all"
  visibilityFilter.value = "all"
}

function onMenu(review: AdminReviewRow, key: string) {
  if (key === "view") selectedReview.value = review
  else askDelete(review)
}

// Deleting always passes through the confirm modal; nothing here calls the API directly.
function askDelete(review: AdminReviewRow) {
  deleteModal.value = { show: true, review, loading: false }
}

async function onConfirmDelete(payload: { reason: string }) {
  const review = deleteModal.value.review
  if (!review) return
  deleteModal.value.loading = true
  try {
    await adminStore.deleteReview(review.id, payload.reason || undefined)
    toast.success(t("admin.reviewModal.deleted", { id: review.id }))
    deleteModal.value.show = false
    if (selectedReview.value?.id === review.id) selectedReview.value = null
    fetchData(adminStore.reviewsPage)
  } catch (err: unknown) {
    toast.error(t(adminErrorKey(err)))
    if ((err as { response?: { status?: number } }).response?.status === 404) {
      deleteModal.value.show = false
      fetchData(adminStore.reviewsPage)
    }
  } finally {
    deleteModal.value.loading = false
  }
}

function formatDate(iso: string) {
  return iso ? new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso)) : "-"
}

onMounted(() => fetchData(1))
onBeforeUnmount(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style scoped>
.cell-id,
.cell-num {
  font-variant-numeric: tabular-nums;
}

.cell-id,
.cell-muted {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.author-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.25rem;
}

.author-name {
  font-weight: 600;
}

.media-link {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  min-width: 11rem;
  color: inherit;
  text-decoration: none;
}

.media-thumb {
  flex-shrink: 0;
  width: 2rem;
  height: 3rem;
  object-fit: cover;
  border-radius: 0.25rem;
  background: var(--color-surface-3);
}

.media-thumb--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
}

.media-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.media-title {
  font-weight: 600;
}

.media-type {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.media-ext {
  flex-shrink: 0;
  color: var(--color-text-muted);
}

.rating {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-variant-numeric: tabular-nums;
}

.rating svg {
  fill: var(--color-rating);
  color: var(--color-rating);
}

.body-cell {
  max-width: 22rem;
}

.body-button {
  display: -webkit-box;
  max-width: 100%;
  padding: 0;
  overflow: hidden;
  background: none;
  border: 0;
  font: inherit;
  text-align: left;
  color: var(--color-text-secondary);
  cursor: pointer;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.detail {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin: 0;
}

.detail__media {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.detail__poster {
  width: 4.5rem;
  border-radius: 0.5rem;
}

.detail__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}

.detail__sub {
  margin: 0.125rem 0 0.5rem;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}

.detail__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
  gap: 0.75rem;
}

.detail dt {
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-muted);
}

.detail dd {
  margin: 0.125rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-primary);
}

.detail__body {
  padding: 0.75rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  line-height: 1.6;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.detail__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  width: 100%;
}
</style>
