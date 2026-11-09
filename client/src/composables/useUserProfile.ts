import { computed, toValue, type MaybeRefOrGetter } from "vue"
import { useQuery } from "@tanstack/vue-query"
import { isAxiosError } from "axios"
import { reviewApi, userApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { useFollowStatus } from "@/composables/useFollow"
import { tasteSummary } from "@/components/profile/tasteProfile"

export type ProfileErrorKind = "notFound" | "error" | null

const REVIEW_LIMIT = 100

/** Profile, stats and reviews for /users/:id. Sections stay disabled while locked. */
export function useUserProfile(id: MaybeRefOrGetter<number>) {
  const auth = useAuthStore()
  const userId = computed(() => toValue(id))
  const isOwner = computed(() => !!auth.user && auth.user.id === userId.value)

  const profile = useQuery({
    queryKey: computed(() => ["user-profile", userId.value]),
    queryFn: () => userApi.getProfile(userId.value).then(res => res.data.user),
    enabled: computed(() => userId.value > 0),
    retry: false,
  })

  const followStatus = useFollowStatus(userId, () => auth.isLoggedIn && !isOwner.value)

  const canViewSections = computed(
    () =>
      isOwner.value ||
      (!!profile.data.value && !profile.data.value.is_private) ||
      followStatus.data.value?.follow_status === "accepted",
  )
  const isLocked = computed(() => !!profile.data.value && !canViewSections.value)

  const stats = useQuery({
    queryKey: computed(() => ["user-stats", userId.value]),
    queryFn: () => userApi.getStats(userId.value).then(res => res.data.stats),
    enabled: computed(() => !!profile.data.value && canViewSections.value),
  })

  const reviews = useQuery({
    queryKey: computed(() => ["user-reviews-taste", userId.value]),
    queryFn: () =>
      reviewApi.getUserReviews(userId.value, { limit: REVIEW_LIMIT }).then(res => res.data.reviews ?? []),
    enabled: computed(() => !!profile.data.value && canViewSections.value),
  })

  const taste = computed(() => tasteSummary(reviews.data.value ?? []))

  const errorKind = computed<ProfileErrorKind>(() => {
    if (!profile.isError.value) return null
    const err = profile.error.value
    return isAxiosError(err) && err.response?.status === 404 ? "notFound" : "error"
  })

  return { userId, isOwner, profile, followStatus, canViewSections, isLocked, stats, reviews, taste, errorKind }
}
