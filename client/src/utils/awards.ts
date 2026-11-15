import type { AwardKind, EventAward } from "@/types/event"

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
