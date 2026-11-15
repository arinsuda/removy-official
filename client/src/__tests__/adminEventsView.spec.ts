import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils"
import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory, createRouter } from "vue-router"
import AdminEventsView from "@/views/admin/AdminEventsView.vue"
import { adminEventApi } from "@/api/endpoints/event"
import type { AdminEvent } from "@/types/event"
import { clickAfterRender } from "./helpers/waitForEnabled"
import { useRealI18n } from "./helpers/realI18n"

vi.mock("@/api/endpoints/event", () => ({
  adminEventApi: {
    list: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
    uploadImage: vi.fn(),
    removeImage: vi.fn(),
    generateMonthly: vi.fn(),
    generateYearlyRound: vi.fn(),
    suggestNominees: vi.fn(),
    setNominees: vi.fn(),
    finalize: vi.fn(),
  },
}))

const i18n = useRealI18n()
const api = vi.mocked(adminEventApi)
const wrappers: VueWrapper[] = []
const tr = (key: string) => i18n.global.t(key)

const makeAdminEvent = (over: Partial<AdminEvent> = {}): AdminEvent => ({
  id: 1,
  type: "announcement",
  media_type: null,
  period: null,
  title: { en: "Admin Event EN", th: "กิจกรรมผู้ดูแล TH" },
  body: { en: "Admin Body EN", th: "เนื้อหาผู้ดูแล TH" },
  image_url: null,
  link_url: "/movies",
  starts_at: "2026-11-15T00:00:00.000Z",
  ends_at: "2026-11-20T00:00:00.000Z",
  show_popup: true,
  popup_priority: 5,
  state: "active",
  status: "published",
  created_at: "2026-11-10T00:00:00.000Z",
  updated_at: "2026-11-10T00:00:00.000Z",
  ...over,
})

const nominee = (id: number, mediaId: number, title: string) => ({
  id,
  event_id: 9,
  media_id: mediaId,
  media_type: "movie" as const,
  position: id,
  media_title: title,
  is_winner: false,
})

/** A promise the test settles by hand, to control response order. */
function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((r) => {
    resolve = r
  })
  return { promise, resolve }
}

const hintText = (min: number) => i18n.global.t("admin.events.saveNomineesEmptyHint", { min }, min)

const listResponse = (events: AdminEvent[]) =>
  ({ data: { events, pagination: { page: 1, limit: 20, total: events.length, total_pages: 1 } } }) as never

/** Modal content is teleported to body, so query the document rather than the wrapper. */
const modalButton = (label: string) =>
  Array.from(document.body.querySelectorAll("button")).find((b) => b.textContent?.trim() === label) as
    | HTMLButtonElement
    | undefined

async function mountView() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/admin/events", name: "admin-events", component: AdminEventsView },
      { path: "/events/:id", name: "event-detail", component: { template: "<div>Detail</div>" } },
    ],
  })
  await router.push("/admin/events")
  await router.isReady()
  const wrapper = mount(AdminEventsView, {
    global: { plugins: [i18n, router] },
    attachTo: document.body,
  })
  wrappers.push(wrapper)
  await flushPromises()
  return wrapper
}

/** Open the row menu for the first row and click the menu item with `label`. */
async function chooseRowAction(wrapper: VueWrapper, label: string) {
  await wrapper.find('button[aria-haspopup="menu"]').trigger("click")
  const item = Array.from(document.body.querySelectorAll('[role="menuitem"]')).find(
    (el) => el.textContent?.trim() === label,
  ) as HTMLElement | undefined
  expect(item).toBeDefined()
  item!.click()
  await flushPromises()
}

describe("AdminEventsView", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    wrappers.forEach((w) => w.unmount())
    wrappers.length = 0
    document.body.innerHTML = ""
    vi.clearAllMocks()
  })

  it("renders event list from api", async () => {
    api.list.mockResolvedValueOnce(
      listResponse([makeAdminEvent({ id: 5, title: { en: "Grand Launch", th: "เปิดตัวยิ่งใหญ่" } })]),
    )

    const wrapper = await mountView()

    expect(wrapper.text()).toContain("Grand Launch")
    expect(wrapper.text()).toContain("#5")
  })

  it("opens create event modal when clicking 'New Event'", async () => {
    api.list.mockResolvedValueOnce(listResponse([]))

    const wrapper = await mountView()

    const newBtn = wrapper.find(".app-btn--primary")
    expect(newBtn.exists()).toBe(true)
    await newBtn.trigger("click")
    await flushPromises()

    expect(document.body.textContent).toContain("ชื่อกิจกรรม (EN)")
    expect(document.body.textContent).toContain("ชื่อกิจกรรม (TH)")
  })

  it("generates a round after validating the period and refreshes the list", async () => {
    api.list.mockResolvedValue(listResponse([]))
    api.generateMonthly.mockResolvedValueOnce({} as never)

    const wrapper = await mountView()
    const generateBtn = wrapper.findAll("button").find((b) => b.text() === tr("admin.events.generateRound"))
    expect(generateBtn).toBeDefined()
    clickAfterRender(generateBtn!.element as HTMLElement)
    await flushPromises()

    // Media type radios sit in a fieldset with a legend.
    const legends = Array.from(document.body.querySelectorAll("fieldset legend")).map((el) => el.textContent?.trim())
    expect(legends).toContain(tr("admin.events.mediaType"))
    const radios = () => Array.from(document.body.querySelectorAll<HTMLInputElement>('input[name="generate-media-type"]'))
    expect(radios()).toHaveLength(2)

    const period = document.body.querySelector<HTMLInputElement>("#gen-period")!
    const submit = () => document.body.querySelector("form.event-form")!.dispatchEvent(new Event("submit"))

    period.value = "2026-13"
    period.dispatchEvent(new Event("input"))
    submit()
    await flushPromises()
    expect(api.generateMonthly).not.toHaveBeenCalled()
    expect(document.body.textContent).toContain(tr("admin.events.errors.periodInvalid"))

    period.value = "2026-11"
    period.dispatchEvent(new Event("input"))
    radios()[1]!.checked = true
    radios()[1]!.dispatchEvent(new Event("change"))
    submit()
    await flushPromises()

    expect(api.generateMonthly).toHaveBeenCalledWith({ media_type: "tv", period: "2026-11", late_start: false })
    await vi.waitFor(() => expect(api.list).toHaveBeenCalledTimes(2))
  })

  it("finalizes a past round after confirmation", async () => {
    api.list.mockResolvedValue(
      listResponse([
        makeAdminEvent({
          id: 3,
          type: "vote_monthly",
          media_type: "movie",
          period: "2026-10",
          state: "past",
          is_finalized: false,
          nominees: [nominee(1, 101, "Alpha")],
        }),
      ]),
    )
    api.finalize.mockResolvedValueOnce({} as never)

    const wrapper = await mountView()
    await chooseRowAction(wrapper, tr("admin.events.finalize"))

    const confirm = document.body.querySelector(".admin-confirm__actions button:last-child") as HTMLButtonElement
    clickAfterRender(confirm)
    await flushPromises()

    expect(api.finalize).toHaveBeenCalledWith(3)
    await vi.waitFor(() => expect(api.list).toHaveBeenCalledTimes(2))
  })

  it("blocks save and shows a hint when the nominee list is empty", async () => {
    api.list.mockResolvedValue(
      listResponse([
        makeAdminEvent({
          id: 7,
          type: "vote_monthly",
          media_type: "movie",
          period: "2026-12",
          state: "upcoming",
          nominees: [nominee(1, 101, "Alpha"), nominee(2, 102, "Beta")],
        }),
      ]),
    )
    api.setNominees.mockResolvedValueOnce({} as never)

    const wrapper = await mountView()
    await chooseRowAction(wrapper, tr("admin.events.editNominees"))

    const removeLabel = tr("admin.events.removeNominee")
    const removeButtons = () =>
      Array.from(document.body.querySelectorAll<HTMLButtonElement>(`[aria-label="${removeLabel}"]`))
    expect(removeButtons()).toHaveLength(2)

    const saveButton = () => modalButton(tr("admin.events.saveNominees"))!
    const removeFirst = () => clickAfterRender(removeButtons()[0]!)
    removeFirst()
    await flushPromises()
    removeFirst()
    await flushPromises()

    expect(removeButtons()).toHaveLength(0)
    expect(saveButton().disabled).toBe(true)
    expect(document.body.textContent).toContain(hintText(2))
  })

  it("keeps a published round at two nominees: save is blocked below the minimum", async () => {
    api.list.mockResolvedValue(
      listResponse([
        makeAdminEvent({
          id: 7,
          type: "vote_monthly",
          media_type: "movie",
          period: "2026-12",
          state: "upcoming",
          nominees: [nominee(1, 101, "Alpha"), nominee(2, 102, "Beta")],
        }),
      ]),
    )

    const wrapper = await mountView()
    await chooseRowAction(wrapper, tr("admin.events.editNominees"))

    clickAfterRender(document.body.querySelector<HTMLButtonElement>(`[aria-label="${tr("admin.events.removeNominee")}"]`)!)
    await flushPromises()

    expect(modalButton(tr("admin.events.saveNominees"))!.disabled).toBe(true)
    expect(document.body.textContent).toContain(hintText(2))
    expect(api.setNominees).not.toHaveBeenCalled()
  })

  it("ignores a slow earlier list response once the filter has changed", async () => {
    const first = deferred<ReturnType<typeof listResponse>>()
    const second = deferred<ReturnType<typeof listResponse>>()
    api.list.mockReturnValueOnce(first.promise as never).mockReturnValueOnce(second.promise as never)

    const wrapper = await mountView()
    wrapper.find(".admin-select__input").setValue("published")
    await flushPromises()

    second.resolve(listResponse([makeAdminEvent({ id: 2, title: { en: "Newest title", th: "Newest title" } })]))
    await flushPromises()
    first.resolve(listResponse([makeAdminEvent({ id: 1, title: { en: "Stale title", th: "Stale title" } })]))
    await flushPromises()

    expect(wrapper.text()).toContain("Newest title")
    expect(wrapper.text()).not.toContain("Stale title")
  })

  it("sends the remaining nominees with renumbered positions on save", async () => {
    api.list.mockResolvedValue(
      listResponse([
        makeAdminEvent({
          id: 7,
          type: "vote_monthly",
          media_type: "movie",
          period: "2026-12",
          state: "upcoming",
          status: "draft",
          nominees: [nominee(1, 101, "Alpha"), nominee(2, 102, "Beta"), nominee(3, 103, "Gamma")],
        }),
      ]),
    )
    api.setNominees.mockResolvedValueOnce({} as never)

    const wrapper = await mountView()
    await chooseRowAction(wrapper, tr("admin.events.editNominees"))

    const removeLabel = tr("admin.events.removeNominee")
    clickAfterRender(document.body.querySelector<HTMLButtonElement>(`[aria-label="${removeLabel}"]`)!)
    await flushPromises()

    const save = modalButton(tr("admin.events.saveNominees"))!
    expect(save.disabled).toBe(false)
    clickAfterRender(save)
    await flushPromises()

    expect(api.setNominees).toHaveBeenCalledWith(
      7,
      [
        expect.objectContaining({ media_id: 102, media_title: "Beta", position: 1 }),
        expect.objectContaining({ media_id: 103, media_title: "Gamma", position: 2 }),
      ],
    )
    await vi.waitFor(() => expect(api.list).toHaveBeenCalledTimes(2))
  })
})

describe("AdminEventsView delete guard", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    wrappers.forEach((w) => w.unmount())
    wrappers.length = 0
    document.body.innerHTML = ""
    vi.clearAllMocks()
  })

  it("disables delete for a finalized voting round and explains why", async () => {
    api.list.mockResolvedValue(
      listResponse([
        makeAdminEvent({
          id: 4,
          type: "vote_monthly",
          media_type: "movie",
          period: "2026-10",
          state: "past",
          is_finalized: true,
          has_votes: true,
          nominees: [nominee(1, 101, "Alpha")],
        }),
      ]),
    )

    const wrapper = await mountView()
    await wrapper.find('button[aria-haspopup="menu"]').trigger("click")
    const del = Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="menuitem"]')).find(
      (el) => el.textContent?.trim() === tr("common.delete"),
    )

    expect(del?.disabled).toBe(true)
    expect(del?.title).toBe(tr("admin.events.deleteBlocked"))
  })

  it("refetches the list when delete is refused because votes arrived", async () => {
    api.list.mockResolvedValue(
      listResponse([
        makeAdminEvent({ id: 4, type: "vote_monthly", media_type: "movie", period: "2026-10", state: "upcoming" }),
      ]),
    )
    api.remove.mockRejectedValueOnce(
      Object.assign(new Error("HTTP 409"), {
        isAxiosError: true,
        response: { status: 409, data: { code: "EVENT_NOT_DELETABLE" } },
      }),
    )

    const wrapper = await mountView()
    await wrapper.find('button[aria-haspopup="menu"]').trigger("click")
    const del = Array.from(document.body.querySelectorAll<HTMLButtonElement>('[role="menuitem"]')).find(
      (el) => el.textContent?.trim() === tr("common.delete"),
    )
    await clickAfterRender(del!)
    await flushPromises()
    await clickAfterRender(modalButton(tr("admin.events.deleteConfirm"))!)

    await vi.waitFor(() => expect(api.remove).toHaveBeenCalledTimes(1))
    await vi.waitFor(() => expect(api.list).toHaveBeenCalledTimes(2))
  })
})
