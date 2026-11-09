<template>
  <div class="ap">
    <AppAvatar :src="preview ?? src" :name="name" size="xl" decorative />
    <div class="ap__side">
      <AppButton variant="outline" size="sm" @click="fileInput?.click()">
        <Camera :size="14" aria-hidden="true" />{{ t("settings.avatar.change") }}
      </AppButton>
      <p class="ap__hint">{{ t("settings.avatar.hint") }}</p>
      <p v-if="fileError" class="ap__error" role="alert">{{ t(`settings.avatar.error.${fileError}`) }}</p>
    </div>
    <input ref="fileInput" type="file" class="sr-only" tabindex="-1" :accept="AVATAR_ACCEPTED_TYPES.join(',')" @change="onFile" />

    <AppModal v-model="cropOpen" :title="t('settings.avatar.cropTitle')" size="sm" @close="cleanup">
      <div
        ref="viewportEl"
        class="ap__viewport"
        tabindex="0"
        role="group"
        :aria-label="t('settings.avatar.cropHelp')"
        :style="{ width: `${VIEWPORT}px`, height: `${VIEWPORT}px` }"
        @pointerdown="startDrag"
        @pointermove="drag"
        @pointerup="endDrag"
        @pointercancel="endDrag"
        @keydown="onKey"
      >
        <img
          v-if="imageUrl"
          ref="imgEl"
          :src="imageUrl"
          alt=""
          draggable="false"
          class="ap__img"
          :style="imgStyle"
          @load="onImageLoad"
        />
      </div>
      <label class="ap__zoom">
        <span>{{ t("settings.avatar.zoom") }}</span>
        <input type="range" min="1" :max="MAX_ZOOM" step="0.05" :value="crop.zoom" @input="setZoom(Number(($event.target as HTMLInputElement).value))" />
      </label>
      <p class="ap__hint">{{ t("settings.avatar.cropHelp") }}</p>
      <div class="ap__actions">
        <AppButton variant="ghost" @click="cropOpen = false">{{ t("common.cancel") }}</AppButton>
        <AppButton variant="primary" :loading="applying" :disabled="!ready" @click="apply">{{ t("settings.avatar.apply") }}</AppButton>
      </div>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Camera } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { AVATAR_ACCEPTED_TYPES, MAX_ZOOM, clampCrop, cropScale, cropToBlob, validateAvatarFile, type AvatarFileError } from "./avatarCrop"

defineProps<{ src: string | null; preview: string | null; name: string }>()
const emit = defineEmits<{ pick: [blob: Blob] }>()
const { t } = useI18n()

const VIEWPORT = 280
const KEY_STEP = 12

const fileInput = ref<HTMLInputElement | null>(null)
const imgEl = ref<HTMLImageElement | null>(null)
const cropOpen = ref(false)
const imageUrl = ref<string | null>(null)
const natural = reactive({ w: 0, h: 0 })
const crop = reactive({ zoom: 1, x: 0, y: 0 })
const fileError = ref<AvatarFileError | null>(null)
const applying = ref(false)
const ready = computed(() => natural.w > 0)

const imgStyle = computed(() => {
  if (!ready.value) return {}
  const scale = cropScale(natural.w, natural.h, VIEWPORT, crop.zoom)
  return {
    width: `${natural.w * scale}px`,
    height: `${natural.h * scale}px`,
    transform: `translate(calc(-50% + ${crop.x}px), calc(-50% + ${crop.y}px))`,
  }
})

function cleanup() {
  if (imageUrl.value) URL.revokeObjectURL(imageUrl.value)
  imageUrl.value = null
  natural.w = natural.h = 0
}

function onFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ""
  if (!file) return
  fileError.value = validateAvatarFile(file)
  if (fileError.value) return
  cleanup()
  Object.assign(crop, { zoom: 1, x: 0, y: 0 })
  imageUrl.value = URL.createObjectURL(file)
  cropOpen.value = true
}

function onImageLoad() {
  natural.w = imgEl.value?.naturalWidth ?? 0
  natural.h = imgEl.value?.naturalHeight ?? 0
}

function update(next: { zoom: number; x: number; y: number }) {
  if (!ready.value) return
  Object.assign(crop, clampCrop(next, natural.w, natural.h, VIEWPORT))
}
const setZoom = (zoom: number) => update({ ...crop, zoom })

let dragFrom: { px: number; py: number; x: number; y: number } | null = null
function startDrag(e: PointerEvent) {
  dragFrom = { px: e.clientX, py: e.clientY, x: crop.x, y: crop.y }
  ;(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)
}
function drag(e: PointerEvent) {
  if (dragFrom) update({ zoom: crop.zoom, x: dragFrom.x + e.clientX - dragFrom.px, y: dragFrom.y + e.clientY - dragFrom.py })
}
function endDrag() {
  dragFrom = null
}

function onKey(e: KeyboardEvent) {
  const moves: Record<string, [number, number]> = {
    ArrowLeft: [-KEY_STEP, 0],
    ArrowRight: [KEY_STEP, 0],
    ArrowUp: [0, -KEY_STEP],
    ArrowDown: [0, KEY_STEP],
  }
  const move = moves[e.key]
  if (move) {
    e.preventDefault()
    update({ zoom: crop.zoom, x: crop.x + move[0], y: crop.y + move[1] })
  } else if (e.key === "+" || e.key === "=") {
    setZoom(crop.zoom + 0.1)
  } else if (e.key === "-") {
    setZoom(crop.zoom - 0.1)
  }
}

async function apply() {
  if (!imgEl.value || !ready.value) return
  applying.value = true
  try {
    emit("pick", await cropToBlob(imgEl.value, crop, VIEWPORT))
    cropOpen.value = false
  } catch {
    fileError.value = "type"
  } finally {
    applying.value = false
  }
}
</script>

<style scoped>
.ap { display: flex; align-items: center; gap: 1rem; }
.ap__side { display: flex; flex-direction: column; align-items: flex-start; gap: 0.375rem; }
.ap__hint { font-size: var(--text-caption); color: var(--color-text-muted); }
.ap__error { font-size: var(--text-caption); color: var(--color-danger); }
.ap__viewport { position: relative; max-width: 100%; margin: 0 auto; overflow: hidden; touch-action: none; cursor: grab; background: var(--color-surface-3); border-radius: var(--radius-card); }
.ap__viewport:active { cursor: grabbing; }
.ap__img { position: absolute; top: 50%; left: 50%; max-width: none; user-select: none; }
.ap__zoom { display: flex; align-items: center; gap: 0.75rem; margin-top: 0.75rem; font-size: 0.8125rem; color: var(--color-text-secondary); }
.ap__zoom input { flex: 1; }
.ap__actions { display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 1rem; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
</style>
