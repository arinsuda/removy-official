import { describe, expect, it } from "vitest"
import { nextTick, ref } from "vue"
import { applyTabList, useTabList } from "@/components/profile/tabs/useTabList"

interface Row {
  name: string
  kind: "movie" | "tv"
  score: number
}
const rows: Row[] = [
  { name: "Dune", kind: "movie", score: 8 },
  { name: "Arrival", kind: "movie", score: 7.5 },
  { name: "Severance", kind: "tv", score: 8.7 },
  { name: "dune messiah", kind: "movie", score: 6 },
]
const cfg = {
  searchText: (r: Row) => r.name,
  filters: { movie: (r: Row) => r.kind === "movie", tv: (r: Row) => r.kind === "tv" },
  sorts: {
    name: (a: Row, b: Row) => a.name.localeCompare(b.name),
    score: (a: Row, b: Row) => b.score - a.score,
  },
}

describe("applyTabList", () => {
  it("searches case-insensitively and ignores surrounding spaces", () => {
    const out = applyTabList(rows, { query: "  DUNE ", filter: "all", sort: "name" }, cfg)
    expect(out.map((r) => r.name)).toEqual(["Dune", "dune messiah"])
  })

  it("filters by key and treats unknown keys as all", () => {
    expect(applyTabList(rows, { query: "", filter: "tv", sort: "name" }, cfg)).toHaveLength(1)
    expect(applyTabList(rows, { query: "", filter: "all", sort: "name" }, cfg)).toHaveLength(4)
  })

  it("sorts without mutating the input", () => {
    const copy = [...rows]
    const out = applyTabList(rows, { query: "", filter: "all", sort: "score" }, cfg)
    expect(out.map((r) => r.score)).toEqual([8.7, 8, 7.5, 6])
    expect(rows).toEqual(copy)
  })

  it("combines search, filter and sort", () => {
    const out = applyTabList(rows, { query: "dune", filter: "movie", sort: "score" }, cfg)
    expect(out.map((r) => r.name)).toEqual(["Dune", "dune messiah"])
  })
})

describe("useTabList", () => {
  it("paginates, resets on change and reports filter counts", async () => {
    const items = ref(Array.from({ length: 5 }, (_, i) => ({ name: `T${i}`, kind: i % 2 ? ("tv" as const) : ("movie" as const), score: i })))
    const list = useTabList({ items, ...cfg, defaultSort: "score", pageSize: 2 })
    expect(list.visible.value.map((r) => r.score)).toEqual([4, 3])
    expect(list.hasMore.value).toBe(true)
    list.loadMore()
    expect(list.visible.value).toHaveLength(4)
    list.filter.value = "tv"
    await nextTick()
    expect(list.visible.value).toHaveLength(2)
    expect(list.hasMore.value).toBe(false)
    expect(list.filterCounts.value).toMatchObject({ all: 5, movie: 3, tv: 2 })
    expect(list.isFiltering.value).toBe(true)
    list.clear()
    expect(list.isFiltering.value).toBe(false)
  })
})
