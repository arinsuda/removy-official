import { beforeEach, describe, expect, it, vi } from "vitest"
import { ref } from "vue"
import { createPinia, setActivePinia } from "pinia"
import { useAuthGate } from "@/composables/useAuthGate"
import { useReviewForm } from "@/composables/useReviewForm"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"

const { createReview } = vi.hoisted(() => ({ createReview: vi.fn() }))
vi.mock("@/api/api", () => ({ reviewApi: { createReview } }))

describe("useReviewForm", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    useAuthGate().close()
    useAuthStore().user = { id: 7 } as never
    createReview.mockReset()
    createReview.mockResolvedValue({ data: { review: { id: 1 } } })
  })

  it("posts to the title shown now after the id ref changes", async () => {
    const id = ref(100)
    const form = useReviewForm(id, "movie")
    id.value = 200
    form.rating.value = 4
    form.body.value = "Great"
    await form.submit()
    expect(createReview).toHaveBeenCalledWith(expect.objectContaining({ media_id: 200, media_type: "movie", rating: 4 }))
  })

  it("accepts plain numbers and getters", async () => {
    const form = useReviewForm(() => 5, () => "tv")
    form.rating.value = 2
    form.body.value = "ok"
    await form.submit()
    expect(createReview).toHaveBeenCalledWith(expect.objectContaining({ media_id: 5, media_type: "tv" }))
  })

  it("starts with no rating and blocks submit until one is chosen", async () => {
    const form = useReviewForm(1, "movie")
    expect(form.rating.value).toBe(0)
    form.body.value = "text"
    expect(form.canSubmit.value).toBe(false)
    expect(await form.submit()).toBeNull()
    expect(createReview).not.toHaveBeenCalled()
    form.rating.value = 3.5
    expect(form.canSubmit.value).toBe(true)
  })

  it("opens the shared auth gate and avoids toasts/API calls when a guest submits", async () => {
    useAuthStore().user = null
    const form = useReviewForm(1, "movie")
    form.rating.value = 4
    form.body.value = "Nice"

    expect(await form.submit()).toBeNull()
    expect(createReview).not.toHaveBeenCalled()
    expect(useAuthGate().isOpen.value).toBe(true)
    expect(useAuthGate().reason.value).toBe("review")
    expect(useUIStore().toasts).toHaveLength(0)
  })
})
