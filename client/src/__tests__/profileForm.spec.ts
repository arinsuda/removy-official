import { describe, expect, it, vi } from "vitest"
import { nextTick, ref } from "vue"
import { createPinia, setActivePinia } from "pinia"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import { createApp } from "vue"
import { changedFormData, snapshotOf, useProfileForm } from "@/components/settings/useProfileForm"

vi.mock("@/api/api", () => ({ userApi: { updateProfile: vi.fn(), updateFavoriteGenres: vi.fn() } }))

const user = (over = {}) => ({
  id: 3, username: "alice", display_name: "Alice", bio: "hello", avatar_url: null, date_of_birth: "1990-05-01T00:00:00Z",
  gender: "male", favorite_genres: "[28,35]", is_private: false, ...over,
}) as never

describe("changedFormData", () => {
  const snap = snapshotOf(user())
  const keys = (fd: FormData) => [...fd.keys()]

  it("sends nothing when nothing changed", () => {
    expect(keys(changedFormData({ ...snap }, snap, null))).toEqual([])
  })

  it("sends the key with an empty value to clear the bio", () => {
    const fd = changedFormData({ ...snap, bio: "" }, snap, null)
    expect(keys(fd)).toEqual(["bio"])
    expect(fd.get("bio")).toBe("")
  })

  it("clears the date of birth with an empty value", () => {
    const fd = changedFormData({ ...snap, date_of_birth: "" }, snap, null)
    expect(fd.get("date_of_birth")).toBe("")
  })

  it("never sends username, even if it appears on the form object", () => {
    const form = { ...snap, display_name: "Bob", username: "mallory" } as typeof snap
    expect(keys(changedFormData(form, snap, null))).not.toContain("username")
  })

  it("sends gender_other only with the 'other' gender and clears it on switching away", () => {
    const other = changedFormData({ ...snap, gender: "other", gender_other: " nb " }, snap, null)
    expect(other.get("gender")).toBe("other")
    expect(other.get("gender_other")).toBe("nb")
    const otherSnap = { ...snap, gender: "other", gender_other: "nb" }
    expect(changedFormData({ ...otherSnap, gender: "female" }, otherSnap, null).get("gender_other")).toBe("")
  })

  it("attaches the cropped avatar", () => {
    expect(keys(changedFormData({ ...snap }, snap, new Blob(["x"])))).toEqual(["avatar"])
  })
})

describe("useProfileForm", () => {
  it("tracks dirtiness and resets", async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const app = createApp({ render: () => null }).use(pinia).use(VueQueryPlugin, { queryClient: new QueryClient() })
    let api!: ReturnType<typeof useProfileForm>
    app.runWithContext(() => {
      api = useProfileForm(ref(user()))
    })
    expect(api.isDirty.value).toBe(false)
    api.form.bio = ""
    await nextTick()
    expect(api.isDirty.value).toBe(true)
    api.reset()
    expect(api.form.bio).toBe("hello")
    expect(api.isDirty.value).toBe(false)
    // reordering genres is not a change
    api.form.favorite_genres = [35, 28]
    expect(api.genresChanged.value).toBe(false)
  })
})
