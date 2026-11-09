import type { ActorCredit } from "./actorCredits"

type Translate = (key: string, named?: Record<string, unknown>) => string

/** "2010 · as Cobb": year plus the role when the credit has one. */
export function creditCaption(credit: ActorCredit, t: Translate): string | undefined {
  const character = credit.character ? t("actor.credit.asCharacter", { character: credit.character }) : ""
  if (credit.year === null) return character || undefined
  return character ? t("actor.credit.caption", { year: credit.year, character }) : String(credit.year)
}
