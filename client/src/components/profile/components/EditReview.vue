<template>
  <AppModal :model-value="true" size="sm" :title="t('reviews.editTitle')" @close="emit('close')">
    <form :id="formId" class="er" @submit.prevent="handleSave" @keydown.enter.meta.prevent="handleSave" @keydown.enter.ctrl.prevent="handleSave">
      <header class="er__media">
        <span class="er__poster">
          <img v-if="review.media.poster_url" :src="review.media.poster_url" :alt="review.media.title" />
          <Film v-else :size="18" :stroke-width="1.4" aria-hidden="true" />
        </span>
        <span class="er__meta">
          <span class="er__eyebrow">{{ metaLine }}</span>
          <span class="er__title">{{ review.media.title }}</span>
          <span class="er__date">{{ t("reviews.edit.writtenOn", { date: writtenOn }) }}</span>
        </span>
      </header>

      <section class="er__rating">
        <span class="er__label">{{ t("reviews.composer.ratingLabel") }}</span>
        <div class="er__rating-row">
          <span class="er__score" aria-hidden="true">{{ n(form.rating, scoreFormat) }}<small>/ {{ n(RATING_MAX) }}</small></span>
          <StarRatingInput v-model="form.rating" :size="STAR_SIZE" :show-label="false" :aria-label="t('reviews.composer.ratingLabel')" />
        </div>
      </section>

      <label class="er__field">
        <span class="er__label">{{ t("reviews.composer.bodyLabel") }}</span>
        <textarea
          v-model="form.body"
          class="er__textarea"
          rows="6"
          :placeholder="t(`reviews.composer.placeholder.${placeholderType}`)"
          data-focus-custom
        />
        <span class="er__hint">{{ t("reviews.edit.shortcut", { keys: shortcutKeys }) }}</span>
      </label>

      <div class="er__row">
        <span class="er__label">{{ t("reviews.composer.visibilityLabel") }}</span>
        <VisibilitySelector v-model="form.visibility" size="sm" />
      </div>
    </form>

    <template #footer>
      <AppButton variant="outline" @click="emit('close')">{{ t("common.cancel") }}</AppButton>
      <AppButton type="submit" :form="formId" :loading="saving" :disabled="saving || !canSave">{{ t("reviews.update") }}</AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, reactive, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { Film } from "lucide-vue-next"
import { reviewApi } from "@/api/api"
import VisibilitySelector from "@/components/common/VisibilitySelector.vue"
import StarRatingInput from "@/components/movie/Starratinginput.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useUIStore } from "@/stores/ui"
import type { ReviewResponse, Visibility } from "@/types"
import { logApiError } from "@/utils/logApiError"

const RATING_MAX = 5
const STAR_SIZE = 30
const scoreFormat = { minimumFractionDigits: 1, maximumFractionDigits: 1 } as const

const props = defineProps<{ userId: number; review: ReviewResponse }>()
const emit = defineEmits<{ close: []; saved: [review: ReviewResponse] }>()

const { t, n, d } = useI18n()
const ui = useUIStore()
const formId = `edit-review-${useId()}`
const saving = ref(false)

const initial = {
  rating: props.review.rating,
  body: props.review.body ?? "",
  visibility: (props.review.visibility || "public") as Visibility,
}
const form = reactive({ ...initial })

const placeholderType = computed(() => (props.review.media.media_type === "tv" ? "tv" : "movie"))
const metaLine = computed(() => {
  const type = t(`reviews.edit.type.${placeholderType.value}`)
  const year = props.review.media.release_date?.slice(0, 4)
  return year ? t("reviews.edit.typeYear", { type, year }) : type
})
const writtenOn = computed(() => d(new Date(props.review.created_at), "short"))
const shortcutKeys = computed(() =>
  t(/mac|iphone|ipad/i.test(navigator.userAgent) ? "reviews.edit.keys.mac" : "reviews.edit.keys.other"),
)

const isDirty = computed(
  () => form.rating !== initial.rating || form.body.trim() !== initial.body.trim() || form.visibility !== initial.visibility,
)
const canSave = computed(() => form.rating > 0 && isDirty.value)

async function handleSave() {
  if (saving.value || !canSave.value) return
  saving.value = true
  try {
    const res = await reviewApi.updateReview(props.review.id, { ...form, body: form.body.trim() })
    emit("saved", res.data.review)
  } catch (err) {
    logApiError("updateReview", err)
    ui.addToast(t("reviews.errors.updateFailed"), "error")
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.er { display: flex; flex-direction: column; gap: 1.25rem; }

.er__media { display: flex; align-items: center; gap: 0.875rem; }
.er__poster { display: grid; flex: none; place-items: center; width: 3.5rem; aspect-ratio: 2 / 3; overflow: hidden; color: var(--color-text-muted); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: calc(var(--radius-card) / 2); }
.er__poster img { width: 100%; height: 100%; object-fit: cover; }
.er__meta { display: flex; flex-direction: column; gap: 0.125rem; min-width: 0; }
.er__eyebrow { font-size: var(--text-caption); font-weight: 600; letter-spacing: 0.04em; color: var(--color-text-muted); text-transform: uppercase; }
.er__title { overflow: hidden; font-size: var(--text-body); font-weight: 700; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.er__date { font-size: var(--text-caption); color: var(--color-text-secondary); }

.er__label { font-size: var(--text-caption); font-weight: 600; color: var(--color-text-secondary); }

.er__rating { display: flex; flex-direction: column; gap: 0.625rem; padding: 0.875rem 1rem; background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.er__rating-row { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.75rem; }
.er__score { font-size: var(--text-heading); font-weight: 800; line-height: 1; color: var(--color-rating); font-variant-numeric: tabular-nums; }
.er__score small { margin-inline-start: 0.25rem; font-size: var(--text-small); font-weight: 600; color: var(--color-text-muted); }

.er__field { display: flex; flex-direction: column; gap: 0.5rem; }
.er__textarea { box-sizing: border-box; width: 100%; min-height: 8.5rem; padding: 0.875rem 1rem; font: inherit; font-size: var(--text-small); line-height: 1.7; color: var(--color-text-primary); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: var(--radius-card); outline: none; resize: vertical; transition: border-color var(--motion-fast) var(--ease-water), box-shadow var(--motion-fast) var(--ease-water); }
.er__textarea::placeholder { color: var(--color-text-muted); }
.er__textarea:focus-visible { border-color: var(--color-border-strong); box-shadow: var(--focus-input-shadow); }
.er__hint { align-self: flex-end; font-size: var(--text-caption); color: var(--color-text-muted); }

.er__row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding-top: 1rem; border-top: 1px solid var(--color-border); }

@media (prefers-reduced-motion: reduce) {
  .er__textarea { transition: none; }
}
</style>
