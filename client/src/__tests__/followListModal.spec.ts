import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { QueryClient, VueQueryPlugin } from "@tanstack/vue-query"
import FollowListModal from "@/components/profile/FollowListModal.vue"
import { useRealI18n } from "./helpers/realI18n"
import { clickAfterRender } from "./helpers/waitForEnabled"

const m = vi.hoisted(() => ({
  getFollowers: vi.fn(),
  getFollowing: vi.fn(),
  getPendingRequests: vi.fn(),
  acceptRequest: vi.fn(),
  rejectRequest: vi.fn(),
}))
vi.mock("@/api/api", () => ({ followApi: m }))

const i18n = useRealI18n()
const t = (k: string) => i18n.global.t(k)
const person = (id: number) => ({ id, username: `user${id}`, display_name: `User ${id}`, avatar_url: null, status: "accepted" })
const wrappers: VueWrapper[] = []

function open(props: { isOwner: boolean; initialTab?: "followers" | "following" }) {
  const w = mount(FollowListModal, {
    props: { modelValue: true, userId: 9, initialTab: props.initialTab ?? "followers", isOwner: props.isOwner },
    global: {
      plugins: [i18n, [VueQueryPlugin, { queryClient: new QueryClient({ defaultOptions: { queries: { retry: false } } }) }]],
      stubs: { RouterLink: { template: "<a><slot /></a>" } },
    },
    attachTo: document.body,
  })
  wrappers.push(w)
  return w
}
const bodyText = () => document.body.textContent ?? ""

beforeEach(() => {
  vi.clearAllMocks()
  m.getFollowers.mockResolvedValue({ data: { followers: [person(1), person(2)] } })
  m.getFollowing.mockResolvedValue({ data: { following: [person(3)] } })
  m.getPendingRequests.mockResolvedValue({ data: { requests: [person(4)] } })
  m.acceptRequest.mockResolvedValue({})
  m.rejectRequest.mockResolvedValue({})
})
afterEach(() => {
  wrappers.splice(0).forEach((w) => w.unmount())
  document.body.innerHTML = ""
})

const tabLabels = () => [...document.body.querySelectorAll("[role=tab]")].map((e) => e.textContent?.trim())

describe("FollowListModal", () => {
  it("lists followers and switches to following", async () => {
    open({ isOwner: false })
    await flushPromises()
    expect(bodyText()).toContain("User 1")
    ;(document.body.querySelectorAll("[role=tab]")[1] as HTMLElement).click()
    await flushPromises()
    expect(bodyText()).toContain("User 3")
  })

  it("shows the requests tab to the owner only", async () => {
    open({ isOwner: false })
    await flushPromises()
    expect(tabLabels()).not.toContain(t("profile.follow.requests"))
    wrappers.pop()!.unmount()
    open({ isOwner: true })
    await flushPromises()
    expect(tabLabels()).toContain(t("profile.follow.requests"))
  })

  it("accepts a request", async () => {
    open({ isOwner: true })
    await flushPromises()
    clickAfterRender(document.body.querySelectorAll("[role=tab]")[2] as HTMLElement)
    await flushPromises()
    const accept = [...document.body.querySelectorAll("button")].find((b) => b.textContent?.trim() === t("profile.follow.accept"))!
    clickAfterRender(accept)
    await flushPromises()
    expect(m.acceptRequest).toHaveBeenCalledWith(9, 4)
  })

  it("explains a private list on 403", async () => {
    m.getFollowers.mockRejectedValue({ isAxiosError: true, response: { status: 403 } })
    open({ isOwner: false })
    await flushPromises()
    expect(bodyText()).toContain(t("profile.follow.privateList"))
  })
})
