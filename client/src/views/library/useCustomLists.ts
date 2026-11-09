import { computed, toValue, type MaybeRefOrGetter } from "vue"
import { useQuery } from "@tanstack/vue-query"
import { listApi } from "@/api/endpoints/list"
import { httpStatus } from "@/components/lists/listErrors"
import { LISTS_KEY } from "@/components/lists/useListActions"

/** Custom lists of one user: all of them for the owner, the visible ones for a visitor. */
export function useCustomLists(userId: MaybeRefOrGetter<number>, isOwner: MaybeRefOrGetter<boolean>) {
  const query = useQuery({
    queryKey: computed(() => [...LISTS_KEY, "user", toValue(userId), toValue(isOwner)] as const),
    queryFn: () => (toValue(isOwner) ? listApi.getOwn() : listApi.getForUser(toValue(userId))).then((r) => r.data.lists ?? []),
    // 403 is a stable answer (locked profile); retrying only delays the locked note.
    retry: (count, err) => httpStatus(err) !== 403 && count < 2,
  })
  return {
    lists: computed(() => query.data.value ?? []),
    isLoading: query.isPending,
    isError: computed(() => query.isError.value && httpStatus(query.error.value) !== 403),
    forbidden: computed(() => httpStatus(query.error.value) === 403),
    refetch: query.refetch,
  }
}
