import { computed, reactive, ref, watch, type Ref } from "vue"
import { useQueryClient } from "@tanstack/vue-query"
import { userApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import type { UserProfile } from "@/types/user"

export const DISPLAY_NAME_MAX = 60
export const BIO_MAX = 200

export interface ProfileFormValues {
  display_name: string
  bio: string
  date_of_birth: string
  gender: string
  gender_other: string
  is_private: boolean
  favorite_genres: number[]
}

export function parseGenres(raw: string | null | undefined): number[] {
  try {
    const parsed: unknown = JSON.parse(raw ?? "[]")
    return Array.isArray(parsed) ? parsed.filter((n): n is number => typeof n === "number") : []
  } catch {
    return []
  }
}

export function snapshotOf(user: UserProfile | null | undefined): ProfileFormValues {
  return {
    display_name: user?.display_name ?? "",
    bio: user?.bio ?? "",
    date_of_birth: user?.date_of_birth ? user.date_of_birth.slice(0, 10) : "",
    gender: user?.gender ?? "",
    gender_other: "",
    is_private: user?.is_private ?? false,
    favorite_genres: parseGenres(user?.favorite_genres),
  }
}

const sameGenres = (a: number[], b: number[]) => {
  const [x, y] = [[...a].sort(), [...b].sort()]
  return x.length === y.length && x.every((g, i) => g === y[i])
}

/**
 * Builds the PATCH body from what changed. A key sent with an empty value
 * clears the field on the server. `username` is never part of the form.
 */
export function changedFormData(form: ProfileFormValues, snap: ProfileFormValues, avatar: Blob | null): FormData {
  const body = new FormData()
  if (form.display_name !== snap.display_name) body.append("display_name", form.display_name.trim())
  if (form.bio !== snap.bio) body.append("bio", form.bio.trim())
  if (form.date_of_birth !== snap.date_of_birth) body.append("date_of_birth", form.date_of_birth)
  if (form.gender !== snap.gender && form.gender) body.append("gender", form.gender)
  const other = form.gender === "other" ? form.gender_other.trim() : ""
  if (other !== snap.gender_other || (form.gender !== snap.gender && snap.gender === "other")) {
    body.append("gender_other", other)
  }
  if (form.is_private !== snap.is_private) body.append("is_private", String(form.is_private))
  if (avatar) body.append("avatar", avatar, "avatar.webp")
  return body
}

export function useProfileForm(user: Ref<UserProfile | null | undefined>) {
  const auth = useAuthStore()
  const queryClient = useQueryClient()

  const snapshot = ref<ProfileFormValues>(snapshotOf(user.value))
  const form = reactive<ProfileFormValues>({ ...snapshot.value })
  const avatarBlob = ref<Blob | null>(null)
  const avatarPreview = ref<string | null>(null)
  const saving = ref(false)
  const error = ref("")

  const profileChanged = computed(() => {
    const s = snapshot.value
    return (
      avatarBlob.value !== null ||
      form.display_name !== s.display_name ||
      form.bio !== s.bio ||
      form.date_of_birth !== s.date_of_birth ||
      form.gender !== s.gender ||
      (form.gender === "other" && form.gender_other !== s.gender_other) ||
      form.is_private !== s.is_private
    )
  })
  const genresChanged = computed(() => !sameGenres(form.favorite_genres, snapshot.value.favorite_genres))
  const isDirty = computed(() => profileChanged.value || genresChanged.value)

  function reset() {
    Object.assign(form, snapshot.value, { favorite_genres: [...snapshot.value.favorite_genres] })
    if (avatarPreview.value) URL.revokeObjectURL(avatarPreview.value)
    avatarBlob.value = null
    avatarPreview.value = null
    error.value = ""
  }

  function setAvatar(blob: Blob) {
    if (avatarPreview.value) URL.revokeObjectURL(avatarPreview.value)
    avatarBlob.value = blob
    avatarPreview.value = URL.createObjectURL(blob)
  }

  // Follow the server copy until the user starts editing.
  watch(
    user,
    next => {
      if (isDirty.value) return
      snapshot.value = snapshotOf(next)
      reset()
    },
    { immediate: true },
  )

  async function save(): Promise<boolean> {
    const id = user.value?.id
    if (!id || saving.value || !isDirty.value) return false
    saving.value = true
    error.value = ""
    try {
      let latest: UserProfile | undefined
      if (profileChanged.value) {
        latest = (await userApi.updateProfile(id, changedFormData(form, snapshot.value, avatarBlob.value))).data.user
      }
      if (genresChanged.value) {
        latest = (await userApi.updateFavoriteGenres(id, form.favorite_genres)).data.user
      }
      if (latest && auth.user) auth.setUser({ ...auth.user, ...latest })
      await queryClient.invalidateQueries({ queryKey: ["user-profile", id] })
      snapshot.value = snapshotOf(latest ?? user.value)
      snapshot.value.gender_other = form.gender === "other" ? form.gender_other : ""
      reset()
      return true
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      error.value = msg ?? "save_failed"
      return false
    } finally {
      saving.value = false
    }
  }

  return { form, avatarBlob, avatarPreview, saving, error, isDirty, profileChanged, genresChanged, reset, setAvatar, save }
}
