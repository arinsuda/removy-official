<template>
  <div>
    <SkeletonList v-if="loading" :count="6" />

    <AppErrorState v-else-if="loadError" :title="t('profile.states.sectionError')" @retry="load" />

    <TabEmptyState
      v-else-if="entries.length === 0"
      :icon="Eye"
      :is-owner="isOwner"
      :owner-title="t('profile.tabContent.watched.emptyOwnerTitle')"
      :owner-body="t('profile.tabContent.watched.emptyOwnerBody')"
      :visitor-title="t('profile.tabContent.watched.emptyVisitorTitle')"
      :visitor-body="t('profile.tabContent.watched.emptyVisitorBody')"
      :cta-label="t('profile.tabContent.watched.browse')"
    />

    <template v-else>
      <p class="uw__stats">
        {{ t("profile.tabContent.watched.thisYear", { n: thisYearCount, year: currentYear }) }}
        <span aria-hidden="true">·</span>
        {{ t("profile.tabContent.watched.total", { n: entries.length }) }}
      </p>
      <TabToolbar
        v-model:query="list.query.value"
        v-model:sort="list.sort.value"
        v-model:filter="list.filter.value"
        v-model:view="view"
        view-toggle
        :view-labels="{ list: t('profile.tabContent.watched.timeline'), grid: t('profile.tabContent.watched.grid') }"
        :count-label="t('profile.tabContent.watched.count', { n: yearEntries.length })"
        :filters="filters"
        :sorts="sorts"
      >
        <template #extra>
          <label class="uw__year">
            <span class="uw__sr">{{ t("profile.tabContent.watched.year") }}</span>
            <select v-model="year" class="uw__select" :aria-label="t('profile.tabContent.watched.year')">
              <option value="all">{{ t("profile.tabContent.watched.allYears") }}</option>
              <option v-for="y in years" :key="y" :value="String(y)">{{ y }}</option>
            </select>
          </label>
        </template>
      </TabToolbar>

      <TabEmptyState v-if="list.filtered.value.length === 0" filtered :icon="Eye" :is-owner="isOwner" @clear="clearAll" />
      <template v-else>
        <WatchedTimeline v-if="view === 'list'" :groups="groups" :is-owner="isOwner">
          <template #actions="{ entry }">
            <WatchedActions :title="entry.media.title" @history="historyTarget = entry.media" @log="logTarget = entry.media" @remove="askRemove(entry)" />
          </template>
        </WatchedTimeline>
        <MediaGrid v-else :items="list.visible.value.map((e) => e.media)" rail="watched" :label="t('profile.tabs.watched')">
          <template #footer="{ item }">
            <span class="uw__date">{{ formatDate(entryOf(item).at) }}</span>
            <span v-if="isOwner" class="uw__acts">
              <WatchedActions :title="item.title" @history="historyTarget = item" @log="logTarget = item" @remove="askRemove(entryOf(item))" />
            </span>
          </template>
        </MediaGrid>
        <div v-if="list.hasMore.value" class="uw__more">
          <AppButton variant="outline" @click="list.loadMore">{{ t("profile.tabContent.common.loadMore") }}</AppButton>
        </div>
        <GuestCapNote v-else-if="capped" />
      </template>
    </template>

    <WatchLogModal v-if="logTarget" :media-id="logTarget.id" :media-type="logTarget.type" @close="logTarget = null" @logged="onWatchLogged" />
    <WatchHistoryPanel v-if="historyTarget" :media-id="historyTarget.id" :media-type="historyTarget.type" @close="historyTarget = null" @changed="refreshCount(historyTarget)" />
    <ConfirmModal v-model="showModal" :list-type="listType" :item-name="selected?.media.title" @confirm="doRemove" @cancel="showModal = false" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Eye } from "lucide-vue-next"
import { libraryApi } from "@/api/api"
import { fetchAllUserReviews } from "@/utils/fetchAllUserReviews"
import { watchLogApi } from "@/api/endpoints/watchLog"
import type { HomeMedia } from "@/components/home/homeMedia"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import WatchHistoryPanel from "@/components/movie/WatchHistoryPanel.vue"
import WatchLogModal from "@/components/movie/WatchLogModal.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useAuthStore } from "@/stores/auth"
import type { ListType } from "@/types"
import GuestCapNote from "./tabs/GuestCapNote.vue"
import MediaGrid from "./tabs/MediaGrid.vue"
import TabEmptyState from "./tabs/TabEmptyState.vue"
import TabToolbar from "./tabs/TabToolbar.vue"
import WatchedActions from "./tabs/WatchedActions.vue"
import WatchedTimeline from "./tabs/WatchedTimeline.vue"
import { ENTRY_FILTERS, ENTRY_SORTS, summaryToHomeMedia } from "./tabs/profileMedia"
import { dedupeWatched, groupByMonth, inYear, yearsOf, type WatchedEntry } from "./tabs/watched"
import { useTabList } from "./tabs/useTabList"
import { useViewMode } from "./tabs/useViewMode"

const props = defineProps<{ userId: number; listType: ListType }>()

const { t, locale } = useI18n()
const auth = useAuthStore()
const isOwner = computed(() => auth.user?.id === props.userId)

const entries = ref<WatchedEntry[]>([])
const capped = ref(false)
const loading = ref(false)
const loadError = ref(false)
const view = useViewMode("watched", "list")

const showModal = ref(false)
const selected = ref<WatchedEntry | null>(null)
const logTarget = ref<HomeMedia | null>(null)
const historyTarget = ref<HomeMedia | null>(null)

const currentYear = new Date().getFullYear()
const year = ref("all")
const years = computed(() => yearsOf(entries.value))
const yearEntries = computed(() => (year.value === "all" ? entries.value : inYear(entries.value, Number(year.value))))
const thisYearCount = computed(() => inYear(entries.value, currentYear).length)

const SORTS = { newest: ENTRY_SORTS.newest, oldest: ENTRY_SORTS.oldest, title: ENTRY_SORTS.title }
const list = useTabList<WatchedEntry>({
  items: yearEntries,
  searchText: (e) => e.media.title,
  filters: ENTRY_FILTERS,
  sorts: SORTS,
  defaultSort: "newest",
  pageSize: 60,
})

const filters = computed(() => [
  { key: "all", label: t("profile.tabContent.common.filters.all"), count: list.filterCounts.value.all },
  { key: "movie", label: t("profile.tabContent.common.filters.movie"), count: list.filterCounts.value.movie },
  { key: "tv", label: t("profile.tabContent.common.filters.tv"), count: list.filterCounts.value.tv },
])
const sorts = computed(() => (["newest", "oldest", "title"] as const).map((key) => ({ key, label: t(`profile.tabContent.common.sorts.${key}`) })))

// Month headers only make sense while the list is in date order.
const groups = computed(() =>
  list.sort.value === "title" ? [{ key: "all", label: "", items: list.visible.value }] : groupByMonth(list.visible.value, locale.value),
)

const byKey = computed(() => new Map(entries.value.map((e) => [e.media.key, e])))
const entryOf = (item: HomeMedia) => byKey.value.get(item.key)!
const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso))

function clearAll() {
  list.clear()
  year.value = "all"
}

async function refreshCount(media: HomeMedia | null) {
  if (!media) return
  try {
    const res = await watchLogApi.getMyWatchLogs(media.type, media.id)
    const count = res.data?.summary?.watch_count ?? res.data?.logs?.length
    const entry = byKey.value.get(media.key)
    if (entry && count !== undefined) entry.count = count
  } catch {
    // The count badge just stays as it was.
  }
}

function onWatchLogged() {
  const media = logTarget.value
  logTarget.value = null
  void refreshCount(media)
}

/** The owner's public-to-viewer ratings, keyed by title; failures just hide the chip. */
async function loadRatings(): Promise<Map<string, number>> {
  const out = new Map<string, number>()
  try {
    for (const r of await fetchAllUserReviews(props.userId)) out.set(`${r.media.media_type}-${r.media.id}`, r.rating)
  } catch {
    // Ratings are optional.
  }
  return out
}

async function load() {
  try {
    loading.value = true
    loadError.value = false
    const [response, ratings] = await Promise.all([libraryApi.getVisibleUserLibrary(props.userId, { list_type: "watched" }), loadRatings()])
    capped.value = response.data.has_more ?? false
    const raw: WatchedEntry[] = response.data.items
      .filter((item) => item.media)
      .map((item) => {
        const media = summaryToHomeMedia(item.media)
        return { id: item.id, media, at: item.watched_at ?? item.created_at, rating: ratings.get(media.key) }
      })
    entries.value = dedupeWatched(raw)
    // Watch logs are the viewer's own, so counts are only fetched for the owner.
    if (isOwner.value) entries.value.forEach((e) => void refreshCount(e.media))
  } catch (err) {
    console.error("Fetch failed:", err)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)

function askRemove(entry: WatchedEntry) {
  selected.value = entry
  showModal.value = true
}

async function doRemove() {
  if (!selected.value) return
  const id = selected.value.id
  try {
    await libraryApi.removeItem(id)
    entries.value = entries.value.filter((e) => e.id !== id)
    showModal.value = false
  } catch (err) {
    console.error("Remove failed:", err)
  }
}
</script>

<style scoped>
.uw__stats { display: flex; flex-wrap: wrap; gap: 0 0.5rem; margin: 0 0 0.75rem; font-size: var(--text-caption); color: var(--color-text-muted); }
.uw__select { min-height: 2.25rem; padding: 0 0.75rem; font: inherit; font-size: var(--text-caption); color: var(--color-text-secondary); background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); cursor: pointer; }
.uw__sr { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
.uw__date { font-size: var(--text-caption); color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.uw__acts { display: inline-flex; flex: none; }
.uw__more { display: flex; justify-content: center; padding: 2rem 0 0.5rem; }
</style>
