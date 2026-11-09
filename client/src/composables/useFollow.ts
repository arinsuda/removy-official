import { computed, toValue, type MaybeRefOrGetter } from "vue"
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query"
import { followApi } from "@/api/endpoints/social"
import type { FollowRelationshipStatus } from "@/types/follow"

export function useFollowStats(id: MaybeRefOrGetter<number>) {
  return useQuery({
    queryKey: computed(() => ["follow-stats", toValue(id)]),
    queryFn: () => followApi.getFollowStats(toValue(id)).then(res => res.data),
    enabled: computed(() => toValue(id) > 0),
  })
}

export function useFollowStatus(id: MaybeRefOrGetter<number>, enabled: MaybeRefOrGetter<boolean> = true) {
  return useQuery({
    queryKey: computed(() => ["follow-status", toValue(id)]),
    queryFn: () => followApi.getFollowStatus(toValue(id)).then(res => res.data),
    enabled: computed(() => toValue(id) > 0 && toValue(enabled)),
  })
}

/** `unfollow` also cancels a pending request (same DELETE /follow). */
export function useFollowActions(id: MaybeRefOrGetter<number>) {
  const queryClient = useQueryClient()
  const target = () => toValue(id)

  function invalidate() {
    const userId = target()
    queryClient.invalidateQueries({ queryKey: ["follow-status", userId] })
    queryClient.invalidateQueries({ queryKey: ["follow-stats", userId] })
    queryClient.invalidateQueries({ queryKey: ["user-profile", userId] })
    queryClient.invalidateQueries({ queryKey: ["user-stats", userId] })
  }

  const follow = useMutation({
    mutationFn: () => followApi.follow(target()),
    onSuccess: res => {
      queryClient.setQueryData<FollowRelationshipStatus>(
        ["follow-status", target()],
        prev => ({
          is_following: res.data.status === "accepted",
          follow_status: res.data.status,
          is_followed_by: prev?.is_followed_by ?? false,
        }),
      )
      invalidate()
    },
  })

  const unfollow = useMutation({
    mutationFn: () => followApi.unfollow(target()),
    onSuccess: () => {
      queryClient.setQueryData<FollowRelationshipStatus>(
        ["follow-status", target()],
        prev => ({
          is_following: false,
          follow_status: undefined,
          is_followed_by: prev?.is_followed_by ?? false,
        }),
      )
      invalidate()
    },
  })

  return { follow, unfollow }
}
