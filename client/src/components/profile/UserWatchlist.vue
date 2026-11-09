<template>
  <div>
    <PosterTab tab="watchlist" :icon="Bookmark" :entries="entries" :loading="loading" :error="loadError" :is-owner="isOwner" @retry="load">
      <template #actions="{ entry }">
        <TabIconButton data-action="mark-watched" :label="t('profile.tabContent.watchlist.markWatchedFor', { title: entry.media.title })" @click="logTarget = entry.media">
          <Check :size="16" aria-hidden="true" />
        </TabIconButton>
        <TabIconButton data-action="remove" danger :label="t('profile.tabContent.common.remove', { title: entry.media.title })" @click="askRemove(entry)">
          <Trash2 :size="15" aria-hidden="true" />
        </TabIconButton>
      </template>
    </PosterTab>

    <WatchLogModal v-if="logTarget" :media-id="logTarget.id" :media-type="logTarget.type" @close="logTarget = null" @logged="onLogged" />
    <ConfirmModal v-model="showModal" :list-type="listType" :item-name="selected?.media.title" @confirm="doRemove" @cancel="showModal = false" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Bookmark, Check, Trash2 } from "lucide-vue-next"
import { libraryApi } from "@/api/api"
import type { HomeMedia } from "@/components/home/homeMedia"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import WatchLogModal from "@/components/movie/WatchLogModal.vue"
import { useAuthStore } from "@/stores/auth"
import type { ListType } from "@/types"
import PosterTab from "./tabs/PosterTab.vue"
import TabIconButton from "./tabs/TabIconButton.vue"
import { summaryToHomeMedia, type ProfileEntry } from "./tabs/profileMedia"

const props = defineProps<{ userId: number; listType: ListType }>()

const { t } = useI18n()
const auth = useAuthStore()
const isOwner = computed(() => auth.user?.id === props.userId)

const entries = ref<ProfileEntry[]>([])
const loading = ref(false)
const loadError = ref(false)

const showModal = ref(false)
const selected = ref<ProfileEntry | null>(null)
const logTarget = ref<HomeMedia | null>(null)

async function load() {
  try {
    loading.value = true
    loadError.value = false
    const response = await libraryApi.getVisibleUserLibrary(props.userId, { list_type: "watchlist" })
    entries.value = response.data.items.map((item) => ({ id: item.id, media: summaryToHomeMedia(item.media), at: item.created_at }))
  } catch (err) {
    console.error("Fetch failed:", err)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)

function askRemove(entry: ProfileEntry) {
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

// Recording a watch can move the title off the watchlist, so reload from the server.
function onLogged() {
  logTarget.value = null
  void load()
}
</script>
