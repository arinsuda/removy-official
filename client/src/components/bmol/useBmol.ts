import { computed, toValue, type MaybeRefOrGetter } from "vue"
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query"
import { useI18n } from "vue-i18n"
import { bmolApi } from "@/api/api"
import { useToast } from "@/composables/useToast"
import type { BMOLItemResponse, MediaType } from "@/types"

/** A title that can be added to the list (from search or suggestions). */
export interface BmolCandidate {
  id: number
  media_type: MediaType
  title: string
  /** TMDB poster path, as the API returns it. */
  poster_url: string
  vote_average: number
}

/** Undoes one optimistic change without touching other in-flight changes. */
type Undo = { undo?: (list: BMOLItemResponse[]) => BMOLItemResponse[] }

function httpStatus(err: unknown): number | undefined {
  return (err as { response?: { status?: number } } | null)?.response?.status
}

let tempId = 0

/**
 * Shared BMOL data for one user. The query key ["bmol", userId] is also read by
 * the profile highlight, so every change shows up in both places at once.
 */
export function useBmol(userId: MaybeRefOrGetter<number>) {
  const { t } = useI18n()
  const toast = useToast()
  const client = useQueryClient()
  const key = computed(() => ["bmol", toValue(userId)] as const)

  const query = useQuery({
    queryKey: key,
    queryFn: () => bmolApi.getUserBMOL(toValue(userId)).then((res) => res.data.items ?? []),
    // A 403 is a stable answer (private account); retrying only delays the locked state.
    retry: (count, err) => httpStatus(err) !== 403 && count < 2,
  })

  const items = computed(() => query.data.value ?? [])
  const forbidden = computed(() => httpStatus(query.error.value) === 403)

  // Mutations on the same item run one after another so a late response can't overwrite a newer one.
  const queues = new Map<string, Promise<unknown>>()
  function enqueue<T>(itemKey: string, task: () => Promise<T>): Promise<T> {
    const run = (queues.get(itemKey) ?? Promise.resolve()).catch(() => undefined).then(task)
    queues.set(itemKey, run)
    void run.catch(() => undefined).then(() => {
      if (queues.get(itemKey) === run) queues.delete(itemKey)
    })
    return run
  }

  async function optimistic(update: (list: BMOLItemResponse[]) => BMOLItemResponse[], undo: (list: BMOLItemResponse[]) => BMOLItemResponse[]): Promise<Undo> {
    await client.cancelQueries({ queryKey: key.value })
    if (!client.getQueryData(key.value)) return {}
    client.setQueryData<BMOLItemResponse[]>(key.value, (list) => update(list ?? []))
    return { undo }
  }
  function rollback(_err: unknown, _vars: unknown, ctx: Undo | undefined) {
    if (ctx?.undo) client.setQueryData<BMOLItemResponse[]>(key.value, (list) => ctx.undo!(list ?? []))
  }
  const settle = () => client.invalidateQueries({ queryKey: key.value })

  const addMutation = useMutation({
    mutationFn: (v: { candidate: BmolCandidate; rank: number }) =>
      enqueue(`media:${v.candidate.media_type}:${v.candidate.id}`, () =>
        bmolApi.addItem({ media_id: v.candidate.id, media_type: v.candidate.media_type, rank: v.rank }),
      ),
    onMutate: ({ candidate, rank }) => {
      const id = --tempId
      return optimistic((list) => [
        ...list,
        {
          id,
          rank,
          note: "",
          media_type: candidate.media_type,
          created_at: new Date().toISOString(),
          media: {
            id: candidate.id,
            title: candidate.title,
            poster_url: candidate.poster_url,
            media_type: candidate.media_type,
            genres: [],
            vote_average: candidate.vote_average,
            created_at: "",
          },
        },
      ], (list) => list.filter((i) => i.id !== id))
    },
    onError: (err, vars, ctx) => {
      rollback(err, vars, ctx)
      toast.error(
        httpStatus(err) === 409
          ? t("library.bmol.toasts.duplicate", { title: vars.candidate.title })
          : t("library.bmol.toasts.addFailed", { title: vars.candidate.title }),
      )
    },
    onSettled: settle,
  })

  const moveMutation = useMutation({
    mutationFn: (v: { id: number; rank: number }) => enqueue(`item:${v.id}`, () => bmolApi.updateItem(v.id, { rank: v.rank })),
    onMutate: ({ id, rank }) => {
      const before = client.getQueryData<BMOLItemResponse[]>(key.value)?.find((i) => i.id === id)?.rank
      return optimistic(
        (list) => list.map((i) => (i.id === id ? { ...i, rank } : i)),
        (list) => list.map((i) => (i.id === id && before !== undefined ? { ...i, rank: before } : i)),
      )
    },
    onError: (err, vars, ctx) => {
      rollback(err, vars, ctx)
      toast.error(t("library.bmol.toasts.moveFailed"))
    },
    onSettled: settle,
  })

  const noteMutation = useMutation({
    mutationFn: (v: { id: number; note: string }) => enqueue(`item:${v.id}`, () => bmolApi.updateItem(v.id, { note: v.note })),
    onMutate: ({ id, note }) => {
      const before = client.getQueryData<BMOLItemResponse[]>(key.value)?.find((i) => i.id === id)?.note
      return optimistic(
        (list) => list.map((i) => (i.id === id ? { ...i, note } : i)),
        (list) => list.map((i) => (i.id === id && before !== undefined ? { ...i, note: before } : i)),
      )
    },
    onError: (err, vars, ctx) => {
      rollback(err, vars, ctx)
      toast.error(t("library.bmol.toasts.noteFailed"))
    },
    onSuccess: () => toast.success(t("library.bmol.toasts.noteSaved")),
    onSettled: settle,
  })

  const removeMutation = useMutation({
    mutationFn: (v: { id: number }) => enqueue(`item:${v.id}`, () => bmolApi.removeItem(v.id)),
    onMutate: ({ id }) => {
      const removed = client.getQueryData<BMOLItemResponse[]>(key.value)?.find((i) => i.id === id)
      return optimistic(
        (list) => list.filter((i) => i.id !== id),
        (list) => (removed && !list.some((i) => i.id === id) ? [...list, removed] : list),
      )
    },
    onError: (err, vars, ctx) => {
      rollback(err, vars, ctx)
      toast.error(t("library.bmol.toasts.removeFailed"))
    },
    onSettled: settle,
  })

  return {
    items,
    forbidden,
    isLoading: query.isLoading,
    isError: computed(() => query.isError.value && !forbidden.value),
    refetch: query.refetch,
    add: (candidate: BmolCandidate, rank: number) => addMutation.mutateAsync({ candidate, rank }).catch(() => undefined),
    move: (id: number, rank: number) => moveMutation.mutateAsync({ id, rank }).catch(() => undefined),
    saveNote: (id: number, note: string) => noteMutation.mutateAsync({ id, note }).then(() => true).catch(() => false),
    remove: (id: number) => removeMutation.mutateAsync({ id }).catch(() => undefined),
  }
}
