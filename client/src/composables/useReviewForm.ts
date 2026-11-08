import { reviewApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { useToast } from "@/composables/useToast"
import { i18n } from "@/i18n"
import { computed, ref, toValue, type MaybeRefOrGetter } from "vue"
import type { ReviewResponse } from "@/types/review"
import type { Visibility } from "@/types"

type MediaType = "movie" | "tv"

/**
 * Create-review form state. The media id and type are read when submitting, so
 * the form stays correct if the page is reused for another title. The rating
 * starts unset (0) and must be chosen.
 */
export function useReviewForm(mediaId: MaybeRefOrGetter<number>, mediaType: MaybeRefOrGetter<MediaType>) {
  const auth = useAuthStore()
  const { error: toastError } = useToast()
  const t = i18n.global.t

  const body = ref("")
  const rating = ref(0)
  const visibility = ref<Visibility>("public")
  const isSubmitting = ref(false)

  const isAuthenticated = computed(() => Boolean(auth.user?.id))
  const canSubmit = computed(
    () =>
      isAuthenticated.value &&
      rating.value > 0 &&
      body.value.trim().length > 0 &&
      !isSubmitting.value,
  )

  function reset() {
    body.value = ""
    rating.value = 0
    visibility.value = "public"
  }

  async function submit(): Promise<ReviewResponse | null> {
    if (!isAuthenticated.value) {
      toastError(t("reviews.errors.loginRequired"))
      return null
    }
    if (rating.value <= 0 || !body.value.trim() || isSubmitting.value) return null

    isSubmitting.value = true
    try {
      const res = await reviewApi.createReview({
        media_id: toValue(mediaId),
        media_type: toValue(mediaType),
        rating: rating.value,
        body: body.value,
        visibility: visibility.value,
      })
      reset()
      return res.data.review
    } catch (err) {
      console.error("submitReview failed:", err)
      toastError(t("reviews.errors.submitFailed"))
      return null
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    body,
    rating,
    visibility,
    isSubmitting,
    isAuthenticated,
    canSubmit,
    reset,
    submit,
  }
}
