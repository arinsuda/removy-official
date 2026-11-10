<template>
  <AppModal v-model="open" size="md" :title="modalTitle">
    <label class="las__label" :for="searchId">{{ t("library.add.searchLabel") }}</label>
    <div class="las__input-wrap">
      <Search :size="16" aria-hidden="true" />
      <input :id="searchId" v-model="query" type="search" class="las__input" data-focus-custom :placeholder="t('library.add.searchPlaceholder')" autocomplete="off" />
    </div>

    <div class="las__results" :aria-busy="searching">
      <SkeletonList v-if="searching" :count="4" />
      <p v-else-if="failed" class="las__note" role="alert">{{ t("library.add.loadFailed") }}</p>
      <p v-else-if="!results.length" class="las__note">{{ query.trim() ? t("library.add.noResults") : t("library.add.hint") }}</p>
      <ul v-else class="las__list">
        <li v-for="r in results" :key="r.key" class="las__row">
          <img v-if="posterOf(r)" :src="posterOf(r)!" alt="" class="las__poster" loading="lazy" />
          <span v-else class="las__poster las__poster--empty" aria-hidden="true"><Film :size="16" /></span>
          <span class="las__info">
            <span class="las__name">{{ r.title }}</span>
            <span class="las__sub">{{ t(r.type === "tv" ? "library.add.typeSeries" : "library.add.typeMovie") }}<template v-if="r.year"> · {{ r.year }}</template></span>
          </span>
          <span v-if="isSaved(r)" class="las__saved"><Check :size="14" aria-hidden="true" />{{ t(justAdded.has(r.key) ? "library.add.added" : "library.add.inList") }}</span>
          <AppButton v-else variant="outline" icon-only data-action="add" :loading="busyKey === r.key" :disabled="busyKey !== null" :aria-label="t('library.add.action', { title: r.title })" @click="add(r)">
            <Plus :size="16" aria-hidden="true" />
          </AppButton>
        </li>
      </ul>
    </div>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Check, Film, Plus, Search } from "lucide-vue-next"
import { libraryApi, movieApi } from "@/api/api"
import { listApi } from "@/api/endpoints/list"
import { listErrorKey } from "@/components/lists/listErrors"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { useUIStore } from "@/stores/ui"
import { getTmdbImageUrl } from "@/utils/image"
import type { MediaType } from "@/types"
import { ADD_LIST_TYPE, ADD_POSTER_SIZE, ADD_SEARCH_DEBOUNCE_MS, ADD_SEARCH_LIMIT } from "../config"

interface Result {
  /** `${type}-${id}`, the same key HomeMedia uses. */
  key: string
  id: number
  type: MediaType
  title: string
  poster: string
  year: string
}

const props = defineProps<{ modelValue: boolean; savedKeys: Set<string>; listId?: number; listName?: string }>()
const emit = defineEmits<{ "update:modelValue": [value: boolean]; added: [] }>()

const { t } = useI18n()
const ui = useUIStore()
const searchId = useId()
const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })
const modalTitle = computed(() =>
  props.listId ? t("library.customLists.detail.addTitle", { name: props.listName ?? "" }) : t("library.add.title"),
)

const query = ref("")
const results = ref<Result[]>([])
const searching = ref(false)
const failed = ref(false)
const busyKey = ref<string | null>(null)
const justAdded = ref(new Set<string>())
let seq = 0
let timer: ReturnType<typeof setTimeout> | undefined

const posterOf = (r: Result) => getTmdbImageUrl(r.poster, ADD_POSTER_SIZE)
const isSaved = (r: Result) => props.savedKeys.has(r.key) || justAdded.value.has(r.key)

async function search(text: string) {
  const mine = ++seq
  searching.value = true
  failed.value = false
  try {
    const half = Math.ceil(ADD_SEARCH_LIMIT / 2)
    const [movies, series] = await Promise.all([movieApi.search(text), movieApi.searchSeries(text)])
    if (mine !== seq) return
    results.value = [
      ...movies.data.results.slice(0, half).map((m): Result => ({ key: `movie-${m.id}`, id: m.id, type: "movie", title: m.title, poster: m.poster_path ?? "", year: m.release_date?.slice(0, 4) ?? "" })),
      ...series.data.results.slice(0, half).map((s): Result => ({ key: `tv-${s.id}`, id: s.id, type: "tv", title: s.name, poster: s.poster_path ?? "", year: s.first_air_date?.slice(0, 4) ?? "" })),
    ]
  } catch {
    if (mine === seq) failed.value = true
  } finally {
    if (mine === seq) searching.value = false
  }
}

watch(query, (value) => {
  clearTimeout(timer)
  const text = value.trim()
  if (!text) {
    seq++
    results.value = []
    searching.value = false
    failed.value = false
    return
  }
  searching.value = true
  timer = setTimeout(() => void search(text), ADD_SEARCH_DEBOUNCE_MS)
})
onBeforeUnmount(() => clearTimeout(timer))

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) return
    query.value = ""
    justAdded.value = new Set()
  },
)

async function add(r: Result) {
  busyKey.value = r.key
  try {
    if (props.listId) {
      await listApi.addItem(props.listId, { media_id: r.id, media_type: r.type })
      justAdded.value = new Set(justAdded.value).add(r.key)
      ui.toastSuccess(t("library.customLists.detail.toastAdded", { title: r.title, name: props.listName ?? "" }))
      emit("added")
    } else {
      await libraryApi.addItem({ media_id: r.id, media_type: r.type, list_type: ADD_LIST_TYPE })
      justAdded.value = new Set(justAdded.value).add(r.key)
      ui.toastSuccess(t("library.add.toastAdded", { title: r.title }))
      emit("added")
    }
  } catch (err) {
    ui.toastError(props.listId ? t(listErrorKey(err)) : t("library.add.toastFailed", { title: r.title }))
  } finally {
    busyKey.value = null
  }
}
</script>

<style scoped>
.las__label { display: block; margin-bottom: 0.375rem; font-size: var(--text-caption); font-weight: 600; color: var(--color-text-secondary); }
.las__input-wrap { display: flex; align-items: center; gap: 0.5rem; min-height: 2.75rem; padding: 0 0.875rem; color: var(--color-text-muted); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-btn); }
.las__input-wrap:focus-within { border-color: var(--color-focus); }
.las__input { flex: 1; min-width: 0; font: inherit; font-size: var(--text-small); color: var(--color-text-primary); background: none; border: 0; outline: 0; }
.las__results { min-height: 12rem; margin-top: 1rem; }
.las__note { margin: 2rem 0; font-size: var(--text-small); text-align: center; color: var(--color-text-muted); }
.las__list { display: flex; flex-direction: column; gap: 0.25rem; margin: 0; padding: 0; list-style: none; }
.las__row { display: flex; align-items: center; gap: 0.75rem; min-height: 3.5rem; padding: 0.25rem; border-radius: var(--radius-card); }
.las__poster { flex: none; width: 2.5rem; aspect-ratio: 2 / 3; object-fit: cover; border-radius: 0.25rem; background: var(--color-surface-3); }
.las__poster--empty { display: grid; place-items: center; color: var(--color-text-muted); }
.las__info { display: flex; flex: 1; flex-direction: column; min-width: 0; }
.las__name { overflow: hidden; font-size: var(--text-small); font-weight: 600; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.las__sub { font-size: var(--text-caption); color: var(--color-text-muted); }
.las__saved { display: inline-flex; flex: none; align-items: center; gap: 0.25rem; font-size: var(--text-caption); color: var(--color-success); }
</style>
