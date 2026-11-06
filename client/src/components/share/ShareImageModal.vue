<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div
        v-if="modelValue"
        class="share-modal-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="$t('share.shareAsImage')"
        @click.self="close"
        @keydown.esc="close"
        tabindex="-1"
        ref="overlayRef"
      >
        <div class="share-modal">
          <!-- Header -->
          <div class="modal-header">
            <h2 class="modal-title">
              <i class="pi pi-share-alt" />
              {{ shareType === 'poster' ? $t('share.posterShare') : $t('share.reviewShare') }}
            </h2>
            <button class="btn-close" @click="close" :aria-label="$t('common.close') || 'Close'">
              <i class="pi pi-times" />
            </button>
          </div>

          <!-- Preview area -->
          <div class="preview-container">
            <div class="preview-scaler" :style="scalerStyle">
              <div ref="exportNodeRef">
                <PosterShareTemplate
                  v-if="shareType === 'poster'"
                  :media="media"
                  :poster-blob-url="posterBlobUrl"
                />
                <ReviewShareTemplate
                  v-if="shareType === 'review' && review"
                  :media="media"
                  :review="review"
                  :poster-blob-url="posterBlobUrl"
                  :avatar-blob-url="avatarBlobUrl"
                />
              </div>
            </div>

            <!-- Loading overlay -->
            <div v-if="isExporting" class="export-overlay">
              <AppSpinner :size="36" class="export-overlay__spinner" />
              <span>{{ $t('share.generating') }}</span>
            </div>
          </div>

          <!-- Actions -->
          <div class="modal-actions">
            <button
              class="btn-action btn-download"
              :disabled="isExporting"
              @click="handleDownload"
            >
              <i class="pi pi-download" />
              {{ $t('share.downloadImage') }}
            </button>
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

          <!-- Format label -->
          <div class="format-label">
            <i class="pi pi-mobile" />
            {{ $t('share.storyFormat') }} · 9:16
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import { ref, watch, computed, onUnmounted, nextTick } from "vue"
  import type { ShareMediaContext, ShareReviewContext } from "@/types/share"
  import { useShareImage } from "@/composables/useShareImage"
  import { getTmdbImageUrl } from "@/utils/image"
  import PosterShareTemplate from "./PosterShareTemplate.vue"
  import ReviewShareTemplate from "./ReviewShareTemplate.vue"

  function getApiBaseUrl(): string {
    const envUrl = import.meta.env.VITE_API_BASE_URL
    if (!envUrl) return "/api"
    const cleanUrl = (envUrl as string).replace(/\/+$/, "")
    return cleanUrl.endsWith("/api") ? cleanUrl : `${cleanUrl}/api`
  }

  const props = defineProps<{
    modelValue: boolean
    shareType: "poster" | "review"
    media: ShareMediaContext
    review?: ShareReviewContext
  }>()

  const emit = defineEmits<{
    "update:modelValue": [value: boolean]
  }>()

  const share = useShareImage()

  const overlayRef = ref<HTMLElement | null>(null)
  const exportNodeRef = ref<HTMLElement | null>(null)

  const posterBlobUrl = ref<string | null>(null)
  const avatarBlobUrl = ref<string | null>(null)
  const isExporting = share.isExporting

  // Scale the 1080px-wide export node to fit the modal preview (~360px)
  const PREVIEW_WIDTH = 360
  const scalerStyle = computed(() => ({
    transform: `scale(${PREVIEW_WIDTH / 1080})`,
    transformOrigin: "top left",
    width: "1080px",
    height: "1920px",
  }))

  function close() {
    emit("update:modelValue", false)
  }

  /** Load images as blob URLs when modal opens */
  async function loadImages() {
    share.resetFallbackTracking()

    // Load poster
    const posterUrl =
      props.media.posterUrl ||
      getTmdbImageUrl(props.media.posterPath, "w500")

    if (posterUrl) {
      const blob = await share.fetchImageAsBlob(posterUrl)
      if (blob) {
        posterBlobUrl.value = blob
      } else {
        posterBlobUrl.value = null
        share.markFallbackUsed()
      }
    } else {
      posterBlobUrl.value = null
    }

    // Load avatar for review shares
    if (props.shareType === "review" && props.review?.authorAvatarUrl) {
      let avatarUrl = props.review.authorAvatarUrl.trim()
      if (
        avatarUrl &&
        !avatarUrl.startsWith("http://") &&
        !avatarUrl.startsWith("https://") &&
        !avatarUrl.startsWith("data:")
      ) {
        const apiBase = getApiBaseUrl()
        if (apiBase.startsWith("http")) {
          const origin = new URL(apiBase).origin
          avatarUrl = `${origin}${avatarUrl.startsWith("/") ? "" : "/"}${avatarUrl}`
        } else {
          avatarUrl = `${window.location.origin}${avatarUrl.startsWith("/") ? "" : "/"}${avatarUrl}`
        }
      }

      if (avatarUrl) {
        const blob = await share.fetchImageAsBlob(avatarUrl)
        if (blob) {
          avatarBlobUrl.value = blob
        } else {
          avatarBlobUrl.value = null
          share.markFallbackUsed()
        }
      }
    }

    share.showFallbackWarningIfNeeded()
  }

  async function handleDownload() {
    if (!exportNodeRef.value) return
    const blob = await share.generatePng(exportNodeRef.value)
    if (blob) {
      const filename = share.sanitizeFilename(
        props.media.title,
        props.shareType === "poster" ? "POPCORN_poster" : "POPCORN_review",
      )
      share.downloadBlob(blob, filename)
    }
  }

  function handleCopyCaption() {
    const caption =
      props.shareType === "poster"
        ? share.buildPosterCaption(
            props.media.title,
            props.media.releaseYear,
            props.media.mediaType,
          )
        : share.buildReviewCaption(
            props.media.title,
            props.review?.authorDisplayName || props.review?.authorUsername || "",
            props.review?.rating || 0,
          )
    share.copyText(caption, "share.captionCopied")
  }

  function handleCopyLink() {
    const link = share.buildShareLink(props.media.mediaType, props.media.id)
    share.copyText(link, "share.linkCopied")
  }

  // Body scroll lock & image loading
  watch(
    () => props.modelValue,
    async (visible) => {
      if (visible) {
        document.body.style.overflow = "hidden"
        await nextTick()
        loadImages()
        overlayRef.value?.focus()
      } else {
        document.body.style.overflow = ""
        posterBlobUrl.value = null
        avatarBlobUrl.value = null
        share.cleanup()
      }
    },
    { immediate: true },
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
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(12px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
  }

  .share-modal {
    background: #1a1a1a;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    max-width: 440px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    box-shadow: 0 24px 80px rgba(0, 0, 0, 0.6);
  }

  /* ── Header ── */
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .modal-title {
    color: #ffffff;
    font-size: 1rem;
    font-weight: 700;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .modal-title i {
    color: #f5432c;
  }

  .btn-close {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    color: #ffffff;
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background 0.15s;
  }

  .btn-close:hover {
    background: rgba(255, 255, 255, 0.12);
  }

  /* ── Preview ── */
  .preview-container {
    position: relative;
    margin: 20px 24px;
    border-radius: 12px;
    overflow: hidden;
    background: #0a0a0a;
    /* Preview dimensions set by scaler */
    width: 360px;
    height: 640px;
    align-self: center;
  }

  .preview-scaler {
    pointer-events: none;
  }

  .export-overlay {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    color: #ffffff;
    font-size: 0.875rem;
    z-index: 2;
  }

  .export-overlay__spinner {
  color: var(--color-brand);
}

  /* ── Actions ── */
  .modal-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 0 24px 16px;
  }

  .btn-action {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 16px;
    border-radius: 10px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.15s;
    border: none;
  }

  .btn-download {
    background: #f5432c;
    color: #ffffff;
  }

  .btn-download:hover:not(:disabled) {
    background: #ff5a42;
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
    background: rgba(255, 255, 255, 0.1);
  }

  /* ── Format label ── */
  .format-label {
    text-align: center;
    padding: 12px 24px 20px;
    color: rgba(255, 255, 255, 0.4);
    font-size: 0.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }

  /* ── Transition ── */
  .modal-fade-enter-active,
  .modal-fade-leave-active {
    transition: opacity 0.2s ease;
  }

  .modal-fade-enter-from,
  .modal-fade-leave-to {
    opacity: 0;
  }
</style>
