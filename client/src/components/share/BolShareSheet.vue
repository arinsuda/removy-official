<template>
  <AppModal v-model="open" size="md" class="bol-sheet" :title="t('library.bmol.share.title')">
    <AppTabs v-model="active" :tabs="tabs" :controls="false" :label="t('share.templatesLabel')" class="bol-sheet__tabs" />

    <div ref="stage" class="bol-sheet__stage" :style="{ minHeight: `${frameHeight}px` }" :aria-busy="status === 'loading'">
      <div v-if="status === 'ready'" class="bol-sheet__frame" :style="{ width: `${def.width * scale}px`, height: `${def.height * scale}px` }">
        <div class="bol-sheet__scaler" :style="{ width: `${def.width}px`, height: `${def.height}px`, transform: `scale(${scale})` }">
          <BolStoryCard
            v-if="active === 'story'"
            ref="card"
            :items="shownItems"
            :media-type="mediaType"
            :username="shownUsername"
            :theme="theme"
            :posters="posters"
            :note="note"
            :short-url="shortUrl"
          />
          <BolSquareCard
            v-else
            ref="card"
            :items="shownItems"
            :media-type="mediaType"
            :username="shownUsername"
            :theme="theme"
            :posters="posters"
            :short-url="shortUrl"
          />
        </div>
      </div>
      <SkeletonBlock v-else class="bol-sheet__skeleton" height="100%" radius="12px" />
      <div v-if="isExporting" class="bol-sheet__busy">
        <AppSpinner :size="32" />
        <span>{{ t("share.generating") }}</span>
      </div>
    </div>

    <div class="bol-sheet__options">
      <ShareThemePicker v-model="theme" />
      <label v-if="username" class="bol-sheet__check">
        <input v-model="showUsername" type="checkbox" />{{ t("share.showUsername") }}
      </label>
      <label v-if="hasNote && active === 'story'" class="bol-sheet__check">
        <input v-model="includeNote" type="checkbox" />{{ t("library.bmol.share.includeNote") }}
      </label>
      <span class="bol-sheet__ratio">{{ def.width }} × {{ def.height }} · {{ def.ratio }}</span>
    </div>

    <template #footer>
      <div class="bol-sheet__actions">
        <AppButton v-if="canNativeShare" block :disabled="status !== 'ready' || isExporting" @click="handleShare">
          <Share2 :size="16" aria-hidden="true" /> {{ t("share.shareTo") }}
        </AppButton>
        <AppButton :variant="canNativeShare ? 'outline' : 'primary'" block :disabled="status !== 'ready' || isExporting" @click="handleDownload">
          <Download :size="16" aria-hidden="true" /> {{ t("share.downloadPng") }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Download, Share2 } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import { flattenGroups, type RankGroup } from "@/components/bmol/bmolRanks"
import { useShareImage } from "@/composables/useShareImage"
import { getTmdbImageUrl } from "@/utils/image"
import type { MediaType } from "@/types"
import type { ShareTheme } from "@/types/share"
import BolSquareCard from "./cards/BolSquareCard.vue"
import BolStoryCard from "./cards/BolStoryCard.vue"
import ShareThemePicker from "./ShareThemePicker.vue"
import { BOL_TEMPLATES, bolPosterKey, bolShareNote, bolTemplateDef, bolTopItems, type BolTemplateKey } from "./shareCards"

const props = defineProps<{
  modelValue: boolean
  mediaType: MediaType
  /** The signed-in owner's own list. Never pass another user's data. */
  groups: RankGroup[]
  username?: string
}>()
const emit = defineEmits<{ "update:modelValue": [value: boolean] }>()

const { t } = useI18n()
const share = useShareImage()
const isExporting = share.isExporting
const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })

const active = ref<BolTemplateKey>("story")
const theme = ref<ShareTheme>("dark")
const showUsername = ref(true)
const includeNote = ref(false)
const status = ref<"loading" | "ready">("loading")
const posters = ref<Record<string, string | null>>({})
const canNativeShare = ref(false)
const card = useTemplateRef<{ $el: HTMLElement }>("card")
const stage = useTemplateRef<HTMLElement>("stage")
const stageWidth = ref(320)
let loadSeq = 0
let resizeObserver: ResizeObserver | null = null

const tabs = computed(() => BOL_TEMPLATES.map((d) => ({ key: d.key, label: t(d.labelKey) })))
const def = computed(() => bolTemplateDef(active.value))
const shownItems = computed(() => bolTopItems(props.groups, def.value.capacity))
const hasNote = computed(() => bolShareNote(flattenGroups(props.groups), true) !== null)
const note = computed(() => bolShareNote(flattenGroups(props.groups), includeNote.value))
const shownUsername = computed(() => (showUsername.value ? (props.username ?? "") : ""))
const shortUrl = computed(() => window.location.host)

// Preview height follows the viewport so the sheet fits on small phones.
const viewportH = ref(typeof window === "undefined" ? 800 : window.innerHeight)
const maxHeight = computed(() => Math.max(260, Math.min(560, viewportH.value - 420)))
const scale = computed(() => Math.min(stageWidth.value / def.value.width, maxHeight.value / def.value.height))
const frameHeight = computed(() => def.value.height * scale.value)
function onResize() {
  viewportH.value = window.innerHeight
  stageWidth.value = stage.value?.clientWidth || stageWidth.value
}

// Posters go through the API proxy as data URLs before the card renders.
async function loadPosters() {
  const seq = ++loadSeq
  status.value = "loading"
  const items = bolTopItems(props.groups, bolTemplateDef("story").capacity)
  const loaded = await Promise.all(
    items.map(async (item) => {
      const url = getTmdbImageUrl(item.media.poster_url, "w342")
      return [bolPosterKey(item), url ? await share.fetchImageAsBlob(url) : null] as const
    }),
  )
  if (seq !== loadSeq || !props.modelValue) return
  posters.value = Object.fromEntries(loaded)
  status.value = "ready"
}

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      active.value = "story"
      includeNote.value = false
      canNativeShare.value = share.canShareFiles()
      window.addEventListener("resize", onResize)
      onResize()
      void loadPosters()
    } else {
      loadSeq++
      window.removeEventListener("resize", onResize)
      status.value = "loading"
      share.cleanup()
    }
  },
  { immediate: true },
)
watch(stage, (el) => {
  resizeObserver?.disconnect()
  if (!el || typeof ResizeObserver === "undefined") return
  resizeObserver = new ResizeObserver(onResize)
  resizeObserver.observe(el)
})
onBeforeUnmount(() => {
  loadSeq++
  window.removeEventListener("resize", onResize)
  resizeObserver?.disconnect()
  share.cleanup()
})

async function exportCard(): Promise<{ blob: Blob; filename: string } | null> {
  const node = card.value?.$el
  if (!node) return null
  const blob = await share.renderPng(node, { width: def.value.width, height: def.value.height })
  return blob ? { blob, filename: share.sanitizeFilename("best_of_life", `POPCORN_${active.value}`) } : null
}

async function handleShare() {
  const out = await exportCard()
  if (out) await share.shareImageFile(out.blob, out.filename, t("library.bmol.share.caption"))
}
async function handleDownload() {
  const out = await exportCard()
  if (out) share.saveBlob(out.blob, out.filename)
}
</script>

<style>
.bol-sheet__tabs { margin-bottom: 12px; overflow-x: auto; }
.bol-sheet__stage { position: relative; display: flex; align-items: flex-start; justify-content: center; }
.bol-sheet__skeleton { flex: 1; min-height: 260px; }
.bol-sheet__busy {
  position: absolute; inset: 0; z-index: 2; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px;
  border-radius: 12px; background: color-mix(in srgb, black 70%, transparent); color: var(--color-text-primary); font-size: 14px;
}
.bol-sheet__ratio { margin-left: auto; color: var(--color-text-muted); font-size: 12px; }
.bol-sheet__actions { display: flex; flex-direction: column; gap: 8px; width: 100%; }
@media (max-width: 767.98px) {
  .app-modal:has(> .bol-sheet) { align-items: flex-end; padding: 0; }
  .bol-sheet { max-width: none; max-height: 94dvh; border-bottom: 0; border-radius: 20px 20px 0 0; padding-bottom: env(safe-area-inset-bottom); }
}
.bol-sheet__frame { margin-inline: auto; overflow: hidden; border-radius: 12px; }
.bol-sheet__scaler { transform-origin: top left; }
.bol-sheet__options { display: flex; flex-wrap: wrap; align-items: center; gap: 12px 20px; margin-top: 16px; }
.bol-sheet__check { display: inline-flex; align-items: center; gap: 8px; font-size: var(--text-small); color: var(--color-text-secondary); cursor: pointer; }
</style>
