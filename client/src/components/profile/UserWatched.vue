<template>
  <div class="wl-root">
    <div class="section-head">
      <span class="eyebrow">Watchlist</span>
      <span class="count-chip">{{ watchlist.length }}</span>
      <div class="rule" />
    </div>

    <SkeletonGrid v-if="loading" :count="8" min-width="130px" />

    <div v-else-if="watchlist.length === 0" class="state-empty">
      <Bookmark :size="28" :stroke-width="1.2" />
      <p>Your watchlist is empty</p>
    </div>

    <div v-else class="poster-grid">
      <div
        v-for="(item, i) in watchlist"
        :key="item.id"
        class="poster-card"
        :style="{ '--i': i }"
      >
        <div
          class="poster-frame"
          tabindex="0"
          @click="goToDetail(item.mediaId, item.mediaType)"
        >
          <img
            v-if="item.coverUrl"
            :src="item.coverUrl"
            :alt="item.title"
            class="poster-img"
            loading="lazy"
          />
          <div v-else class="poster-fallback">
            <Film :size="18" :stroke-width="1.4" />
          </div>

          <span class="cat-badge">{{ item.category }}</span>

          <button
            class="rm-btn"
            :aria-label="`Remove ${item.title}`"
            @click.stop="handleRemove(item.id)"
          >
            <X :size="11" />
          </button>

          <ConfirmModal
            v-model="showModal"
            :list-type="listType"
            :item-name="selectedItem?.title"
            @confirm="doRemove"
            @cancel="showModal = false"
          />

          <div class="poster-overlay">
            <p class="overlay-name">{{ item.title }}</p>
            <span class="overlay-genre">{{
              item.genres.slice(0, 2).join(" · ")
            }}</span>
            <div class="overlay-bottom">
              <span class="overlay-rating flex items-center gap-0.5"
                ><Star :size="9" class="fill-amber-500 text-amber-500" /> {{ item.rating.toFixed(1) }}</span
              >
              <span class="overlay-date"
                ><Clock :size="9" /> {{ item.addedAt }}</span
              >
            </div>
          </div>
        </div>

        <div class="poster-meta">
          <h4
            class="poster-name"
            @click="goToDetail(item.mediaId, item.mediaType)"
          >
            {{ item.title }}
          </h4>

          <!-- Watch Log Bar -->
          <div class="watch-log-bar">
            <button
              class="watch-count-chip"
              :title="$t('watchLog.watchHistory')"
              @click.stop="openWatchHistory(item)"
            >
              <Eye :size="11" />
              <span>{{ getWatchCountText(item) }}</span>
            </button>
            <button
              v-if="isOwner"
              class="btn-log-watch-mini"
              :title="$t('watchLog.recordAnother')"
              @click.stop="openWatchLogModal(item)"
            >
              <Plus :size="11" />
              <span>{{ $t('watchLog.recordWatch') }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals for Watch Log & History -->
    <WatchLogModal
      v-if="showLogModal && activeLogMedia"
      :media-id="activeLogMedia.id"
      :media-type="activeLogMedia.type"
      @close="showLogModal = false"
      @logged="onWatchLogged"
    />

    <WatchHistoryPanel
      v-if="showHistoryPanel && activeHistoryMedia"
      :media-id="activeHistoryMedia.id"
      :media-type="activeHistoryMedia.type"
      @close="showHistoryPanel = false"
      @changed="onWatchHistoryChanged"
    />
  </div>
</template>

<script setup lang="ts">
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"
  import { onMounted, ref, computed } from "vue"
  import { useI18n } from "vue-i18n"
  import { Bookmark, Film, X, Clock, Star, Eye, Plus } from "lucide-vue-next"
  import { libraryApi } from "@/api/api"
  import { watchLogApi } from "@/api/endpoints/watchLog"
  import { useAuthStore } from "@/stores/auth"
  import { useRouter } from "vue-router"
  import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
  import WatchLogModal from "@/components/movie/WatchLogModal.vue"
  import WatchHistoryPanel from "@/components/movie/WatchHistoryPanel.vue"
  import type { ListType } from "@/types"

  const props = defineProps<{
    userId: number
    listType: ListType
  }>()

  const router = useRouter()
  const { t } = useI18n()
  const auth = useAuthStore()
  const isOwner = computed(() => auth.user?.id === props.userId)

  const loading = ref(false)
  const TMDB_IMG = "https://image.tmdb.org/t/p/w342"

  const showModal = ref(false)
  const pendingId = ref<number | null>(null)
  const pendingType = ref<"movie" | "tv" | null>(null)
  const selectedItem = ref<WatchlistItem | null>(null)

  const watchSummaries = ref<Record<string, number>>({})
  const showLogModal = ref(false)
  const activeLogMedia = ref<{ id: number; type: "movie" | "tv" } | null>(null)
  const showHistoryPanel = ref(false)
  const activeHistoryMedia = ref<{ id: number; type: "movie" | "tv" } | null>(null)

  interface WatchlistItem {
    id: number
    mediaId: number
    mediaType: string
    title: string
    category: string
    coverUrl: string
    addedAt: string
    genres: string[]
    rating: number
  }

  const watchlist = ref<WatchlistItem[]>([])

  function getWatchMediaKey(mediaId: number, mediaType: string): string {
    return `${mediaType}_${mediaId}`
  }

  function getWatchCount(item: WatchlistItem): number {
    const key = getWatchMediaKey(item.mediaId, item.mediaType)
    return watchSummaries.value[key] ?? 1
  }

  function getWatchCountText(item: WatchlistItem): string {
    const count = getWatchCount(item)
    if (count === 1) {
      return t("watchLog.watchedOnce")
    }
    return t("watchLog.watchedTimes", { count })
  }

  async function fetchWatchSummaryForItem(mediaId: number, mediaType: "movie" | "tv") {
    try {
      const res = await watchLogApi.getMyWatchLogs(mediaType, mediaId)
      const key = getWatchMediaKey(mediaId, mediaType)
      if (res.data) {
        const count = res.data.summary?.watch_count ?? res.data.logs?.length
        if (count !== undefined) {
          watchSummaries.value[key] = count
        }
      }
    } catch (err) {
      // Fallback gracefully
    }
  }

  function openWatchLogModal(item: WatchlistItem) {
    activeLogMedia.value = { id: item.mediaId, type: item.mediaType as "movie" | "tv" }
    showLogModal.value = true
  }

  function onWatchLogged() {
    showLogModal.value = false
    if (activeLogMedia.value) {
      const key = getWatchMediaKey(activeLogMedia.value.id, activeLogMedia.value.type)
      watchSummaries.value[key] = (watchSummaries.value[key] || 1) + 1
      fetchWatchSummaryForItem(activeLogMedia.value.id, activeLogMedia.value.type)
    }
  }

  function openWatchHistory(item: WatchlistItem) {
    activeHistoryMedia.value = { id: item.mediaId, type: item.mediaType as "movie" | "tv" }
    showHistoryPanel.value = true
  }

  function onWatchHistoryChanged() {
    if (activeHistoryMedia.value) {
      fetchWatchSummaryForItem(activeHistoryMedia.value.id, activeHistoryMedia.value.type)
    }
  }

  onMounted(async () => {
    try {
      loading.value = true
      const response = await libraryApi.getVisibleUserLibrary(props.userId, {
        list_type: "watched",
      })

      const seen = new Set<string>()
      const uniqueItems: WatchlistItem[] = []
      for (const item of response.data.items) {
        if (!item.media) continue
        const key = `${item.media.media_type}_${item.media.id}`
        if (!seen.has(key)) {
          seen.add(key)
          uniqueItems.push({
            id: item.id,
            mediaId: item.media.id,
            mediaType: item.media.media_type,
            title: item.media.title,
            category: item.media.media_type,
            coverUrl: item.media.poster_url
              ? `${TMDB_IMG}${item.media.poster_url}`
              : "",
            addedAt: new Date(item.created_at).toLocaleDateString(),
            genres: item.media.genres ? item.media.genres.map(g => g.name) : [],
            rating: item.media.vote_average,
          })
        }
      }
      watchlist.value = uniqueItems

      for (const item of watchlist.value) {
        fetchWatchSummaryForItem(item.mediaId, item.mediaType as "movie" | "tv")
      }
    } catch (err) {
      console.error("Fetch failed:", err)
    } finally {
      loading.value = false
    }
  })

  function goToDetail(mediaId: number, mediaType: string) {
    router.push({
      name: mediaType === "tv" ? "tv-detail" : "movie-detail",
      params: { id: mediaId },
    })
  }

  function handleRemove(id: number) {
    const item = watchlist.value.find(i => i.id === id)
    if (!item) return

    pendingId.value = id
    pendingType.value = item.category as "movie" | "tv"
    selectedItem.value = item
    showModal.value = true
  }

  async function doRemove() {
    if (pendingId.value == null || !pendingType.value) return

    try {
      await libraryApi.removeItem(pendingId.value)

      watchlist.value = watchlist.value.filter(i => i.id !== pendingId.value)
      showModal.value = false
    } catch (err) {
      console.error("Remove failed:", err)
    }
  }
</script>

<style scoped>
  .wl-root {
    --c-card: var(--color-surface-2);
    --c-border: var(--color-border);
    --c-border-h: var(--color-border-strong);
    --c-red: var(--color-brand);
    --c-text: var(--color-text-primary);
    --c-sub: var(--color-text-secondary);
    --c-muted: var(--color-text-muted);
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
    font-size: 0.6rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--c-muted);
    white-space: nowrap;
  }
  .count-chip {
    font-size: 0.65rem;
    font-weight: 700;
    background: var(--c-red);
    color: #fff;
    padding: 2px 7px;
    border-radius: 4px;
  }
  .rule {
    flex: 1;
    height: 1px;
    background: var(--c-border);
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

  /* Grid */
  .poster-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 16px 12px;
  }

  .poster-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    animation: fadeUp 0.4s cubic-bezier(0.16, 1, 0.3, 1) calc(var(--i) * 60ms)
      both;
  }

  @keyframes fadeUp {
    from {
      opacity: 0;
      transform: translateY(12px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  /* Frame */
  .poster-frame {
    position: relative;
    aspect-ratio: 2/3;
    background: var(--c-card);
    border: 1px solid var(--c-border);
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    outline: none;
    transition:
      transform 0.3s var(--ease),
      border-color 0.3s,
      box-shadow 0.3s;
  }
  .poster-frame:hover,
  .poster-frame:focus-visible {
    transform: translateY(-4px);
    border-color: var(--c-border-h);
    box-shadow: 0 10px 28px rgba(0, 0, 0, 0.5);
  }

  .poster-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.4s var(--ease);
  }
  .poster-frame:hover .poster-img {
    transform: scale(1.06);
  }

  .poster-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--c-muted);
  }

  .cat-badge {
    position: absolute;
    top: 7px;
    left: 7px;
    font-size: 0.5rem;
    font-weight: 700;
    letter-spacing: 0.06em;
    background: rgba(0, 0, 0, 0.75);
    color: #fff;
    padding: 3px 6px;
    border-radius: 4px;
    backdrop-filter: blur(4px);
    z-index: 2;
  }

  .rm-btn {
    position: absolute;
    top: 7px;
    right: 7px;
    width: 22px;
    height: 22px;
    background: rgba(0, 0, 0, 0.65);
    backdrop-filter: blur(4px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 50%;
    color: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0;
    transform: scale(0.85);
    transition: all 0.2s var(--ease);
    z-index: 3;
  }
  .poster-frame:hover .rm-btn {
    opacity: 1;
    transform: scale(1);
  }
  .rm-btn:hover {
    background: var(--c-red);
    border-color: var(--c-red);
  }

  .poster-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to top,
      rgba(0, 0, 0, 0.9) 0%,
      rgba(0, 0, 0, 0.4) 45%,
      transparent 80%
    );
    opacity: 0;
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    padding: 10px;
    gap: 3px;
    transition: opacity 0.2s;
    z-index: 1;
  }
  .poster-frame:hover .poster-overlay {
    opacity: 1;
  }

  .overlay-name {
    font-size: 0.62rem;
    font-weight: 600;
    color: #fff;
    margin: 0;
    line-height: 1.3;
  }
  .overlay-date {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    font-size: 0.58rem;
    color: rgba(255, 255, 255, 0.6);
  }

  /* Meta */
  .poster-meta {
    padding: 0 2px;
  }
  .poster-name {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--c-text);
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    transition: color 0.2s;
  }
  .poster-name:hover {
    color: var(--c-red);
    cursor: pointer;
  }
  .overlay-genre {
    font-size: 0.55rem;
    color: rgba(255, 255, 255, 0.5);
    margin: 0;
  }
  .overlay-bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .overlay-rating {
    font-size: 0.58rem;
    color: rgba(255, 255, 255, 0.75);
  }

  /* Watch Log Controls */
  .watch-log-bar {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-top: 4px;
    flex-wrap: wrap;
  }

  .watch-count-chip {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    background: rgba(79, 70, 229, 0.12);
    border: 1px solid rgba(99, 102, 241, 0.25);
    color: #a5b4fc;
    padding: 2px 6px;
    border-radius: 6px;
    font-size: 0.65rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s var(--ease);
  }

  .watch-count-chip:hover {
    background: rgba(79, 70, 229, 0.22);
    border-color: rgba(99, 102, 241, 0.45);
    color: #c7d2fe;
  }

  .btn-log-watch-mini {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: #e4e4e7;
    padding: 2px 6px;
    border-radius: 6px;
    font-size: 0.65rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s var(--ease);
  }

  .btn-log-watch-mini:hover {
    background: rgba(225, 37, 27, 0.15);
    border-color: rgba(225, 37, 27, 0.4);
    color: #ffffff;
  }

  @media (hover: none), (max-width: 640px) {
    .rm-btn {
      opacity: 1;
      transform: scale(1);
      background: rgba(0, 0, 0, 0.75);
    }
  }

  @media (max-width: 480px) {
    .poster-grid {
      grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
      gap: 12px 10px;
    }
    .watch-log-bar {
      gap: 3px;
    }
    .watch-count-chip,
    .btn-log-watch-mini {
      padding: 2px 4px;
      font-size: 0.6rem;
    }
  }

  @media (max-width: 360px) {
    .poster-grid {
      grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
      gap: 10px 8px;
    }
  }
</style>
