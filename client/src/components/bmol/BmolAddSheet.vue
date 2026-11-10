<template>
  <AppModal v-model="open" size="md" class="add-sheet" :title="t('library.bmol.add.title', { type: typeLabel })" :dismiss-disabled="busy">
    <AppTabs v-model="source" :tabs="tabs" :controls="false" :label="t('library.bmol.add.sourceLabel')" class="add-sheet__tabs" />

    <div v-if="source === 'search'" class="add-sheet__search">
      <label class="add-sheet__label" :for="searchId">{{ t("library.bmol.add.searchLabel") }}</label>
      <div class="add-sheet__input-wrap">
        <Search :size="16" aria-hidden="true" />
        <input :id="searchId" v-model="query" type="search" class="add-sheet__input" :placeholder="t('library.bmol.add.searchPlaceholder')" autocomplete="off" />
      </div>
    </div>

    <div class="add-sheet__results" :aria-busy="listLoading">
      <SkeletonList v-if="listLoading" :count="4" />
      <p v-else-if="listError" class="add-sheet__empty" role="alert">{{ t("library.bmol.add.loadFailed") }}</p>
      <p v-else-if="!candidates.length" class="add-sheet__empty">
        {{ source === "search" ? (query.trim() ? t("library.bmol.add.noResults") : t("library.bmol.add.searchHint")) : t("library.bmol.suggestions.empty") }}
      </p>
      <ul v-else class="add-sheet__list">
        <li v-for="c in candidates" :key="candidateKey(c)">
          <label class="add-sheet__row" :class="{ 'is-disabled': ranked.has(candidateKey(c)), 'is-selected': selected.has(candidateKey(c)) }">
            <input
              type="checkbox"
              class="sr-only"
              :checked="selected.has(candidateKey(c))"
              :disabled="ranked.has(candidateKey(c))"
              @change="toggle(c)"
            />
            <img v-if="posterOf(c)" :src="posterOf(c)!" alt="" class="add-sheet__poster" loading="lazy" />
            <span v-else class="add-sheet__poster add-sheet__poster--empty" aria-hidden="true"><Film :size="16" /></span>
            <span class="add-sheet__info">
              <span class="add-sheet__name">{{ c.title }}</span>
              <span v-if="c.year" class="add-sheet__year">{{ c.year }}</span>
            </span>
            <span v-if="ranked.has(candidateKey(c))" class="add-sheet__badge">{{ t("library.bmol.add.alreadyRanked") }}</span>
            <Check v-else-if="selected.has(candidateKey(c))" :size="18" class="add-sheet__check" aria-hidden="true" />
          </label>
        </li>
      </ul>
    </div>

    <template #footer>
      <div class="add-sheet__footer">
        <div class="add-sheet__rank">
          <label :for="rankId">{{ t("library.bmol.add.targetRank") }}</label>
          <input :id="rankId" v-model.number="targetRank" type="number" min="1" max="999" inputmode="numeric" class="add-sheet__rank-input" />
        </div>
        <p class="add-sheet__hint">{{ t("library.bmol.add.tieHint") }}</p>
        <AppButton block :loading="busy" :disabled="!selected.size || !rankValid" @click="submit">
          {{ t("library.bmol.add.submit", { count: selected.size }) }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useQuery } from "@tanstack/vue-query"
import { Check, Film, Search } from "lucide-vue-next"
import { libraryApi, movieApi } from "@/api/api"
import { fetchAllUserReviews } from "@/utils/fetchAllUserReviews"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import SkeletonList from "@/components/ui/skeleton/SkeletonList.vue"
import { getTmdbImageUrl } from "@/utils/image"
import type { BMOLItemResponse, MediaType } from "@/types"
import { nextFreeRank } from "./bmolRanks"
import type { BmolCandidate } from "./useBmol"

interface Candidate extends BmolCandidate {
  year: string
}

const FAVORITE_MIN_RATING = 4.5
const SEARCH_DEBOUNCE_MS = 300

const props = defineProps<{
  modelValue: boolean
  userId: number
  mediaType: MediaType
  /** Everything already ranked (both types), for the "already ranked" flag and the default rank. */
  items: BMOLItemResponse[]
  /** Rank to preselect, e.g. 1 from the onboarding call to action. */
  defaultRank?: number
}>()
const emit = defineEmits<{
  "update:modelValue": [value: boolean]
  add: [candidates: BmolCandidate[], rank: number]
}>()

const { t } = useI18n()
const searchId = useId()
const rankId = useId()
const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })
const typeLabel = computed(() => t(props.mediaType === "tv" ? "library.bmol.tabs.tv" : "library.bmol.tabs.movie"))

const source = ref<"search" | "favorites">("search")
const tabs = computed(() => [
  { key: "search" as const, label: t("library.bmol.add.tabSearch") },
  { key: "favorites" as const, label: t("library.bmol.suggestions.tab") },
])

const query = ref("")
const searchResults = ref<Candidate[]>([])
const searching = ref(false)
const searchFailed = ref(false)
let searchSeq = 0
let timer: ReturnType<typeof setTimeout> | undefined

const selected = ref(new Map<string, Candidate>())
const targetRank = ref(1)
const busy = ref(false)

const candidateKey = (c: Candidate) => `${c.media_type}:${c.id}`
const posterOf = (c: Candidate) => getTmdbImageUrl(c.poster_url, "w92")
const ranked = computed(() => new Set(props.items.map((i) => `${i.media_type}:${i.media.id}`)))
const rankValid = computed(() => Number.isInteger(targetRank.value) && targetRank.value >= 1 && targetRank.value <= 999)

function fromMovie(m: { id: number; title: string; poster_path: string; vote_average: number; release_date?: string }): Candidate {
  return { id: m.id, media_type: "movie", title: m.title, poster_url: m.poster_path ?? "", vote_average: m.vote_average ?? 0, year: m.release_date?.slice(0, 4) ?? "" }
}
function fromTv(m: { id: number; name: string; poster_path: string; vote_average: number; first_air_date?: string }): Candidate {
  return { id: m.id, media_type: "tv", title: m.name, poster_url: m.poster_path ?? "", vote_average: m.vote_average ?? 0, year: m.first_air_date?.slice(0, 4) ?? "" }
}

async function runSearch(text: string) {
  const seq = ++searchSeq
  searching.value = true
  searchFailed.value = false
  try {
    const list =
      props.mediaType === "tv"
        ? (await movieApi.searchSeries(text)).data.results.slice(0, 8).map(fromTv)
        : (await movieApi.search(text)).data.results.slice(0, 8).map(fromMovie)
    if (seq === searchSeq) searchResults.value = list
  } catch {
    if (seq === searchSeq) searchFailed.value = true
  } finally {
    if (seq === searchSeq) searching.value = false
  }
}

watch(query, (value) => {
  clearTimeout(timer)
  const text = value.trim()
  if (!text) {
    searchSeq++
    searchResults.value = []
    searching.value = false
    return
  }
  searching.value = true
  timer = setTimeout(() => runSearch(text), SEARCH_DEBOUNCE_MS)
})
onBeforeUnmount(() => clearTimeout(timer))

// Suggestions: the owner's own 4.5+ reviews and watched titles of this type.
const favorites = useQuery({
  queryKey: computed(() => ["bmol-suggestions", props.userId, props.mediaType]),
  enabled: computed(() => props.modelValue && source.value === "favorites"),
  retry: false,
  staleTime: 60_000,
  queryFn: async (): Promise<Candidate[]> => {
    const [reviews, watched] = await Promise.all([
      fetchAllUserReviews(props.userId),
      libraryApi.getVisibleUserLibrary(props.userId, { list_type: "watched", media_type: props.mediaType }),
    ])
    const out = new Map<string, Candidate>()
    const push = (m: { id: number; title: string; poster_url: string; vote_average: number; media_type: MediaType; created_at?: string }) => {
      if (m.media_type !== props.mediaType) return
      const c: Candidate = { id: m.id, media_type: m.media_type, title: m.title, poster_url: m.poster_url, vote_average: m.vote_average, year: "" }
      if (!out.has(candidateKey(c))) out.set(candidateKey(c), c)
    }
    reviews.filter((r) => r.rating >= FAVORITE_MIN_RATING).forEach((r) => push(r.media))
    watched.data.items.forEach((i) => push(i.media))
    return [...out.values()]
  },
})

const candidates = computed(() => (source.value === "search" ? searchResults.value : (favorites.data.value ?? [])))
const listLoading = computed(() => (source.value === "search" ? searching.value : favorites.isLoading.value))
const listError = computed(() => (source.value === "search" ? searchFailed.value : favorites.isError.value))

function toggle(c: Candidate) {
  const next = new Map(selected.value)
  if (next.has(candidateKey(c))) next.delete(candidateKey(c))
  else next.set(candidateKey(c), c)
  selected.value = next
}

watch(
  () => props.modelValue,
  (visible) => {
    if (!visible) return
    source.value = "search"
    query.value = ""
    searchResults.value = []
    selected.value = new Map()
    targetRank.value = props.defaultRank ?? nextFreeRank(props.items)
  },
  { immediate: true },
)
watch(() => props.mediaType, () => {
  query.value = ""
  searchResults.value = []
  selected.value = new Map()
})

async function submit() {
  if (!selected.value.size || !rankValid.value) return
  busy.value = true
  emit("add", [...selected.value.values()], targetRank.value)
  // The board closes the sheet; reset here so a reopened sheet is clean.
  busy.value = false
  open.value = false
}
</script>

<style>
.add-sheet__tabs { margin-bottom: 0.75rem; }
.add-sheet__label { display: block; margin-bottom: 0.375rem; font-size: var(--text-caption); color: var(--color-text-secondary); }
.add-sheet__input-wrap {
  display: flex; align-items: center; gap: 0.5rem; padding: 0 0.75rem; color: var(--color-text-muted);
  background: var(--color-surface); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card);
}
.add-sheet__input { flex: 1; min-width: 0; padding: 0.625rem 0; font: inherit; font-size: var(--text-small); color: var(--color-text-primary); background: none; border: 0; }
.add-sheet__results { min-height: 12rem; margin-top: 0.75rem; }
.add-sheet__empty { padding: 2rem 0; font-size: var(--text-small); text-align: center; color: var(--color-text-muted); }
.add-sheet__list { display: flex; flex-direction: column; gap: 0.25rem; margin: 0; padding: 0; list-style: none; }
.add-sheet__row {
  display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem; cursor: pointer;
  border: 1px solid transparent; border-radius: var(--radius-card);
}
.add-sheet__row:hover { background: var(--color-surface-3); }
.add-sheet__row.is-selected { border-color: var(--color-brand); background: color-mix(in srgb, var(--color-brand) 10%, transparent); }
.add-sheet__row.is-disabled { cursor: not-allowed; opacity: 0.55; }
.add-sheet__row:has(input:focus-visible) { outline: 2px solid var(--focus-ring-color); outline-offset: 2px; }
.add-sheet__poster { flex: none; width: 2.25rem; aspect-ratio: 2 / 3; object-fit: cover; background: var(--color-surface-3); border-radius: 4px; }
.add-sheet__poster--empty { display: grid; place-items: center; color: var(--color-text-muted); }
.add-sheet__info { display: flex; flex: 1; flex-direction: column; min-width: 0; }
.add-sheet__name { overflow: hidden; font-size: var(--text-small); font-weight: 600; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.add-sheet__year { font-size: var(--text-caption); color: var(--color-text-muted); }
.add-sheet__badge { font-size: var(--text-caption); color: var(--color-text-muted); }
.add-sheet__check { color: var(--color-brand); }
.add-sheet__footer { display: flex; flex-direction: column; gap: 0.5rem; width: 100%; }
.add-sheet__rank { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; font-size: var(--text-small); color: var(--color-text-secondary); }
.add-sheet__rank-input {
  width: 5rem; padding: 0.375rem 0.5rem; font: inherit; color: var(--color-text-primary); text-align: center;
  background: var(--color-surface); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card);
}
.add-sheet__hint { font-size: var(--text-caption); color: var(--color-text-muted); }

@media (max-width: 767.98px) {
  .app-modal:has(> .add-sheet) { align-items: flex-end; padding: 0; }
  .add-sheet { max-width: none; max-height: 94dvh; border-bottom: 0; border-radius: 20px 20px 0 0; padding-bottom: env(safe-area-inset-bottom); }
}
</style>
