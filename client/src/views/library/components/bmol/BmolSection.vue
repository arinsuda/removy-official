<template>
  <div class="bmol-wrapper">
    <!-- Media Type Navigation Subtabs (For View Mode) -->
    <div v-if="selectedRankDetail === null" class="bmol-view-header">
      <h2 class="bmol-section-title">
        {{ bmolSubTab === 'movie' ? $t("library.bmol.movieTitle") : $t("library.bmol.tvTitle") }}
      </h2>
      <div class="bmol-header-actions">
        <div class="bmol-view-selector">
          <button
            class="view-selector-btn"
            :class="{ 'view-selector-btn--active': bmolSubTab === 'movie' }"
            @click="bmolSubTab = 'movie'"
          >
            {{ $t("library.filters.movies") }}
          </button>
          <button
            class="view-selector-btn"
            :class="{ 'view-selector-btn--active': bmolSubTab === 'tv' }"
            @click="bmolSubTab = 'tv'"
          >
            {{ $t("library.filters.tvSeries") }}
          </button>
        </div>
        <button
          class="btn-share-bol"
          @click="showShareModal = true"
          :title="$t('share.shareAsImage')"
        >
          <Share2 :size="16" />
          <span>{{ $t("share.shareAsImage") }}</span>
        </button>
      </div>
    </div>

    <!-- BREADCRUMB HEADER (When viewing a specific Rank's detail list) -->
    <div v-if="selectedRankDetail !== null" class="bmol-breadcrumb-card">
      <div class="breadcrumb-left">
        <button class="btn-back-circle" :title="$t('library.bmol.backToList')" @click="goBackToList">
          <ChevronLeft :size="16" />
        </button>
        <div class="breadcrumb-path">
          <span class="path-link" @click="goBackToList">{{ $t("library.bmol.title") }}</span>
          <span class="path-separator">/</span>
          <span class="path-current">
            <span class="rank-badge-glow">
              {{ $t("library.bmol.rank", { rank: selectedRankDetail }) }}
            </span>
          </span>
        </div>
      </div>
      <div class="breadcrumb-right">
        <span class="detail-count-chip">
          {{ bmolFilteredItems.length }} {{ bmolFilteredItems.length === 1 ? 'item' : 'items' }}
        </span>
      </div>
    </div>

    <!-- DETAIL VIEW: When a specific Rank is selected -->
    <div v-if="selectedRankDetail !== null" class="bmol-detail-view fade-in-up">
      <!-- Detail Toolbar: Search + Actions -->
      <div class="detail-toolbar">
        <div class="detail-search-box">
          <Search :size="15" class="detail-search-icon" />
          <input
            v-model="rankFilterQuery"
            type="text"
            :placeholder="$t('library.bmol.filterPlaceholder')"
            class="detail-search-input"
          />
          <button
            v-if="rankFilterQuery.length > 0"
            class="detail-search-clear"
            @click="rankFilterQuery = ''"
          >
            ✕
          </button>
        </div>
        <div v-if="isOwner" class="detail-toolbar-actions">
          <button
            class="btn-quick-add"
            @click="openSpotlight(selectedRankDetail!)"
          >
            + {{ $t("library.bmol.addMedia") }}
          </button>
        </div>
      </div>

      <!-- Ranks List inside Rank Detail -->
      <div v-if="bmolFilteredItems.length > 0" class="bmol-detail-grid">
        <div
          v-for="item in bmolFilteredItems"
          :key="item.id"
          class="bmol-item-card"
        >
          <router-link
            :to="item.media_type === 'tv' ? `/tv/${item.media.id}` : `/movies/${item.media.id}`"
            class="bmol-poster-frame"
          >
            <img
              v-if="item.media.poster_url"
              :src="`${TMDB_IMG}${item.media.poster_url}`"
              :alt="item.media.title"
              loading="lazy"
            />
            <div class="poster-fallback" v-else>
              <Film :size="18" />
            </div>
          </router-link>
          <div class="bmol-meta">
            <router-link
              :to="item.media_type === 'tv' ? `/tv/${item.media.id}` : `/movies/${item.media.id}`"
              class="bmol-title"
            >
              {{ item.media.title }}
            </router-link>
            <span class="bmol-rating">★ {{ item.media.vote_average.toFixed(1) }}</span>

            <!-- Actions (Only for Owner) -->
            <div v-if="isOwner" class="bmol-actions">
              <button
                class="bmol-action-btn"
                :title="$t('library.bmol.increaseRank')"
                :disabled="item.rank <= 1"
                @click="increaseRank(item)"
              >
                ▲
              </button>
              <button
                class="bmol-action-btn"
                :title="$t('library.bmol.decreaseRank')"
                @click="decreaseRank(item)"
              >
                ▼
              </button>
              <button
                class="bmol-action-btn bmol-action-btn--danger"
                @click="triggerRemoveBmolItem(item)"
              >
                ✕
              </button>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="empty-state-card fade-in-up">
        <div class="empty-state-visual">
          <div class="empty-glow-backdrop" />
          <div class="empty-icon-badge">
            <Search :size="32" class="empty-trophy-icon" />
          </div>
        </div>
        <h3 class="empty-state-title">{{ $t("library.bmol.errorEmptySearch") }}</h3>
        <p class="empty-state-description">ไม่พบรายการตรงกับคำค้นหาในอันดับนี้ ลองพิมพ์คำค้นหาใหม่อีกครั้ง</p>
      </div>
    </div>

    <!-- DEFAULT VIEW: Grouped Ranks Grid List -->
    <div v-else>
      <!-- ADD FORM: Search bar to add media to any rank (Owner only) -->
      <div v-if="isOwner" class="bmol-add-card">
        <h3 class="bmol-add-title">
          {{ bmolSubTab === 'movie' ? $t("library.bmol.addMovie") : $t("library.bmol.addTV") }}
        </h3>
        <div class="bmol-add-form">
          <!-- Search Input -->
          <div v-if="!bmolAddSelected" class="bmol-add-search-wrapper">
            <Search :size="15" class="bmol-add-search-icon" />
            <input
              v-model="bmolAddQuery"
              type="text"
              :placeholder="$t('library.bmol.searchPlaceholder')"
              class="bmol-add-search-input"
            />
            <div v-if="bmolAddLoading" class="bmol-add-spinner">
              <AppSpinner :size="16" />
            </div>

            <!-- Search Results Dropdown -->
            <div v-if="bmolAddResults.length > 0" class="bmol-add-dropdown">
              <div
                v-for="res in bmolAddResults"
                :key="res.id"
                class="bmol-add-dropdown-item"
                @click="selectBmolAddResult(res)"
              >
                <img
                  v-if="res.poster_path"
                  :src="`${TMDB_IMG}${res.poster_path}`"
                  :alt="getMediaTitle(res)"
                  class="bmol-add-dropdown-poster"
                />
                <div v-else class="bmol-add-dropdown-poster-fallback">
                  <Film :size="12" />
                </div>
                <div class="bmol-add-dropdown-info">
                  <p class="bmol-add-dropdown-title">{{ getMediaTitle(res) }}</p>
                  <p class="bmol-add-dropdown-date">{{ getMediaDate(res) || '—' }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Selected Preview -->
          <div v-if="bmolAddSelected" class="bmol-add-selected">
            <div class="bmol-add-selected-info">
              <img
                v-if="bmolAddSelected.poster_path"
                :src="`${TMDB_IMG}${bmolAddSelected.poster_path}`"
                :alt="getMediaTitle(bmolAddSelected)"
                class="bmol-add-selected-poster"
              />
              <div class="bmol-add-selected-detail">
                <p class="bmol-add-selected-title">{{ getMediaTitle(bmolAddSelected) }}</p>
                <p class="bmol-add-selected-date">{{ getMediaDate(bmolAddSelected) || '—' }}</p>
              </div>
              <button class="bmol-add-clear-btn" @click="clearBmolAddSelection">✕</button>
            </div>
            <div class="bmol-add-rank-row">
              <label class="bmol-add-rank-label">{{ $t("library.bmol.rank", { rank: '' }).replace('#', '') }}</label>
              <input
                v-model.number="bmolAddRank"
                type="number"
                min="1"
                class="bmol-add-rank-input"
              />
              <button class="bmol-add-submit-btn" @click="submitBmolAdd">
                {{ $t("library.bmol.addMedia") }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Paginated Ranks -->
      <div v-if="groupedBmolItems.length > 0" class="bmol-ranks-list">
        <div
          v-for="group in paginatedBmolGroups"
          :key="group.rank"
          class="bmol-rank-group"
        >
          <div class="bmol-rank-header">
            <span class="rank-number-tag">{{ $t("library.bmol.rank", { rank: group.rank }) }}</span>
            <div class="rank-line" />
            <!-- Quick Add Button (Only for Owner) -->
            <button
              v-if="isOwner"
              class="btn-quick-add"
              :title="$t('library.bmol.addMedia')"
              @click="openSpotlight(group.rank)"
            >
              + {{ $t("library.bmol.addMedia") }}
            </button>
          </div>

          <div class="bmol-rank-items-row">
            <div
              v-for="item in getRankItemsToShow(group.items)"
              :key="item.id"
              class="bmol-item-card"
            >
              <router-link
                :to="item.media_type === 'tv' ? `/tv/${item.media.id}` : `/movies/${item.media.id}`"
                class="bmol-poster-frame"
              >
                <img
                  v-if="item.media.poster_url"
                  :src="`${TMDB_IMG}${item.media.poster_url}`"
                  :alt="item.media.title"
                  loading="lazy"
                />
                <div v-else class="poster-fallback">
                  <Film :size="24" />
                </div>
              </router-link>

              <div class="bmol-meta">
                <router-link
                  :to="item.media_type === 'tv' ? `/tv/${item.media.id}` : `/movies/${item.media.id}`"
                  class="bmol-title"
                >
                  {{ item.media.title }}
                </router-link>
                <span class="bmol-rating">★ {{ item.media.vote_average.toFixed(1) }}</span>

                <!-- Actions (Only for Owner) -->
                <div v-if="isOwner" class="bmol-actions">
                  <button
                    class="bmol-action-btn"
                    :title="$t('library.bmol.increaseRank')"
                    :disabled="item.rank <= 1"
                    @click="increaseRank(item)"
                  >
                    ▲
                  </button>
                  <button
                    class="bmol-action-btn"
                    :title="$t('library.bmol.decreaseRank')"
                    @click="decreaseRank(item)"
                  >
                    ▼
                  </button>
                  <button
                    class="bmol-action-btn bmol-action-btn--danger"
                    @click="triggerRemoveBmolItem(item)"
                  >
                    ✕
                  </button>
                </div>
              </div>
            </div>

            <!-- "+ X More" Card -->
            <div
              v-if="group.items.length > maxSlots"
              class="bmol-item-card bmol-more-card"
              @click="openShowAllDetail(group.rank)"
            >
              <div class="more-card-inner">
                <span class="more-count">
                  {{ $t("library.bmol.moreItems", { count: getMoreCount(group.items) }) }}
                </span>
                <span class="more-sub">Click to view all</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination Controls -->
        <div v-if="bmolTotalPages > 1" class="bmol-pagination">
          <button
            class="bmol-page-btn"
            :disabled="bmolPage <= 1"
            @click="setBmolPage(bmolPage - 1)"
          >
            ‹
          </button>
          <button
            v-for="p in bmolTotalPages"
            :key="p"
            class="bmol-page-btn"
            :class="{ 'bmol-page-btn--active': p === bmolPage }"
            @click="setBmolPage(p)"
          >
            {{ p }}
          </button>
          <button
            class="bmol-page-btn"
            :disabled="bmolPage >= bmolTotalPages"
            @click="setBmolPage(bmolPage + 1)"
          >
            ›
          </button>
        </div>
      </div>

      <div v-else class="empty-state-card fade-in-up">
        <div class="empty-state-visual">
          <div class="empty-glow-backdrop" />
          <div class="empty-icon-badge">
            <Trophy :size="36" class="empty-trophy-icon" />
          </div>
        </div>
        <h3 class="empty-state-title">
          {{ bmolSubTab === 'movie' ? $t("library.bmol.emptyMovieTitle") : $t("library.bmol.emptyTVTitle") }}
        </h3>
        <p class="empty-state-description">{{ $t("library.bmol.empty") }}</p>
        <div v-if="isOwner" class="empty-state-hint">
          <Sparkles :size="15" class="hint-sparkle-icon" />
          <span>{{ $t("library.bmol.emptyHint") }}</span>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <BmolSpotlightModal
      v-model="spotlightActive"
      :rank="spotlightRank"
      :media-type="bmolSubTab"
      :saving="spotlightSaving"
      @save="saveSpotlightItems"
    />

    <BmolDeleteConfirmModal
      v-model="showDeleteConfirm"
      :item-name="deleteItemName"
      :confirm-disabled="deleteLoading"
      @confirm="confirmDelete"
      @cancel="showDeleteConfirm = false"
    />

    <BolShareModal
      v-model="showShareModal"
      :media-type="bmolSubTab"
      :ranks="groupedBmolItems"
      :username="authStore.user?.username || authStore.user?.display_name || ''"
    />
  </div>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue"
import { Search, Trophy, ChevronLeft, Film, Sparkles, Share2 } from "lucide-vue-next"
import { bmolApi, movieApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import type { BMOLItemResponse, Movie, TVSeries } from "@/types"
import BmolSpotlightModal from "./BmolSpotlightModal.vue"
import BmolDeleteConfirmModal from "./BmolDeleteConfirmModal.vue"
import BolShareModal from "@/components/share/BolShareModal.vue"

const props = defineProps<{
  isOwner: boolean
  userId: number
  bmolItems: BMOLItemResponse[]
}>()

const emit = defineEmits<{
  (e: "update:bmolItems", items: BMOLItemResponse[]): void
}>()

const authStore = useAuthStore()
const showShareModal = ref(false)

const TMDB_IMG = "https://image.tmdb.org/t/p/w342"

// Responsive Viewport Slots calculation
const windowWidth = ref(typeof window !== "undefined" ? window.innerWidth : 1200)

function updateWindowWidth() {
  windowWidth.value = window.innerWidth
}

onMounted(() => {
  window.addEventListener("resize", updateWindowWidth, { passive: true })
  updateWindowWidth()
})

onBeforeUnmount(() => {
  window.removeEventListener("resize", updateWindowWidth)
})

// Subtab & Detail selection
const bmolSubTab = ref<"movie" | "tv">("movie")
const selectedRankDetail = ref<number | null>(null)
const rankFilterQuery = ref("")

// Ranks calculations
const groupedBmolItems = computed(() => {
  const filtered = props.bmolItems.filter(
    item => item.media_type === (bmolSubTab.value === "movie" ? "movie" : "tv")
  )
  const groups: Record<number, BMOLItemResponse[]> = {}
  filtered.forEach(item => {
    if (!groups[item.rank]) {
      groups[item.rank] = []
    }
    groups[item.rank]?.push(item)
  })

  return Object.entries(groups)
    .map(([rankStr, items]) => ({
      rank: Number(rankStr),
      items
    }))
    .sort((a, b) => a.rank - b.rank)
})

const bmolFilteredItems = computed(() => {
  if (selectedRankDetail.value === null) return []
  const rankItems = props.bmolItems.filter(
    item => item.rank === selectedRankDetail.value &&
            item.media_type === (bmolSubTab.value === "movie" ? "movie" : "tv")
  )
  const query = rankFilterQuery.value.trim().toLowerCase()
  if (!query) return rankItems
  return rankItems.filter(item =>
    item.media.title.toLowerCase().includes(query)
  )
})

// Dynamic maximum visible item slots in a rank row
const maxSlots = computed(() => {
  if (windowWidth.value >= 1400) return 6
  if (windowWidth.value >= 1200) return 5
  if (windowWidth.value >= 960) return 4
  if (windowWidth.value >= 480) return 3
  return 2
})

function getRankItemsToShow(items: BMOLItemResponse[]) {
  const limit = maxSlots.value
  if (items.length <= limit) {
    return items
  }
  return items.slice(0, limit - 1)
}

function getMoreCount(items: BMOLItemResponse[]) {
  const limit = maxSlots.value
  if (items.length <= limit) return 0
  return items.length - (limit - 1)
}

function openShowAllDetail(rank: number) {
  selectedRankDetail.value = rank
  rankFilterQuery.value = ""
}

function goBackToList() {
  selectedRankDetail.value = null
  rankFilterQuery.value = ""
}

// Reset BMOL state on user ID change
function resetLocalBmolState() {
  selectedRankDetail.value = null
  rankFilterQuery.value = ""
  bmolPage.value = 1
  spotlightActive.value = false
  spotlightRank.value = null
  showDeleteConfirm.value = false
  deleteItemId.value = null
  deleteItemName.value = ""
  bmolAddQuery.value = ""
  bmolAddResults.value = []
  bmolAddSelected.value = null
}

watch(() => props.userId, resetLocalBmolState)

// Watch subtab changes to reset views & query searches
watch(bmolSubTab, () => {
  selectedRankDetail.value = null
  bmolPage.value = 1
  bmolAddQuery.value = ""
  bmolAddResults.value = []
  bmolAddSelected.value = null
})

// --- Pagination ---
const BMOL_PER_PAGE = 5
const bmolPage = ref(1)

const bmolTotalPages = computed(() =>
  Math.max(1, Math.ceil(groupedBmolItems.value.length / BMOL_PER_PAGE))
)

const paginatedBmolGroups = computed(() => {
  const start = (bmolPage.value - 1) * BMOL_PER_PAGE
  return groupedBmolItems.value.slice(start, start + BMOL_PER_PAGE)
})

function setBmolPage(page: number) {
  if (page >= 1 && page <= bmolTotalPages.value) {
    bmolPage.value = page
  }
}

// --- Add Form ---
const bmolAddQuery = ref("")
const bmolAddResults = ref<Array<Movie | TVSeries>>([])
const bmolAddLoading = ref(false)
const bmolAddSelected = ref<(Movie | TVSeries) | null>(null)
const bmolAddRank = ref(1)
let bmolAddSearchTimeout: ReturnType<typeof setTimeout> | null = null

watch(bmolAddQuery, (newVal) => {
  if (bmolAddSearchTimeout) clearTimeout(bmolAddSearchTimeout)
  if (!newVal.trim()) {
    bmolAddResults.value = []
    return
  }
  bmolAddLoading.value = true
  bmolAddSearchTimeout = setTimeout(async () => {
    try {
      if (bmolSubTab.value === "movie") {
        const res = await movieApi.search(newVal)
        bmolAddResults.value = res.data.results.slice(0, 6)
      } else {
        const res = await movieApi.searchSeries(newVal)
        bmolAddResults.value = res.data.results.slice(0, 6)
      }
    } catch (err) {
      console.error("BMOL add search failed:", err)
    } finally {
      bmolAddLoading.value = false
    }
  }, 350)
})

function selectBmolAddResult(media: Movie | TVSeries) {
  bmolAddSelected.value = media
  bmolAddQuery.value = ""
  bmolAddResults.value = []
}

function clearBmolAddSelection() {
  bmolAddSelected.value = null
}

async function submitBmolAdd() {
  if (!bmolAddSelected.value) return
  const media = bmolAddSelected.value
  const rank = bmolAddRank.value

  const tempItem: BMOLItemResponse = {
    id: Date.now() + Math.random(),
    rank,
    media_type: bmolSubTab.value,
    created_at: new Date().toISOString(),
    media: {
      id: media.id,
      title: getMediaTitle(media),
      poster_url: media.poster_path || "",
      vote_average: media.vote_average || 0
    }
  } as BMOLItemResponse

  // Optimistic emit
  emit("update:bmolItems", [...props.bmolItems, tempItem])
  clearBmolAddSelection()

  try {
    await bmolApi.addItem({
      media_id: media.id,
      media_type: bmolSubTab.value,
      rank
    })
    const res = await bmolApi.getUserBMOL(props.userId)
    emit("update:bmolItems", res.data.items)
  } catch (err: unknown) {
    const errorWithResponse = err as { response?: { status: number } }
    if (errorWithResponse.response?.status === 409) {
      emit("update:bmolItems", props.bmolItems.filter(i => i.media.id !== media.id || i.id !== tempItem.id))
      window.$toast?.error(`"${getMediaTitle(media)}" ถูกจัดอันดับในที่สุดของชีวิตแล้ว`, "ข้อผิดพลาด")
    } else {
      emit("update:bmolItems", props.bmolItems.filter(i => i.id !== tempItem.id))
      window.$toast?.error(`ไม่สามารถเพิ่ม "${getMediaTitle(media)}" ได้`, "ข้อผิดพลาด")
      console.error("Failed to add BMOL item:", err)
    }
  }
}

// --- Spotlight Quick Add Modal logic ---
const spotlightActive = ref(false)
const spotlightRank = ref<number | null>(null)
const spotlightSaving = ref(false)

function openSpotlight(rank: number) {
  spotlightRank.value = rank
  spotlightActive.value = true
}

async function saveSpotlightItems(itemsToSave: Array<Movie | TVSeries>) {
  if (itemsToSave.length === 0 || spotlightRank.value === null) return
  const rank = spotlightRank.value
  spotlightSaving.value = true
  spotlightActive.value = false

  const newItems = itemsToSave.map(media => ({
    id: Date.now() + Math.random(),
    rank: rank,
    media_type: bmolSubTab.value,
    created_at: new Date().toISOString(),
    media: {
      id: media.id,
      title: getMediaTitle(media),
      poster_url: media.poster_path || "",
      vote_average: media.vote_average || 0
    }
  } as BMOLItemResponse))

  // Optimistic UI emission
  const initialBackup = [...props.bmolItems]
  emit("update:bmolItems", [...props.bmolItems, ...newItems])

  let hasErrors = false

  for (const media of itemsToSave) {
    try {
      await bmolApi.addItem({
        media_id: media.id,
        media_type: bmolSubTab.value,
        rank: rank
      })
    } catch (err: unknown) {
      hasErrors = true
      const errorWithResponse = err as { response?: { status: number } }
      if (errorWithResponse.response?.status === 409) {
        emit("update:bmolItems", props.bmolItems.filter(i => i.media.id !== media.id))
        window.$toast?.error(`"${getMediaTitle(media)}" ถูกจัดอันดับในที่สุดของชีวิตแล้ว`, "ข้อผิดพลาด")
      } else {
        emit("update:bmolItems", props.bmolItems.filter(i => i.media.id !== media.id))
        window.$toast?.error(`ไม่สามารถเพิ่ม "${getMediaTitle(media)}" ได้`, "ข้อผิดพลาด")
        console.error("Failed to add spotlight item:", err)
      }
    }
  }

  // Refresh final state from backend
  try {
    const res = await bmolApi.getUserBMOL(props.userId)
    emit("update:bmolItems", res.data.items)
  } catch (err) {
    console.error("Refresh failed:", err)
    if (hasErrors) {
      emit("update:bmolItems", initialBackup)
    }
  } finally {
    spotlightSaving.value = false
  }
}

// --- Deletion Flow ---
const showDeleteConfirm = ref(false)
const deleteItemId = ref<number | null>(null)
const deleteItemName = ref("")
const deleteLoading = ref(false)

function triggerRemoveBmolItem(item: BMOLItemResponse) {
  deleteItemId.value = item.id
  deleteItemName.value = item.media.title
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (deleteItemId.value === null) return
  const itemId = deleteItemId.value
  deleteLoading.value = true
  showDeleteConfirm.value = false

  const backup = [...props.bmolItems]
  emit("update:bmolItems", props.bmolItems.filter(i => i.id !== itemId))

  try {
    await bmolApi.removeItem(itemId)
  } catch (err) {
    emit("update:bmolItems", backup)
    console.error("Failed to remove item:", err)
  } finally {
    deleteLoading.value = false
    deleteItemId.value = null
    deleteItemName.value = ""
  }
}

// --- Rank Swap Flow ---
async function increaseRank(item: BMOLItemResponse) {
  if (item.rank <= 1) return
  const oldRank = item.rank
  const newRank = item.rank - 1

  // Optimistic local change
  const updatedItems = props.bmolItems.map(i => {
    if (i.id === item.id) {
      return { ...i, rank: newRank }
    }
    return i
  })
  emit("update:bmolItems", updatedItems)

  try {
    await bmolApi.updateItem(item.id, { rank: newRank })
  } catch (err) {
    // Revert on error
    const revertedItems = props.bmolItems.map(i => {
      if (i.id === item.id) {
        return { ...i, rank: oldRank }
      }
      return i
    })
    emit("update:bmolItems", revertedItems)
    console.error("Failed to increase rank:", err)
  }
}

async function decreaseRank(item: BMOLItemResponse) {
  const oldRank = item.rank
  const newRank = item.rank + 1

  // Optimistic local change
  const updatedItems = props.bmolItems.map(i => {
    if (i.id === item.id) {
      return { ...i, rank: newRank }
    }
    return i
  })
  emit("update:bmolItems", updatedItems)

  try {
    await bmolApi.updateItem(item.id, { rank: newRank })
  } catch (err) {
    // Revert on error
    const revertedItems = props.bmolItems.map(i => {
      if (i.id === item.id) {
        return { ...i, rank: oldRank }
      }
      return i
    })
    emit("update:bmolItems", revertedItems)
    console.error("Failed to decrease rank:", err)
  }
}

// Helpers
function getMediaTitle(media: Movie | TVSeries): string {
  if ("title" in media) {
    return media.title
  }
  return media.name
}

function getMediaDate(media: Movie | TVSeries): string {
  if ("release_date" in media) {
    return media.release_date
  }
  return media.first_air_date
}
</script>

<style scoped>
/* BMOL section styling */
.bmol-section {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

/* ==============================
   BMOL Add Form
   ============================== */
.bmol-add-card {
  position: relative;
  z-index: 30;
  background: rgba(20, 20, 22, 0.55);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 20px;
  backdrop-filter: blur(12px);
  margin-bottom: 1.5rem;
}

.bmol-add-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: #e4e4e7;
  margin: 0 0 14px 0;
  padding-left: 10px;
  border-left: 2px solid #e1251b;
}

.bmol-add-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* Search wrapper */
.bmol-add-search-wrapper {
  position: relative;
  z-index: 40;
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 10px;
  padding: 0 14px;
  height: 42px;
  transition: border-color 0.2s, background-color 0.2s;
}

.bmol-add-search-wrapper:focus-within {
  border-color: rgba(225, 37, 27, 0.35);
  background: rgba(255, 255, 255, 0.04);
}

.bmol-add-search-icon {
  color: #52525b;
  flex-shrink: 0;
  transition: color 0.2s;
}

.bmol-add-search-wrapper:focus-within .bmol-add-search-icon {
  color: #e1251b;
}

.bmol-add-search-input {
  background: transparent;
  border: none;
  outline: none;
  color: #e4e4e7;
  font-size: 0.85rem;
  font-weight: 500;
  width: 100%;
}

.bmol-add-search-input::placeholder {
  color: #52525b;
}

.bmol-add-spinner {
  flex-shrink: 0;
}

/* Dropdown */
.bmol-add-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  background: #18181b;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  z-index: 50;
  max-height: 280px;
  overflow-y: auto;
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.5);
}

.bmol-add-dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  cursor: pointer;
  transition: background-color 0.15s;
  border-bottom: 1px solid rgba(255, 255, 255, 0.02);
}

.bmol-add-dropdown-item:last-child {
  border-bottom: none;
}

.bmol-add-dropdown-item:hover {
  background: rgba(255, 255, 255, 0.04);
}

.bmol-add-dropdown-poster {
  width: 30px;
  height: 45px;
  border-radius: 4px;
  object-fit: cover;
  flex-shrink: 0;
}

.bmol-add-dropdown-poster-fallback {
  width: 30px;
  height: 45px;
  border-radius: 4px;
  background: #27272a;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #52525b;
  flex-shrink: 0;
}

.bmol-add-dropdown-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.bmol-add-dropdown-title {
  font-size: 0.82rem;
  font-weight: 650;
  color: #ffffff;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bmol-add-dropdown-date {
  font-size: 0.7rem;
  color: #71717a;
  margin: 3px 0 0;
}

/* Selected preview */
.bmol-add-selected {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  padding: 14px;
}

.bmol-add-selected-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.bmol-add-selected-poster {
  width: 40px;
  height: 60px;
  border-radius: 6px;
  object-fit: cover;
  flex-shrink: 0;
}

.bmol-add-selected-detail {
  flex: 1;
  min-width: 0;
}

.bmol-add-selected-title {
  font-size: 0.88rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bmol-add-selected-date {
  font-size: 0.72rem;
  color: #71717a;
  margin: 4px 0 0;
}

.bmol-add-clear-btn {
  background: rgba(255, 255, 255, 0.05);
  border: none;
  color: #71717a;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 0.7rem;
  flex-shrink: 0;
  transition: all 0.2s;
}

.bmol-add-clear-btn:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}

/* Rank selector row */
.bmol-add-rank-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bmol-add-rank-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: #a1a1aa;
  white-space: nowrap;
}

.bmol-add-rank-input {
  width: 60px;
  background: #1c1c1e;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  color: #ffffff;
  font-size: 0.85rem;
  font-weight: 600;
  padding: 6px 10px;
  text-align: center;
  outline: none;
  transition: border-color 0.2s;
}

.bmol-add-rank-input:focus {
  border-color: rgba(225, 37, 27, 0.4);
}

/* Chrome number input arrows */
.bmol-add-rank-input::-webkit-inner-spin-button,
.bmol-add-rank-input::-webkit-outer-spin-button {
  opacity: 1;
}

.bmol-add-submit-btn {
  margin-left: auto;
  background: #e1251b;
  color: #ffffff;
  border: none;
  border-radius: 8px;
  font-size: 0.78rem;
  font-weight: 650;
  padding: 8px 18px;
  cursor: pointer;
  transition: background-color 0.2s, transform 0.1s;
  white-space: nowrap;
}

.bmol-add-submit-btn:hover {
  background: #b81d15;
}

.bmol-add-submit-btn:active {
  transform: scale(0.97);
}

/* ==============================
   BMOL Pagination
   ============================== */
.bmol-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 1.5rem;
  padding: 10px 0;
}

.bmol-page-btn {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
  color: #a1a1aa;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.bmol-page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.06);
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.12);
}

.bmol-page-btn--active {
  background: rgba(225, 37, 27, 0.12) !important;
  color: #e1251b !important;
  border-color: rgba(225, 37, 27, 0.3) !important;
}

.bmol-page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

/* Quick Add Button */
.btn-quick-add {
  background: rgba(225, 37, 27, 0.08);
  border: 1px solid rgba(225, 37, 27, 0.2);
  border-radius: 8px;
  color: #e1251b;
  font-size: 0.75rem;
  font-weight: 650;
  padding: 4px 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-quick-add:hover {
  background: #e1251b;
  color: #ffffff;
  border-color: #e1251b;
}

/* + X More Card styling */
.bmol-more-card {
  cursor: pointer;
  background: rgba(255, 255, 255, 0.01) !important;
  border: 2px dashed rgba(255, 255, 255, 0.1) !important;
  display: flex;
  align-items: center;
  justify-content: center;
  aspect-ratio: 2/3;
  transition: all 0.25s;
}

.bmol-more-card:hover {
  background: rgba(225, 37, 27, 0.02) !important;
  border-color: rgba(225, 37, 27, 0.35) !important;
}

.more-card-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 100%;
  width: 100%;
}

.more-count {
  font-size: 1.1rem;
  font-weight: 750;
  color: #e1251b;
}

.more-sub {
  font-size: 0.68rem;
  color: #71717a;
}

.bmol-wrapper {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  overflow-x: hidden;
}

/* View Header area */
.bmol-view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  padding-bottom: 1rem;
  gap: 12px;
  flex-wrap: wrap;
}

.bmol-section-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
}

.bmol-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-share-bol {
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(135deg, rgba(245, 67, 44, 0.9) 0%, rgba(180, 0, 10, 0.9) 100%);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 10px;
  padding: 0.45rem 0.9rem;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(245, 67, 44, 0.3);
  transition: all 0.2s ease;
  touch-action: manipulation;
}

.btn-share-bol:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(245, 67, 44, 0.45);
  background: linear-gradient(135deg, rgba(245, 10, 20, 1) 0%, rgba(200, 0, 10, 1) 100%);
}

.bmol-view-selector {
  display: flex;
  background: rgba(20, 20, 22, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 10px;
  padding: 2px;
}

.view-selector-btn {
  background: transparent;
  border: none;
  color: #9ca3af;
  padding: 0.375rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  border-radius: 8px;
  transition: background-color 0.2s, color 0.2s;
  touch-action: manipulation;
}

.view-selector-btn--active {
  background: rgba(225, 37, 27, 0.15);
  color: #e1251b;
}

/* Rank groups */
.bmol-ranks-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 2rem;
}

.bmol-rank-group {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.bmol-rank-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.rank-number-tag {
  font-size: 0.875rem;
  font-weight: 700;
  color: #e1251b;
  background: rgba(225, 37, 27, 0.1);
  border: 1px solid rgba(225, 37, 27, 0.2);
  padding: 4px 10px;
  border-radius: 8px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.rank-line {
  flex: 1;
  height: 1px;
  background: rgba(255, 255, 255, 0.05);
}

/* Scrollable row of items in same rank */
.bmol-rank-items-row {
  display: flex;
  gap: clamp(10px, 2vw, 20px);
  flex-wrap: nowrap;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  width: 100%;
  max-width: 100%;
  padding-bottom: 0.75rem;
  box-sizing: border-box;
}

.bmol-rank-items-row--scrollable::-webkit-scrollbar {
  height: 6px;
}
.bmol-rank-items-row--scrollable::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.02);
  border-radius: 3px;
}
.bmol-rank-items-row--scrollable::-webkit-scrollbar-thumb {
  background: rgba(225, 37, 27, 0.3);
  border-radius: 3px;
}
.bmol-rank-items-row--scrollable::-webkit-scrollbar-thumb:hover {
  background: rgba(225, 37, 27, 0.5);
}

.bmol-rank-items-row .bmol-item-card {
  flex: 0 0 clamp(104px, 22vw, 150px);
  min-width: 0;
}

.btn-toggle-expand {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  color: #9ca3af;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 4px 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-toggle-expand:hover {
  background: rgba(255, 255, 255, 0.04);
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.15);
}

.bmol-item-card {
  background: rgba(20, 20, 22, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.04);
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: transform 0.25s, border-color 0.25s;
}

.bmol-item-card:hover {
  transform: translateY(-4px);
  border-color: rgba(255, 255, 255, 0.1);
}

.bmol-poster-frame {
  aspect-ratio: 2/3;
  background: #1c1c1e;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bmol-poster-frame img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.5s;
}

.bmol-item-card:hover .bmol-poster-frame img {
  transform: scale(1.05);
}

.bmol-meta {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
}

.bmol-title {
  font-size: 0.875rem;
  font-weight: 700;
  color: #ffffff;
  text-decoration: none;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  transition: color 0.2s;
  flex: 1;
}

.bmol-title:hover {
  color: #e1251b;
}

.bmol-rating {
  font-size: 0.75rem;
  color: #ffb800;
  font-weight: 600;
}

.bmol-actions {
  display: flex;
  gap: 0.375rem;
  margin-top: 0.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
  padding-top: 0.5rem;
}

.bmol-action-btn {
  flex: 1;
  background: #27272a;
  border: 1px solid rgba(255, 255, 255, 0.05);
  color: #a1a1aa;
  font-size: 0.75rem;
  padding: 4px 0;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s, color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bmol-action-btn:hover:not(:disabled) {
  background: #3f3f46;
  color: #ffffff;
}

.bmol-action-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.bmol-action-btn--danger:hover {
  background: rgba(239, 68, 68, 0.1) !important;
  color: #ef4444 !important;
  border-color: rgba(239, 68, 68, 0.2) !important;
}/* Premium Breadcrumb Header Card */
.bmol-breadcrumb-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 14px;
  padding: 12px 18px;
  margin-bottom: 20px;
  backdrop-filter: blur(10px);
}

.breadcrumb-left {
  display: flex;
  align-items: center;
  gap: 14px;
}

.btn-back-circle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #a1a1aa;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-back-circle:hover {
  background: rgba(225, 37, 27, 0.1);
  border-color: rgba(225, 37, 27, 0.3);
  color: #e1251b;
  transform: translateX(-2px);
}

.breadcrumb-path {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  font-weight: 500;
}

.path-link {
  color: #8a8a8e;
  cursor: pointer;
  transition: color 0.2s;
}

.path-link:hover {
  color: #ffffff;
}

.path-separator {
  color: rgba(255, 255, 255, 0.15);
  font-weight: 300;
}

.path-current {
  display: flex;
  align-items: center;
}

.rank-badge-glow {
  background: linear-gradient(135deg, #e1251b 0%, #b81d15 100%);
  color: #ffffff;
  font-size: 0.78rem;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 6px;
  box-shadow: 0 0 12px rgba(225, 37, 27, 0.35);
  letter-spacing: 0.5px;
}

.detail-count-chip {
  font-size: 0.75rem;
  font-weight: 600;
  color: #8a8a8e;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.05);
  padding: 4px 10px;
  border-radius: 20px;
}

/* Detail View Styles */
.bmol-detail-view {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* --- Detail Toolbar --- */
.detail-toolbar {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 12px 16px;
  background: rgba(20, 20, 24, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25);
  margin-bottom: 1.5rem;
}

.detail-search-box {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  gap: 10px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  padding: 0 14px;
  height: 42px;
  transition: all 0.2s ease;
}

.detail-search-box:focus-within {
  border-color: rgba(225, 37, 27, 0.45);
  background: rgba(255, 255, 255, 0.06);
  box-shadow: 0 0 16px rgba(225, 37, 27, 0.15);
}

.detail-search-icon {
  color: #71717a;
  flex-shrink: 0;
  transition: color 0.2s;
}

.detail-search-box:focus-within .detail-search-icon {
  color: #e1251b;
}

.detail-search-input {
  background: transparent;
  border: none;
  outline: none;
  color: #ffffff;
  font-size: 0.85rem;
  font-weight: 500;
  width: 100%;
  min-width: 0;
}

.detail-search-input::placeholder {
  color: #71717a;
  font-weight: 400;
}

.detail-search-clear {
  background: rgba(255, 255, 255, 0.08);
  border: none;
  color: #a1a1aa;
  font-size: 0.7rem;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.2s;
}

.detail-search-clear:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.detail-toolbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.detail-result-count {
  font-size: 0.78rem;
  font-weight: 600;
  color: #a1a1aa;
  background: rgba(255, 255, 255, 0.04);
  padding: 5px 12px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  white-space: nowrap;
}

@media (max-width: 640px) {
  .bmol-view-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  .bmol-section-title {
    font-size: 1.1rem;
  }
  .bmol-header-actions {
    flex-direction: column;
    width: 100%;
    gap: 8px;
  }
  .bmol-view-selector {
    width: 100%;
  }
  .view-selector-btn {
    flex: 1;
    text-align: center;
  }
  .btn-share-bol {
    width: 100%;
    justify-content: center;
  }
  .bmol-breadcrumb-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 14px;
  }
  .breadcrumb-right {
    width: 100%;
    display: flex;
    justify-content: flex-end;
  }
  .detail-toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
    padding: 12px;
  }
  .detail-toolbar-actions {
    justify-content: space-between;
  }
}

.bmol-detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(clamp(100px, 24vw, 150px), 1fr));
  gap: clamp(10px, 2vw, 16px);
}
.bmol-detail-grid .bmol-item-card {
  width: 100% !important;
  min-width: 0 !important;
}

/* ==============================
   Rich Empty State Styling
   ============================== */
.empty-state-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background: rgba(20, 20, 24, 0.65);
  border: 1px dashed rgba(225, 37, 27, 0.25);
  border-radius: 24px;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.35);
  text-align: center;
  margin: 1rem 0;
  position: relative;
  z-index: 1;
  overflow: hidden;
}

.empty-state-visual {
  position: relative;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-glow-backdrop {
  position: absolute;
  width: 100px;
  height: 100px;
  background: radial-gradient(circle, rgba(225, 37, 27, 0.35) 0%, rgba(225, 37, 27, 0) 70%);
  border-radius: 50%;
  filter: blur(12px);
  animation: pulseGlow 3.5s ease-in-out infinite alternate;
}

@keyframes pulseGlow {
  0% {
    transform: scale(0.85);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.15);
    opacity: 1;
  }
}

.empty-icon-badge {
  width: 76px;
  height: 76px;
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(35, 35, 42, 0.9), rgba(20, 20, 24, 0.95));
  border: 1px solid rgba(225, 37, 27, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 2;
  box-shadow: 0 8px 24px rgba(225, 37, 27, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.empty-trophy-icon {
  color: #e1251b;
  filter: drop-shadow(0 2px 8px rgba(225, 37, 27, 0.5));
}

.empty-state-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 0.5rem 0;
  letter-spacing: -0.01em;
}

.empty-state-description {
  font-size: 0.925rem;
  color: #9ca3af;
  max-width: 440px;
  margin: 0 0 1.5rem 0;
  line-height: 1.6;
}

.empty-state-hint {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(225, 37, 27, 0.08);
  border: 1px solid rgba(225, 37, 27, 0.2);
  border-radius: 99px;
  padding: 0.5rem 1.25rem;
  color: #f3f4f6;
  font-size: 0.825rem;
  font-weight: 500;
}

.hint-sparkle-icon {
  color: #e1251b;
}
</style>
