<template>
  <AppModal v-model="open" size="md" class="ach-sheet" :title="t('achievements.share.title')">
    <AppTabs v-model="active" :tabs="tabs" :controls="false" :label="t('share.templatesLabel')" class="ach-sheet__tabs" />

    <div ref="stage" class="ach-sheet__stage" :style="{ minHeight: `${frameHeight}px` }" :aria-busy="status === 'loading'">
      <div
        v-if="status === 'ready' && achievement"
        class="ach-sheet__frame"
        :style="{ width: `${def.width * scale}px`, height: `${def.height * scale}px` }"
      >
        <div
          class="ach-sheet__scaler"
          :style="{ width: `${def.width}px`, height: `${def.height}px`, transform: `scale(${scale})` }"
        >
          <AchievementStoryCard
            v-if="active === 'story'"
            ref="card"
            :achievement="achievement"
            :tiers="tiers"
            :username="shownUsername"
            :avatar-url="avatarDataUrl"
            :theme="theme"
            :short-url="shortUrl"
          />
          <AchievementSquareCard
            v-else
            ref="card"
            :achievement="achievement"
            :tiers="tiers"
            :username="shownUsername"
            :avatar-url="avatarDataUrl"
            :theme="theme"
            :short-url="shortUrl"
          />
        </div>
      </div>
      <SkeletonBlock v-else class="ach-sheet__skeleton" height="100%" radius="var(--radius-card)" />
      <div v-if="isExporting" class="ach-sheet__busy">
        <AppSpinner :size="32" />
        <span>{{ t("share.generating") }}</span>
      </div>
    </div>

    <div class="ach-sheet__options">
      <ShareThemePicker v-model="theme" />
      <label v-if="username" class="ach-sheet__check">
        <input v-model="showUsername" type="checkbox" />{{ t("share.showUsername") }}
      </label>
      <span class="ach-sheet__ratio">{{ def.width }} × {{ def.height }} · {{ def.ratio }}</span>
    </div>

    <template #footer>
      <div class="ach-sheet__actions">
        <AppButton v-if="canNativeShare" block :disabled="status !== 'ready' || isExporting" @click="handleShare">
          <Share2 :size="16" aria-hidden="true" /> {{ t("share.shareTo") }}
        </AppButton>
        <AppButton
          :variant="canNativeShare ? 'outline' : 'primary'"
          block
          :disabled="status !== 'ready' || isExporting"
          @click="handleDownload"
        >
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
import { resolveApiAssetUrl, useShareImage } from "@/composables/useShareImage"
import type { UserAchievement } from "@/types/achievement"
import type { ShareTheme } from "@/types/share"
import AchievementSquareCard from "./cards/AchievementSquareCard.vue"
import AchievementStoryCard from "./cards/AchievementStoryCard.vue"
import ShareThemePicker from "./ShareThemePicker.vue"
import { ACHIEVEMENT_TEMPLATES, achievementTemplateDef, type AchievementTemplateKey } from "./shareCards"

const DEFAULT_STAGE_WIDTH = 320
const DEFAULT_VIEWPORT_H = 800
const MIN_PREVIEW_H = 260
const MAX_PREVIEW_H = 560
const VIEWPORT_OFFSET_H = 420
const FALLBACK_FILENAME = "achievement"

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    /** Unlocked achievement owned by the signed-in user. */
    achievement: UserAchievement | null
    tiers?: UserAchievement[]
    username?: string
    avatarUrl?: string | null
  }>(),
  {
    tiers: () => [],
    username: "",
    avatarUrl: null,
  },
)
const emit = defineEmits<{ "update:modelValue": [value: boolean] }>()

const { t } = useI18n()
const share = useShareImage()
const isExporting = share.isExporting
const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })

const active = ref<AchievementTemplateKey>("story")
const theme = ref<ShareTheme>("dark")
const showUsername = ref(true)
const status = ref<"loading" | "ready">("loading")
const avatarDataUrl = ref<string | null>(null)
const canNativeShare = ref(false)
const card = useTemplateRef<{ $el: HTMLElement }>("card")
const stage = useTemplateRef<HTMLElement>("stage")
const stageWidth = ref(DEFAULT_STAGE_WIDTH)
let loadSeq = 0
let resizeObserver: ResizeObserver | null = null

const tabs = computed(() => ACHIEVEMENT_TEMPLATES.map((d) => ({ key: d.key, label: t(d.labelKey) })))
const def = computed(() => achievementTemplateDef(active.value))
const shownUsername = computed(() => (showUsername.value ? (props.username ?? "") : ""))
const shortUrl = computed(() => window.location.host)

const viewportH = ref(typeof window === "undefined" ? DEFAULT_VIEWPORT_H : window.innerHeight)
const maxHeight = computed(() => Math.max(MIN_PREVIEW_H, Math.min(MAX_PREVIEW_H, viewportH.value - VIEWPORT_OFFSET_H)))
const scale = computed(() => Math.min(stageWidth.value / def.value.width, maxHeight.value / def.value.height))
const frameHeight = computed(() => def.value.height * scale.value)

function onResize() {
  viewportH.value = window.innerHeight
  stageWidth.value = stage.value?.clientWidth || stageWidth.value
}

// The avatar is inlined as a data URL so the export canvas is never tainted.
// It loads the same way as on review share cards; on failure the card falls back to initials.
async function loadAvatar() {
  const seq = ++loadSeq
  status.value = "loading"
  const source = resolveApiAssetUrl(props.avatarUrl)
  const resolved = source ? await share.fetchImageAsBlob(source) : null
  if (seq !== loadSeq || !props.modelValue) return
  avatarDataUrl.value = resolved
  status.value = "ready"
}

watch(
  () => [props.modelValue, props.avatarUrl] as const,
  ([visible]) => {
    if (visible) {
      active.value = "story"
      canNativeShare.value = share.canShareFiles()
      window.addEventListener("resize", onResize)
      onResize()
      void loadAvatar()
    } else {
      loadSeq++
      window.removeEventListener("resize", onResize)
      avatarDataUrl.value = null
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
  const name = props.achievement?.achievement.name ?? FALLBACK_FILENAME
  if (!node) return null
  const blob = await share.renderPng(node, { width: def.value.width, height: def.value.height })
  return blob ? { blob, filename: share.sanitizeFilename(name, `POPCORN_${active.value}`) } : null
}

async function handleShare() {
  const out = await exportCard()
  const name = props.achievement?.achievement.name ?? ""
  if (out) await share.shareImageFile(out.blob, out.filename, t("achievements.share.caption", { name }))
}

async function handleDownload() {
  const out = await exportCard()
  if (out) share.saveBlob(out.blob, out.filename)
}
</script>

<style>
.ach-sheet__tabs { margin-bottom: 0.75rem; overflow-x: auto; }
.ach-sheet__stage { position: relative; display: flex; align-items: flex-start; justify-content: center; }
.ach-sheet__skeleton { flex: 1; min-height: 16.25rem; }
.ach-sheet__busy {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  border-radius: var(--radius-card);
  background: color-mix(in srgb, var(--color-surface) 80%, transparent);
  color: var(--color-text-primary);
  font-size: var(--text-small);
}
.ach-sheet__ratio { margin-inline-start: auto; color: var(--color-text-muted); font-size: var(--text-caption); }
.ach-sheet__actions { display: flex; flex-direction: column; gap: 0.5rem; width: 100%; }
.ach-sheet__frame { margin-inline: auto; overflow: hidden; border-radius: var(--radius-card); }
.ach-sheet__scaler { transform-origin: top left; }
.ach-sheet__options { display: flex; flex-wrap: wrap; align-items: center; gap: 0.75rem 1.25rem; margin-top: 1rem; }
.ach-sheet__check {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  cursor: pointer;
}
</style>
