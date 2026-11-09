<template>
  <div class="lib">
    <SkeletonProfile v-if="data.loading.value" />

    <PrivateSectionLock v-else-if="data.locked.value" :user-id="userId" @reload="data.load()" />

    <AppErrorState
      v-else-if="data.profileFailed.value && !data.profile.value"
      :title="t('library.states.errorTitle')"
      :description="t('library.states.errorBody')"
      @retry="data.load()"
    />

    <div v-else class="lib__layout">
      <LibraryHeader :user-id="userId" :profile="data.profile.value" :is-owner="isOwner" :counts="counts" :list-to="listTo" />

      <p v-if="data.listsFailed.value" class="lib__warn" role="status">
        <strong>{{ t("library.states.partialTitle") }}</strong> {{ t("library.states.partialBody") }}
        <AppButton variant="ghost" size="sm" @click="data.load(true)">{{ t("common.retry") }}</AppButton>
      </p>

      <div class="lib__tabs">
        <AppTabs :model-value="tab" :tabs="tabs" :label="t('library.tabs.label')" @update:model-value="setTab" />
      </div>

      <section :id="panelId" role="tabpanel" class="lib__panel" :aria-labelledby="tabId">
        <DashboardSection
          v-if="tab === 'dashboard'"
          :is-owner="isOwner"
          :watchlist="data.watchlist.value"
          :watched="data.watched.value"
          :likes="data.likes.value"
          :reviews="data.reviews.value"
          :list-to="listTo"
          :tab-to="tabTo"
          @add="openAdd"
        />
        <ListsSection
          v-else-if="tab === 'lists'"
          :active-list="list"
          :user-id="userId"
          :is-owner="isOwner"
          :watchlist="data.watchlist.value"
          :watched="data.watched.value"
          :likes="data.likes.value"
          :reviews="data.reviews.value"
          :reload-key="reloadKey"
          :custom-list-id="customListId"
          :list-to="listTo"
          :custom-list-to="customListTo"
          @open="openList"
          @open-custom="openCustomList"
          @add="showAdd = true"
        />
        <BmolSection v-else :user-id="userId" />
      </section>
    </div>

    <LibraryAddSheet v-if="isOwner" v-model="showAdd" :saved-keys="savedKeys" @added="onAdded" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useRoute } from "vue-router"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import PrivateSectionLock from "@/components/profile/PrivateSectionLock.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import SkeletonProfile from "@/components/ui/skeleton/SkeletonProfile.vue"
import { useAuthStore } from "@/stores/auth"
import type { LibraryListKey, LibraryTab } from "@/types"
import { ADD_LIST_TYPE, LIBRARY_TABS } from "./config"
import { useLibraryData } from "./useLibraryData"
import { useLibraryRoute } from "./useLibraryRoute"
import BmolSection from "./components/bmol/BmolSection.vue"
import DashboardSection from "./components/DashboardSection.vue"
import LibraryAddSheet from "./components/LibraryAddSheet.vue"
import LibraryHeader from "./components/LibraryHeader.vue"
import ListsSection from "./components/ListsSection.vue"

const { t } = useI18n()
const route = useRoute()
const auth = useAuthStore()

const userId = computed(() => {
  const raw = route.params.userId
  return Number(Array.isArray(raw) ? raw[0] : raw)
})
const isOwner = computed(() => auth.user?.id === userId.value)

const data = useLibraryData(userId)
const { tab, list, customListId, setTab, openList, openCustomList, listTo, customListTo, tabTo } = useLibraryRoute()

const tabsRef = useTemplateRef<{ tabId: (key: LibraryTab) => string; panelId: (key: LibraryTab) => string }>("tabsRef")
const tabId = computed(() => tabsRef.value?.tabId(tab.value))
const panelId = computed(() => tabsRef.value?.panelId(tab.value))

const tabs = computed(() => LIBRARY_TABS.map((key: LibraryTab) => ({ key, label: t(`library.tabs.${key}`) })))

const counts = computed<Record<LibraryListKey, number>>(() => ({
  watchlist: data.watchlist.value.length,
  watched: data.watched.value.length,
  likes: data.likes.value.length,
  reviews: data.reviews.value.length,
}))

const showAdd = ref(false)
const reloadKey = ref(0)
const savedKeys = computed(() => new Set(data.watchlist.value.map((e) => e.media.key)))

function openAdd() {
  showAdd.value = true
}
function onAdded() {
  reloadKey.value++
  void data.load(true)
}

// Counts follow changes made inside a list (remove, log a watch) once the user leaves it.
watch(list, (_next, prev) => {
  if (prev) void data.load(true)
})
// Adding from the dashboard happens against the watchlist, so show it afterwards.
watch(showAdd, (open) => {
  if (!open && tab.value === "dashboard" && reloadKey.value) void data.load(true)
})
void ADD_LIST_TYPE
</script>

<style scoped>
.lib { width: 100%; min-height: 100vh; color: var(--color-text-primary); background: var(--color-surface); }
.lib__layout { display: flex; flex-direction: column; gap: 1.25rem; max-width: 75rem; margin: 0 auto; padding: 0 1rem 5rem; }
.lib__tabs { overflow-x: auto; scrollbar-width: none; padding-bottom: 0.25rem; }
.lib__tabs::-webkit-scrollbar { display: none; }
.lib__tabs :deep(.app-tabs__tab) { flex: none; min-height: 2.75rem; white-space: nowrap; }
.lib__warn { display: flex; flex-wrap: wrap; align-items: center; gap: 0.25rem 0.5rem; margin: 0; padding: 0.75rem 1rem; font-size: var(--text-small); color: var(--color-text-secondary); background: var(--color-surface-2); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card); }
.lib__warn strong { color: var(--color-warning); }
@media (min-width: 768px) {
  .lib__layout { padding: 0 1.5rem 5rem; }
}
</style>
