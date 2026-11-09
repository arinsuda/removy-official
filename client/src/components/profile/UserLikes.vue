<template>
  <div>
    <PosterTab tab="likes" :icon="Heart" :entries="entries" :loading="loading" :error="loadError" :is-owner="isOwner" @retry="load">
      <template #actions="{ entry }">
        <TabIconButton data-action="remove" danger :label="t('profile.tabContent.common.remove', { title: entry.media.title })" @click="askRemove(entry)">
          <Trash2 :size="15" aria-hidden="true" />
        </TabIconButton>
      </template>
    </PosterTab>

    <ConfirmModal v-model="showModal" :list-type="listType" :item-name="selected?.media.title" @confirm="doRemove" @cancel="showModal = false" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Heart, Trash2 } from "lucide-vue-next"
import { mediaApi } from "@/api/endpoints/media"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"
import { useAuthStore } from "@/stores/auth"
import type { LikeItem, ListType } from "@/types"
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

async function load() {
  try {
    loading.value = true
    loadError.value = false
    const response = await mediaApi.getLikeByUserId(props.userId)
    // Likes are keyed by media id: the unlike call takes (type, media id).
    entries.value = response.data.likes.map((item: LikeItem) => ({ id: item.media.id, media: summaryToHomeMedia(item.media), at: item.created_at }))
  } catch (err) {
    console.error("Fetch favorites failed:", err)
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
  const entry = selected.value
  if (!entry) return
  try {
    await mediaApi.unlikeMedia(entry.media.type, entry.id)
    entries.value = entries.value.filter((e) => e.media.key !== entry.media.key)
    showModal.value = false
  } catch (err) {
    console.error("Remove failed:", err)
  }
}
</script>
