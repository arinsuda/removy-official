import { describe, expect, it } from "vitest"
import { flattenGroups, groupByRank, moveTarget, nextFreeRank, tierOf } from "@/components/bmol/bmolRanks"
import type { BMOLItemResponse } from "@/types"

const item = (id: number, rank: number): BMOLItemResponse =>
  ({ id, rank, note: "", media_type: "movie", media: { id: id * 10, title: `T${id}` }, created_at: "" }) as BMOLItemResponse

describe("groupByRank", () => {
  it("keeps ties together, sorted by rank then insertion", () => {
    const groups = groupByRank([item(1, 3), item(2, 1), item(3, 3), item(4, 1), item(5, 2)])
    expect(groups.map((g) => [g.rank, g.items.map((i) => i.id)])).toEqual([
      [1, [2, 4]],
      [2, [5]],
      [3, [1, 3]],
    ])
    expect(flattenGroups(groups).map((i) => i.id)).toEqual([2, 4, 5, 1, 3])
  })

  it("returns nothing for an empty list", () => {
    expect(groupByRank([])).toEqual([])
  })
})

describe("tierOf", () => {
  it("maps ranks to tiers", () => {
    expect([1, 2, 3, 4, 20].map(tierOf)).toEqual(["gold", "silver", "bronze", "shelf", "shelf"])
  })
})

describe("nextFreeRank", () => {
  it("starts at 1 and fills the first gap", () => {
    expect(nextFreeRank([])).toBe(1)
    expect(nextFreeRank([item(1, 1), item(2, 2)])).toBe(3)
    expect(nextFreeRank([item(1, 1), item(2, 3)])).toBe(2)
  })
})

describe("moveTarget", () => {
  const groups = groupByRank([item(1, 1), item(2, 2), item(3, 2), item(4, 5)])
  it("joins the neighbouring group", () => {
    expect(moveTarget(groups, 2, "up")).toBe(1)
    expect(moveTarget(groups, 2, "down")).toBe(5)
    expect(moveTarget(groups, 1, "down")).toBe(2)
  })
  it("stops at the edges", () => {
    expect(moveTarget(groups, 1, "up")).toBeNull()
    expect(moveTarget(groups, 5, "down")).toBeNull()
  })
  it("opens a new rank below a shared last group", () => {
    expect(moveTarget(groupByRank([item(1, 1), item(2, 1)]), 1, "down")).toBe(2)
  })
  it("ignores an unknown rank", () => {
    expect(moveTarget(groups, 9, "up")).toBeNull()
  })

  it("keeps legacy rank < 1 rows last and never moves them or others upward by 'down'", () => {
    const legacy = groupByRank([item(1, 0), item(2, 1), item(3, 2), item(4, -1)])
    expect(legacy.map((g) => g.rank)).toEqual([1, 2, 0, -1])
    expect(moveTarget(legacy, 2, "down")).toBeNull()
    expect(moveTarget(legacy, 0, "up")).toBe(2)
    expect(moveTarget(groupByRank([item(1, 999), item(2, 999)]), 999, "down")).toBeNull()
  })
})
