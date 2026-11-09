<template>
  <div class="pt">
    <SkeletonGrid v-if="loading" :count="10" min-width="140px" />

    <AppErrorState v-else-if="error" :title="t('profile.states.sectionError')" @retry="emit('retry')" />

    <TabEmptyState
      v-else-if="entries.length === 0"
      :icon="icon"
      :is-owner="isOwner"
      :owner-title="t(`profile.tabContent.${tab}.emptyOwnerTitle`)"
      :owner-body="t(`profile.tabContent.${tab}.emptyOwnerBody`)"
      :visitor-title="t(`profile.tabContent.${tab}.emptyVisitorTitle`)"
      :visitor-body="t(`profile.tabContent.${tab}.emptyVisitorBody`)"
      :cta-label="t(`profile.tabContent.${tab}.browse`)"
    />

    <template v-else>
      <TabToolbar
        v-model:query="list.query.value"
        v-model:sort="list.sort.value"
        v-model:filter="list.filter.value"
        :count-label="t(`profile.tabContent.${tab}.count`, { n: entries.length })"
        :filters="filters"
        :sorts="sorts"
      />
      <TabEmptyState
        v-if="list.filtered.value.length === 0"
        filtered
        :icon="icon"
        :is-owner="isOwner"
        @clear="list.clear"
      />
      <template v-else>
        <MediaGrid :items="list.visible.value.map((e) => e.media)" :rail="tab" :label="t(`profile.tabs.${tab}`)">
          <template #footer="{ item }">
            <span class="pt__meta">
              <Heart v-if="tab === 'likes'" :size="12" fill="currentColor" class="pt__heart" role="img" :aria-label="t('profile.tabContent.likes.heart')" />
              <time :datetime="entryOf(item).at">{{ formatDate(entryOf(item).at) }}</time>
            </span>
            <span v-if="isOwner" class="pt__actions"><slot name="actions" :entry="entryOf(item)" /></span>
          </template>
        </MediaGrid>
        <div v-if="list.hasMore.value" class="pt__more">
          <AppButton variant="outline" @click="list.loadMore">{{ t("profile.tabContent.common.loadMore") }}</AppButton>
          <span class="pt__shown">{{ t("profile.tabContent.common.showing", { shown: list.visible.value.length, total: list.filtered.value.length }) }}</span>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, toRef, type Component } from "vue"
import { Heart } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"
import type { HomeMedia } from "@/components/home/homeMedia"
import MediaGrid from "./MediaGrid.vue"
import TabEmptyState from "./TabEmptyState.vue"
import TabToolbar from "./TabToolbar.vue"
import { ENTRY_FILTERS, ENTRY_SORTS, ENTRY_SORT_KEYS, type ProfileEntry } from "./profileMedia"
import { useTabList } from "./useTabList"

const props = defineProps<{
  /** Selects the profile.tabContent.<tab> copy. */
  tab: "watchlist" | "likes"
  entries: ProfileEntry[]
  loading: boolean
  error: boolean
  isOwner: boolean
  icon: Component
}>()
const emit = defineEmits<{ retry: [] }>()
defineSlots<{ actions?: (p: { entry: ProfileEntry }) => unknown }>()

const { t, locale } = useI18n()

const list = useTabList({
  items: toRef(props, "entries"),
  searchText: (e) => e.media.title,
  filters: ENTRY_FILTERS,
  sorts: ENTRY_SORTS,
  defaultSort: "newest",
  pageSize: 48,
})

const byKey = computed(() => new Map(props.entries.map((e) => [e.media.key, e])))
// Only items from `visible` reach the footer slot, and each has an entry.
const entryOf = (item: HomeMedia) => byKey.value.get(item.key)!
const filters = computed(() => [
  { key: "all", label: t("profile.tabContent.common.filters.all"), count: list.filterCounts.value.all },
  { key: "movie", label: t("profile.tabContent.common.filters.movie"), count: list.filterCounts.value.movie },
  { key: "tv", label: t("profile.tabContent.common.filters.tv"), count: list.filterCounts.value.tv },
])
const sorts = computed(() => ENTRY_SORT_KEYS.map((key) => ({ key, label: t(`profile.tabContent.common.sorts.${key}`) })))

const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso))
</script>

<style scoped>
.pt__meta { display: inline-flex; align-items: center; gap: 0.3rem; min-width: 0; font-size: var(--text-caption); color: var(--color-text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pt__heart { flex: none; color: var(--color-brand); }
.pt__actions { display: inline-flex; flex: none; align-items: center; }
.pt__more { display: flex; flex-direction: column; align-items: center; gap: 0.5rem; padding: 2rem 0 0.5rem; }
.pt__shown { font-size: var(--text-caption); color: var(--color-text-muted); }
</style>
