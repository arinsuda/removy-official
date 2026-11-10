<template>
  <article class="own" data-test="own-review">
    <header class="own__head">
      <h3 class="own__title">{{ t("reviews.own.title") }}</h3>
      <AppButton v-if="!editing" variant="ghost" size="sm" data-test="own-edit" @click="startEdit">
        <Pencil :size="14" aria-hidden="true" />
        {{ t("common.edit") }}
      </AppButton>
    </header>

    <template v-if="!editing">
      <div class="own__rating">
        <StarRatingInput :model-value="review.rating" readonly :size="18" :show-label="false" :aria-label="t('reviews.composer.ratingLabel')" />
        <span class="own__score">{{ review.rating }}/5</span>
        <span v-if="review.visibility !== 'public'" class="own__badge">{{ t(`visibility.${review.visibility}`) }}</span>
      </div>
      <p class="own__body">{{ review.body }}</p>
    </template>

    <form v-else class="own__form" @submit.prevent="save">
      <StarRatingInput v-model="draft.rating" :size="26" :aria-label="t('reviews.composer.ratingLabel')" />
      <textarea ref="field" v-model="draft.body" class="own__input" rows="3" :aria-label="t('reviews.composer.bodyLabel')" />
      <div class="own__footer">
        <VisibilitySelector v-model="draft.visibility" size="sm" />
        <div class="own__buttons">
          <AppButton variant="ghost" :disabled="saving" @click="editing = false">{{ t("common.cancel") }}</AppButton>
          <AppButton type="submit" :loading="saving" :disabled="!canSave">{{ t("reviews.update") }}</AppButton>
        </div>
      </div>
    </form>
  </article>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import { Pencil } from "lucide-vue-next"
import { reviewApi } from "@/api/api"
import AppButton from "@/components/ui/AppButton.vue"
import VisibilitySelector from "@/components/common/VisibilitySelector.vue"
import StarRatingInput from "@/components/movie/StarRatingInput.vue"
import { useToast } from "@/composables/useToast"
import type { ReviewResponse } from "@/types/review"
import type { Visibility } from "@/types"

const props = defineProps<{ review: ReviewResponse }>()
const emit = defineEmits<{ updated: [review: ReviewResponse] }>()

const { t } = useI18n()
const { error: toastError, success: toastSuccess } = useToast()
const field = useTemplateRef<HTMLTextAreaElement>("field")
const editing = ref(false)
const saving = ref(false)
const draft = reactive<{ rating: number; body: string; visibility: Visibility }>({ rating: 0, body: "", visibility: "public" })
const canSave = computed(() => draft.rating > 0 && draft.body.trim().length > 0 && !saving.value)

async function startEdit() {
  Object.assign(draft, { rating: props.review.rating, body: props.review.body, visibility: props.review.visibility })
  editing.value = true
  await nextTick()
  field.value?.focus()
}

async function save() {
  if (!canSave.value) return
  saving.value = true
  try {
    const res = await reviewApi.updateReview(props.review.id, { ...draft })
    emit("updated", res.data.review)
    editing.value = false
    toastSuccess(t("reviews.own.saved"))
  } catch (err) {
    console.error("updateReview failed:", err)
    toastError(t("reviews.errors.updateFailed"))
  } finally {
    saving.value = false
  }
}

defineExpose({ startEdit })
</script>

<style scoped>
.own { display: flex; flex-direction: column; gap: 0.6rem; padding: 1rem; background: color-mix(in srgb, var(--color-brand) 8%, var(--color-surface-2)); border: 1px solid color-mix(in srgb, var(--color-brand) 40%, transparent); border-radius: var(--radius-card); }
.own__head { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
.own__title { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.own__rating { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
.own__score { font-size: var(--text-small); font-weight: 700; color: var(--color-rating); }
.own__badge { padding: 0.1rem 0.5rem; font-size: var(--text-caption); color: var(--color-text-secondary); background: var(--color-surface-3); border-radius: var(--radius-btn); }
.own__body { font-size: var(--text-small); line-height: 1.6; color: var(--color-text-secondary); white-space: pre-line; overflow-wrap: anywhere; }
.own__form { display: flex; flex-direction: column; gap: 0.75rem; }
.own__input { width: 100%; padding: 0.75rem; font: inherit; font-size: var(--text-small); color: var(--color-text-primary); background: var(--color-surface); border: 1px solid var(--color-border-strong); border-radius: var(--radius-card); resize: vertical; box-sizing: border-box; }
.own__footer { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.75rem; }
.own__buttons { display: flex; gap: 0.5rem; }
</style>
