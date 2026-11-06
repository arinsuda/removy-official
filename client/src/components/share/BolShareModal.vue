<template>
  <Teleport to="body">
    <Transition
      name="modal-fade"
      @after-enter="onModalEntered"
      @after-leave="onModalLeft"
    >
      <div
        v-if="modelValue"
        class="share-modal-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="$t('share.bolShareTitle')"
        @click.self="close"
        @keydown.esc="close"
        tabindex="-1"
        ref="overlayRef"
      >
        <div class="share-modal">
          <!-- Header -->
          <div class="modal-header">
            <h2 class="modal-title">
              <i class="pi pi-trophy" />
              {{ $t('share.bolShareTitle') }}
            </h2>
            <button class="btn-close" @click="close" :aria-label="$t('common.close') || 'Close'">
              <i class="pi pi-times" />
            </button>
          </div>

          <!-- Controls Section: Username Toggle -->
          <div class="controls-section">
            <div class="toggle-row">
              <span class="toggle-label">{{ $t('share.showUsername') }}</span>
              <label class="toggle-switch">
                <input type="checkbox" v-model="showUsername" />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <!-- Preview Area -->
          <div class="preview-container">
            <div class="preview-scaler" :style="scalerStyle">
              <div ref="exportNodeRef">
                <BolShareTemplate
                  :media-type="mediaType"
                  :ranks="ranks"
                  :username="username"
                  :show-username="showUsername"
                  :poster-blob-urls="posterBlobUrls"
                  :dominant-colors="dominantColors"
                  :english-titles="englishTitles"
                />
              </div>
            </div>

            <!-- Loading overlay -->
            <div v-if="isPreparing || isExporting" class="export-overlay">
              <i class="pi pi-spin pi-spinner" />
              <span>{{ isPreparing ? $t('share.preparingImage') : $t('share.generating') }}</span>
            </div>
          </div>

          <!-- Actions -->
          <div class="modal-actions">
            <button
              class="btn-action btn-download"
              :disabled="isPreparing || isExporting"
              @click="handleDownload"
            >
              <i class="pi pi-download" />
              {{ $t('share.downloadImage') }}
            </button>
            <div class="actions-row">
              <button
                class="btn-action btn-secondary"
                @click="handleCopyCaption"
              >
                <i class="pi pi-copy" />
                {{ $t('share.copyCaption') }}
              </button>
              <button
                class="btn-action btn-secondary"
                @click="handleCopyLink"
              >
                <i class="pi pi-link" />
                {{ $t('share.copyLink') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, computed, onUnmounted, nextTick } from "vue"
import type { RankGroup } from "./BolShareTemplate.vue"
import type { MediaType } from "@/types/common"
import type { ExtractedColor } from "@/utils/extractDominantColor"
import { extractDominantColor } from "@/utils/extractDominantColor"
import { getTmdbImageUrl } from "@/utils/image"
import { movieApi } from "@/api/endpoints/movie"
import { useShareImage } from "@/composables/useShareImage"
import BolShareTemplate from "./BolShareTemplate.vue"

const props = defineProps<{
  modelValue: boolean
  mediaType: MediaType
  ranks: RankGroup[]
  username?: string
}>()

const emit = defineEmits<{
  "update:modelValue": [value: boolean]
}>()

const share = useShareImage()

const overlayRef = ref<HTMLElement | null>(null)
const exportNodeRef = ref<HTMLElement | null>(null)

const showUsername = ref(true)

const posterBlobUrls = ref<Record<string, string>>({})
const dominantColors = ref<Record<string, ExtractedColor>>({})
const englishTitles = ref<Record<string, string>>({})
const isPreparing = ref(false)
const isExporting = share.isExporting

// Scale the 1080×1920 export node to 460px height for perfect compact fit without scrollbars
const PREVIEW_HEIGHT = 460
const PREVIEW_SCALE = PREVIEW_HEIGHT / 1920
const scalerStyle = computed(() => ({
  transform: `scale(${PREVIEW_SCALE})`,
  transformOrigin: "top left",
  width: "1080px",
  height: "1920px",
}))

function close() {
  emit("update:modelValue", false)
}

// Persistent session cache across modal opens
const sessionBlobCache: Record<string, string> = {}
const sessionColorCache: Record<string, ExtractedColor> = {}
const sessionEnglishTitleCache: Record<string, string> = {}

/** Pre-load poster blobs, English titles & extract dominant colors with single-pass batching */
async function loadMediaData() {
  share.resetFallbackTracking()

  try {
    const topRanks = props.ranks.slice(0, 5)
    const loadPromises: Promise<void>[] = []
    const newBlobUrls: Record<string, string> = { ...sessionBlobCache }
    const newColors: Record<string, ExtractedColor> = { ...sessionColorCache }
    const newEnglishTitles: Record<string, string> = { ...sessionEnglishTitleCache }

    let needsFetch = false

    for (const group of topRanks) {
      for (const item of group.items) {
        const key = `${item.media_type}:${item.media.id}`

        // Pre-fetch English title if not cached
        if (!sessionEnglishTitleCache[key]) {
          needsFetch = true
          const titlePromise = (async () => {
            try {
              if (item.media_type === "movie") {
                const res = await movieApi.getById(item.media.id)
                const m = res.data?.movie
                if (m) {
                  const enTitle = m.english_title || m.original_title || m.title || ""
                  sessionEnglishTitleCache[key] = enTitle
                  newEnglishTitles[key] = enTitle
                }
              } else {
                const res = await movieApi.getSeriesById(item.media.id)
                const s = res.data?.series
                if (s) {
                  const enTitle = s.english_title || s.original_name || s.name || ""
                  sessionEnglishTitleCache[key] = enTitle
                  newEnglishTitles[key] = enTitle
                }
              }
            } catch {
              /* ignore fetch error */
            }
          })()
          loadPromises.push(titlePromise)
        }

        // Pre-load poster blob & color if not cached
        if (sessionBlobCache[key] && sessionColorCache[key]) continue

        let posterUrl = item.media.poster_url
        if (!posterUrl) continue

        if (!posterUrl.startsWith("http://") && !posterUrl.startsWith("https://") && !posterUrl.startsWith("data:")) {
          posterUrl = getTmdbImageUrl(posterUrl, "w500") || posterUrl
        }

        needsFetch = true
        const promise = (async () => {
          const blobUrl = await share.fetchImageAsBlob(posterUrl)
          if (blobUrl) {
            sessionBlobCache[key] = blobUrl
            newBlobUrls[key] = blobUrl
            const color = await extractDominantColor(blobUrl)
            sessionColorCache[key] = color
            newColors[key] = color
          } else {
            share.markFallbackUsed()
          }
        })()

        loadPromises.push(promise)
      }
    }

    if (needsFetch) {
      isPreparing.value = true
      await Promise.all(loadPromises)
    }

    // Batch reactive update in 1 single pass
    posterBlobUrls.value = newBlobUrls
    dominantColors.value = newColors
    englishTitles.value = newEnglishTitles
  } catch (err) {
    console.error("Error loading media for BOL share:", err)
  } finally {
    isPreparing.value = false
  }
}

async function handleDownload() {
  if (!exportNodeRef.value || isPreparing.value || isExporting.value) return

  // Yield to main thread so pointer event paints loading state immediately (<16ms INP)
  isPreparing.value = true
  await new Promise((resolve) => setTimeout(resolve, 0))

  try {
    const blob = await share.generatePng(exportNodeRef.value)
    if (blob) {
      const label = props.mediaType === "movie" ? "Movies" : "TV"
      const filename = share.sanitizeFilename(
        `BOL_${label}_Top5`,
        "POPCORN_BOL"
      )
      share.downloadBlob(blob, filename)
    }
  } finally {
    isPreparing.value = false
  }
}

function handleCopyCaption() {
  const mediaLabel = props.mediaType === "movie" ? "Movies" : "TV Series"
  let text = `🏆 My Top Best ${mediaLabel} of Life on POPCORN ✨\n\n`

  const topRanks = props.ranks.slice(0, 5)
  topRanks.forEach((group) => {
    const titles = group.items.map((i) => i.media.title).join(", ")
    text += `#${group.rank}: ${titles}\n`
  })

  if (showUsername.value && props.username) {
    text += `\nCurated by @${props.username}\n`
  }

  text += `\nShared from POPCORN ✨`
  share.copyText(text, "share.captionCopied")
}

function handleCopyLink() {
  const link = `${window.location.origin}/library`
  share.copyText(link, "share.linkCopied")
}

// Transition hooks to prevent JS execution during CSS animation frames
function onModalEntered() {
  overlayRef.value?.focus()
  loadMediaData()
}

function onModalLeft() {
  share.cleanup()
}

// Watch modal state
watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      document.body.style.overflow = "hidden"
    } else {
      document.body.style.overflow = ""
    }
  },
  { immediate: true }
)

onUnmounted(() => {
  document.body.style.overflow = ""
  share.cleanup()
})
</script>

<style scoped>
.share-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(10, 10, 14, 0.88);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  touch-action: manipulation;
  will-change: opacity;
}

.share-modal {
  background: #18181c;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 20px;
  max-width: 440px;
  width: 100%;
  display: flex;
  flex-direction: column;
  box-shadow: 0 28px 90px rgba(0, 0, 0, 0.7);
  touch-action: manipulation;
  will-change: transform, opacity;
  transform: translateZ(0);
  overflow: hidden;
}

/* Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.modal-title {
  color: #ffffff;
  font-size: 1.05rem;
  font-weight: 700;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-title i {
  color: #ffd700;
}

.btn-close {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  color: #ffffff;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  touch-action: manipulation;
}

.btn-close:hover {
  background: rgba(255, 255, 255, 0.16);
  transform: scale(1.05);
}

.btn-close:active {
  transform: scale(0.95);
}

/* Controls */
.controls-section {
  padding: 12px 20px 4px 20px;
  display: flex;
  flex-direction: column;
}

.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2px;
}

.toggle-label {
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.8);
  font-weight: 500;
}

/* Custom Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 42px;
  height: 22px;
  touch-action: manipulation;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background-color: rgba(255, 255, 255, 0.15);
  transition: background-color 0.3s ease;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  border-radius: 50%;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
}

input:checked + .slider {
  background-color: #f5432c;
}

input:checked + .slider:before {
  transform: translateX(20px);
}

/* Preview Area */
.preview-container {
  position: relative;
  margin: 6px auto;
  border-radius: 14px;
  overflow: hidden;
  background: #08080a;
  width: 259px;
  height: 460px;
  align-self: center;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  flex-shrink: 0;
  contain: strict;
  contain-intrinsic-size: 259px 460px;
  transform: translateZ(0);
  will-change: transform;
}

.preview-scaler {
  pointer-events: none;
  transform-origin: top left;
  will-change: transform;
}

.export-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #ffffff;
  font-size: 0.9rem;
  z-index: 5;
}

.export-overlay .pi-spinner {
  font-size: 2.2rem;
  color: #f5432c;
}

/* Actions */
.modal-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 20px 20px 20px;
}

.actions-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.actions-row .btn-action {
  flex: 1;
}

.btn-action {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1);
  border: none;
  position: relative;
  touch-action: manipulation;
}

.btn-download {
  background: #f5432c;
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(245, 67, 44, 0.3);
  width: 100%;
}

.btn-download:hover:not(:disabled) {
  background: #ff5a42;
  transform: translateY(-2px);
  box-shadow: 0 6px 18px rgba(245, 67, 44, 0.45);
}

.btn-download:active:not(:disabled) {
  transform: translateY(0) scale(0.98);
}

.btn-download:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.05);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.2);
  transform: translateY(-1px);
}

.btn-secondary:active {
  transform: translateY(0) scale(0.98);
}

/* High-Performance 60 FPS Modal Transitions */
.modal-fade-enter-active {
  transition: opacity 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.modal-fade-leave-active {
  transition: opacity 0.2s cubic-bezier(0.7, 0, 0.84, 0);
}
.modal-fade-enter-active .share-modal {
  transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.25s ease;
}
.modal-fade-leave-active .share-modal {
  transition: transform 0.2s cubic-bezier(0.7, 0, 0.84, 0), opacity 0.2s ease;
}
.modal-fade-enter-from {
  opacity: 0;
}
.modal-fade-enter-from .share-modal {
  opacity: 0;
  transform: scale(0.93) translateY(8px);
}
.modal-fade-leave-to {
  opacity: 0;
}
.modal-fade-leave-to .share-modal {
  opacity: 0;
  transform: scale(0.95) translateY(4px);
}
</style>
