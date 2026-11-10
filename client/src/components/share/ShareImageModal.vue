<template>
  <AppModal
    v-model="open"
    size="md"
    class="share-sheet"
    :title="t('share.shareAsImage')"
    initial-focus=".share-sheet__tabs [role='tab'][aria-selected='true']"
  >
    <div class="share-sheet__tabs">
      <AppTabs v-model="active" :tabs="tabs" :controls="false" :label="t('share.templatesLabel')" />
    </div>

    <div class="share-sheet__stage" :style="{ minHeight: `${stageMinHeight}px` }" :aria-busy="status === 'loading'">
      <ShareTemplateCarousel
        v-if="status === 'ready'"
        ref="carousel"
        v-model="active"
        :templates="templates"
        :media="media"
        :poster-url="posterUrl"
        :avatar-url="avatarUrl"
        :theme="theme"
        :short-url="shortUrl"
        :stats="ownStats"
        :review="shareableReview"
        :max-height="previewMaxHeight"
      />
      <SkeletonBlock v-else-if="status === 'loading'" class="share-sheet__skeleton" height="100%" radius="12px" />
      <div v-else class="share-sheet__error" role="alert">
        <p>{{ t("share.imageLoadFailed") }}</p>
        <AppButton variant="outline" size="sm" @click="loadImages">{{ t("share.retry") }}</AppButton>
      </div>

      <div v-if="isExporting" class="share-sheet__busy">
        <AppSpinner :size="32" />
        <span>{{ t("share.generating") }}</span>
      </div>
    </div>

    <div class="share-sheet__meta">
      <ShareThemePicker v-if="activeDef.key !== 'review'" v-model="theme" />
      <span class="share-sheet__ratio">{{ activeDef.width }} × {{ activeDef.height }} · {{ activeDef.ratio }}</span>
    </div>

    <template #footer>
      <div class="share-sheet__actions">
        <AppButton v-if="canNativeShare" variant="primary" block :disabled="!ready || isExporting" @click="handleShare">
          <Share2 :size="16" aria-hidden="true" /> {{ t("share.shareTo") }}
        </AppButton>
        <AppButton :variant="canNativeShare ? 'outline' : 'primary'" block :disabled="!ready || isExporting" @click="handleDownload">
          <Download :size="16" aria-hidden="true" /> {{ t("share.downloadPng") }}
        </AppButton>
        <div class="share-sheet__secondary">
          <AppButton v-if="canCopyImg" variant="ghost" size="sm" :disabled="!ready || isExporting" @click="handleCopyImage">
            <Copy :size="15" aria-hidden="true" /> {{ t("share.copyImage") }}
          </AppButton>
          <AppButton variant="ghost" size="sm" @click="handleCopyLink">
            <Link2 :size="15" aria-hidden="true" /> {{ t("share.copyLink") }}
          </AppButton>
          <AppButton variant="ghost" size="sm" @click="handleCopyCaption">
            <MessageSquareText :size="15" aria-hidden="true" /> {{ t("share.copyCaption") }}
          </AppButton>
        </div>
      </div>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Copy, Download, Link2, MessageSquareText, Share2 } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import { resolveApiAssetUrl, useShareImage } from "@/composables/useShareImage"
import { getTmdbImageUrl } from "@/utils/image"
import type { ShareMediaContext, ShareReviewContext, ShareStatsContext, ShareTemplateKey, ShareTheme } from "@/types/share"
import ShareTemplateCarousel from "./ShareTemplateCarousel.vue"
import ShareThemePicker from "./ShareThemePicker.vue"
import { availableTemplates, hasOwnStats, publicExcerpt, shortShareUrl, templateDef } from "./shareCards"

const props = defineProps<{
  modelValue: boolean
  shareType: "poster" | "review"
  media: ShareMediaContext
  review?: ShareReviewContext
  /** The signed-in user's own stats. Ignored when sharing a review. */
  stats?: ShareStatsContext | null
}>()
const emit = defineEmits<{ "update:modelValue": [value: boolean] }>()

const DEFAULT_VIEWPORT_H = 800
const MIN_PREVIEW_H = 260
const MAX_PREVIEW_H = 560
/** Space taken by the sheet header, tabs, options and footer around the preview. */
const VIEWPORT_OFFSET_H = 400

const { t } = useI18n()
const share = useShareImage()
const isExporting = share.isExporting

const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })
const carousel = useTemplateRef<InstanceType<typeof ShareTemplateCarousel>>("carousel")

// Privacy: another user's stats never appear, and review text needs a public review.
const ownStats = computed(() => (props.shareType === "poster" && hasOwnStats(props.stats) ? props.stats : null))
const shareableReview = computed(() =>
  props.shareType === "review" && props.review && publicExcerpt(props.review.visibility, props.review.body) ? props.review : null,
)

const templates = computed(() => availableTemplates(props.shareType, ownStats.value, shareableReview.value))
const tabs = computed(() => templates.value.map((d) => ({ key: d.key, label: t(d.labelKey) })))
const active = ref<ShareTemplateKey>("poster")
const activeDef = computed(() => templates.value.find((d) => d.key === active.value) ?? templateDef("poster"))
const theme = ref<ShareTheme>("dark")

const shortUrl = computed(() => shortShareUrl(props.media))
const canNativeShare = ref(false)
const canCopyImg = ref(false)

// Preview height follows the viewport so the sheet fits on small phones.
const viewportH = ref(typeof window === "undefined" ? DEFAULT_VIEWPORT_H : window.innerHeight)
const previewMaxHeight = computed(() => Math.max(MIN_PREVIEW_H, Math.min(MAX_PREVIEW_H, viewportH.value - VIEWPORT_OFFSET_H)))
const stageMinHeight = computed(() => previewMaxHeight.value)
function onResize() {
  viewportH.value = window.innerHeight
}

// Images: TMDB goes through the API proxy as data URLs before any card renders.
const status = ref<"loading" | "ready" | "error">("loading")
const posterUrl = ref<string | null>(null)
const avatarUrl = ref<string | null>(null)
const ready = computed(() => status.value === "ready")
let loadSeq = 0

async function loadImages() {
  const seq = ++loadSeq
  status.value = "loading"
  const source = props.media.posterUrl || getTmdbImageUrl(props.media.posterPath, "w780")

  const avatar = props.shareType === "review" ? resolveApiAssetUrl(props.review?.authorAvatarUrl) : ""

  const [poster, avatarData] = await Promise.all([
    source ? share.fetchImageAsBlob(source) : Promise.resolve(null),
    avatar ? share.fetchImageAsBlob(avatar) : Promise.resolve(null),
  ])
  if (seq !== loadSeq || !props.modelValue) return
  posterUrl.value = poster
  avatarUrl.value = avatarData
  // A title with no poster at all renders the fallback; a poster that failed to load is an error.
  status.value = source && !poster ? "error" : "ready"
}

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      active.value = templates.value[0]!.key
      canNativeShare.value = share.canShareFiles()
      canCopyImg.value = share.canCopyImage()
      onResize()
      window.addEventListener("resize", onResize)
      loadImages()
    } else {
      loadSeq++
      window.removeEventListener("resize", onResize)
      posterUrl.value = null
      avatarUrl.value = null
      status.value = "loading"
      share.cleanup()
    }
  },
  { immediate: true },
)

// Keep the selection valid if the offered templates change while open.
watch(templates, (list) => {
  if (!list.some((d) => d.key === active.value)) active.value = list[0]!.key
})

onBeforeUnmount(() => {
  loadSeq++
  window.removeEventListener("resize", onResize)
  share.cleanup()
})

async function exportActive(): Promise<{ blob: Blob; filename: string } | null> {
  const def = activeDef.value
  const node = carousel.value?.getNode(def.key)
  if (!node) return null
  const blob = await share.renderPng(node, { width: def.width, height: def.height })
  if (!blob) return null
  return { blob, filename: share.sanitizeFilename(props.media.title, `POPCORN_${def.key}`) }
}

function caption(): string {
  const r = shareableReview.value
  if (active.value === "review" && r) {
    return share.buildReviewCaption(props.media.title, r.authorDisplayName || r.authorUsername, r.rating)
  }
  return share.buildPosterCaption(props.media.title, props.media.releaseYear, props.media.mediaType)
}

async function handleShare() {
  const out = await exportActive()
  if (!out) return
  const link = share.buildShareLink(props.media.mediaType, props.media.id)
  await share.shareImageFile(out.blob, out.filename, `${caption()}\n${link}`)
}

async function handleDownload() {
  const out = await exportActive()
  if (out) share.saveBlob(out.blob, out.filename)
}

async function handleCopyImage() {
  const out = await exportActive()
  if (out) await share.copyImage(out.blob)
}

function handleCopyCaption() {
  share.copyText(caption(), "share.captionCopied")
}

function handleCopyLink() {
  share.copyText(share.buildShareLink(props.media.mediaType, props.media.id), "share.linkCopied")
}
</script>

<style>
.share-sheet__tabs { margin-bottom: 12px; overflow-x: auto; }
.share-sheet__stage { position: relative; display: flex; align-items: stretch; justify-content: center; }
.share-sheet__stage > * { flex: 1; min-width: 0; }
.share-sheet__skeleton { flex: 1; }
.share-sheet__error {
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; padding: 24px;
  border-radius: 12px; background: var(--color-surface); color: var(--color-text-secondary); font-size: 14px; text-align: center;
}
.share-sheet__busy {
  position: absolute; inset: 0; z-index: 2; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px;
  border-radius: 12px; background: color-mix(in srgb, black 70%, transparent); color: var(--color-text-primary); font-size: 14px;
}
.share-sheet__meta { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: 16px; }
.share-sheet__ratio { margin-left: auto; color: var(--color-text-muted); font-size: 12px; }
.share-sheet__actions { display: flex; flex-direction: column; gap: 8px; width: 100%; }
.share-sheet__secondary { display: flex; flex-wrap: wrap; justify-content: center; gap: 4px; }

/* Bottom sheet below 768px. */
@media (max-width: 767px) {
  .app-modal:has(> .share-sheet) { align-items: flex-end; padding: 0; }
  .share-sheet {
    max-width: none;
    max-height: 94dvh;
    border-bottom: 0;
    border-radius: 20px 20px 0 0;
    padding-bottom: env(safe-area-inset-bottom);
  }
}
</style>
