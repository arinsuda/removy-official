<template>
  <section class="ld" :aria-labelledby="titleId">
    <header class="ld__head">
      <AppButton variant="ghost" size="sm" @click="emit('back')">
        <ChevronLeft :size="16" aria-hidden="true" />{{ t("library.lists.back") }}
      </AppButton>
      <h2 :id="titleId" class="ld__title"><component :is="LIST_ICONS[listKey]" :size="20" aria-hidden="true" />{{ t(`library.lists.names.${listKey}`) }}</h2>
      <AppButton v-if="canAdd" class="ld__add" size="sm" data-action="add-titles" @click="emit('add')">
        <Plus :size="16" aria-hidden="true" />{{ t("library.lists.addTitles") }}
      </AppButton>
    </header>
    <UserReviews v-if="listKey === 'reviews'" :user-id="userId" />
    <component :is="LIST_VIEWS[listKey]" v-else :key="reloadKey" :user-id="userId" :list-type="listKey" />
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import { ChevronLeft, Plus } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import UserLikes from "@/components/profile/UserLikes.vue"
import UserReviews from "@/components/profile/UserReviews.vue"
import UserWatched from "@/components/profile/UserWatched.vue"
import UserWatchlist from "@/components/profile/UserWatchlist.vue"
import type { LibraryListKey } from "@/types"
import { ADD_LIST_TYPE, LIST_ICONS } from "../config"

const LIST_VIEWS = { watchlist: UserWatchlist, watched: UserWatched, likes: UserLikes }

const props = defineProps<{
  listKey: LibraryListKey
  userId: number
  isOwner: boolean
  /** Bump to remount the list after titles were added. */
  reloadKey: number
}>()
const emit = defineEmits<{ back: []; add: [] }>()

const { t } = useI18n()
const titleId = useId()
const canAdd = computed(() => props.isOwner && props.listKey === ADD_LIST_TYPE)
</script>

<style scoped>
.ld__head { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem 0.75rem; margin-bottom: 1.25rem; }
.ld__title { display: flex; flex: 1 1 auto; align-items: center; gap: 0.5rem; order: 2; margin: 0; font-size: var(--text-title); font-weight: 700; color: var(--color-text-primary); }
.ld__add { order: 3; }
@media (min-width: 640px) {
  .ld__title { order: 0; }
}
</style>
