import type { AwardKind, EventAward, MediaAward } from "@/types/event"

/** One awards section: every winner of one round (co-winners share kind and period). */
export interface AwardGroup {
  kind: AwardKind
  period: string
  awards: EventAward[]
}

/** Yearly sections come first, then newest period first. */
function compareGroups(a: AwardGroup, b: AwardGroup): number {
  if (a.kind !== b.kind) return a.kind === "yearly" ? -1 : 1
  return b.period.localeCompare(a.period)
}

/** Group winners by round so co-winners sit together, then order the sections for display. */
export function groupAwards(awards: EventAward[]): AwardGroup[] {
  const groups = new Map<string, AwardGroup>()
  for (const award of awards) {
    const key = `${award.kind}:${award.period}`
    const group = groups.get(key) ?? { kind: award.kind, period: award.period, awards: [] }
    group.awards.push(award)
    groups.set(key, group)
  }
  return [...groups.values()].sort(compareGroups)
}

/** Newest period first; "YYYY-MM" and "YYYY" both sort correctly as text within one kind. */
const byNewestPeriod = (a: MediaAward, b: MediaAward) => b.period.localeCompare(a.period)

/**
 * The award a poster shows: the newest yearly award when there is one, else the newest monthly award.
 * Returns null when the title has no awards.
 */
export function topAward(awards: MediaAward[] | null | undefined): MediaAward | null {
  if (!awards?.length) return null
  const yearly = awards.filter((a) => a.kind === "yearly").sort(byNewestPeriod)
  if (yearly.length) return yearly[0] ?? null
  return [...awards].sort(byNewestPeriod)[0] ?? null
}
